package leriantest

import (
	"context"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// newFakeMatcherClient constructs a [matcher.Client] with all service fields
// backed by in-memory fakes.
func newFakeMatcherClient(cfg *fakeConfig) *matcher.Client {
	return &matcher.Client{
		Contexts:        &fakeMatcherContexts{store: newFakeStore[matcher.Context](), cfg: cfg},
		Rules:           &fakeMatcherRules{store: newFakeStore[matcher.Rule](), cfg: cfg},
		Schedules:       &fakeMatcherSchedules{store: newFakeStore[matcher.Schedule](), cfg: cfg},
		Sources:         &fakeMatcherSources{store: newFakeStore[matcher.Source](), cfg: cfg},
		SourceFieldMaps: &fakeMatcherSourceFieldMaps{store: newFakeStore[matcher.SourceFieldMap](), cfg: cfg},
		FeeSchedules:    &fakeMatcherFeeSchedules{store: newFakeStore[matcher.FeeSchedule](), cfg: cfg},
		FieldMaps:       &fakeMatcherFieldMaps{store: newFakeStore[matcher.FieldMap](), cfg: cfg},
		ExportJobs:      &fakeMatcherExportJobs{store: newFakeStore[matcher.ExportJob](), cfg: cfg},
		Disputes:        &fakeMatcherDisputes{store: newFakeStore[matcher.Dispute](), cfg: cfg},
		Exceptions:      &fakeMatcherExceptions{store: newFakeStore[matcher.Exception](), cfg: cfg},
		Governance:      &fakeMatcherGovernance{archives: newFakeStore[matcher.Archive](), auditLogs: newFakeStore[matcher.AuditLog](), cfg: cfg},
		Imports:         &fakeMatcherImports{store: newFakeStore[matcher.Import](), cfg: cfg},
		Matching:        &fakeMatcherMatching{cfg: cfg},
		Reports:         &fakeMatcherReports{cfg: cfg},
	}
}

// ---------------------------------------------------------------------------
// Contexts
// ---------------------------------------------------------------------------

type fakeMatcherContexts struct {
	store *fakeStore[matcher.Context]
	cfg   *fakeConfig
}

var _ matcher.ContextsService = (*fakeMatcherContexts)(nil)

func (f *fakeMatcherContexts) Create(_ context.Context, input *matcher.CreateContextInput) (*matcher.Context, error) {
	if err := f.cfg.injectedError("matcher.Contexts.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	c := matcher.Context{ID: generateID("mctx"), Name: input.Name, CreatedAt: now, UpdatedAt: now}
	f.store.Set(c.ID, c)

	return &c, nil
}

func (f *fakeMatcherContexts) Get(_ context.Context, id string) (*matcher.Context, error) {
	if err := f.cfg.injectedError("matcher.Contexts.Get"); err != nil {
		return nil, err
	}

	c, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Contexts.Get", "Context", id)
	}

	return &c, nil
}

func (f *fakeMatcherContexts) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.Context] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherContexts) Update(_ context.Context, id string, input *matcher.UpdateContextInput) (*matcher.Context, error) {
	if err := f.cfg.injectedError("matcher.Contexts.Update"); err != nil {
		return nil, err
	}

	c, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Contexts.Update", "Context", id)
	}

	if input.Name != nil {
		c.Name = *input.Name
	}

	c.UpdatedAt = time.Now()
	f.store.Set(id, c)

	return &c, nil
}

func (f *fakeMatcherContexts) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("matcher.Contexts.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Contexts.Delete", "Context", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeMatcherContexts) Clone(_ context.Context, id string) (*matcher.Context, error) {
	if err := f.cfg.injectedError("matcher.Contexts.Clone"); err != nil {
		return nil, err
	}

	c, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Contexts.Clone", "Context", id)
	}

	now := time.Now()
	clone := c
	clone.ID = generateID("mctx")
	clone.CreatedAt = now
	clone.UpdatedAt = now
	f.store.Set(clone.ID, clone)

	return &clone, nil
}

// ---------------------------------------------------------------------------
// Rules
// ---------------------------------------------------------------------------

type fakeMatcherRules struct {
	store *fakeStore[matcher.Rule]
	cfg   *fakeConfig
}

