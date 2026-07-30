package fees

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/shopspring/decimal"
)

// packagesServiceAPI manages fee packages. A package groups one or more
// [Fee] definitions that are evaluated together when calculating
// fees for a transaction.
type packagesServiceAPI interface {
	// Create creates a new fee package from the provided input.
	Create(ctx context.Context, input *CreatePackageInput) (*Package, error)

	// Get retrieves a fee package by its unique identifier.
	Get(ctx context.Context, id string) (*Package, error)

	// List returns a normalized page of fee packages matching the given filter
	// options.
	List(ctx context.Context, opts *PackageListOptions) (*PackagePage, error)

	// Update partially updates an existing fee package.
	Update(ctx context.Context, id string, input *UpdatePackageInput) (*Package, error)

	// Delete removes a fee package by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// packagesService is the concrete implementation of [packagesServiceAPI].
type packagesService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ packagesServiceAPI = (*packagesService)(nil)

// newPackagesService constructs a [packagesServiceAPI] backed by the given
// [core.Backend].
func newPackagesService(backend core.Backend) packagesServiceAPI {
	return &packagesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Create creates a new fee package.
func (s *packagesService) Create(ctx context.Context, input *CreatePackageInput) (*Package, error) {
	if err := ensureService(s); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	return core.Create[Package, CreatePackageInput](ctx, &s.BaseService, "/packages", input)
}

// Get retrieves a fee package by ID.
func (s *packagesService) Get(ctx context.Context, id string) (*Package, error) {
	const operation = "Packages.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Package", "id is required")
	}

	return core.Get[Package](ctx, &s.BaseService, "/packages/"+url.PathEscape(id))
}

