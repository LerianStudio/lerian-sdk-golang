// governance.go implements the governanceServiceAPI for accessing audit logs and
// archive operations within the Matcher reconciliation service. All methods
// are read-only.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// governanceServiceAPI provides access to audit logs and archive operations
// for reconciliation governance. All methods are read-only.
type governanceServiceAPI interface {
	// ListArchives returns a paginated iterator over archived reconciliation
	// record batches.
	ListArchives(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Archive]

	// GetArchive retrieves a single archive by its unique identifier.
	GetArchive(ctx context.Context, id string) (*Archive, error)

	// ListAuditLogs returns a paginated iterator over audit trail entries.
	ListAuditLogs(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[AuditLog]

	// GetAuditLog retrieves a single audit log entry by its unique identifier.
	GetAuditLog(ctx context.Context, id string) (*AuditLog, error)
}

// governanceService is the concrete implementation of [governanceServiceAPI].
type governanceService struct {
	core.BaseService
}

// newGovernanceService creates a new [governanceServiceAPI] backed by the given
// [core.Backend].
func newGovernanceService(backend core.Backend) governanceServiceAPI {
	return &governanceService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ governanceServiceAPI = (*governanceService)(nil)

// ListArchives returns a paginated iterator over archived reconciliation
// record batches.
func (s *governanceService) ListArchives(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Archive] {
	return core.List[Archive](ctx, &s.BaseService, "/archives", opts)
}

// GetArchive retrieves a single archive by its unique identifier.
func (s *governanceService) GetArchive(ctx context.Context, id string) (*Archive, error) {
	const operation = "Governance.GetArchive"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Archive", "id is required")
	}

	return core.Get[Archive](ctx, &s.BaseService, "/archives/"+url.PathEscape(id))
}

// ListAuditLogs returns a paginated iterator over audit trail entries.
func (s *governanceService) ListAuditLogs(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[AuditLog] {
	return core.List[AuditLog](ctx, &s.BaseService, "/audit-logs", opts)
}

// GetAuditLog retrieves a single audit log entry by its unique identifier.
func (s *governanceService) GetAuditLog(ctx context.Context, id string) (*AuditLog, error) {
	const operation = "Governance.GetAuditLog"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "AuditLog", "id is required")
	}

	return core.Get[AuditLog](ctx, &s.BaseService, "/audit-logs/"+url.PathEscape(id))
}
