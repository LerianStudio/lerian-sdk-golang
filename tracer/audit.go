package tracer

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// AuditEventsService provides read-only access to the audit trail and
// integrity verification of audit events. Every mutation in the system
// is tracked as an audit event with a timestamped record of who did
// what to which resource.
type AuditEventsService interface {
	// Get retrieves an audit event by its unique identifier.
	Get(ctx context.Context, id string) (*AuditEvent, error)

	// List returns a paginated iterator over audit events.
	// The server returns cursor-based pagination for audit events.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[AuditEvent]

	// Verify checks the integrity of an audit event by comparing its
	// content hash against the expected value, providing tamper-evidence
	// for the audit trail.
	Verify(ctx context.Context, id string) (*AuditVerification, error)
}

// auditEventsService is the concrete implementation of [AuditEventsService].
type auditEventsService struct {
	core.BaseService
}

// newAuditEventsService creates a new [AuditEventsService] backed by the given [core.Backend].
func newAuditEventsService(backend core.Backend) AuditEventsService {
	return &auditEventsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface check.
var _ AuditEventsService = (*auditEventsService)(nil)

func (s *auditEventsService) Get(ctx context.Context, id string) (*AuditEvent, error) {
	const operation = "AuditEvents.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "AuditEvent", "id is required")
	}

	return core.Get[AuditEvent](ctx, &s.BaseService, "/audit-events/"+url.PathEscape(id))
}

func (s *auditEventsService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[AuditEvent] {
	return core.List[AuditEvent](ctx, &s.BaseService, "/audit-events", opts)
}

func (s *auditEventsService) Verify(ctx context.Context, id string) (*AuditVerification, error) {
	const operation = "AuditEvents.Verify"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "AuditEvent", "id is required")
	}

	return core.Action[AuditVerification](ctx, &s.BaseService, "/audit-events/"+url.PathEscape(id)+"/verify", nil)
}
