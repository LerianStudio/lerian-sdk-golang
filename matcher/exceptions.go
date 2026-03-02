// exceptions.go implements the ExceptionsService for managing reconciliation
// exceptions in the Matcher service. Exceptions represent anomalies or
// discrepancies detected during reconciliation that require human review,
// and support approval/rejection workflows, reassignment, and bulk operations.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// ExceptionsService provides CRUD, approval, rejection, reassignment,
// bulk operations, and analytics for reconciliation exceptions.
type ExceptionsService interface {
	// Create creates a new reconciliation exception from the given input.
	Create(ctx context.Context, input *CreateExceptionInput) (*Exception, error)

	// Get retrieves a reconciliation exception by its unique identifier.
	Get(ctx context.Context, id string) (*Exception, error)

	// List returns a paginated iterator over reconciliation exceptions.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Exception]

	// Update partially updates an existing reconciliation exception.
	Update(ctx context.Context, id string, input *UpdateExceptionInput) (*Exception, error)

	// Delete removes a reconciliation exception by its unique identifier.
	Delete(ctx context.Context, id string) error

	// Approve approves a reconciliation exception, marking it as resolved.
	Approve(ctx context.Context, id string) (*Exception, error)

	// Reject rejects a reconciliation exception with a reason.
	Reject(ctx context.Context, id string, input *RejectExceptionInput) (*Exception, error)

	// Reassign reassigns a reconciliation exception to a different user.
	Reassign(ctx context.Context, id string, input *ReassignExceptionInput) (*Exception, error)

	// BulkApprove approves multiple exceptions at once.
	BulkApprove(ctx context.Context, input *BulkExceptionInput) (*BulkExceptionResult, error)

	// BulkReject rejects multiple exceptions at once with a shared reason.
	BulkReject(ctx context.Context, input *BulkRejectInput) (*BulkExceptionResult, error)

	// BulkReassign reassigns multiple exceptions at once to a target user.
	BulkReassign(ctx context.Context, input *BulkReassignInput) (*BulkExceptionResult, error)

	// ListByContext returns a paginated iterator over exceptions within
	// a specific reconciliation context.
	ListByContext(ctx context.Context, contextID string, opts *models.ListOptions) *pagination.Iterator[Exception]

	// GetStatistics retrieves aggregate statistics about exceptions.
	GetStatistics(ctx context.Context) (*ExceptionStatistics, error)
}

// exceptionsService is the concrete implementation of [ExceptionsService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type exceptionsService struct {
	core.BaseService
}

// newExceptionsService creates a new [ExceptionsService] backed by the given
// Matcher [core.Backend].
func newExceptionsService(backend core.Backend) ExceptionsService {
	return &exceptionsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ ExceptionsService = (*exceptionsService)(nil)

// Create creates a new reconciliation exception from the given input.
func (s *exceptionsService) Create(ctx context.Context, input *CreateExceptionInput) (*Exception, error) {
	const operation = "Exceptions.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Exception", "input is required")
	}

	return core.Create[Exception, CreateExceptionInput](ctx, &s.BaseService, "/exceptions", input)
}

// Get retrieves a reconciliation exception by its unique identifier.
func (s *exceptionsService) Get(ctx context.Context, id string) (*Exception, error) {
	const operation = "Exceptions.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Exception", "id is required")
	}

	return core.Get[Exception](ctx, &s.BaseService, "/exceptions/"+url.PathEscape(id))
}

// List returns a paginated iterator over reconciliation exceptions.
func (s *exceptionsService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Exception] {
	return core.List[Exception](ctx, &s.BaseService, "/exceptions", opts)
}

// Update partially updates an existing reconciliation exception.
func (s *exceptionsService) Update(ctx context.Context, id string, input *UpdateExceptionInput) (*Exception, error) {
	const operation = "Exceptions.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Exception", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Exception", "input is required")
	}

	return core.Update[Exception, UpdateExceptionInput](ctx, &s.BaseService, "/exceptions/"+url.PathEscape(id), input)
}

// Delete removes a reconciliation exception by its unique identifier.
func (s *exceptionsService) Delete(ctx context.Context, id string) error {
	const operation = "Exceptions.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Exception", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/exceptions/"+url.PathEscape(id))
}

// Approve approves a reconciliation exception, marking it as resolved.
func (s *exceptionsService) Approve(ctx context.Context, id string) (*Exception, error) {
	const operation = "Exceptions.Approve"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Exception", "id is required")
	}

	return core.Action[Exception](ctx, &s.BaseService, "/exceptions/"+url.PathEscape(id)+"/approve", nil)
}

// Reject rejects a reconciliation exception with a reason.
func (s *exceptionsService) Reject(ctx context.Context, id string, input *RejectExceptionInput) (*Exception, error) {
	const operation = "Exceptions.Reject"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Exception", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Exception", "input is required")
	}

	return core.Action[Exception](ctx, &s.BaseService, "/exceptions/"+url.PathEscape(id)+"/reject", input)
}

// Reassign reassigns a reconciliation exception to a different user.
func (s *exceptionsService) Reassign(ctx context.Context, id string, input *ReassignExceptionInput) (*Exception, error) {
	const operation = "Exceptions.Reassign"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Exception", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Exception", "input is required")
	}

	return core.Action[Exception](ctx, &s.BaseService, "/exceptions/"+url.PathEscape(id)+"/reassign", input)
}

// BulkApprove approves multiple exceptions at once.
func (s *exceptionsService) BulkApprove(ctx context.Context, input *BulkExceptionInput) (*BulkExceptionResult, error) {
	const operation = "Exceptions.BulkApprove"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Exception", "input is required")
	}

	return core.Create[BulkExceptionResult, BulkExceptionInput](ctx, &s.BaseService, "/exceptions/bulk/approve", input)
}

// BulkReject rejects multiple exceptions at once with a shared reason.
func (s *exceptionsService) BulkReject(ctx context.Context, input *BulkRejectInput) (*BulkExceptionResult, error) {
	const operation = "Exceptions.BulkReject"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Exception", "input is required")
	}

	return core.Create[BulkExceptionResult, BulkRejectInput](ctx, &s.BaseService, "/exceptions/bulk/reject", input)
}

// BulkReassign reassigns multiple exceptions at once to a target user.
func (s *exceptionsService) BulkReassign(ctx context.Context, input *BulkReassignInput) (*BulkExceptionResult, error) {
	const operation = "Exceptions.BulkReassign"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Exception", "input is required")
	}

	return core.Create[BulkExceptionResult, BulkReassignInput](ctx, &s.BaseService, "/exceptions/bulk/reassign", input)
}

// ListByContext returns a paginated iterator over exceptions within
// a specific reconciliation context.
func (s *exceptionsService) ListByContext(ctx context.Context, contextID string, opts *models.ListOptions) *pagination.Iterator[Exception] {
	if contextID == "" {
		return pagination.NewIterator[Exception](func(_ context.Context, _ string) ([]Exception, string, error) {
			return nil, "", sdkerrors.NewValidation("Exceptions.ListByContext", "Exception", "context ID is required")
		})
	}

	return core.List[Exception](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/exceptions", opts)
}

// GetStatistics retrieves aggregate statistics about exceptions.
func (s *exceptionsService) GetStatistics(ctx context.Context) (*ExceptionStatistics, error) {
	return core.Get[ExceptionStatistics](ctx, &s.BaseService, "/exceptions/statistics")
}
