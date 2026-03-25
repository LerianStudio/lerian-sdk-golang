package tracer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock Backend (shared across tracer service tests)
// ---------------------------------------------------------------------------

type mockBackend struct {
	callFn         func(ctx context.Context, method, path string, body, result any) error
	callWithHdrsFn func(ctx context.Context, method, path string, headers map[string]string, body, result any) error
	callRawFn      func(ctx context.Context, method, path string, body any) ([]byte, error)
}

func (m *mockBackend) Do(ctx context.Context, req core.Request) (*core.Response, error) {
	if len(req.Headers) > 0 && m.callWithHdrsFn != nil {
		var result any

		resultArg := any(&result)
		if req.ExpectNoResponse {
			resultArg = nil
		}

		if err := m.callWithHdrsFn(ctx, req.Method, req.Path, req.Headers, reqBody(req), resultArg); err != nil {
			return nil, err
		}

		if req.ExpectNoResponse {
			return &core.Response{}, nil
		}

		return jsonResponse(result)
	}

	if m.callFn != nil {
		var result any

		resultArg := any(&result)
		if req.ExpectNoResponse {
			resultArg = nil
		}

		if err := m.callFn(ctx, req.Method, req.Path, reqBody(req), resultArg); err != nil {
			return nil, err
		}

		if req.ExpectNoResponse {
			return &core.Response{}, nil
		}

		return jsonResponse(result)
	}

	if m.callRawFn != nil {
		body, err := m.callRawFn(ctx, req.Method, req.Path, reqBody(req))
		if err != nil {
			return nil, err
		}

		return &core.Response{Body: body}, nil
	}

	return nil, fmt.Errorf("mockBackend.Do not configured")
}

var _ core.Backend = (*mockBackend)(nil)

// unmarshalInto simulates backend JSON round-trip.
func unmarshalInto(src any, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dst)
}

func jsonResponse(result any) (*core.Response, error) {
	if result == nil {
		return &core.Response{}, nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &core.Response{Body: data}, nil
}

func reqBody(req core.Request) any {
	if len(req.BodyBytes) > 0 {
		return req.BodyBytes
	}

	return req.Body
}

// sampleRule returns a test Rule with sensible defaults.
func sampleRule(id string, status RuleStatus) Rule {
	return Rule{
		ID:     id,
		Name:   "Test Rule",
		Status: status,
		Conditions: []RuleCondition{
			{Field: "amount", Operator: "gt", Value: 1000},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestRulesCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/rules", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateRuleInput)
			require.True(t, ok)
			assert.Equal(t, "High Value Rule", input.Name)

			return unmarshalInto(sampleRule("rule-1", RuleStatusDraft), result)
		},
	}

	svc := newRulesService(mock)
	rule, err := svc.Create(context.Background(), &CreateRuleInput{
		Name:     "High Value Rule",
		Priority: 1,
		Conditions: []RuleCondition{
			{Field: "amount", Operator: "gt", Value: 1000},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "rule-1", rule.ID)
	assert.Equal(t, RuleStatusDraft, rule.Status)
}

func TestRulesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newRulesService(&mockBackend{})
	rule, err := svc.Create(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestRulesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: create failed")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newRulesService(mock)
	rule, err := svc.Create(context.Background(), &CreateRuleInput{Name: "x"})

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestRulesGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/rules/rule-42", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleRule("rule-42", RuleStatusActive), result)
		},
	}

	svc := newRulesService(mock)
	rule, err := svc.Get(context.Background(), "rule-42")

	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "rule-42", rule.ID)
	assert.Equal(t, RuleStatusActive, rule.Status)
}

func TestRulesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newRulesService(&mockBackend{})
	rule, err := svc.Get(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestRulesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newRulesService(mock)
	rule, err := svc.Get(context.Background(), "rule-missing")

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestRulesList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/rules")
			assert.Nil(t, body)

			resp := models.ListResponse[Rule]{
				Items: []Rule{
					sampleRule("rule-1", RuleStatusActive),
					sampleRule("rule-2", RuleStatusDraft),
				},
				Pagination: models.Pagination{
					Total: 2,
					Limit: 10,
				},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newRulesService(mock)
	iter := svc.List(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "rule-1", items[0].ID)
	assert.Equal(t, "rule-2", items[1].ID)
}

func TestRulesListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			receivedPath = path

			resp := models.ListResponse[Rule]{
				Items:      []Rule{sampleRule("rule-1", RuleStatusActive)},
				Pagination: models.Pagination{Total: 1, Limit: 25},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newRulesService(mock)
	opts := &models.CursorListOptions{
		Limit:     25,
		Cursor:    "cursor-2",
		SortBy:    "name",
		SortOrder: "asc",
	}

	iter := svc.List(context.Background(), opts)

	ctx := context.Background()
	require.True(t, iter.Next(ctx))

	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "cursor=cursor-2")
	assert.Contains(t, receivedPath, "sortBy=name")
	assert.Contains(t, receivedPath, "sortOrder=asc")
}

func TestRulesListEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			resp := models.ListResponse[Rule]{
				Items:      []Rule{},
				Pagination: models.Pagination{Total: 0, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newRulesService(mock)
	iter := svc.List(context.Background(), nil)

	ctx := context.Background()
	assert.False(t, iter.Next(ctx))
	require.NoError(t, iter.Err())
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestRulesUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/rules/rule-1", path)
			assert.NotNil(t, body)

			input, ok := body.(*UpdateRuleInput)
			require.True(t, ok)
			require.NotNil(t, input.Name)
			assert.Equal(t, "Updated Rule", *input.Name)

			return unmarshalInto(sampleRule("rule-1", RuleStatusDraft), result)
		},
	}

	svc := newRulesService(mock)
	name := "Updated Rule"
	rule, err := svc.Update(context.Background(), "rule-1", &UpdateRuleInput{Name: &name})

	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "rule-1", rule.ID)
}

func TestRulesUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newRulesService(&mockBackend{})
	rule, err := svc.Update(context.Background(), "", &UpdateRuleInput{})

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestRulesUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newRulesService(&mockBackend{})
	rule, err := svc.Update(context.Background(), "rule-1", nil)

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestRulesDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/rules/rule-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newRulesService(mock)
	err := svc.Delete(context.Background(), "rule-1")

	require.NoError(t, err)
}

func TestRulesDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newRulesService(&mockBackend{})
	err := svc.Delete(context.Background(), "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestRulesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newRulesService(mock)
	err := svc.Delete(context.Background(), "rule-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Activate
// ---------------------------------------------------------------------------

func TestRulesActivate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/rules/rule-1/activate", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleRule("rule-1", RuleStatusActive), result)
		},
	}

	svc := newRulesService(mock)
	rule, err := svc.Activate(context.Background(), "rule-1")

	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "rule-1", rule.ID)
	assert.Equal(t, RuleStatusActive, rule.Status)
}

func TestRulesActivateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newRulesService(&mockBackend{})
	rule, err := svc.Activate(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestRulesActivateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newRulesService(mock)
	rule, err := svc.Activate(context.Background(), "rule-1")

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Deactivate
// ---------------------------------------------------------------------------

func TestRulesDeactivate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/rules/rule-1/deactivate", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleRule("rule-1", RuleStatusInactive), result)
		},
	}

	svc := newRulesService(mock)
	rule, err := svc.Deactivate(context.Background(), "rule-1")

	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "rule-1", rule.ID)
	assert.Equal(t, RuleStatusInactive, rule.Status)
}

func TestRulesDeactivateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newRulesService(&mockBackend{})
	rule, err := svc.Deactivate(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Draft
// ---------------------------------------------------------------------------

func TestRulesDraft(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/rules/rule-1/draft", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleRule("rule-1", RuleStatusDraft), result)
		},
	}

	svc := newRulesService(mock)
	rule, err := svc.Draft(context.Background(), "rule-1")

	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "rule-1", rule.ID)
	assert.Equal(t, RuleStatusDraft, rule.Status)
}

func TestRulesDraftEmptyID(t *testing.T) {
	t.Parallel()

	svc := newRulesService(&mockBackend{})
	rule, err := svc.Draft(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Lifecycle: Create -> Activate -> Deactivate -> Draft
// ---------------------------------------------------------------------------

func TestRulesLifecycle(t *testing.T) {
	t.Parallel()

	currentStatus := RuleStatusDraft

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			switch {
			case method == "POST" && path == "/rules":
				currentStatus = RuleStatusDraft
			case path == "/rules/rule-lc/activate":
				currentStatus = RuleStatusActive
			case path == "/rules/rule-lc/deactivate":
				currentStatus = RuleStatusInactive
			case path == "/rules/rule-lc/draft":
				currentStatus = RuleStatusDraft
			}

			return unmarshalInto(sampleRule("rule-lc", currentStatus), result)
		},
	}

	svc := newRulesService(mock)
	ctx := context.Background()

	// Create (DRAFT)
	rule, err := svc.Create(ctx, &CreateRuleInput{
		Name:       "Lifecycle Rule",
		Conditions: []RuleCondition{{Field: "amount", Operator: "gt", Value: 500}},
	})
	require.NoError(t, err)
	assert.Equal(t, RuleStatusDraft, rule.Status)

	// Activate (DRAFT -> ACTIVE)
	rule, err = svc.Activate(ctx, "rule-lc")
	require.NoError(t, err)
	assert.Equal(t, RuleStatusActive, rule.Status)

	// Deactivate (ACTIVE -> INACTIVE)
	rule, err = svc.Deactivate(ctx, "rule-lc")
	require.NoError(t, err)
	assert.Equal(t, RuleStatusInactive, rule.Status)

	// Draft (INACTIVE -> DRAFT)
	rule, err = svc.Draft(ctx, "rule-lc")
	require.NoError(t, err)
	assert.Equal(t, RuleStatusDraft, rule.Status)
}
