// balances.go implements the balancesServiceAPI for managing account balance
// entries within a ledger.
//
// IMPORTANT: Unlike other services in this package that use the onboarding
// backend, balancesServiceAPI routes all requests through the **transaction
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

// balancesServiceAPI provides CRUD and explicit plural lookup operations for balances.
type balancesServiceAPI interface {
	// Get retrieves a balance by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Balance, error)

	// List returns a paginated iterator over balances in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Balance]

	// Update partially updates an existing balance.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateBalanceInput) (*Balance, error)

	// Delete removes a balance by its unique identifier.
	Delete(ctx context.Context, orgID, ledgerID, id string) error

	// CreateForAccount creates a new additional balance for the specified
	// account using the nested account route required by the Midaz API.
	CreateForAccount(ctx context.Context, orgID, ledgerID, accountID string, input *CreateBalanceInput) (*Balance, error)

	// ListByAlias retrieves all balances associated with an account alias.
	ListByAlias(ctx context.Context, orgID, ledgerID, alias string) ([]Balance, error)

	// ListByExternalCode retrieves all balances associated with an account
	// external code.
	ListByExternalCode(ctx context.Context, orgID, ledgerID, code string) ([]Balance, error)

	// ListByAccountID retrieves all balances associated with an account ID.
	ListByAccountID(ctx context.Context, orgID, ledgerID, accountID string) ([]Balance, error)
}

// balancesService is the concrete implementation of [balancesServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
// Note: this service is wired to the **transaction** backend.
type balancesService struct {
	core.BaseService
}

// newBalancesService creates a new [balancesServiceAPI] backed by the given
// transaction [core.Backend].
func newBalancesService(backend core.Backend) balancesServiceAPI {
	return &balancesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ balancesServiceAPI = (*balancesService)(nil)

// balancesBasePath builds the base path for balance operations scoped to
// an organization and ledger.
func balancesBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/balances"
}

// balancesAccountBasePath builds the base path for balance operations nested
// under a specific account, as required by the Transaction API.
func balancesAccountBasePath(orgID, ledgerID, accountID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/accounts/" + url.PathEscape(accountID) + "/balances"
}

// balancesAliasPath builds the path for retrieving balances via account alias.
func balancesAliasPath(orgID, ledgerID, alias string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/accounts/alias/" + url.PathEscape(alias) + "/balances"
}

// balancesExternalCodePath builds the path for retrieving balances via external code.
func balancesExternalCodePath(orgID, ledgerID, code string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/accounts/external/" + url.PathEscape(code) + "/balances"
}

const balanceResource = "Balance"

type createAdditionalBalanceRequest struct {
	Key            string `json:"key"`
	AllowSending   *bool  `json:"allowSending,omitempty"`
	AllowReceiving *bool  `json:"allowReceiving,omitempty"`
}

type balancesLookupResponse struct {
	Items []Balance `json:"items"`
}

// CreateForAccount creates a new additional balance for the specified account.
func (s *balancesService) CreateForAccount(ctx context.Context, orgID, ledgerID, accountID string, input *CreateBalanceInput) (*Balance, error) {
	const operation = "Balances.CreateForAccount"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if accountID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "account id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "input is required")
	}

	if input.Key == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "key is required")
	}

	request := createAdditionalBalanceRequest{
		Key:            input.Key,
		AllowSending:   input.AllowSending,
		AllowReceiving: input.AllowReceiving,
	}

	return core.Create[Balance, createAdditionalBalanceRequest](ctx, &s.BaseService, balancesAccountBasePath(orgID, ledgerID, accountID), &request)
}

// Get retrieves a balance by its unique identifier.
func (s *balancesService) Get(ctx context.Context, orgID, ledgerID, id string) (*Balance, error) {
	const operation = "Balances.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

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
func (s *balancesService) List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Balance] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[Balance](err)
	}

	if orgID == "" || ledgerID == "" {
		return newErrorIterator[Balance](sdkerrors.NewValidation("Balances.List", balanceResource, "organization ID and ledger ID are required"))
	}

	return core.List[Balance](ctx, &s.BaseService, balancesBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing balance.
func (s *balancesService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateBalanceInput) (*Balance, error) {
	const operation = "Balances.Update"

	if err := ensureService(s); err != nil {
		return nil, err
	}

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

	if err := ensureService(s); err != nil {
		return err
	}

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

func (s *balancesService) listByLookupPath(ctx context.Context, path string) ([]Balance, error) {
	if err := ensureService(s); err != nil {
		return nil, err
	}

	response, err := core.Get[balancesLookupResponse](ctx, &s.BaseService, path)
	if err != nil {
		return nil, err
	}

	if response.Items == nil {
		return []Balance{}, nil
	}

	return response.Items, nil
}

// ListByAlias retrieves all balances by account alias.
func (s *balancesService) ListByAlias(ctx context.Context, orgID, ledgerID, alias string) ([]Balance, error) {
	const operation = "Balances.ListByAlias"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if alias == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "alias is required")
	}

	return s.listByLookupPath(ctx, balancesAliasPath(orgID, ledgerID, alias))
}

// ListByExternalCode retrieves all balances by account external code.
func (s *balancesService) ListByExternalCode(ctx context.Context, orgID, ledgerID, code string) ([]Balance, error) {
	const operation = "Balances.ListByExternalCode"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if code == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "external code is required")
	}

	return s.listByLookupPath(ctx, balancesExternalCodePath(orgID, ledgerID, code))
}

// ListByAccountID retrieves all balances by account ID.
func (s *balancesService) ListByAccountID(ctx context.Context, orgID, ledgerID, accountID string) ([]Balance, error) {
	const operation = "Balances.ListByAccountID"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "ledger id is required")
	}

	if accountID == "" {
		return nil, sdkerrors.NewValidation(operation, balanceResource, "account id is required")
	}

	return s.listByLookupPath(ctx, balancesAccountBasePath(orgID, ledgerID, accountID))
}
