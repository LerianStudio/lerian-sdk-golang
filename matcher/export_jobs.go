// export_jobs.go implements the exportJobsServiceAPI for managing data export
// jobs in the Matcher service. Export jobs produce downloadable files in
// various formats (CSV, XLSX, etc.) containing reconciliation data.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// exportJobsServiceAPI provides operations for creating, monitoring, cancelling,
// and downloading data export jobs.
type exportJobsServiceAPI interface {
	// Create creates a new data export job from the given input.
	Create(ctx context.Context, input *CreateExportJobInput) (*ExportJob, error)

	// Get retrieves a data export job by its unique identifier.
	Get(ctx context.Context, id string) (*ExportJob, error)

	// List returns a paginated iterator over data export jobs.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[ExportJob]

	// Cancel cancels a pending or in-progress export job.
	Cancel(ctx context.Context, id string) (*ExportJob, error)

	// Download retrieves the raw file content of a completed export job.
	// The returned byte slice contains the export in whatever format was
	// requested (CSV, XLSX, etc.). The caller is responsible for writing
	// the bytes to a file or stream.
	Download(ctx context.Context, id string) ([]byte, error)
}

// exportJobsService is the concrete implementation of [exportJobsServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type exportJobsService struct {
	core.BaseService
}

// newExportJobsService creates a new [exportJobsServiceAPI] backed by the given
// Matcher [core.Backend].
func newExportJobsService(backend core.Backend) exportJobsServiceAPI {
	return &exportJobsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ exportJobsServiceAPI = (*exportJobsService)(nil)

// Create creates a new data export job from the given input.
func (s *exportJobsService) Create(ctx context.Context, input *CreateExportJobInput) (*ExportJob, error) {
	const operation = "ExportJobs.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "ExportJob", "input is required")
	}

	return core.Create[ExportJob, CreateExportJobInput](ctx, &s.BaseService, "/export-jobs", input)
}

// Get retrieves a data export job by its unique identifier.
func (s *exportJobsService) Get(ctx context.Context, id string) (*ExportJob, error) {
	const operation = "ExportJobs.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "ExportJob", "id is required")
	}

	return core.Get[ExportJob](ctx, &s.BaseService, "/export-jobs/"+url.PathEscape(id))
}

// List returns a paginated iterator over data export jobs.
func (s *exportJobsService) List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[ExportJob] {
	return core.List[ExportJob](ctx, &s.BaseService, "/export-jobs", opts)
}

// Cancel cancels a pending or in-progress export job.
func (s *exportJobsService) Cancel(ctx context.Context, id string) (*ExportJob, error) {
	const operation = "ExportJobs.Cancel"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "ExportJob", "id is required")
	}

	return core.Action[ExportJob](ctx, &s.BaseService, "/export-jobs/"+url.PathEscape(id)+"/cancel", nil)
}

// Download retrieves the raw file bytes of a completed export job. The
// returned byte slice contains the export in whatever format was requested
// (CSV, XLSX, etc.). The caller is responsible for writing the bytes to a
// file or stream.
func (s *exportJobsService) Download(ctx context.Context, id string) ([]byte, error) {
	const operation = "ExportJobs.Download"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "ExportJob", "id is required")
	}

	backend, err := core.ResolveBackend(&s.BaseService)
	if err != nil {
		return nil, err
	}

	res, err := backend.Do(ctx, core.Request{Method: "GET", Path: "/export-jobs/" + url.PathEscape(id) + "/download"})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
