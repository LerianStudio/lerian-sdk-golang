package leriantest

import (
	"context"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
)

// newFakeTracerClient constructs a [tracer.Client] with all service fields
// backed by in-memory fakes.
func newFakeTracerClient(cfg *fakeConfig) *tracer.Client {
	return &tracer.Client{
		Rules:       &fakeTracerRules{store: newFakeStore[tracer.Rule](), cfg: cfg},
		Limits:      &fakeTracerLimits{store: newFakeStore[tracer.Limit](), cfg: cfg},
		Validations: &fakeTracerValidations{store: newFakeStore[tracer.Validation](), cfg: cfg},
		AuditEvents: &fakeTracerAuditEvents{store: newFakeStore[tracer.AuditEvent](), cfg: cfg},
	}
}

// ---------------------------------------------------------------------------
// Rules
// ---------------------------------------------------------------------------

type fakeTracerRules struct {
	store *fakeStore[tracer.Rule]
	cfg   *fakeConfig
}

var _ tracer.RulesService = (*fakeTracerRules)(nil)

func (f *fakeTracerRules) Create(_ context.Context, input *tracer.CreateRuleInput) (*tracer.Rule, error) {
	if err := f.cfg.injectedError("tracer.Rules.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	r := tracer.Rule{ID: generateID("trule"), Name: input.Name, Status: "DRAFT", CreatedAt: now, UpdatedAt: now}
	f.store.Set(r.ID, r)

	return &r, nil
}

func (f *fakeTracerRules) Get(_ context.Context, id string) (*tracer.Rule, error) {
	r, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Rules.Get", "Rule", id)
	}

	return &r, nil
}

func (f *fakeTracerRules) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[tracer.Rule] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeTracerRules) Update(_ context.Context, id string, input *tracer.UpdateRuleInput) (*tracer.Rule, error) {
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

func (f *fakeTracerRules) Delete(_ context.Context, id string) error {
	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Rules.Delete", "Rule", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeTracerRules) Activate(_ context.Context, id string) (*tracer.Rule, error) {
	r, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Rules.Activate", "Rule", id)
	}

	r.Status = "ACTIVE"
	r.UpdatedAt = time.Now()
	f.store.Set(id, r)

	return &r, nil
}

func (f *fakeTracerRules) Deactivate(_ context.Context, id string) (*tracer.Rule, error) {
	r, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Rules.Deactivate", "Rule", id)
	}

	r.Status = "INACTIVE"
	r.UpdatedAt = time.Now()
	f.store.Set(id, r)

	return &r, nil
}

func (f *fakeTracerRules) Draft(_ context.Context, id string) (*tracer.Rule, error) {
	r, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Rules.Draft", "Rule", id)
	}

	r.Status = "DRAFT"
	r.UpdatedAt = time.Now()
	f.store.Set(id, r)

	return &r, nil
}

// ---------------------------------------------------------------------------
// Limits
// ---------------------------------------------------------------------------

type fakeTracerLimits struct {
	store *fakeStore[tracer.Limit]
	cfg   *fakeConfig
}

var _ tracer.LimitsService = (*fakeTracerLimits)(nil)

func (f *fakeTracerLimits) Create(_ context.Context, input *tracer.CreateLimitInput) (*tracer.Limit, error) {
	if err := f.cfg.injectedError("tracer.Limits.Create"); err != nil {
		return nil, err
	}

	now := time.Now()
	l := tracer.Limit{ID: generateID("tlimit"), Name: input.Name, Status: "DRAFT", CreatedAt: now, UpdatedAt: now}
	f.store.Set(l.ID, l)

	return &l, nil
}

func (f *fakeTracerLimits) Get(_ context.Context, id string) (*tracer.Limit, error) {
	l, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Limits.Get", "Limit", id)
	}

	return &l, nil
}

func (f *fakeTracerLimits) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[tracer.Limit] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeTracerLimits) Update(_ context.Context, id string, input *tracer.UpdateLimitInput) (*tracer.Limit, error) {
	l, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Limits.Update", "Limit", id)
	}

	if input.Name != nil {
		l.Name = *input.Name
	}

	l.UpdatedAt = time.Now()
	f.store.Set(id, l)

	return &l, nil
}

func (f *fakeTracerLimits) Delete(_ context.Context, id string) error {
	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Limits.Delete", "Limit", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeTracerLimits) Activate(_ context.Context, id string) (*tracer.Limit, error) {
	l, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Limits.Activate", "Limit", id)
	}

	l.Status = "ACTIVE"
	l.UpdatedAt = time.Now()
	f.store.Set(id, l)

	return &l, nil
}

func (f *fakeTracerLimits) Deactivate(_ context.Context, id string) (*tracer.Limit, error) {
	l, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Limits.Deactivate", "Limit", id)
	}

	l.Status = "INACTIVE"
	l.UpdatedAt = time.Now()
	f.store.Set(id, l)

	return &l, nil
}

func (f *fakeTracerLimits) Draft(_ context.Context, id string) (*tracer.Limit, error) {
	l, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Limits.Draft", "Limit", id)
	}

	l.Status = "DRAFT"
	l.UpdatedAt = time.Now()
	f.store.Set(id, l)

	return &l, nil
}

// ---------------------------------------------------------------------------
// Validations
// ---------------------------------------------------------------------------

type fakeTracerValidations struct {
	store *fakeStore[tracer.Validation]
	cfg   *fakeConfig
}

var _ tracer.ValidationsService = (*fakeTracerValidations)(nil)

func (f *fakeTracerValidations) Create(_ context.Context, _ *tracer.CreateValidationInput) (*tracer.Validation, error) {
	now := time.Now()
	v := tracer.Validation{ID: generateID("tval"), Status: "PASSED", CreatedAt: now}
	f.store.Set(v.ID, v)

	return &v, nil
}

func (f *fakeTracerValidations) Get(_ context.Context, id string) (*tracer.Validation, error) {
	v, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Validations.Get", "Validation", id)
	}

	return &v, nil
}

func (f *fakeTracerValidations) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[tracer.Validation] {
	return f.store.PaginatedIterator(opts)
}

// ---------------------------------------------------------------------------
// AuditEvents
// ---------------------------------------------------------------------------

type fakeTracerAuditEvents struct {
	store *fakeStore[tracer.AuditEvent]
	cfg   *fakeConfig
}

var _ tracer.AuditEventsService = (*fakeTracerAuditEvents)(nil)

func (f *fakeTracerAuditEvents) Get(_ context.Context, id string) (*tracer.AuditEvent, error) {
	ev, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("AuditEvents.Get", "AuditEvent", id)
	}

	return &ev, nil
}

func (f *fakeTracerAuditEvents) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[tracer.AuditEvent] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeTracerAuditEvents) Verify(_ context.Context, id string) (*tracer.AuditVerification, error) {
	if _, ok := f.store.Get(id); !ok {
		return nil, sdkerrors.NewNotFound("AuditEvents.Verify", "AuditEvent", id)
	}

	return &tracer.AuditVerification{EventID: id, Valid: true}, nil
}
