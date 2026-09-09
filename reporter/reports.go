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
// Reporter serves four report operations: create, get, list and download. A
// generated report is retained, not edited: the service exposes no way to
// change or remove one. The Update and Delete methods below predate that
// finding and fail against any deployment.
type reportsServiceAPI interface {
	// Create generates a new report from the given input parameters.
	Create(ctx context.Context, input *CreateReportInput) (*Report, error)

	// Get retrieves a single report by its unique identifier.
	Get(ctx context.Context, id string) (*Report, error)

	// List returns a paginated iterator over all reports.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Report]

	// Update sends a PATCH request that Reporter does not answer. The service
	// registers no update operation for a report, so this call fails against
	// every deployment. Do not build on it.
	//
	// Deprecated: Reporter has no report update operation. This method cannot
	// succeed.
	Update(ctx context.Context, id string, input *UpdateReportInput) (*Report, error)

	// Delete sends a DELETE request that Reporter does not answer. The service
	// registers no delete operation for a report, so this call fails against
	// every deployment. A retention or correction flow cannot be built on it.
	//
	// Deprecated: Reporter has no report delete operation. This method cannot
	// succeed.
	Delete(ctx context.Context, id string) error

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

// Update sends PATCH /reports/{id}, an operation Reporter does not register.
//
// Deprecated: Reporter has no report update operation. This method cannot
// succeed.
func (s *reportsService) Update(ctx context.Context, id string, input *UpdateReportInput) (*Report, error) {
	const operation = "Reports.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Report", "input is required")
	}

	return core.Update[Report, UpdateReportInput](ctx, &s.BaseService, "/reports/"+url.PathEscape(id), input)
}

// Delete sends DELETE /reports/{id}, an operation Reporter does not register.
//
// Deprecated: Reporter has no report delete operation. This method cannot
// succeed.
func (s *reportsService) Delete(ctx context.Context, id string) error {
	const operation = "Reports.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Report", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/reports/"+url.PathEscape(id))
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
