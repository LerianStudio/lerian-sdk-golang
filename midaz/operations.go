// operations.go implements the operationsServiceAPI for querying and updating
// individual debit/credit legs within transactions. Operations are created
// as side-effects of transactions and cannot be created or deleted
// independently. They support retrieval, filtered listing by transaction
// or account, and partial updates to metadata and description.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// operationsServiceAPI provides query and update operations for individual
// debit/credit legs within a ledger. Operations are created as side-effects
// of transactions and cannot be created or deleted directly, but their
// metadata and description can be updated.
type operationsServiceAPI interface {
	// Get retrieves an operation by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Operation, error)

	// List returns a paginated iterator over all operations in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Operation]

	// ListByTransaction returns a paginated iterator over operations
	// belonging to a specific transaction.
	ListByTransaction(ctx context.Context, orgID, ledgerID, transactionID string, opts *models.CursorListOptions) *pagination.Iterator[Operation]

	// ListByAccount returns a paginated iterator over operations
	// affecting a specific account.
	ListByAccount(ctx context.Context, orgID, ledgerID, accountID string, opts *models.CursorListOptions) *pagination.Iterator[Operation]

	// Update partially updates an existing operation's metadata.
	Update(ctx context.Context, orgID, ledgerID, transactionID, operationID string, input *UpdateOperationInput) (*Operation, error)
}

// operationsService is the concrete implementation of [operationsServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type operationsService struct {
	core.BaseService
}

// newOperationsService creates a new [operationsServiceAPI] backed by the
// given transaction [core.Backend].
func newOperationsService(backend core.Backend) operationsServiceAPI {
	return &operationsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ operationsServiceAPI = (*operationsService)(nil)

// operationsBasePath builds the operations collection path for the given
// organization and ledger.
func operationsBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/operations"
}

// operationsItemPath builds the path for a specific operation.
func operationsItemPath(orgID, ledgerID, id string) string {
	return operationsBasePath(orgID, ledgerID) + "/" + url.PathEscape(id)
}

// operationsByTransactionPath builds the path for listing operations
// belonging to a specific transaction.
func operationsByTransactionPath(orgID, ledgerID, transactionID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/transactions/" + url.PathEscape(transactionID) + "/operations"
}

// operationsByAccountPath builds the path for listing operations
// affecting a specific account.
func operationsByAccountPath(orgID, ledgerID, accountID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/accounts/" + url.PathEscape(accountID) + "/operations"
}

const operationResource = "Operation"

// Get retrieves an operation by its unique identifier.
func (s *operationsService) Get(ctx context.Context, orgID, ledgerID, id string) (*Operation, error) {
	const operation = "Operations.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, operationResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, operationResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, operationResource, "operation id is required")
	}

	return core.Get[Operation](ctx, &s.BaseService, operationsItemPath(orgID, ledgerID, id))
}

// List returns a paginated iterator over all operations in a ledger.
func (s *operationsService) List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Operation] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[Operation](err)
	}

	if orgID == "" || ledgerID == "" {
		return newErrorIterator[Operation](sdkerrors.NewValidation("Operations.List", operationResource, "organization ID and ledger ID are required"))
	}

	return core.List[Operation](ctx, &s.BaseService, operationsBasePath(orgID, ledgerID), opts)
}

// ListByTransaction returns a paginated iterator over operations
// belonging to a specific transaction.
func (s *operationsService) ListByTransaction(ctx context.Context, orgID, ledgerID, transactionID string, opts *models.CursorListOptions) *pagination.Iterator[Operation] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[Operation](err)
	}

	if orgID == "" || ledgerID == "" || transactionID == "" {
		return newErrorIterator[Operation](sdkerrors.NewValidation("Operations.ListByTransaction", operationResource, "organization ID, ledger ID, and transaction ID are required"))
	}

	return core.List[Operation](ctx, &s.BaseService, operationsByTransactionPath(orgID, ledgerID, transactionID), opts)
}

// ListByAccount returns a paginated iterator over operations
// affecting a specific account.
func (s *operationsService) ListByAccount(ctx context.Context, orgID, ledgerID, accountID string, opts *models.CursorListOptions) *pagination.Iterator[Operation] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[Operation](err)
	}

	if orgID == "" || ledgerID == "" || accountID == "" {
		return newErrorIterator[Operation](sdkerrors.NewValidation("Operations.ListByAccount", operationResource, "organization ID, ledger ID, and account ID are required"))
	}

	return core.List[Operation](ctx, &s.BaseService, operationsByAccountPath(orgID, ledgerID, accountID), opts)
}

// operationsByTransactionItemPath builds the path for a specific operation
// within a transaction.
func operationsByTransactionItemPath(orgID, ledgerID, transactionID, operationID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/transactions/" + url.PathEscape(transactionID) + "/operations/" + url.PathEscape(operationID)
}

// Update partially updates an existing operation's metadata.
func (s *operationsService) Update(ctx context.Context, orgID, ledgerID, transactionID, operationID string, input *UpdateOperationInput) (*Operation, error) {
	const operation = "Operations.Update"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, operationResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, operationResource, "ledger id is required")
	}

	if transactionID == "" {
		return nil, sdkerrors.NewValidation(operation, operationResource, "transaction id is required")
	}

	if operationID == "" {
		return nil, sdkerrors.NewValidation(operation, operationResource, "operation id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, operationResource, "input is required")
	}

	return core.Update[Operation, UpdateOperationInput](ctx, &s.BaseService, operationsByTransactionItemPath(orgID, ledgerID, transactionID, operationID), input)
}
