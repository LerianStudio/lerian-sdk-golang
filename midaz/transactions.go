// transactions.go implements the TransactionsService for managing atomic
// financial movements within a ledger. Transactions are composed of one or
// more operations (debit/credit legs) and follow a state machine: once
// created they can be committed, cancelled, or reverted but never deleted.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// TransactionsService provides CRUD operations (minus Delete) and state
// machine actions for transactions scoped to an organization and ledger.
// Transactions are immutable once created and cannot be deleted; instead
// they transition through states via Commit, Cancel, and Revert actions.
type TransactionsService interface {
	// Create creates a new transaction within the specified organization and ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateTransactionInput) (*Transaction, error)

	// Get retrieves a transaction by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error)

	// List returns a paginated iterator over transactions in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[Transaction]

	// Update partially updates an existing transaction (e.g., description or metadata).
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateTransactionInput) (*Transaction, error)

	// Commit transitions the transaction to committed state, finalizing
	// all its operations and applying balance changes.
	Commit(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error)

	// Cancel transitions the transaction to cancelled state, releasing
	// any held balances without applying the operations.
	Cancel(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error)

	// Revert creates a reversal of a previously committed transaction,
	// undoing its balance effects with compensating operations.
	Revert(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error)
}

// transactionsService is the concrete implementation of [TransactionsService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type transactionsService struct {
	core.BaseService
}

// newTransactionsService creates a new [TransactionsService] backed by the
// given transaction [core.Backend].
func newTransactionsService(backend core.Backend) TransactionsService {
	return &transactionsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ TransactionsService = (*transactionsService)(nil)

// transactionsBasePath builds the transactions collection path for the given
// organization and ledger.
func transactionsBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/transactions"
}

// transactionsItemPath builds the path for a specific transaction.
func transactionsItemPath(orgID, ledgerID, id string) string {
	return transactionsBasePath(orgID, ledgerID) + "/" + url.PathEscape(id)
}

const transactionResource = "Transaction"

// Create creates a new transaction within the specified organization and ledger.
func (s *transactionsService) Create(ctx context.Context, orgID, ledgerID string, input *CreateTransactionInput) (*Transaction, error) {
	const operation = "Transactions.Create"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "input is required")
	}

	return core.Create[Transaction, CreateTransactionInput](ctx, &s.BaseService, transactionsBasePath(orgID, ledgerID), input)
}

// Get retrieves a transaction by its unique identifier.
func (s *transactionsService) Get(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error) {
	const operation = "Transactions.Get"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	return core.Get[Transaction](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id))
}

// List returns a paginated iterator over transactions in a ledger.
func (s *transactionsService) List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[Transaction] {
	if orgID == "" || ledgerID == "" {
		return pagination.NewIterator[Transaction](func(_ context.Context, _ string) ([]Transaction, string, error) {
			return nil, "", sdkerrors.NewValidation("Transactions.List", transactionResource, "organization ID and ledger ID are required")
		})
	}

	return core.List[Transaction](ctx, &s.BaseService, transactionsBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing transaction.
func (s *transactionsService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateTransactionInput) (*Transaction, error) {
	const operation = "Transactions.Update"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "input is required")
	}

	return core.Update[Transaction, UpdateTransactionInput](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id), input)
}

// Commit transitions the transaction to committed state.
func (s *transactionsService) Commit(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error) {
	const operation = "Transactions.Commit"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	return core.Action[Transaction](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id)+"/commit", nil)
}

// Cancel transitions the transaction to cancelled state.
func (s *transactionsService) Cancel(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error) {
	const operation = "Transactions.Cancel"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	return core.Action[Transaction](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id)+"/cancel", nil)
}

// Revert creates a reversal of a previously committed transaction.
func (s *transactionsService) Revert(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error) {
	const operation = "Transactions.Revert"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	return core.Action[Transaction](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id)+"/revert", nil)
}
