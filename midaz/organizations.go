// organizations.go implements the organizationsServiceAPI for managing
// top-level organizational entities in Midaz. Organizations are the root
// scope for ledgers, accounts, and all other financial domain objects.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// organizationsServiceAPI provides CRUD operations for organizations.
type organizationsServiceAPI interface {
	// Create creates a new organization from the given input.
	Create(ctx context.Context, input *CreateOrganizationInput) (*Organization, error)

	// Get retrieves an organization by its unique identifier.
	Get(ctx context.Context, id string) (*Organization, error)

	// List returns a paginated iterator over organizations.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Organization]

	// Update partially updates an existing organization.
	Update(ctx context.Context, id string, input *UpdateOrganizationInput) (*Organization, error)

	// Delete removes an organization by its unique identifier.
	Delete(ctx context.Context, id string) error

	// Count returns the total number of organizations.
	Count(ctx context.Context) (int, error)
}

// organizationsService is the concrete implementation of [organizationsServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type organizationsService struct {
	core.BaseService
}

// newOrganizationsService creates a new [organizationsServiceAPI] backed by the
// given onboarding [core.Backend].
func newOrganizationsService(backend core.Backend) organizationsServiceAPI {
	return &organizationsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ organizationsServiceAPI = (*organizationsService)(nil)

// Create creates a new organization from the given input.
func (s *organizationsService) Create(ctx context.Context, input *CreateOrganizationInput) (*Organization, error) {
	const operation = "Organizations.Create"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Organization", "input is required")
	}

	return core.Create[Organization, CreateOrganizationInput](ctx, &s.BaseService, "/organizations", input)
}

// Get retrieves an organization by its unique identifier.
func (s *organizationsService) Get(ctx context.Context, id string) (*Organization, error) {
	const operation = "Organizations.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Organization", "id is required")
	}

	return core.Get[Organization](ctx, &s.BaseService, "/organizations/"+url.PathEscape(id))
}

// List returns a paginated iterator over organizations.
func (s *organizationsService) List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Organization] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[Organization](err)
	}

	return core.List[Organization](ctx, &s.BaseService, "/organizations", opts)
}

// Update partially updates an existing organization.
func (s *organizationsService) Update(ctx context.Context, id string, input *UpdateOrganizationInput) (*Organization, error) {
	const operation = "Organizations.Update"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Organization", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Organization", "input is required")
	}

	return core.Update[Organization, UpdateOrganizationInput](ctx, &s.BaseService, "/organizations/"+url.PathEscape(id), input)
}

// Delete removes an organization by its unique identifier.
func (s *organizationsService) Delete(ctx context.Context, id string) error {
	const operation = "Organizations.Delete"

	if err := ensureService(s); err != nil {
		return err
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, "Organization", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/organizations/"+url.PathEscape(id))
}

// Count returns the total number of organizations.
func (s *organizationsService) Count(ctx context.Context) (int, error) {
	if err := ensureService(s); err != nil {
		return 0, err
	}

	return core.Count(ctx, &s.BaseService, "/organizations/metrics/count")
}
