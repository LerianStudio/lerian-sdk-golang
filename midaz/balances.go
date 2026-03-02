// balances.go implements the BalancesService for managing account balance
// entries within a ledger.
//
// IMPORTANT: Unlike other services in this package that use the onboarding
// backend, BalancesService routes all requests through the **transaction
// backend** because balances are managed by the Midaz transaction
// microservice.
//
// In addition to standard CRUD, the service provides lookup-by-alias,
// lookup-by-external-code, and lookup-by-account-id endpoints for resolving
// balances via business identifiers rather than the Midaz-internal UUID.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// BalancesService provides CRUD and lookup operations for balances.
type BalancesService interface {
	// Create creates a new balance entry within the specified ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateBalanceInput) (*Balance, error)

	// Get retrieves a balance by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Balance, error)

	// List returns a paginated iterator over balances in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[Balance]

	// Update partially updates an existing balance.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateBalanceInput) (*Balance, error)

	// Delete removes a balance by its unique identifier.
	Delete(ctx context.Context, orgID, ledgerID, id string) error

	// GetByAlias retrieves a balance by its account alias.
	GetByAlias(ctx context.Context, orgID, ledgerID, alias string) (*Balance, error)

	// GetByExternalCode retrieves a balance by its external code.
	GetByExternalCode(ctx context.Context, orgID, ledgerID, code string) (*Balance, error)

	// GetByAccountID retrieves a balance by the associated account ID.
	GetByAccountID(ctx context.Context, orgID, ledgerID, accountID string) (*Balance, error)
}

// balancesService is the concrete implementation of [BalancesService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
// Note: this service is wired to the **transaction** backend.
type balancesService struct {
	core.BaseService
}

// newBalancesService creates a new [BalancesService] backed by the given
// transaction [core.Backend].
func newBalancesService(backend core.Backend) BalancesService {
	return &balancesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ BalancesService = (*balancesService)(nil)

// balancesBasePath builds the base path for balance operations scoped to
// an organization and ledger.
func balancesBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/balances"
}

const balanceResource = "Balance"

// Create creates a new balance entry within the specified ledger.
func (s *balancesService) Create(ctx context.Context, orgID, ledgerID string, input *CreateBalanceInput) (*Balance, error) {
	const operation = "Balances.Create"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "input is required")
	}

	return core.Create[Balance, CreateBalanceInput](ctx, &s.BaseService, balancesBasePath(orgID, ledgerID), input)
}

// Get retrieves a balance by its unique identifier.
func (s *balancesService) Get(ctx context.Context, orgID, ledgerID, id string) (*Balance, error) {
	const operation = "Balances.Get"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "id is required")
	}

	return core.Get[Balance](ctx, &s.BaseService, balancesBasePath(orgID, ledgerID)+"/"+url.PathEscape(id))
}

// List returns a paginated iterator over balances in a ledger.
func (s *balancesService) List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[Balance] {
	if orgID == "" || ledgerID == "" {
		return pagination.NewIterator[Balance](func(_ context.Context, _ string) ([]Balance, string, error) {
			return nil, "", sdkerrors.NewValidation("Balances.List", balanceResource, "organization ID and ledger ID are required")
		})
	}

	return core.List[Balance](ctx, &s.BaseService, balancesBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing balance.
func (s *balancesService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateBalanceInput) (*Balance, error) {
	const operation = "Balances.Update"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "input is required")
	}

	return core.Update[Balance, UpdateBalanceInput](ctx, &s.BaseService, balancesBasePath(orgID, ledgerID)+"/"+url.PathEscape(id), input)
}

// Delete removes a balance by its unique identifier.
func (s *balancesService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "Balances.Delete"

	if orgID == "" {
		return sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, balanceResource, "id is required")
	}

	return core.Delete(ctx, &s.BaseService, balancesBasePath(orgID, ledgerID)+"/"+url.PathEscape(id))
}

// GetByAlias retrieves a balance by its account alias.
func (s *balancesService) GetByAlias(ctx context.Context, orgID, ledgerID, alias string) (*Balance, error) {
	const operation = "Balances.GetByAlias"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if alias == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "alias is required")
	}

	return core.Get[Balance](ctx, &s.BaseService, balancesBasePath(orgID, ledgerID)+"/alias/"+url.PathEscape(alias))
}

// GetByExternalCode retrieves a balance by its external code.
func (s *balancesService) GetByExternalCode(ctx context.Context, orgID, ledgerID, code string) (*Balance, error) {
	const operation = "Balances.GetByExternalCode"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if code == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "external code is required")
	}

	return core.Get[Balance](ctx, &s.BaseService, balancesBasePath(orgID, ledgerID)+"/external-code/"+url.PathEscape(code))
}

// GetByAccountID retrieves a balance by the associated account ID.
func (s *balancesService) GetByAccountID(ctx context.Context, orgID, ledgerID, accountID string) (*Balance, error) {
	const operation = "Balances.GetByAccountID"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if accountID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "account id is required")
	}

	return core.Get[Balance](ctx, &s.BaseService, balancesBasePath(orgID, ledgerID)+"/account/"+url.PathEscape(accountID))
}