var _ matcher.RulesService = (*fakeMatcherRules)(nil)

func (f *fakeMatcherRules) Create(_ context.Context, input *matcher.CreateRuleInput) (*matcher.Rule, error) {
	if err := f.cfg.injectedError("matcher.Rules.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	r := matcher.Rule{ID: generateID("mrule"), Name: input.Name, CreatedAt: now, UpdatedAt: now}
	f.store.Set(r.ID, r)

	return &r, nil
}

func (f *fakeMatcherRules) Get(_ context.Context, id string) (*matcher.Rule, error) {
	if err := f.cfg.injectedError("matcher.Rules.Get"); err != nil {
		return nil, err
	}

	r, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Rules.Get", "Rule", id)
	}

	return &r, nil
}

func (f *fakeMatcherRules) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.Rule] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherRules) Update(_ context.Context, id string, input *matcher.UpdateRuleInput) (*matcher.Rule, error) {
	if err := f.cfg.injectedError("matcher.Rules.Update"); err != nil {
		return nil, err
	}

	r, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Rules.Update", "Rule", id)
	}

	if input.Name != nil {
		r.Name = *input.Name
	}

	r.UpdatedAt = time.Now()
	f.store.Set(id, r)

	return &r, nil
}

func (f *fakeMatcherRules) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("matcher.Rules.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Rules.Delete", "Rule", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeMatcherRules) Reorder(_ context.Context, _ string, _ *matcher.ReorderRulesInput) error {
	if err := f.cfg.injectedError("matcher.Rules.Reorder"); err != nil {
		return err
	}

	return nil // no-op in fake
}

// ---------------------------------------------------------------------------
// Schedules
// ---------------------------------------------------------------------------

type fakeMatcherSchedules struct {
	store *fakeStore[matcher.Schedule]
	cfg   *fakeConfig
}

var _ matcher.SchedulesService = (*fakeMatcherSchedules)(nil)

func (f *fakeMatcherSchedules) Create(_ context.Context, input *matcher.CreateScheduleInput) (*matcher.Schedule, error) {
	if err := f.cfg.injectedError("matcher.Schedules.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	s := matcher.Schedule{ID: generateID("msched"), Name: input.Name, CreatedAt: now, UpdatedAt: now}
	f.store.Set(s.ID, s)

	return &s, nil
}

func (f *fakeMatcherSchedules) Get(_ context.Context, id string) (*matcher.Schedule, error) {
	if err := f.cfg.injectedError("matcher.Schedules.Get"); err != nil {
		return nil, err
	}

	s, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Schedules.Get", "Schedule", id)
	}

	return &s, nil
}

func (f *fakeMatcherSchedules) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.Schedule] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherSchedules) Update(_ context.Context, id string, _ *matcher.UpdateScheduleInput) (*matcher.Schedule, error) {
	if err := f.cfg.injectedError("matcher.Schedules.Update"); err != nil {
		return nil, err
	}

	s, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Schedules.Update", "Schedule", id)
	}

	s.UpdatedAt = time.Now()
	f.store.Set(id, s)

	return &s, nil
}

func (f *fakeMatcherSchedules) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("matcher.Schedules.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Schedules.Delete", "Schedule", id)
	}

	f.store.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// Sources
// ---------------------------------------------------------------------------

type fakeMatcherSources struct {
	store *fakeStore[matcher.Source]
	cfg   *fakeConfig
}

var _ matcher.SourcesService = (*fakeMatcherSources)(nil)

func (f *fakeMatcherSources) Create(_ context.Context, input *matcher.CreateSourceInput) (*matcher.Source, error) {
	if err := f.cfg.injectedError("matcher.Sources.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	s := matcher.Source{ID: generateID("msrc"), Name: input.Name, CreatedAt: now, UpdatedAt: now}
	f.store.Set(s.ID, s)

	return &s, nil
}

func (f *fakeMatcherSources) Get(_ context.Context, id string) (*matcher.Source, error) {
	if err := f.cfg.injectedError("matcher.Sources.Get"); err != nil {
		return nil, err
	}

	s, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Sources.Get", "Source", id)
	}

	return &s, nil
}

