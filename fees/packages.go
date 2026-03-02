package fees

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// PackagesService manages fee packages. A package groups one or more
// [FeeRule] definitions that are evaluated together when calculating
// fees for a transaction.
type PackagesService interface {
	// Create creates a new fee package from the provided input.
	Create(ctx context.Context, input *CreatePackageInput) (*Package, error)

	// Get retrieves a fee package by its unique identifier.
	Get(ctx context.Context, id string) (*Package, error)

	// List returns a paginated iterator over all fee packages.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Package]

	// Update partially updates an existing fee package.
	Update(ctx context.Context, id string, input *UpdatePackageInput) (*Package, error)

	// Delete removes a fee package by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// packagesService is the concrete implementation of [PackagesService].
type packagesService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ PackagesService = (*packagesService)(nil)

// newPackagesService constructs a [PackagesService] backed by the given
// [core.Backend].
func newPackagesService(backend core.Backend) PackagesService {
	return &packagesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Create creates a new fee package.
func (s *packagesService) Create(ctx context.Context, input *CreatePackageInput) (*Package, error) {
	const operation = "Packages.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Package", "input is required")
	}

	return core.Create[Package, CreatePackageInput](ctx, &s.BaseService, "/packages", input)
}

// Get retrieves a fee package by ID.
func (s *packagesService) Get(ctx context.Context, id string) (*Package, error) {
	const operation = "Packages.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Package", "id is required")
	}

	return core.Get[Package](ctx, &s.BaseService, "/packages/"+url.PathEscape(id))
}

// List returns a paginated iterator over fee packages.
func (s *packagesService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Package] {
	return core.List[Package](ctx, &s.BaseService, "/packages", opts)
}

// Update partially updates an existing fee package.
func (s *packagesService) Update(ctx context.Context, id string, input *UpdatePackageInput) (*Package, error) {
	const operation = "Packages.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Package", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Package", "input is required")
	}

	return core.Update[Package, UpdatePackageInput](ctx, &s.BaseService, "/packages/"+url.PathEscape(id), input)
}

// Delete removes a fee package by ID.
func (s *packagesService) Delete(ctx context.Context, id string) error {
	const operation = "Packages.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Package", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/packages/"+url.PathEscape(id))
}
