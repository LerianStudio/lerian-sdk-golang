// holders.go implements the holdersServiceAPI for managing customer/entity
// records in the Midaz CRM system. Holders are the root objects for
// customer relationship management and are linked to ledger accounts
// through aliases.
//
// Unlike onboarding and transaction services, the CRM API uses the
// X-Organization-Id header instead of URL path parameters for organization
// context. All methods therefore require an orgID that is injected via the
// CRM helper functions defined in crm.go.
package midaz

import (
	"context"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// holdersServiceAPI provides CRUD operations for holders in the CRM system.
type holdersServiceAPI interface {
	// Create creates a new holder in the specified organization.
	Create(ctx context.Context, orgID string, input *CreateHolderInput) (*Holder, error)

	// Get retrieves a holder by ID.
	Get(ctx context.Context, orgID, id string, opts *CRMGetOptions) (*Holder, error)

	// List returns a lazy iterator over holders visible to the organization.
	List(ctx context.Context, orgID string, opts *CRMListOptions) *pagination.Iterator[Holder]

	// Update partially updates an existing holder.
	Update(ctx context.Context, orgID, id string, input *UpdateHolderInput) (*Holder, error)

	// Delete removes a holder.
	Delete(ctx context.Context, orgID, id string, opts *CRMDeleteOptions) error
}

// holdersService is the concrete implementation of [holdersServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type holdersService struct {
	core.BaseService
}

// newHoldersService creates a new [holdersServiceAPI] backed by the given
// CRM [core.Backend].
func newHoldersService(backend core.Backend) holdersServiceAPI {
	return &holdersService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ holdersServiceAPI = (*holdersService)(nil)

const holderResource = "Holder"

// holdersBasePath returns the collection path for holder operations.
func holdersBasePath() string {
	return crmCollectionPath("holders")
}

// holdersItemPath returns the path for a specific holder.
func holdersItemPath(id string) string {
	return crmItemPath("holders", id)
}

// Create creates a new holder in the specified organization.
func (s *holdersService) Create(ctx context.Context, orgID string, input *CreateHolderInput) (*Holder, error) {
	const operation = "Holders.Create"

	if s == nil {
		return nil, core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, holderResource, orgID)
	if err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, holderResource, "input is required")
	}

	return core.CreateWithHeaders[Holder, CreateHolderInput](ctx, &s.BaseService, holdersBasePath(), crmHeaders(orgID), input)
}

// Get retrieves a holder by ID.
func (s *holdersService) Get(ctx context.Context, orgID, id string, opts *CRMGetOptions) (*Holder, error) {
	const operation = "Holders.Get"

	if s == nil {
		return nil, core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, holderResource, orgID)
	if err != nil {
		return nil, err
	}

	id, err = validateCRMIdentifier(operation, holderResource, id, "id")
	if err != nil {
		return nil, err
	}

	return core.GetWithHeaders[Holder](ctx, &s.BaseService, applyCRMGetOptions(holdersItemPath(id), opts), crmHeaders(orgID))
}

// List returns a lazy iterator over holders visible to the organization.
func (s *holdersService) List(ctx context.Context, orgID string, opts *CRMListOptions) *pagination.Iterator[Holder] {
	const operation = "Holders.List"

	if s == nil {
		return pagination.NewErrorIterator[Holder](core.ErrNilService)
	}

	orgID, err := validateCRMOrgID(operation, holderResource, orgID)
	if err != nil {
		return pagination.NewErrorIterator[Holder](err)
	}

	return core.ListPageWithHeaders[Holder](ctx, &s.BaseService, crmHeaders(orgID), initialCRMPage(opts), func(page int) string {
		return buildCRMListPath(holdersBasePath(), opts, page)
	})
}

// Update partially updates an existing holder.
func (s *holdersService) Update(ctx context.Context, orgID, id string, input *UpdateHolderInput) (*Holder, error) {
	const operation = "Holders.Update"

	if s == nil {
		return nil, core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, holderResource, orgID)
	if err != nil {
		return nil, err
	}

	id, err = validateCRMIdentifier(operation, holderResource, id, "id")
	if err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, holderResource, "input is required")
	}

	return core.UpdateWithHeaders[Holder, UpdateHolderInput](ctx, &s.BaseService, holdersItemPath(id), crmHeaders(orgID), input)
}

// Delete removes a holder.
func (s *holdersService) Delete(ctx context.Context, orgID, id string, opts *CRMDeleteOptions) error {
	const operation = "Holders.Delete"

	if s == nil {
		return core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, holderResource, orgID)
	if err != nil {
		return err
	}

	id, err = validateCRMIdentifier(operation, holderResource, id, "id")
	if err != nil {
		return err
	}

	return core.DeleteWithHeaders(ctx, &s.BaseService, applyCRMDeleteOptions(holdersItemPath(id), opts), crmHeaders(orgID))
}