func (f *fakeMatcherSources) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.Source] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherSources) Update(_ context.Context, id string, _ *matcher.UpdateSourceInput) (*matcher.Source, error) {
	if err := f.cfg.injectedError("matcher.Sources.Update"); err != nil {
		return nil, err
	}

	s, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Sources.Update", "Source", id)
	}

	s.UpdatedAt = time.Now()
	f.store.Set(id, s)

	return &s, nil
}

func (f *fakeMatcherSources) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("matcher.Sources.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Sources.Delete", "Source", id)
	}

	f.store.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// SourceFieldMaps
// ---------------------------------------------------------------------------

type fakeMatcherSourceFieldMaps struct {
	store *fakeStore[matcher.SourceFieldMap]
	cfg   *fakeConfig
}

var _ matcher.SourceFieldMapsService = (*fakeMatcherSourceFieldMaps)(nil)

func (f *fakeMatcherSourceFieldMaps) Create(_ context.Context, _ *matcher.CreateSourceFieldMapInput) (*matcher.SourceFieldMap, error) {
	if err := f.cfg.injectedError("matcher.SourceFieldMaps.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	sfm := matcher.SourceFieldMap{ID: generateID("msfm"), CreatedAt: now, UpdatedAt: now}
	f.store.Set(sfm.ID, sfm)

	return &sfm, nil
}

func (f *fakeMatcherSourceFieldMaps) Get(_ context.Context, id string) (*matcher.SourceFieldMap, error) {
	if err := f.cfg.injectedError("matcher.SourceFieldMaps.Get"); err != nil {
		return nil, err
	}

	sfm, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("SourceFieldMaps.Get", "SourceFieldMap", id)
	}

	return &sfm, nil
}

func (f *fakeMatcherSourceFieldMaps) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.SourceFieldMap] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherSourceFieldMaps) Update(_ context.Context, id string, _ *matcher.UpdateSourceFieldMapInput) (*matcher.SourceFieldMap, error) {
	if err := f.cfg.injectedError("matcher.SourceFieldMaps.Update"); err != nil {
		return nil, err
	}

	sfm, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("SourceFieldMaps.Update", "SourceFieldMap", id)
	}

	sfm.UpdatedAt = time.Now()
	f.store.Set(id, sfm)

	return &sfm, nil
}

func (f *fakeMatcherSourceFieldMaps) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("matcher.SourceFieldMaps.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("SourceFieldMaps.Delete", "SourceFieldMap", id)
	}

	f.store.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// FeeSchedules
// ---------------------------------------------------------------------------

type fakeMatcherFeeSchedules struct {
	store *fakeStore[matcher.FeeSchedule]
	cfg   *fakeConfig
}

var _ matcher.FeeSchedulesService = (*fakeMatcherFeeSchedules)(nil)

func (f *fakeMatcherFeeSchedules) Create(_ context.Context, input *matcher.CreateFeeScheduleInput) (*matcher.FeeSchedule, error) {
	if err := f.cfg.injectedError("matcher.FeeSchedules.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	fs := matcher.FeeSchedule{ID: generateID("mfs"), Name: input.Name, CreatedAt: now, UpdatedAt: now}
	f.store.Set(fs.ID, fs)

	return &fs, nil
}

func (f *fakeMatcherFeeSchedules) Get(_ context.Context, id string) (*matcher.FeeSchedule, error) {
	if err := f.cfg.injectedError("matcher.FeeSchedules.Get"); err != nil {
		return nil, err
	}

	fs, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("FeeSchedules.Get", "FeeSchedule", id)
	}

	return &fs, nil
}

func (f *fakeMatcherFeeSchedules) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.FeeSchedule] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherFeeSchedules) Update(_ context.Context, id string, _ *matcher.UpdateFeeScheduleInput) (*matcher.FeeSchedule, error) {
	if err := f.cfg.injectedError("matcher.FeeSchedules.Update"); err != nil {
		return nil, err
	}

	fs, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("FeeSchedules.Update", "FeeSchedule", id)
	}

	fs.UpdatedAt = time.Now()
	f.store.Set(id, fs)

	return &fs, nil
}

