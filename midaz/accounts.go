// accounts.go implements the AccountsService for managing financial accounts
// within a ledger. Accounts hold balances denominated in a specific asset and
// participate in transactions via operations.
//
// In addition to standard CRUD, the service provides lookup-by-alias and
// lookup-by-external-code endpoints that resolve accounts via their unique
// business identifiers rather than the Midaz-internal UUID.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// AccountsService provides CRUD and lookup operations for accounts.
type AccountsService interface {
	// Create creates a new account within the specified ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateAccountInput) (*Account, error)

	// Get retrieves an account by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Account, error)

	// List returns a paginated iterator over accounts in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[Account]

	// Update partially updates an existing account.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateAccountInput) (*Account, error)

	// Delete removes an account by its unique identifier.
	Delete(ctx context.Context, orgID, ledgerID, id string) error

	// GetByAlias retrieves an account by its unique alias.
	GetByAlias(ctx context.Context, orgID, ledgerID, alias string) (*Account, error)

	// GetByExternalCode retrieves an account by its external code.
	GetByExternalCode(ctx context.Context, orgID, ledgerID, code string) (*Account, error)
}

// accountsService is the concrete implementation of [AccountsService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type accountsService struct {
	core.BaseService
}

// newAccountsService creates a new [AccountsService] backed by the given
// onboarding [core.Backend].
func newAccountsService(backend core.Backend) AccountsService {
	return &accountsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ AccountsService = (*accountsService)(nil)

// accountsBasePath builds the base path for account operations scoped to
// an organization and ledger.
func accountsBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/accounts"
}

const accountResource = "Account"

// Create creates a new account within the specified ledger.
func (s *accountsService) Create(ctx context.Context, orgID, ledgerID string, input *CreateAccountInput) (*Account, error) {
	const operation = "Accounts.Create"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, accountResource, "input is required")
	}

	return core.Create[Account, CreateAccountInput](ctx, &s.BaseService, accountsBasePath(orgID, ledgerID), input)
}

// Get retrieves an account by its unique identifier.
func (s *accountsService) Get(ctx context.Context, orgID, ledgerID, id string) (*Account, error) {
	const operation = "Accounts.Get"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "id is required")
	}

	return core.Get[Account](ctx, &s.BaseService, accountsBasePath(orgID, ledgerID)+"/"+url.PathEscape(id))
}

// List returns a paginated iterator over accounts in a ledger.
func (s *accountsService) List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[Account] {
	if orgID == "" || ledgerID == "" {
		return pagination.NewIterator[Account](func(_ context.Context, _ string) ([]Account, string, error) {
			return nil, "", sdkerrors.NewValidation("Accounts.List", accountResource, "organization ID and ledger ID are required")
		})
	}

	return core.List[Account](ctx, &s.BaseService, accountsBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing account.
func (s *accountsService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateAccountInput) (*Account, error) {
	const operation = "Accounts.Update"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, accountResource, "input is required")
	}

	return core.Update[Account, UpdateAccountInput](ctx, &s.BaseService, accountsBasePath(orgID, ledgerID)+"/"+url.PathEscape(id), input)
}

// Delete removes an account by its unique identifier.
func (s *accountsService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "Accounts.Delete"

	if orgID == "" {
		return sdkerrors.NewValidation(operation, accountResource, "organization id is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, accountResource, "ledger id is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, accountResource, "id is required")
	}

	return core.Delete(ctx, &s.BaseService, accountsBasePath(orgID, ledgerID)+"/"+url.PathEscape(id))
}

// GetByAlias retrieves an account by its unique alias.
func (s *accountsService) GetByAlias(ctx context.Context, orgID, ledgerID, alias string) (*Account, error) {
	const operation = "Accounts.GetByAlias"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "ledger id is required")
	}

	if alias == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "alias is required")
	}

	return core.Get[Account](ctx, &s.BaseService, accountsBasePath(orgID, ledgerID)+"/alias/"+url.PathEscape(alias))
}

// GetByExternalCode retrieves an account by its external code.
func (s *accountsService) GetByExternalCode(ctx context.Context, orgID, ledgerID, code string) (*Account, error) {
	const operation = "Accounts.GetByExternalCode"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "ledger id is required")
	}

	if code == "" {
		return nil, sdkerrors.NewValidation(operation, accountResource, "external code is required")
	}

	return core.Get[Account](ctx, &s.BaseService, accountsBasePath(orgID, ledgerID)+"/external-code/"+url.PathEscape(code))
}
