// account_types.go implements the accountTypesServiceAPI for managing
// account type classifications within a ledger (e.g., "deposit",
// "savings", "external"). Account types define the categories that
// accounts can belong to.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// accountTypesServiceAPI provides CRUD operations for account types scoped to
// an organization and ledger.
type accountTypesServiceAPI interface {
	// Create creates a new account type within the specified organization and ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateAccountTypeInput) (*AccountType, error)

	// Get retrieves an account type by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*AccountType, error)

	// List returns a paginated iterator over account types in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[AccountType]

	// Update partially updates an existing account type.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateAccountTypeInput) (*AccountType, error)

	// Delete removes an account type by its unique identifier.
	Delete(ctx context.Context, orgID, ledgerID, id string) error
}

// accountTypesService is the concrete implementation of [accountTypesServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type accountTypesService struct {
	core.BaseService
}

// newAccountTypesService creates a new [accountTypesServiceAPI] backed by the
// given onboarding [core.Backend].
func newAccountTypesService(backend core.Backend) accountTypesServiceAPI {
	return &accountTypesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ accountTypesServiceAPI = (*accountTypesService)(nil)

// accountTypesBasePath builds the account-types collection path for the
// given organization and ledger.
func accountTypesBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/account-types"
}

// accountTypesItemPath builds the path for a specific account type
// within an organization and ledger.
func accountTypesItemPath(orgID, ledgerID, id string) string {
	return accountTypesBasePath(orgID, ledgerID) + "/" + url.PathEscape(id)
}

const accountTypeResource = "AccountType"

// Create creates a new account type within the specified organization and ledger.
func (s *accountTypesService) Create(ctx context.Context, orgID, ledgerID string, input *CreateAccountTypeInput) (*AccountType, error) {
	const operation = "AccountTypes.Create"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "input is required")
	}

	return core.Create[AccountType, CreateAccountTypeInput](ctx, &s.BaseService, accountTypesBasePath(orgID, ledgerID), input)
}

// Get retrieves an account type by its unique identifier.
func (s *accountTypesService) Get(ctx context.Context, orgID, ledgerID, id string) (*AccountType, error) {
	const operation = "AccountTypes.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "id is required")
	}

	return core.Get[AccountType](ctx, &s.BaseService, accountTypesItemPath(orgID, ledgerID, id))
}

// List returns a paginated iterator over account types in a ledger.
func (s *accountTypesService) List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[AccountType] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[AccountType](err)
	}

	if orgID == "" || ledgerID == "" {
		return newErrorIterator[AccountType](sdkerrors.NewValidation("AccountTypes.List", accountTypeResource, "organization ID and ledger ID are required"))
	}

	return core.List[AccountType](ctx, &s.BaseService, accountTypesBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing account type.
func (s *accountTypesService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateAccountTypeInput) (*AccountType, error) {
	const operation = "AccountTypes.Update"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, accountTypeResource, "input is required")
	}

	return core.Update[AccountType, UpdateAccountTypeInput](ctx, &s.BaseService, accountTypesItemPath(orgID, ledgerID, id), input)
}

// Delete removes an account type by its unique identifier.
func (s *accountTypesService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "AccountTypes.Delete"

	if err := ensureService(s); err != nil {
		return err
	}

	if orgID == "" {
		return sdkerrors.NewValidation(operation, accountTypeResource, "organization id is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, accountTypeResource, "ledger id is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, accountTypeResource, "id is required")
	}

	return core.Delete(ctx, &s.BaseService, accountTypesItemPath(orgID, ledgerID, id))
}
