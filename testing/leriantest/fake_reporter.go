package leriantest

import (
	"context"
	"io"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
	"github.com/LerianStudio/lerian-sdk-golang/reporter"
)

// newFakeReporterClient constructs a [reporter.Client] with all service
// fields backed by in-memory fakes.
func newFakeReporterClient(cfg *fakeConfig) *reporter.Client {
	return &reporter.Client{
		DataSources: &fakeReporterDataSources{store: newFakeStore[reporter.DataSource](), cfg: cfg},
		Reports:     &fakeReporterReports{store: newFakeStore[reporter.Report](), cfg: cfg},
		Templates:   &fakeReporterTemplates{store: newFakeStore[reporter.Template](), cfg: cfg},
	}
}

// ---------------------------------------------------------------------------
// DataSources (read-only)
// ---------------------------------------------------------------------------

type fakeReporterDataSources struct {
	store *fakeStore[reporter.DataSource]
	cfg   *fakeConfig
}

func (f *fakeReporterDataSources) Get(_ context.Context, id string) (*reporter.DataSource, error) {
	return fakeGetStored(f.cfg, "", "DataSources.Get", "DataSource", id, f.store)
}

func (f *fakeReporterDataSources) List(_ context.Context, opts *models.CursorListOptions) *pagination.Iterator[reporter.DataSource] {
	return fakeListStored(f.cfg, "", f.store, opts)
}

// ---------------------------------------------------------------------------
// Reports
// ---------------------------------------------------------------------------

type fakeReporterReports struct {
	store *fakeStore[reporter.Report]
	cfg   *fakeConfig
}

func (f *fakeReporterReports) Create(_ context.Context, input *reporter.CreateReportInput) (*reporter.Report, error) {
	if err := f.cfg.injectedError("reporter.Reports.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

	r := reporter.Report{
		ID:        generateID("rpt"),
		Name:      input.Name,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	f.store.Set(r.ID, r)

	return &r, nil
}

func (f *fakeReporterReports) Get(_ context.Context, id string) (*reporter.Report, error) {
	return fakeGetStored(f.cfg, "", "Reports.Get", "Report", id, f.store)
}

func (f *fakeReporterReports) List(_ context.Context, opts *models.CursorListOptions) *pagination.Iterator[reporter.Report] {
	return fakeListStored(f.cfg, "", f.store, opts)
}

func (f *fakeReporterReports) Update(_ context.Context, id string, _ *reporter.UpdateReportInput) (*reporter.Report, error) {
	return fakeMutateStored(f.cfg, "", "Reports.Update", "Report", id, f.store, func(r *reporter.Report) {
		r.UpdatedAt = time.Now()
	})
}

func (f *fakeReporterReports) Delete(_ context.Context, id string) error {
	return fakeDeleteStored(f.cfg, "", "Reports.Delete", "Report", id, f.store)
}

func (f *fakeReporterReports) Download(_ context.Context, id string) ([]byte, error) {
	if _, ok := f.store.Get(id); !ok {
		return nil, sdkerrors.NewNotFound("Reports.Download", "Report", id)
	}

	return []byte("fake-report-data"), nil
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

type fakeReporterTemplates struct {
	store *fakeStore[reporter.Template]
	cfg   *fakeConfig
}

func (f *fakeReporterTemplates) Create(_ context.Context, input *reporter.CreateTemplateInput, _ io.Reader) (*reporter.Template, error) {
	if err := f.cfg.injectedError("reporter.Templates.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

	t := reporter.Template{
		ID:        generateID("tmpl"),
		Name:      input.Name,
		Format:    input.Format,
		CreatedAt: now,
		UpdatedAt: now,
	}

	f.store.Set(t.ID, t)

	return &t, nil
}

func (f *fakeReporterTemplates) Get(_ context.Context, id string) (*reporter.Template, error) {
	return fakeGetStored(f.cfg, "", "Templates.Get", "Template", id, f.store)
}

func (f *fakeReporterTemplates) List(_ context.Context, opts *models.CursorListOptions) *pagination.Iterator[reporter.Template] {
	return fakeListStored(f.cfg, "", f.store, opts)
}

func (f *fakeReporterTemplates) Delete(_ context.Context, id string) error {
	return fakeDeleteStored(f.cfg, "", "Templates.Delete", "Template", id, f.store)
}
