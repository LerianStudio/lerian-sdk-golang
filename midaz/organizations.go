// organizations.go implements the OrganizationsService for managing
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

// OrganizationsService provides CRUD operations for organizations.
type OrganizationsService interface {
	// Create creates a new organization from the given input.
	Create(ctx context.Context, input *CreateOrganizationInput) (*Organization, error)

	// Get retrieves an organization by its unique identifier.
	Get(ctx context.Context, id string) (*Organization, error)

	// List returns a paginated iterator over organizations.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Organization]

	// Update partially updates an existing organization.
	Update(ctx context.Context, id string, input *UpdateOrganizationInput) (*Organization, error)

	// Delete removes an organization by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// organizationsService is the concrete implementation of [OrganizationsService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type organizationsService struct {
	core.BaseService
}

// newOrganizationsService creates a new [OrganizationsService] backed by the
// given onboarding [core.Backend].
func newOrganizationsService(backend core.Backend) OrganizationsService {
	return &organizationsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ OrganizationsService = (*organizationsService)(nil)

// Create creates a new organization from the given input.
func (s *organizationsService) Create(ctx context.Context, input *CreateOrganizationInput) (*Organization, error) {
	const operation = "Organizations.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Organization", "input is required")
	}

	return core.Create[Organization, CreateOrganizationInput](ctx, &s.BaseService, "/organizations", input)
}

// Get retrieves an organization by its unique identifier.
func (s *organizationsService) Get(ctx context.Context, id string) (*Organization, error) {
	const operation = "Organizations.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Organization", "id is required")
	}

	return core.Get[Organization](ctx, &s.BaseService, "/organizations/"+url.PathEscape(id))
}

// List returns a paginated iterator over organizations.
func (s *organizationsService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Organization] {
	return core.List[Organization](ctx, &s.BaseService, "/organizations", opts)
}

// Update partially updates an existing organization.
func (s *organizationsService) Update(ctx context.Context, id string, input *UpdateOrganizationInput) (*Organization, error) {
	const operation = "Organizations.Update"

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

	if id == "" {
		return sdkerrors.NewValidation(operation, "Organization", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/organizations/"+url.PathEscape(id))
}