func (f *fakeMatcherFeeSchedules) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("matcher.FeeSchedules.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("FeeSchedules.Delete", "FeeSchedule", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeMatcherFeeSchedules) Simulate(_ context.Context, _ string, _ *matcher.SimulateFeeScheduleInput) (*matcher.FeeSimulationResult, error) {
	if err := f.cfg.injectedError("matcher.FeeSchedules.Simulate"); err != nil {
		return nil, err
	}

	return &matcher.FeeSimulationResult{}, nil
}

// ---------------------------------------------------------------------------
// FieldMaps
// ---------------------------------------------------------------------------

type fakeMatcherFieldMaps struct {
	store *fakeStore[matcher.FieldMap]
	cfg   *fakeConfig
}

var _ matcher.FieldMapsService = (*fakeMatcherFieldMaps)(nil)

func (f *fakeMatcherFieldMaps) Create(_ context.Context, _ *matcher.CreateFieldMapInput) (*matcher.FieldMap, error) {
	if err := f.cfg.injectedError("matcher.FieldMaps.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	fm := matcher.FieldMap{ID: generateID("mfm"), CreatedAt: now, UpdatedAt: now}
	f.store.Set(fm.ID, fm)

	return &fm, nil
}

func (f *fakeMatcherFieldMaps) Get(_ context.Context, id string) (*matcher.FieldMap, error) {
	if err := f.cfg.injectedError("matcher.FieldMaps.Get"); err != nil {
		return nil, err
	}

	fm, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("FieldMaps.Get", "FieldMap", id)
	}

	return &fm, nil
}

func (f *fakeMatcherFieldMaps) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.FieldMap] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherFieldMaps) Update(_ context.Context, id string, _ *matcher.UpdateFieldMapInput) (*matcher.FieldMap, error) {
	if err := f.cfg.injectedError("matcher.FieldMaps.Update"); err != nil {
		return nil, err
	}

	fm, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("FieldMaps.Update", "FieldMap", id)
	}

	fm.UpdatedAt = time.Now()
	f.store.Set(id, fm)

	return &fm, nil
}

func (f *fakeMatcherFieldMaps) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("matcher.FieldMaps.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("FieldMaps.Delete", "FieldMap", id)
	}

	f.store.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// ExportJobs
// ---------------------------------------------------------------------------

type fakeMatcherExportJobs struct {
	store *fakeStore[matcher.ExportJob]
	cfg   *fakeConfig
}

var _ matcher.ExportJobsService = (*fakeMatcherExportJobs)(nil)

func (f *fakeMatcherExportJobs) Create(_ context.Context, _ *matcher.CreateExportJobInput) (*matcher.ExportJob, error) {
	if err := f.cfg.injectedError("matcher.ExportJobs.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	ej := matcher.ExportJob{ID: generateID("mej"), CreatedAt: now, UpdatedAt: now}
	f.store.Set(ej.ID, ej)

	return &ej, nil
}

func (f *fakeMatcherExportJobs) Get(_ context.Context, id string) (*matcher.ExportJob, error) {
	if err := f.cfg.injectedError("matcher.ExportJobs.Get"); err != nil {
		return nil, err
	}

	ej, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("ExportJobs.Get", "ExportJob", id)
	}

	return &ej, nil
}

func (f *fakeMatcherExportJobs) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.ExportJob] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherExportJobs) Cancel(_ context.Context, id string) (*matcher.ExportJob, error) {
	if err := f.cfg.injectedError("matcher.ExportJobs.Cancel"); err != nil {
		return nil, err
	}

	ej, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("ExportJobs.Cancel", "ExportJob", id)
	}

	return &ej, nil
}

func (f *fakeMatcherExportJobs) Download(_ context.Context, id string) ([]byte, error) {
	if err := f.cfg.injectedError("matcher.ExportJobs.Download"); err != nil {
		return nil, err
	}

	if _, ok := f.store.Get(id); !ok {
		return nil, sdkerrors.NewNotFound("ExportJobs.Download", "ExportJob", id)
	}

	return []byte("fake-export-data"), nil
}

// ---------------------------------------------------------------------------
// Disputes
// ---------------------------------------------------------------------------

type fakeMatcherDisputes struct {
	store *fakeStore[matcher.Dispute]
	cfg   *fakeConfig
}

var _ matcher.DisputesService = (*fakeMatcherDisputes)(nil)