// List returns a paginated list of fee packages. The plugin uses a flat
// pagination envelope and supports domain-specific query parameters
// (segmentId, ledgerId, transactionRoute, enable) alongside standard
// pagination params.
func (s *packagesService) List(ctx context.Context, opts *PackageListOptions) (*PackagePage, error) {
	if err := ensureService(s); err != nil {
		return nil, err
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	backend, err := core.ResolveBackend(&s.BaseService)
	if err != nil {
		return nil, err
	}

	queryPath := buildPackagesListPath(opts)

	res, err := backend.Do(ctx, core.Request{Method: "GET", Path: queryPath})
	if err != nil {
		return nil, err
	}

	var raw struct {
		Items json.RawMessage `json:"items"`
		Page  int             `json:"page"`
		Limit int             `json:"limit"`
		Total int             `json:"total"`
	}
	if err := json.Unmarshal(res.Body, &raw); err != nil {
		return nil, sdkerrors.NewInternal("fees", "Packages.List", "failed to unmarshal response body", err)
	}

	if raw.Items == nil {
		return nil, sdkerrors.NewInternal("fees", "Packages.List", "response contained no items payload", nil)
	}

	items, err := normalizePackageItems(raw.Items)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if raw.Limit > 0 {
		totalPages = (raw.Total + raw.Limit - 1) / raw.Limit
	}

	return &PackagePage{
		Items:      items,
		PageNumber: raw.Page,
		PageSize:   raw.Limit,
		TotalItems: raw.Total,
		TotalPages: totalPages,
	}, nil
}

// Update partially updates an existing fee package.
func (s *packagesService) Update(ctx context.Context, id string, input *UpdatePackageInput) (*Package, error) {
	if err := ensureService(s); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, sdkerrors.NewValidation("Packages.Update", "Package", "id is required")
	}

	normalized, err := s.prepareUpdateInput(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return core.Update[Package, UpdatePackageInput](ctx, &s.BaseService, "/packages/"+url.PathEscape(id), normalized)
}

// Delete removes a fee package by ID.
func (s *packagesService) Delete(ctx context.Context, id string) error {
	const operation = "Packages.Delete"

	if err := ensureService(s); err != nil {
		return err
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, "Package", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/packages/"+url.PathEscape(id))
}

func normalizePackageItems(raw json.RawMessage) ([]Package, error) {
	if string(raw) == "null" {
		return []Package{}, nil
	}

	var items []Package
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, sdkerrors.NewInternal("fees", "Packages.List", "failed to unmarshal items payload", err)
	}

	if items == nil {
		items = []Package{}
	}

	for i := range items {
		if items[i].Fees == nil {
			items[i].Fees = map[string]Fee{}
		}
	}

	return items, nil
}

func (s *packagesService) prepareUpdateInput(ctx context.Context, id string, input *UpdatePackageInput) (*UpdatePackageInput, error) {
	if input == nil {
		return nil, sdkerrors.NewValidation("Packages.Update", "Package", "input is required")
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	current, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	normalized := cloneUpdatePackageInput(input)
	if err := validateEffectivePackageBounds(normalized, current); err != nil {
		return nil, err
	}

	if normalized.Fees != nil {
		normalized.Fees = mergeFeeUpdates(current.Fees, normalized.Fees)

		minAmount, err := resolveEffectiveMinimumAmount(current, normalized)
		if err != nil {
			return nil, err
		}

		for key, fee := range normalized.Fees {
			if err := validateCreateFeeDefinition("Packages.Update", key, fee, minAmount); err != nil {
				return nil, err
			}
		}
	}

	return normalized, nil
}

func cloneUpdatePackageInput(input *UpdatePackageInput) *UpdatePackageInput {
	if input == nil {
		return nil
	}

	cloned := *input
	if input.Fees != nil {
		cloned.Fees = make(map[string]Fee, len(input.Fees))
		for key, fee := range input.Fees {
			cloned.Fees[key] = fee
		}
	}

	return &cloned
}

func validateEffectivePackageBounds(input *UpdatePackageInput, current *Package) error {
	const operation = "Packages.Update"

	minimum := current.MinimumAmount
	if input.MinimumAmount != nil {
		minimum = *input.MinimumAmount
	}

	maximum := current.MaximumAmount
	if input.MaximumAmount != nil {
		maximum = *input.MaximumAmount
	}

	minAmount, err := parsePackageAmount(operation, "minimum amount", minimum)
	if err != nil {
		return err
	}

	maxAmount, err := parsePackageAmount(operation, "maximum amount", maximum)
	if err != nil {
		return err
	}

	if minAmount.GreaterThan(maxAmount) {
		return sdkerrors.NewValidation(operation, packageResource, "minimum amount must be less than or equal to maximum amount")
	}

	return nil
}

func resolveEffectiveMinimumAmount(current *Package, input *UpdatePackageInput) (decimal.Decimal, error) {
	minimum := current.MinimumAmount
	if input.MinimumAmount != nil {
		minimum = *input.MinimumAmount
	}

	return parsePackageAmount("Packages.Update", "minimum amount", minimum)
}

func mergeFeeUpdates(current, patch map[string]Fee) map[string]Fee {
	if patch == nil {
		return nil
	}

	merged := make(map[string]Fee, len(current)+len(patch))
	for key, fee := range current {
		merged[key] = fee
	}

	for key, feePatch := range patch {
		merged[key] = mergeFeeUpdate(merged[key], feePatch)
	}

	return merged
}

func mergeFeeUpdate(base, patch Fee) Fee {
	merged := base

	if strings.TrimSpace(patch.FeeLabel) != "" {
		merged.FeeLabel = patch.FeeLabel
	}

	if patch.CalculationModel != nil {
		merged.CalculationModel = patch.CalculationModel
	}

	if strings.TrimSpace(patch.ReferenceAmount) != "" {
		merged.ReferenceAmount = patch.ReferenceAmount
	}

	if patch.Priority != 0 || base.Priority == 0 {
		merged.Priority = patch.Priority
	}

	if patch.IsDeductibleFrom != nil {
		merged.IsDeductibleFrom = patch.IsDeductibleFrom
	}

	if strings.TrimSpace(patch.CreditAccount) != "" {
		merged.CreditAccount = patch.CreditAccount
	}

	if patch.RouteFrom != nil {
		merged.RouteFrom = patch.RouteFrom
	}

	if patch.RouteTo != nil {
		merged.RouteTo = patch.RouteTo
	}

	return merged
}

// buildPackagesListPath constructs the query string for the packages list
// endpoint. The plugin expects bare query parameters (not filter[key]=value)
// and uses snake_case for some pagination params (sort_order, start_date,
// end_date).
func buildPackagesListPath(opts *PackageListOptions) string {
	const basePath = "/packages"

	if opts == nil {
		return basePath
	}

	params := url.Values{}

	// Domain-specific filters
	if opts.SegmentID != "" {
		params.Set("segmentId", opts.SegmentID)
	}

	if opts.LedgerID != "" {
		params.Set("ledgerId", opts.LedgerID)
	}

	if opts.TransactionRoute != "" {
		params.Set("transactionRoute", opts.TransactionRoute)
	}

	if opts.Enabled != nil {
		params.Set("enable", strconv.FormatBool(*opts.Enabled))
	}

	// Pagination params (plugin uses snake_case for some)
	if opts.PageSize > 0 {
		params.Set("limit", strconv.Itoa(opts.PageSize))
	}

	if opts.PageNumber > 0 {
		params.Set("page", strconv.Itoa(opts.PageNumber))
	}

	if opts.SortOrder != "" {
		params.Set("sort_order", opts.SortOrder)
	}

	if opts.CreatedFrom != nil {
		params.Set("start_date", opts.CreatedFrom.Format("2006-01-02"))
	}

	if opts.CreatedTo != nil {
		params.Set("end_date", opts.CreatedTo.Format("2006-01-02"))
	}

	if len(params) == 0 {
		return basePath
	}

	return basePath + "?" + params.Encode()
}
