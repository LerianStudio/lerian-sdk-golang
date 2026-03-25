// aliases.go implements the aliasesServiceAPI for managing alias accounts in
// the Midaz CRM system. An alias links a holder to a ledger account and
// carries banking details, regulatory fields, and related-party information.
//
// Most alias operations are nested under a specific holder
// (e.g. /holders/{holder_id}/aliases/{alias_id}), while the top-level
// list endpoint (/aliases) returns all aliases across holders.
//
// Like all CRM services, organization context is provided via the
// X-Organization-Id header rather than URL path parameters.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// aliasesServiceAPI provides CRUD operations for alias accounts in the CRM system.
type aliasesServiceAPI interface {
	// Create creates a new alias account under the specified holder.
	Create(ctx context.Context, orgID, holderID string, input *CreateAliasInput) (*Alias, error)

	// Get retrieves an alias by its unique identifier under a holder.
	Get(ctx context.Context, orgID, holderID, aliasID string, opts *CRMGetOptions) (*Alias, error)

	// List returns all aliases visible to the organization (top-level,
	// not scoped to a specific holder).
	List(ctx context.Context, orgID string, opts *AliasListOptions) *pagination.Iterator[Alias]

	// Update partially updates an existing alias.
	Update(ctx context.Context, orgID, holderID, aliasID string, input *UpdateAliasInput) (*Alias, error)

	// Delete removes an alias.
	Delete(ctx context.Context, orgID, holderID, aliasID string, opts *CRMDeleteOptions) error

	// DeleteRelatedParty removes a related party from an alias.
	DeleteRelatedParty(ctx context.Context, orgID, holderID, aliasID, relatedPartyID string) error
}

// aliasesService is the concrete implementation of [aliasesServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type aliasesService struct {
	core.BaseService
}

// newAliasesService creates a new [aliasesServiceAPI] backed by the given
// CRM [core.Backend].
func newAliasesService(backend core.Backend) aliasesServiceAPI {
	return &aliasesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ aliasesServiceAPI = (*aliasesService)(nil)

const aliasResource = "Alias"

// aliasesBasePath builds the base path for alias operations scoped to a holder.
func aliasesBasePath(holderID string) string {
	return crmNestedCollectionPath("holders", holderID, "aliases")
}

// aliasesItemPath builds the path for a specific alias under a holder.
func aliasesItemPath(holderID, aliasID string) string {
	return crmNestedItemPath("holders", holderID, "aliases", aliasID)
}

// Create creates a new alias account under the specified holder.
func (s *aliasesService) Create(ctx context.Context, orgID, holderID string, input *CreateAliasInput) (*Alias, error) {
	const operation = "Aliases.Create"

	if s == nil {
		return nil, core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, aliasResource, orgID)
	if err != nil {
		return nil, err
	}

	holderID, err = validateCRMIdentifier(operation, aliasResource, holderID, "holder id")
	if err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, aliasResource, "input is required")
	}

	return core.CreateWithHeaders[Alias, CreateAliasInput](ctx, &s.BaseService, aliasesBasePath(holderID), crmHeaders(orgID), input)
}

// Get retrieves an alias by its unique identifier under a holder.
func (s *aliasesService) Get(ctx context.Context, orgID, holderID, aliasID string, opts *CRMGetOptions) (*Alias, error) {
	const operation = "Aliases.Get"

	if s == nil {
		return nil, core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, aliasResource, orgID)
	if err != nil {
		return nil, err
	}

	holderID, err = validateCRMIdentifier(operation, aliasResource, holderID, "holder id")
	if err != nil {
		return nil, err
	}

	aliasID, err = validateCRMIdentifier(operation, aliasResource, aliasID, "alias id")
	if err != nil {
		return nil, err
	}

	return core.GetWithHeaders[Alias](ctx, &s.BaseService, applyCRMGetOptions(aliasesItemPath(holderID, aliasID), opts), crmHeaders(orgID))
}

// List returns all aliases visible to the organization. This is the
// top-level list endpoint (/aliases), not scoped to a specific holder.
func (s *aliasesService) List(ctx context.Context, orgID string, opts *AliasListOptions) *pagination.Iterator[Alias] {
	const operation = "Aliases.List"

	if s == nil {
		return pagination.NewErrorIterator[Alias](core.ErrNilService)
	}

	orgID, err := validateCRMOrgID(operation, aliasResource, orgID)
	if err != nil {
		return pagination.NewErrorIterator[Alias](err)
	}

	normalizedOpts, err := normalizeAliasListOptions(operation, aliasResource, opts)
	if err != nil {
		return pagination.NewErrorIterator[Alias](err)
	}

	return core.ListPageWithHeaders[Alias](ctx, &s.BaseService, crmHeaders(orgID), initialCRMAliasPage(normalizedOpts), func(page int) string {
		return buildCRMAliasListPath("/aliases", normalizedOpts, page)
	})
}

// Update partially updates an existing alias.
func (s *aliasesService) Update(ctx context.Context, orgID, holderID, aliasID string, input *UpdateAliasInput) (*Alias, error) {
	const operation = "Aliases.Update"

	if s == nil {
		return nil, core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, aliasResource, orgID)
	if err != nil {
		return nil, err
	}

	holderID, err = validateCRMIdentifier(operation, aliasResource, holderID, "holder id")
	if err != nil {
		return nil, err
	}

	aliasID, err = validateCRMIdentifier(operation, aliasResource, aliasID, "alias id")
	if err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, aliasResource, "input is required")
	}

	return core.UpdateWithHeaders[Alias, UpdateAliasInput](ctx, &s.BaseService, aliasesItemPath(holderID, aliasID), crmHeaders(orgID), input)
}

// Delete removes an alias.
func (s *aliasesService) Delete(ctx context.Context, orgID, holderID, aliasID string, opts *CRMDeleteOptions) error {
	const operation = "Aliases.Delete"

	if s == nil {
		return core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, aliasResource, orgID)
	if err != nil {
		return err
	}

	holderID, err = validateCRMIdentifier(operation, aliasResource, holderID, "holder id")
	if err != nil {
		return err
	}

	aliasID, err = validateCRMIdentifier(operation, aliasResource, aliasID, "alias id")
	if err != nil {
		return err
	}

	return core.DeleteWithHeaders(ctx, &s.BaseService, applyCRMDeleteOptions(aliasesItemPath(holderID, aliasID), opts), crmHeaders(orgID))
}

// DeleteRelatedParty removes a related party from an alias.
func (s *aliasesService) DeleteRelatedParty(ctx context.Context, orgID, holderID, aliasID, relatedPartyID string) error {
	const operation = "Aliases.DeleteRelatedParty"

	if s == nil {
		return core.ErrNilService
	}

	orgID, err := validateCRMOrgID(operation, aliasResource, orgID)
	if err != nil {
		return err
	}

	holderID, err = validateCRMIdentifier(operation, aliasResource, holderID, "holder id")
	if err != nil {
		return err
	}

	aliasID, err = validateCRMIdentifier(operation, aliasResource, aliasID, "alias id")
	if err != nil {
		return err
	}

	relatedPartyID, err = validateCRMIdentifier(operation, aliasResource, relatedPartyID, "related party id")
	if err != nil {
		return err
	}

	path := aliasesItemPath(holderID, aliasID) + "/related-parties/" + url.PathEscape(relatedPartyID)

	return core.DeleteWithHeaders(ctx, &s.BaseService, path, crmHeaders(orgID))
}