func (f *fakeMatcherDisputes) Create(_ context.Context, _ *matcher.CreateDisputeInput) (*matcher.Dispute, error) {
	if err := f.cfg.injectedError("matcher.Disputes.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	d := matcher.Dispute{ID: generateID("mdisp"), CreatedAt: now, UpdatedAt: now}
	f.store.Set(d.ID, d)

	return &d, nil
}

func (f *fakeMatcherDisputes) Get(_ context.Context, id string) (*matcher.Dispute, error) {
	if err := f.cfg.injectedError("matcher.Disputes.Get"); err != nil {
		return nil, err
	}

	d, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Disputes.Get", "Dispute", id)
	}

	return &d, nil
}

func (f *fakeMatcherDisputes) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.Dispute] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherDisputes) Update(_ context.Context, id string, _ *matcher.UpdateDisputeInput) (*matcher.Dispute, error) {
	if err := f.cfg.injectedError("matcher.Disputes.Update"); err != nil {
		return nil, err
	}

	d, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Disputes.Update", "Dispute", id)
	}

	d.UpdatedAt = time.Now()
	f.store.Set(id, d)

	return &d, nil
}

func (f *fakeMatcherDisputes) Resolve(_ context.Context, id string, _ *matcher.ResolveDisputeInput) (*matcher.Dispute, error) {
	if err := f.cfg.injectedError("matcher.Disputes.Resolve"); err != nil {
		return nil, err
	}

	d, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Disputes.Resolve", "Dispute", id)
	}

	return &d, nil
}

func (f *fakeMatcherDisputes) Escalate(_ context.Context, id string) (*matcher.Dispute, error) {
	if err := f.cfg.injectedError("matcher.Disputes.Escalate"); err != nil {
		return nil, err
	}

	d, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Disputes.Escalate", "Dispute", id)
	}

	return &d, nil
}

// ---------------------------------------------------------------------------
// Exceptions
// ---------------------------------------------------------------------------

type fakeMatcherExceptions struct {
	store *fakeStore[matcher.Exception]
	cfg   *fakeConfig
}

var _ matcher.ExceptionsService = (*fakeMatcherExceptions)(nil)

func (f *fakeMatcherExceptions) Create(_ context.Context, _ *matcher.CreateExceptionInput) (*matcher.Exception, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	e := matcher.Exception{ID: generateID("mexc"), CreatedAt: now, UpdatedAt: now}
	f.store.Set(e.ID, e)

	return &e, nil
}

func (f *fakeMatcherExceptions) Get(_ context.Context, id string) (*matcher.Exception, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.Get"); err != nil {
		return nil, err
	}

	e, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Exceptions.Get", "Exception", id)
	}

	return &e, nil
}

func (f *fakeMatcherExceptions) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.Exception] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherExceptions) Update(_ context.Context, id string, _ *matcher.UpdateExceptionInput) (*matcher.Exception, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.Update"); err != nil {
		return nil, err
	}

	e, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Exceptions.Update", "Exception", id)
	}

	e.UpdatedAt = time.Now()
	f.store.Set(id, e)

	return &e, nil
}

func (f *fakeMatcherExceptions) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("matcher.Exceptions.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Exceptions.Delete", "Exception", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeMatcherExceptions) Approve(_ context.Context, id string) (*matcher.Exception, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.Approve"); err != nil {
		return nil, err
	}

	e, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Exceptions.Approve", "Exception", id)
	}

	return &e, nil
}

func (f *fakeMatcherExceptions) Reject(_ context.Context, id string, _ *matcher.RejectExceptionInput) (*matcher.Exception, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.Reject"); err != nil {
		return nil, err
	}

	e, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Exceptions.Reject", "Exception", id)
	}

	return &e, nil
}

func (f *fakeMatcherExceptions) Reassign(_ context.Context, id string, _ *matcher.ReassignExceptionInput) (*matcher.Exception, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.Reassign"); err != nil {
		return nil, err
	}

	e, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Exceptions.Reassign", "Exception", id)
	}

	return &e, nil
}

func (f *fakeMatcherExceptions) BulkApprove(_ context.Context, _ *matcher.BulkExceptionInput) (*matcher.BulkExceptionResult, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.BulkApprove"); err != nil {
		return nil, err
	}

	return &matcher.BulkExceptionResult{}, nil
}

