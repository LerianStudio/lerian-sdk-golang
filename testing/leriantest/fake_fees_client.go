package leriantest

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// newFakeFeesClient constructs a [fees.Client] with all service fields
// backed by in-memory fakes.
func newFakeFeesClient(cfg *fakeConfig) *fees.Client {
	state := &fakeFeesState{packages: newFakeStore[fees.Package]()}

	return &fees.Client{
		Packages:  &fakeFeesPackages{state: state, cfg: cfg},
		Estimates: &fakeFeesEstimates{state: state, cfg: cfg},
		Fees:      &fakeFeesFees{state: state, cfg: cfg},
	}
}

type fakeFeesState struct {
	packages *fakeStore[fees.Package]
}

// ---------------------------------------------------------------------------
// Packages
// ---------------------------------------------------------------------------

type fakeFeesPackages struct {
	state *fakeFeesState
	cfg   *fakeConfig
}

func (f *fakeFeesPackages) Create(_ context.Context, input *fees.CreatePackageInput) (*fees.Package, error) {
	if err := f.cfg.injectedError("fees.Packages.Create"); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	p := fees.Package{
		ID:               generateID("fpkg"),
		FeeGroupLabel:    input.FeeGroupLabel,
		Description:      cloneStringPointer(input.Description),
		SegmentID:        cloneStringPointer(input.SegmentID),
		LedgerID:         input.LedgerID,
		TransactionRoute: cloneStringPointer(input.TransactionRoute),
		MinimumAmount:    input.MinimumAmount,
		MaximumAmount:    input.MaximumAmount,
		WaivedAccounts:   cloneStringSlicePointer(input.WaivedAccounts),
		Fees:             cloneFeeMap(input.Fees),
		Enable:           cloneBoolPointer(input.Enable),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	f.state.packages.Set(p.ID, p)

	return clonePackagePointer(&p), nil
}

func (f *fakeFeesPackages) Get(_ context.Context, id string) (*fees.Package, error) {
	p, ok := f.state.packages.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Packages.Get", "Package", id)
	}

	return clonePackagePointer(&p), nil
}

