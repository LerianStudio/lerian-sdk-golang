package reporter

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// ReportsService provides full CRUD access to Reporter report endpoints,
// plus a Download method for retrieving the generated report file.
type ReportsService interface {
	// Create generates a new report from the given input parameters.
	Create(ctx context.Context, input *CreateReportInput) (*Report, error)

	// Get retrieves a single report by its unique identifier.
	Get(ctx context.Context, id string) (*Report, error)

	// List returns a paginated iterator over all reports.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Report]

	// Update modifies an existing report's mutable fields.
	Update(ctx context.Context, id string, input *UpdateReportInput) (*Report, error)

	// Delete removes a report by ID.
	Delete(ctx context.Context, id string) error

	// Download retrieves the raw file content (PDF, CSV, XLSX, etc.)
	// of a generated report.
	Download(ctx context.Context, id string) ([]byte, error)
}

// reportsService is the concrete implementation of [ReportsService].
type reportsService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ ReportsService = (*reportsService)(nil)

// newReportsService constructs a [ReportsService] backed by the given
// [core.Backend].
func newReportsService(backend core.Backend) ReportsService {
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
func (s *reportsService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Report] {
	return core.List[Report](ctx, &s.BaseService, "/reports", opts)
}

// Update modifies an existing report.
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

// Delete removes a report by ID.
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

	return s.Backend.CallRaw(ctx, "GET", "/reports/"+url.PathEscape(id)+"/download", nil)
}
