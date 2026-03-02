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

// sampleValidation returns a test Validation with sensible defaults.
func sampleValidation(id, result string) Validation {
	return Validation{
		ID:           id,
		Status:       "completed",
		Result:       result,
		RulesApplied: []string{"rule-1", "rule-2"},
		Transaction:  map[string]any{"amount": 5000, "currency": "USD"},
		CreatedAt:    time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestValidationsCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/validations", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateValidationInput)
			require.True(t, ok)
			assert.NotNil(t, input.Transaction)
			assert.Equal(t, float64(5000), input.Transaction["amount"])

			return unmarshalInto(sampleValidation("val-1", "approved"), result)
		},
	}

	svc := newValidationsService(mock)
	val, err := svc.Create(context.Background(), &CreateValidationInput{
		Transaction: map[string]any{"amount": float64(5000), "currency": "USD"},
		RuleIDs:     []string{"rule-1"},
	})

	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, "val-1", val.ID)
	assert.Equal(t, "approved", val.Result)
}

func TestValidationsCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newValidationsService(&mockBackend{})
	val, err := svc.Create(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, val)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestValidationsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: validation rejected")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newValidationsService(mock)
	val, err := svc.Create(context.Background(), &CreateValidationInput{
		Transaction: map[string]any{"amount": 100},
	})

	require.Error(t, err)
	assert.Nil(t, val)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestValidationsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/validations/val-42", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleValidation("val-42", "rejected"), result)
		},
	}

	svc := newValidationsService(mock)
	val, err := svc.Get(context.Background(), "val-42")

	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, "val-42", val.ID)
	assert.Equal(t, "rejected", val.Result)
}

func TestValidationsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newValidationsService(&mockBackend{})
	val, err := svc.Get(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, val)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestValidationsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newValidationsService(mock)
	val, err := svc.Get(context.Background(), "val-missing")

	require.Error(t, err)
	assert.Nil(t, val)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestValidationsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/validations")
			assert.Nil(t, body)

			resp := models.ListResponse[Validation]{
				Items: []Validation{
					sampleValidation("val-1", "approved"),
					sampleValidation("val-2", "rejected"),
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newValidationsService(mock)
	iter := svc.List(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "val-1", items[0].ID)
	assert.Equal(t, "val-2", items[1].ID)
}

func TestValidationsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, _ string, path string, _ any, result any) error {
			receivedPath = path

			resp := models.ListResponse[Validation]{
				Items:      []Validation{sampleValidation("val-1", "approved")},
				Pagination: models.Pagination{Total: 1, Limit: 50},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newValidationsService(mock)
	opts := &models.ListOptions{
		Limit:     50,
		SortBy:    "createdAt",
		SortOrder: "desc",
	}

	iter := svc.List(context.Background(), opts)

	ctx := context.Background()
	require.True(t, iter.Next(ctx))

	assert.Contains(t, receivedPath, "limit=50")
	assert.Contains(t, receivedPath, "sortBy=createdAt")
	assert.Contains(t, receivedPath, "sortOrder=desc")
}

func TestValidationsListEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			resp := models.ListResponse[Validation]{
				Items:      []Validation{},
				Pagination: models.Pagination{Total: 0, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newValidationsService(mock)
	iter := svc.List(context.Background(), nil)

	ctx := context.Background()
	assert.False(t, iter.Next(ctx))
	require.NoError(t, iter.Err())
}

func TestValidationsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: internal server error")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newValidationsService(mock)
	iter := svc.List(context.Background(), nil)

	ctx := context.Background()
	assert.False(t, iter.Next(ctx))
	assert.Equal(t, expectedErr, iter.Err())
}