func (f *fakeFeesPackages) List(_ context.Context, opts *fees.PackageListOptions) (*fees.PackagePage, error) {
	if err := f.cfg.injectedError("fees.Packages.List"); err != nil {
		return nil, err
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	all := f.state.packages.List()
	filtered := filterFakePackages(all, opts)
	sorted := sortFakePackages(filtered, opts)

	// Determine page size and page number.
	limit := 10
	page := 1

	if opts != nil {
		if opts.PageSize > 0 {
			limit = opts.PageSize
		}

		if opts.PageNumber > 0 {
			page = opts.PageNumber
		}
	}

	total := len(sorted)

	start := (page - 1) * limit
	if start > total {
		start = total
	}

	end := start + limit
	if end > total {
		end = total
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return &fees.PackagePage{
		Items:      clonePackageSlice(sorted[start:end]),
		PageNumber: page,
		PageSize:   limit,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (f *fakeFeesPackages) Update(_ context.Context, id string, input *fees.UpdatePackageInput) (*fees.Package, error) {
	if err := f.cfg.injectedError("fees.Packages.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Packages.Update", "Package", "input is required")
	}

	hasFeesPatch := input.Fees != nil

	if err := input.Validate(); err != nil {
		return nil, err
	}

	updated, err := fakeMutateStored(f.cfg, "", "Packages.Update", "Package", id, f.state.packages, func(p *fees.Package) {
		if input.FeeGroupLabel != "" {
			p.FeeGroupLabel = input.FeeGroupLabel
		}
		if input.Description != "" {
			p.Description = &input.Description
		}
		if input.MinimumAmount != nil {
			p.MinimumAmount = *input.MinimumAmount
		}
		if input.MaximumAmount != nil {
			p.MaximumAmount = *input.MaximumAmount
		}
		if input.WaivedAccounts != nil {
			p.WaivedAccounts = cloneStringSlicePointer(input.WaivedAccounts)
		}
		if hasFeesPatch {
			p.Fees = cloneFeeMap(input.Fees)
		}
		if input.Enable != nil {
			p.Enable = cloneBoolPointer(input.Enable)
		}
		p.UpdatedAt = time.Now()
	})
	if err != nil {
		return nil, err
	}

	return clonePackagePointer(updated), nil
}

func (f *fakeFeesPackages) Delete(_ context.Context, id string) error {
	if _, ok := f.state.packages.Get(id); !ok {
		return sdkerrors.NewNotFound("Packages.Delete", "Package", id)
	}

	f.state.packages.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// Estimates (RPC-style, no store)
// ---------------------------------------------------------------------------

type fakeFeesEstimates struct {
	state *fakeFeesState
	cfg   *fakeConfig
}

func (f *fakeFeesEstimates) Calculate(_ context.Context, input *fees.FeeEstimateInput) (*fees.FeeEstimateResponse, error) {
	if err := f.cfg.injectedError("fees.Estimates.Calculate"); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	pack, ok := f.state.packages.Get(input.PackageID)
	if !ok {
		return nil, sdkerrors.NewNotFound("Estimates.Calculate", "Package", input.PackageID)
	}

	if !packageMatchesEstimate(pack, input) {
		return &fees.FeeEstimateResponse{
			Message:     "No fee or gratuity rules were found for the given parameters.",
			FeesApplied: nil,
		}, nil
	}

	result, applied := applyFakeFeesToTransaction(pack, input.Transaction)
	if !applied {
		return &fees.FeeEstimateResponse{
			Message:     "No fee or gratuity rules were found for the given parameters.",
			FeesApplied: nil,
		}, nil
	}

	return &fees.FeeEstimateResponse{
		Message: "Successfully estimated fee.",
		FeesApplied: &fees.FeeCalculate{
			LedgerID:    input.LedgerID,
			Transaction: result,
		},
	}, nil
}

// ---------------------------------------------------------------------------
// Fees (RPC-style, no store)
// ---------------------------------------------------------------------------

type fakeFeesFees struct {
	state *fakeFeesState
	cfg   *fakeConfig
}

func (f *fakeFeesFees) Calculate(_ context.Context, input *fees.FeeCalculate) (*fees.FeeCalculate, error) {
	if err := f.cfg.injectedError("fees.Fees.Calculate"); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	pack, matched, err := findMatchingFakePackage(f.state.packages.List(), input)
	if err != nil {
		return nil, err
	}
	result := cloneTransactionDSL(input.Transaction)
	if matched && pack != nil {
		if appliedResult, applied := applyFakeFeesToTransaction(*pack, input.Transaction); applied {
			result = appliedResult
		}
	}

	return &fees.FeeCalculate{
		SegmentID:   cloneStringPointer(input.SegmentID),
		LedgerID:    input.LedgerID,
		Transaction: result,
	}, nil
}

func filterFakePackages(all []fees.Package, opts *fees.PackageListOptions) []fees.Package {
	if opts == nil {
		return clonePackageSlice(all)
	}

	filtered := make([]fees.Package, 0, len(all))
	for _, pack := range all {
		if opts.LedgerID != "" && pack.LedgerID != opts.LedgerID {
			continue
		}

		if opts.SegmentID != "" && (pack.SegmentID == nil || *pack.SegmentID != opts.SegmentID) {
			continue
		}

		if opts.TransactionRoute != "" && (pack.TransactionRoute == nil || *pack.TransactionRoute != opts.TransactionRoute) {
			continue
		}

		if opts.Enabled != nil {
			if pack.Enable == nil || *pack.Enable != *opts.Enabled {
				continue
			}
		}

		if opts.CreatedFrom != nil && pack.CreatedAt.Before(startOfDay(*opts.CreatedFrom)) {
			continue
		}

		if opts.CreatedTo != nil && pack.CreatedAt.After(endOfDay(*opts.CreatedTo)) {
			continue
		}

		filtered = append(filtered, pack)
	}

	return clonePackageSlice(filtered)
}

func sortFakePackages(all []fees.Package, opts *fees.PackageListOptions) []fees.Package {
	ordered := clonePackageSlice(all)
	if len(ordered) < 2 {
		return ordered
	}

	sortOrder := "desc"
	if opts != nil && opts.SortOrder != "" {
		sortOrder = strings.ToLower(opts.SortOrder)
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		left := ordered[i].CreatedAt
		right := ordered[j].CreatedAt

		if sortOrder == "asc" {
			return left.Before(right)
		}

		return left.After(right)
	})

	return ordered
}
