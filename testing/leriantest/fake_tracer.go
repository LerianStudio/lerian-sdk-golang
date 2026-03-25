package leriantest

import (
	"context"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
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
	return fakeGetStored(f.cfg, "", "Rules.Get", "Rule", id, f.store)
}

func (f *fakeTracerRules) List(_ context.Context, opts *models.CursorListOptions) *pagination.Iterator[tracer.Rule] {
	return fakeListStored(f.cfg, "", f.store, opts)
}

func (f *fakeTracerRules) Update(_ context.Context, id string, input *tracer.UpdateRuleInput) (*tracer.Rule, error) {
	return fakeMutateStored(f.cfg, "", "Rules.Update", "Rule", id, f.store, func(r *tracer.Rule) {
		if input.Name != nil {
			r.Name = *input.Name
		}
		r.UpdatedAt = time.Now()
	})
}

func (f *fakeTracerRules) Delete(_ context.Context, id string) error {
	return fakeDeleteStored(f.cfg, "", "Rules.Delete", "Rule", id, f.store)
}

func (f *fakeTracerRules) Activate(_ context.Context, id string) (*tracer.Rule, error) {
	return fakeMutateStored(f.cfg, "", "Rules.Activate", "Rule", id, f.store, func(r *tracer.Rule) {
		r.Status = "ACTIVE"
		r.UpdatedAt = time.Now()
	})
}

func (f *fakeTracerRules) Deactivate(_ context.Context, id string) (*tracer.Rule, error) {
	return fakeMutateStored(f.cfg, "", "Rules.Deactivate", "Rule", id, f.store, func(r *tracer.Rule) {
		r.Status = "INACTIVE"
		r.UpdatedAt = time.Now()
	})
}

func (f *fakeTracerRules) Draft(_ context.Context, id string) (*tracer.Rule, error) {
	return fakeMutateStored(f.cfg, "", "Rules.Draft", "Rule", id, f.store, func(r *tracer.Rule) {
		r.Status = "DRAFT"
		r.UpdatedAt = time.Now()
	})
}

// ---------------------------------------------------------------------------
// Limits
// ---------------------------------------------------------------------------

type fakeTracerLimits struct {
	store *fakeStore[tracer.Limit]
	cfg   *fakeConfig
}

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
	return fakeGetStored(f.cfg, "", "Limits.Get", "Limit", id, f.store)
}

func (f *fakeTracerLimits) List(_ context.Context, opts *models.CursorListOptions) *pagination.Iterator[tracer.Limit] {
	return fakeListStored(f.cfg, "", f.store, opts)
}

func (f *fakeTracerLimits) Update(_ context.Context, id string, input *tracer.UpdateLimitInput) (*tracer.Limit, error) {
	return fakeMutateStored(f.cfg, "", "Limits.Update", "Limit", id, f.store, func(l *tracer.Limit) {
		if input.Name != nil {
			l.Name = *input.Name
		}
		l.UpdatedAt = time.Now()
	})
}

func (f *fakeTracerLimits) Delete(_ context.Context, id string) error {
	return fakeDeleteStored(f.cfg, "", "Limits.Delete", "Limit", id, f.store)
}

func (f *fakeTracerLimits) Activate(_ context.Context, id string) (*tracer.Limit, error) {
	return fakeMutateStored(f.cfg, "", "Limits.Activate", "Limit", id, f.store, func(l *tracer.Limit) {
		l.Status = "ACTIVE"
		l.UpdatedAt = time.Now()
	})
}

func (f *fakeTracerLimits) Deactivate(_ context.Context, id string) (*tracer.Limit, error) {
	return fakeMutateStored(f.cfg, "", "Limits.Deactivate", "Limit", id, f.store, func(l *tracer.Limit) {
		l.Status = "INACTIVE"
		l.UpdatedAt = time.Now()
	})
}

func (f *fakeTracerLimits) Draft(_ context.Context, id string) (*tracer.Limit, error) {
	return fakeMutateStored(f.cfg, "", "Limits.Draft", "Limit", id, f.store, func(l *tracer.Limit) {
		l.Status = "DRAFT"
		l.UpdatedAt = time.Now()
	})
}

// ---------------------------------------------------------------------------
// Validations
// ---------------------------------------------------------------------------

type fakeTracerValidations struct {
	store *fakeStore[tracer.Validation]
	cfg   *fakeConfig
}

func (f *fakeTracerValidations) Create(_ context.Context, _ *tracer.CreateValidationInput) (*tracer.Validation, error) {
	now := time.Now()
	v := tracer.Validation{ID: generateID("tval"), Status: "PASSED", CreatedAt: now}
	f.store.Set(v.ID, v)

	return &v, nil
}

func (f *fakeTracerValidations) Get(_ context.Context, id string) (*tracer.Validation, error) {
	return fakeGetStored(f.cfg, "", "Validations.Get", "Validation", id, f.store)
}

func (f *fakeTracerValidations) List(_ context.Context, opts *models.CursorListOptions) *pagination.Iterator[tracer.Validation] {
	return fakeListStored(f.cfg, "", f.store, opts)
}

// ---------------------------------------------------------------------------
// AuditEvents
// ---------------------------------------------------------------------------

type fakeTracerAuditEvents struct {
	store *fakeStore[tracer.AuditEvent]
	cfg   *fakeConfig
}

func (f *fakeTracerAuditEvents) Get(_ context.Context, id string) (*tracer.AuditEvent, error) {
	return fakeGetStored(f.cfg, "", "AuditEvents.Get", "AuditEvent", id, f.store)
}

func (f *fakeTracerAuditEvents) List(_ context.Context, opts *models.CursorListOptions) *pagination.Iterator[tracer.AuditEvent] {
	return fakeListStored(f.cfg, "", f.store, opts)
}

func (f *fakeTracerAuditEvents) Verify(_ context.Context, id string) (*tracer.AuditVerification, error) {
	if _, err := fakeActionStored(f.cfg, "", "AuditEvents.Verify", "AuditEvent", id, f.store); err != nil {
		return nil, err
	}
	return &tracer.AuditVerification{EventID: id, Valid: true}, nil
}
