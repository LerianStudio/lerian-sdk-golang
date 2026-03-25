// imports.go implements the importsServiceAPI for creating and monitoring data
// import jobs in the Matcher reconciliation service. Import jobs load records
// from external files for reconciliation processing.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// importsServiceAPI provides operations for creating and monitoring data
// import jobs.
type importsServiceAPI interface {
	// Create initiates a new data import job.
	Create(ctx context.Context, input *CreateImportInput) (*Import, error)

	// Get retrieves an import job by its unique identifier.
	Get(ctx context.Context, id string) (*Import, error)

	// List returns a paginated iterator over import jobs.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Import]

	// Cancel requests cancellation of an in-flight import job.
	Cancel(ctx context.Context, id string) (*Import, error)

	// GetStatus retrieves the current progress of an import job.
	GetStatus(ctx context.Context, id string) (*ImportStatus, error)
}

// importsService is the concrete implementation of [importsServiceAPI].
type importsService struct {
	core.BaseService
}

// newImportsService creates a new [importsServiceAPI] backed by the given
// [core.Backend].
func newImportsService(backend core.Backend) importsServiceAPI {
	return &importsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ importsServiceAPI = (*importsService)(nil)

// Create initiates a new data import job.
func (s *importsService) Create(ctx context.Context, input *CreateImportInput) (*Import, error) {
	const operation = "Imports.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Import", "input is required")
	}

	return core.Create[Import, CreateImportInput](ctx, &s.BaseService, "/imports", input)
}

// Get retrieves an import job by its unique identifier.
func (s *importsService) Get(ctx context.Context, id string) (*Import, error) {
	const operation = "Imports.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Import", "id is required")
	}

	return core.Get[Import](ctx, &s.BaseService, "/imports/"+url.PathEscape(id))
}

// List returns a paginated iterator over import jobs.
func (s *importsService) List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Import] {
	return core.List[Import](ctx, &s.BaseService, "/imports", opts)
}

// Cancel requests cancellation of an in-flight import job.
func (s *importsService) Cancel(ctx context.Context, id string) (*Import, error) {
	const operation = "Imports.Cancel"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Import", "id is required")
	}

	return core.Action[Import](ctx, &s.BaseService, "/imports/"+url.PathEscape(id)+"/cancel", nil)
}

// GetStatus retrieves the current progress of an import job.
func (s *importsService) GetStatus(ctx context.Context, id string) (*ImportStatus, error) {
	const operation = "Imports.GetStatus"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "ImportStatus", "id is required")
	}

	return core.Get[ImportStatus](ctx, &s.BaseService, "/imports/"+url.PathEscape(id)+"/status")
}
