// operation_routes.go implements the operationRoutesServiceAPI for managing
// routing rules within transaction routes. An operation route maps a specific
// operation leg (debit or credit) to an account, defining how money flows
// when transactions are processed through a given transaction route.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// operationRoutesServiceAPI provides CRUD operations for operation routes
// scoped to an organization and ledger.
type operationRoutesServiceAPI interface {
	// Create creates a new operation route within the specified organization and ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateOperationRouteInput) (*OperationRoute, error)

	// Get retrieves an operation route by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*OperationRoute, error)

	// List returns a paginated iterator over operation routes in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[OperationRoute]

	// Update partially updates an existing operation route.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateOperationRouteInput) (*OperationRoute, error)

	// Delete removes an operation route by its unique identifier.
	Delete(ctx context.Context, orgID, ledgerID, id string) error
}

// operationRoutesService is the concrete implementation of [operationRoutesServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type operationRoutesService struct {
	core.BaseService
}

// newOperationRoutesService creates a new [operationRoutesServiceAPI] backed
// by the given transaction [core.Backend].
func newOperationRoutesService(backend core.Backend) operationRoutesServiceAPI {
	return &operationRoutesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ operationRoutesServiceAPI = (*operationRoutesService)(nil)

// operationRoutesBasePath builds the operation-routes collection path for
// the given organization and ledger.
func operationRoutesBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/operation-routes"
}

// operationRoutesItemPath builds the path for a specific operation route.
func operationRoutesItemPath(orgID, ledgerID, id string) string {
	return operationRoutesBasePath(orgID, ledgerID) + "/" + url.PathEscape(id)
}

const operationRouteResource = "OperationRoute"

// Create creates a new operation route within the specified organization and ledger.
func (s *operationRoutesService) Create(ctx context.Context, orgID, ledgerID string, input *CreateOperationRouteInput) (*OperationRoute, error) {
	const operation = "OperationRoutes.Create"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "input is required")
	}

	return core.Create[OperationRoute, CreateOperationRouteInput](ctx, &s.BaseService, operationRoutesBasePath(orgID, ledgerID), input)
}

// Get retrieves an operation route by its unique identifier.
func (s *operationRoutesService) Get(ctx context.Context, orgID, ledgerID, id string) (*OperationRoute, error) {
	const operation = "OperationRoutes.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "operation route id is required")
	}

	return core.Get[OperationRoute](ctx, &s.BaseService, operationRoutesItemPath(orgID, ledgerID, id))
}

// List returns a paginated iterator over operation routes in a ledger.
func (s *operationRoutesService) List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[OperationRoute] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[OperationRoute](err)
	}

	if orgID == "" || ledgerID == "" {
		return newErrorIterator[OperationRoute](sdkerrors.NewValidation("OperationRoutes.List", operationRouteResource, "organization ID and ledger ID are required"))
	}

	return core.List[OperationRoute](ctx, &s.BaseService, operationRoutesBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing operation route.
func (s *operationRoutesService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateOperationRouteInput) (*OperationRoute, error) {
	const operation = "OperationRoutes.Update"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "operation route id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, operationRouteResource, "input is required")
	}

	return core.Update[OperationRoute, UpdateOperationRouteInput](ctx, &s.BaseService, operationRoutesItemPath(orgID, ledgerID, id), input)
}

// Delete removes an operation route by its unique identifier.
func (s *operationRoutesService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "OperationRoutes.Delete"

	if err := ensureService(s); err != nil {
		return err
	}

	if orgID == "" {
		return sdkerrors.NewValidation(operation, operationRouteResource, "organization id is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, operationRouteResource, "ledger id is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, operationRouteResource, "operation route id is required")
	}

	return core.Delete(ctx, &s.BaseService, operationRoutesItemPath(orgID, ledgerID, id))
}