func (f *fakeMatcherExceptions) BulkReject(_ context.Context, _ *matcher.BulkRejectInput) (*matcher.BulkExceptionResult, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.BulkReject"); err != nil {
		return nil, err
	}

	return &matcher.BulkExceptionResult{}, nil
}

func (f *fakeMatcherExceptions) BulkReassign(_ context.Context, _ *matcher.BulkReassignInput) (*matcher.BulkExceptionResult, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.BulkReassign"); err != nil {
		return nil, err
	}

	return &matcher.BulkExceptionResult{}, nil
}

func (f *fakeMatcherExceptions) ListByContext(_ context.Context, _ string, opts *models.ListOptions) *pagination.Iterator[matcher.Exception] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherExceptions) GetStatistics(_ context.Context) (*matcher.ExceptionStatistics, error) {
	if err := f.cfg.injectedError("matcher.Exceptions.GetStatistics"); err != nil {
		return nil, err
	}

	return &matcher.ExceptionStatistics{}, nil
}

// ---------------------------------------------------------------------------
// Governance
// ---------------------------------------------------------------------------

type fakeMatcherGovernance struct {
	archives  *fakeStore[matcher.Archive]
	auditLogs *fakeStore[matcher.AuditLog]
	cfg       *fakeConfig
}

var _ matcher.GovernanceService = (*fakeMatcherGovernance)(nil)

func (f *fakeMatcherGovernance) ListArchives(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.Archive] {
	return f.archives.PaginatedIterator(opts)
}

func (f *fakeMatcherGovernance) GetArchive(_ context.Context, id string) (*matcher.Archive, error) {
	if err := f.cfg.injectedError("matcher.Governance.GetArchive"); err != nil {
		return nil, err
	}

	a, ok := f.archives.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Governance.GetArchive", "Archive", id)
	}

	return &a, nil
}

func (f *fakeMatcherGovernance) ListAuditLogs(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.AuditLog] {
	return f.auditLogs.PaginatedIterator(opts)
}

func (f *fakeMatcherGovernance) GetAuditLog(_ context.Context, id string) (*matcher.AuditLog, error) {
	if err := f.cfg.injectedError("matcher.Governance.GetAuditLog"); err != nil {
		return nil, err
	}

	al, ok := f.auditLogs.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Governance.GetAuditLog", "AuditLog", id)
	}

	return &al, nil
}

// ---------------------------------------------------------------------------
// Imports
// ---------------------------------------------------------------------------

type fakeMatcherImports struct {
	store *fakeStore[matcher.Import]
	cfg   *fakeConfig
}

var _ matcher.ImportsService = (*fakeMatcherImports)(nil)

func (f *fakeMatcherImports) Create(_ context.Context, _ *matcher.CreateImportInput) (*matcher.Import, error) {
	if err := f.cfg.injectedError("matcher.Imports.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	imp := matcher.Import{ID: generateID("mimp"), CreatedAt: now, UpdatedAt: now}
	f.store.Set(imp.ID, imp)

	return &imp, nil
}

func (f *fakeMatcherImports) Get(_ context.Context, id string) (*matcher.Import, error) {
	if err := f.cfg.injectedError("matcher.Imports.Get"); err != nil {
		return nil, err
	}

	imp, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Imports.Get", "Import", id)
	}

	return &imp, nil
}

func (f *fakeMatcherImports) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[matcher.Import] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeMatcherImports) Cancel(_ context.Context, id string) (*matcher.Import, error) {
	if err := f.cfg.injectedError("matcher.Imports.Cancel"); err != nil {
		return nil, err
	}

	imp, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Imports.Cancel", "Import", id)
	}

	return &imp, nil
}

func (f *fakeMatcherImports) GetStatus(_ context.Context, id string) (*matcher.ImportStatus, error) {
	if err := f.cfg.injectedError("matcher.Imports.GetStatus"); err != nil {
		return nil, err
	}

	if _, ok := f.store.Get(id); !ok {
		return nil, sdkerrors.NewNotFound("Imports.GetStatus", "Import", id)
	}

	return &matcher.ImportStatus{ID: id}, nil
}

// ---------------------------------------------------------------------------
// Matching
// ---------------------------------------------------------------------------

