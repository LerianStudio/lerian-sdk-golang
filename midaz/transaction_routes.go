// transaction_routes.go implements the TransactionRoutesService for managing
// routing templates that govern how transactions of a given type are processed.
// Transaction routes define the rules and account mappings applied when a
// transaction is created with a matching transaction type.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// TransactionRoutesService provides CRUD operations for transaction routes
// scoped to an organization and ledger.
type TransactionRoutesService interface {
	// Create creates a new transaction route within the specified organization and ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateTransactionRouteInput) (*TransactionRoute, error)

	// Get retrieves a transaction route by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*TransactionRoute, error)

	// List returns a paginated iterator over transaction routes in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[TransactionRoute]

	// Update partially updates an existing transaction route.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateTransactionRouteInput) (*TransactionRoute, error)

	// Delete removes a transaction route by its unique identifier.
	Delete(ctx context.Context, orgID, ledgerID, id string) error
}

// transactionRoutesService is the concrete implementation of [TransactionRoutesService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type transactionRoutesService struct {
	core.BaseService
}

// newTransactionRoutesService creates a new [TransactionRoutesService] backed
// by the given transaction [core.Backend].
func newTransactionRoutesService(backend core.Backend) TransactionRoutesService {
	return &transactionRoutesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ TransactionRoutesService = (*transactionRoutesService)(nil)

// transactionRoutesBasePath builds the transaction-routes collection path for
// the given organization and ledger.
func transactionRoutesBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/transaction-routes"
}

// transactionRoutesItemPath builds the path for a specific transaction route.
func transactionRoutesItemPath(orgID, ledgerID, id string) string {
	return transactionRoutesBasePath(orgID, ledgerID) + "/" + url.PathEscape(id)
}

const transactionRouteResource = "TransactionRoute"

// Create creates a new transaction route within the specified organization and ledger.
func (s *transactionRoutesService) Create(ctx context.Context, orgID, ledgerID string, input *CreateTransactionRouteInput) (*TransactionRoute, error) {
	const operation = "TransactionRoutes.Create"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "input is required")
	}

	return core.Create[TransactionRoute, CreateTransactionRouteInput](ctx, &s.BaseService, transactionRoutesBasePath(orgID, ledgerID), input)
}

// Get retrieves a transaction route by its unique identifier.
func (s *transactionRoutesService) Get(ctx context.Context, orgID, ledgerID, id string) (*TransactionRoute, error) {
	const operation = "TransactionRoutes.Get"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "transaction route id is required")
	}

	return core.Get[TransactionRoute](ctx, &s.BaseService, transactionRoutesItemPath(orgID, ledgerID, id))
}

// List returns a paginated iterator over transaction routes in a ledger.
func (s *transactionRoutesService) List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[TransactionRoute] {
	if orgID == "" || ledgerID == "" {
		return pagination.NewIterator[TransactionRoute](func(_ context.Context, _ string) ([]TransactionRoute, string, error) {
			return nil, "", sdkerrors.NewValidation("TransactionRoutes.List", transactionRouteResource, "organization ID and ledger ID are required")
		})
	}

	return core.List[TransactionRoute](ctx, &s.BaseService, transactionRoutesBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing transaction route.
func (s *transactionRoutesService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateTransactionRouteInput) (*TransactionRoute, error) {
	const operation = "TransactionRoutes.Update"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "transaction route id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionRouteResource, "input is required")
	}

	return core.Update[TransactionRoute, UpdateTransactionRouteInput](ctx, &s.BaseService, transactionRoutesItemPath(orgID, ledgerID, id), input)
}

// Delete removes a transaction route by its unique identifier.
func (s *transactionRoutesService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "TransactionRoutes.Delete"

	if orgID == "" {
		return sdkerrors.NewValidation(operation, transactionRouteResource, "organization id is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, transactionRouteResource, "ledger id is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, transactionRouteResource, "transaction route id is required")
	}

	return core.Delete(ctx, &s.BaseService, transactionRoutesItemPath(orgID, ledgerID, id))
}
