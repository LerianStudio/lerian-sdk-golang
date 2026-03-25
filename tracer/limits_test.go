package tracer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleLimit returns a test Limit with sensible defaults.
func sampleLimit(id string, status RuleStatus) Limit {
	return Limit{
		ID:        id,
		Name:      "Test Limit",
		Status:    status,
		Type:      "amount",
		MaxAmount: 100000,
		Period:    "daily",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestLimitsCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/limits", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateLimitInput)
			require.True(t, ok)
			assert.Equal(t, "Daily Cap", input.Name)
			assert.Equal(t, int64(50000), input.MaxAmount)

			return unmarshalInto(sampleLimit("lim-1", RuleStatusDraft), result)
		},
	}

	svc := newLimitsService(mock)
	limit, err := svc.Create(context.Background(), &CreateLimitInput{
		Name:      "Daily Cap",
		Type:      "amount",
		MaxAmount: 50000,
		Period:    "daily",
	})

	require.NoError(t, err)
	require.NotNil(t, limit)
	assert.Equal(t, "lim-1", limit.ID)
	assert.Equal(t, RuleStatusDraft, limit.Status)
}

func TestLimitsCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newLimitsService(&mockBackend{})
	limit, err := svc.Create(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLimitsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: create failed")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLimitsService(mock)
	limit, err := svc.Create(context.Background(), &CreateLimitInput{Name: "x"})

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestLimitsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/limits/lim-42", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleLimit("lim-42", RuleStatusActive), result)
		},
	}

	svc := newLimitsService(mock)
	limit, err := svc.Get(context.Background(), "lim-42")

	require.NoError(t, err)
	require.NotNil(t, limit)
	assert.Equal(t, "lim-42", limit.ID)
	assert.Equal(t, RuleStatusActive, limit.Status)
}

func TestLimitsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newLimitsService(&mockBackend{})
	limit, err := svc.Get(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLimitsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLimitsService(mock)
	limit, err := svc.Get(context.Background(), "lim-missing")

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestLimitsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/limits")
			assert.Nil(t, body)

			resp := models.ListResponse[Limit]{
				Items: []Limit{
					sampleLimit("lim-1", RuleStatusActive),
					sampleLimit("lim-2", RuleStatusDraft),
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newLimitsService(mock)
	iter := svc.List(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "lim-1", items[0].ID)
	assert.Equal(t, "lim-2", items[1].ID)
}

func TestLimitsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, _ string, path string, _ any, result any) error {
			receivedPath = path

			resp := models.ListResponse[Limit]{
				Items:      []Limit{sampleLimit("lim-1", RuleStatusActive)},
				Pagination: models.Pagination{Total: 1, Limit: 25},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newLimitsService(mock)
	opts := &models.CursorListOptions{
		Limit:     25,
		SortBy:    "name",
		SortOrder: "desc",
	}

	iter := svc.List(context.Background(), opts)

	ctx := context.Background()
	require.True(t, iter.Next(ctx))

	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortBy=name")
	assert.Contains(t, receivedPath, "sortOrder=desc")
}

func TestLimitsListEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			resp := models.ListResponse[Limit]{
				Items:      []Limit{},
				Pagination: models.Pagination{Total: 0, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newLimitsService(mock)
	iter := svc.List(context.Background(), nil)

	ctx := context.Background()
	assert.False(t, iter.Next(ctx))
	require.NoError(t, iter.Err())
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestLimitsUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/limits/lim-1", path)
			assert.NotNil(t, body)

			input, ok := body.(*UpdateLimitInput)
			require.True(t, ok)
			require.NotNil(t, input.Name)
			assert.Equal(t, "Updated Limit", *input.Name)

			return unmarshalInto(sampleLimit("lim-1", RuleStatusDraft), result)
		},
	}

	svc := newLimitsService(mock)
	name := "Updated Limit"
	limit, err := svc.Update(context.Background(), "lim-1", &UpdateLimitInput{Name: &name})

	require.NoError(t, err)
	require.NotNil(t, limit)
	assert.Equal(t, "lim-1", limit.ID)
}

func TestLimitsUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newLimitsService(&mockBackend{})
	limit, err := svc.Update(context.Background(), "", &UpdateLimitInput{})

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLimitsUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newLimitsService(&mockBackend{})
	limit, err := svc.Update(context.Background(), "lim-1", nil)

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestLimitsDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/limits/lim-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newLimitsService(mock)
	err := svc.Delete(context.Background(), "lim-1")

	require.NoError(t, err)
}

func TestLimitsDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newLimitsService(&mockBackend{})
	err := svc.Delete(context.Background(), "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLimitsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLimitsService(mock)
	err := svc.Delete(context.Background(), "lim-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Activate
// ---------------------------------------------------------------------------

func TestLimitsActivate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/limits/lim-1/activate", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleLimit("lim-1", RuleStatusActive), result)
		},
	}

	svc := newLimitsService(mock)
	limit, err := svc.Activate(context.Background(), "lim-1")

	require.NoError(t, err)
	require.NotNil(t, limit)
	assert.Equal(t, "lim-1", limit.ID)
	assert.Equal(t, RuleStatusActive, limit.Status)
}

func TestLimitsActivateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newLimitsService(&mockBackend{})
	limit, err := svc.Activate(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLimitsActivateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLimitsService(mock)
	limit, err := svc.Activate(context.Background(), "lim-1")

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Deactivate
// ---------------------------------------------------------------------------

func TestLimitsDeactivate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/limits/lim-1/deactivate", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleLimit("lim-1", RuleStatusInactive), result)
		},
	}

	svc := newLimitsService(mock)
	limit, err := svc.Deactivate(context.Background(), "lim-1")

	require.NoError(t, err)
	require.NotNil(t, limit)
	assert.Equal(t, "lim-1", limit.ID)
	assert.Equal(t, RuleStatusInactive, limit.Status)
}

func TestLimitsDeactivateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newLimitsService(&mockBackend{})
	limit, err := svc.Deactivate(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Draft
// ---------------------------------------------------------------------------

func TestLimitsDraft(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/limits/lim-1/draft", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleLimit("lim-1", RuleStatusDraft), result)
		},
	}

	svc := newLimitsService(mock)
	limit, err := svc.Draft(context.Background(), "lim-1")

	require.NoError(t, err)
	require.NotNil(t, limit)
	assert.Equal(t, "lim-1", limit.ID)
	assert.Equal(t, RuleStatusDraft, limit.Status)
}

func TestLimitsDraftEmptyID(t *testing.T) {
	t.Parallel()

	svc := newLimitsService(&mockBackend{})
	limit, err := svc.Draft(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, limit)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Lifecycle: Create -> Activate -> Deactivate -> Draft
// ---------------------------------------------------------------------------

func TestLimitsLifecycle(t *testing.T) {
	t.Parallel()

	currentStatus := RuleStatusDraft

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, _ any, result any) error {
			switch {
			case method == "POST" && path == "/limits":
				currentStatus = RuleStatusDraft
			case path == "/limits/lim-lc/activate":
				currentStatus = RuleStatusActive
			case path == "/limits/lim-lc/deactivate":
				currentStatus = RuleStatusInactive
			case path == "/limits/lim-lc/draft":
				currentStatus = RuleStatusDraft
			}

			return unmarshalInto(sampleLimit("lim-lc", currentStatus), result)
		},
	}

	svc := newLimitsService(mock)
	ctx := context.Background()

	// Create (DRAFT)
	limit, err := svc.Create(ctx, &CreateLimitInput{
		Name:      "Lifecycle Limit",
		Type:      "amount",
		MaxAmount: 10000,
		Period:    "daily",
	})
	require.NoError(t, err)
	assert.Equal(t, RuleStatusDraft, limit.Status)

	// Activate (DRAFT -> ACTIVE)
	limit, err = svc.Activate(ctx, "lim-lc")
	require.NoError(t, err)
	assert.Equal(t, RuleStatusActive, limit.Status)

	// Deactivate (ACTIVE -> INACTIVE)
	limit, err = svc.Deactivate(ctx, "lim-lc")
	require.NoError(t, err)
	assert.Equal(t, RuleStatusInactive, limit.Status)

	// Draft (INACTIVE -> DRAFT)
	limit, err = svc.Draft(ctx, "lim-lc")
	require.NoError(t, err)
	assert.Equal(t, RuleStatusDraft, limit.Status)
}
