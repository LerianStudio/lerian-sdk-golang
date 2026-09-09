package reporter

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// reportsServiceAPI provides access to Reporter report endpoints, plus a
// Download method for retrieving the generated report file.
//
// Reporter serves four report operations and only four: create, get, list and
// download. A generated report is retained, not edited -- the service registers
// no update and no delete operation for one, so this interface offers neither.
type reportsServiceAPI interface {
	// Create generates a new report from the given input parameters.
	Create(ctx context.Context, input *CreateReportInput) (*Report, error)

	// Get retrieves a single report by its unique identifier.
	Get(ctx context.Context, id string) (*Report, error)

	// List returns a paginated iterator over all reports.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Report]

	// Download retrieves the raw file content (PDF, CSV, XLSX, etc.)
	// of a generated report.
	Download(ctx context.Context, id string) ([]byte, error)
}

// reportsService is the concrete implementation of [reportsServiceAPI].
type reportsService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ reportsServiceAPI = (*reportsService)(nil)

// newReportsService constructs a [reportsServiceAPI] backed by the given
// [core.Backend].
func newReportsService(backend core.Backend) reportsServiceAPI {
	return &reportsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Create generates a new report.
func (s *reportsService) Create(ctx context.Context, input *CreateReportInput) (*Report, error) {
	const operation = "Reports.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Report", "input is required")
	}

	return core.Create[Report, CreateReportInput](ctx, &s.BaseService, "/reports", input)
}

// Get retrieves a single report by ID.
func (s *reportsService) Get(ctx context.Context, id string) (*Report, error) {
	const operation = "Reports.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "id is required")
	}

	return core.Get[Report](ctx, &s.BaseService, "/reports/"+url.PathEscape(id))
}

// List returns a paginated iterator over reports.
func (s *reportsService) List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Report] {
	return core.List[Report](ctx, &s.BaseService, "/reports", opts)
}

// Download retrieves the raw file bytes of a generated report. The returned
// byte slice contains the report in whatever format it was generated (PDF,
// CSV, XLSX, etc.). The caller is responsible for writing the bytes to a
// file or stream.
func (s *reportsService) Download(ctx context.Context, id string) ([]byte, error) {
	const operation = "Reports.Download"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "id is required")
	}

	backend, err := core.ResolveBackend(&s.BaseService)
	if err != nil {
		return nil, err
	}

	res, err := backend.Do(ctx, core.Request{Method: "GET", Path: "/reports/" + url.PathEscape(id) + "/download"})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
