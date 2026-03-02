package leriantest

import (
	"context"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// newFakeFeesClient constructs a [fees.Client] with all service fields
// backed by in-memory fakes.
func newFakeFeesClient(cfg *fakeConfig) *fees.Client {
	return &fees.Client{
		Packages:  &fakeFeesPackages{store: newFakeStore[fees.Package](), cfg: cfg},
		Estimates: &fakeFeesEstimates{cfg: cfg},
		Fees:      &fakeFeesFees{cfg: cfg},
	}
}

// ---------------------------------------------------------------------------
// Packages
// ---------------------------------------------------------------------------

type fakeFeesPackages struct {
	store *fakeStore[fees.Package]
	cfg   *fakeConfig
}

var _ fees.PackagesService = (*fakeFeesPackages)(nil)

func (f *fakeFeesPackages) Create(_ context.Context, input *fees.CreatePackageInput) (*fees.Package, error) {
	if err := f.cfg.injectedError("fees.Packages.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

	p := fees.Package{
		ID:        generateID("fpkg"),
		Name:      input.Name,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	f.store.Set(p.ID, p)

	return &p, nil
}

func (f *fakeFeesPackages) Get(_ context.Context, id string) (*fees.Package, error) {
	p, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Packages.Get", "Package", id)
	}

	return &p, nil
}

func (f *fakeFeesPackages) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[fees.Package] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeFeesPackages) Update(_ context.Context, id string, _ *fees.UpdatePackageInput) (*fees.Package, error) {
	p, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Packages.Update", "Package", id)
	}

	p.UpdatedAt = time.Now()
	f.store.Set(id, p)

	return &p, nil
}

func (f *fakeFeesPackages) Delete(_ context.Context, id string) error {
	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Packages.Delete", "Package", id)
	}

	f.store.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// Estimates (RPC-style, no store)
// ---------------------------------------------------------------------------

type fakeFeesEstimates struct {
	cfg *fakeConfig
}

var _ fees.EstimatesService = (*fakeFeesEstimates)(nil)

func (f *fakeFeesEstimates) Calculate(_ context.Context, _ *fees.CalculateEstimateInput) (*fees.Estimate, error) {
	if err := f.cfg.injectedError("fees.Estimates.Calculate"); err != nil {
		return nil, err
	}

	return &fees.Estimate{
		ID: generateID("fest"),
	}, nil
}

// ---------------------------------------------------------------------------
// Fees (RPC-style, no store)
// ---------------------------------------------------------------------------

type fakeFeesFees struct {
	cfg *fakeConfig
}

var _ fees.FeesService = (*fakeFeesFees)(nil)

func (f *fakeFeesFees) Calculate(_ context.Context, _ *fees.CalculateFeeInput) (*fees.Fee, error) {
	if err := f.cfg.injectedError("fees.Fees.Calculate"); err != nil {
		return nil, err
	}

	return &fees.Fee{
		ID: generateID("ffee"),
	}, nil
}