type fakeMatcherMatching struct {
	cfg *fakeConfig
}

var _ matcher.MatchingService = (*fakeMatcherMatching)(nil)

func (f *fakeMatcherMatching) Run(_ context.Context, contextID string) (*matcher.MatchResult, error) {
	if err := f.cfg.injectedError("matcher.Matching.Run"); err != nil {
		return nil, err
	}

	return &matcher.MatchResult{ContextID: contextID}, nil
}

func (f *fakeMatcherMatching) Manual(_ context.Context, _ *matcher.ManualMatchInput) (*matcher.MatchResult, error) {
	if err := f.cfg.injectedError("matcher.Matching.Manual"); err != nil {
		return nil, err
	}

	return &matcher.MatchResult{}, nil
}

func (f *fakeMatcherMatching) Adjust(_ context.Context, _ *matcher.AdjustmentInput) (*matcher.Adjustment, error) {
	if err := f.cfg.injectedError("matcher.Matching.Adjust"); err != nil {
		return nil, err
	}

	return &matcher.Adjustment{ID: generateID("madj")}, nil
}

// ---------------------------------------------------------------------------
// Reports
// ---------------------------------------------------------------------------

type fakeMatcherReports struct {
	cfg *fakeConfig
}

var _ matcher.ReportsService = (*fakeMatcherReports)(nil)

func (f *fakeMatcherReports) GetSummary(_ context.Context, _ string) (*matcher.ReconciliationSummary, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetSummary"); err != nil {
		return nil, err
	}

	return &matcher.ReconciliationSummary{}, nil
}

func (f *fakeMatcherReports) GetMatchRate(_ context.Context, _ string) (*matcher.MatchRateReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetMatchRate"); err != nil {
		return nil, err
	}

	return &matcher.MatchRateReport{}, nil
}

func (f *fakeMatcherReports) GetExceptionTrend(_ context.Context, _ string) (*matcher.ExceptionTrendReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetExceptionTrend"); err != nil {
		return nil, err
	}

	return &matcher.ExceptionTrendReport{}, nil
}

func (f *fakeMatcherReports) GetAgingAnalysis(_ context.Context, _ string) (*matcher.AgingAnalysisReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetAgingAnalysis"); err != nil {
		return nil, err
	}

	return &matcher.AgingAnalysisReport{}, nil
}

func (f *fakeMatcherReports) GetSourceComparison(_ context.Context, _ string) (*matcher.SourceComparisonReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetSourceComparison"); err != nil {
		return nil, err
	}

	return &matcher.SourceComparisonReport{}, nil
}

func (f *fakeMatcherReports) GetVolumeAnalysis(_ context.Context, _ string) (*matcher.VolumeAnalysisReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetVolumeAnalysis"); err != nil {
		return nil, err
	}

	return &matcher.VolumeAnalysisReport{}, nil
}

func (f *fakeMatcherReports) GetDisputeMetrics(_ context.Context, _ string) (*matcher.DisputeMetricsReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetDisputeMetrics"); err != nil {
		return nil, err
	}

	return &matcher.DisputeMetricsReport{}, nil
}

func (f *fakeMatcherReports) GetFeeAnalysis(_ context.Context, _ string) (*matcher.FeeAnalysisReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetFeeAnalysis"); err != nil {
		return nil, err
	}

	return &matcher.FeeAnalysisReport{}, nil
}

func (f *fakeMatcherReports) GetReconciliationHistory(_ context.Context, _ string, _ *models.ListOptions) *pagination.Iterator[matcher.ReconciliationHistoryEntry] {
	return pagination.NewIteratorFromSlice[matcher.ReconciliationHistoryEntry](nil)
}

func (f *fakeMatcherReports) GetPerformanceMetrics(_ context.Context, _ string) (*matcher.PerformanceMetricsReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetPerformanceMetrics"); err != nil {
		return nil, err
	}

	return &matcher.PerformanceMetricsReport{}, nil
}

func (f *fakeMatcherReports) GetDashboard(_ context.Context, _ string) (*matcher.DashboardReport, error) {
	if err := f.cfg.injectedError("matcher.Reports.GetDashboard"); err != nil {
		return nil, err
	}

	return &matcher.DashboardReport{}, nil
}
