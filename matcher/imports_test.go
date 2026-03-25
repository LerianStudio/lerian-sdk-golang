package matcher

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

// ---------------------------------------------------------------------------
// Shared fixture
// ---------------------------------------------------------------------------

var testImport = Import{
	ID:        "imp-1",
	ContextID: "ctx-1",
	Status:    "processing",
	FileName:  "records-2026-01.csv",
	CreatedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2026, 1, 15, 10, 5, 0, 0, time.UTC),
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestImportsCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/imports", path)
			assert.NotNil(t, body)

			return unmarshalInto(testImport, result)
		}}

		svc := newImportsService(mb)
		got, err := svc.Create(context.Background(), &CreateImportInput{
			ContextID: "ctx-1",
			FileName:  "records-2026-01.csv",
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "imp-1", got.ID)
		assert.Equal(t, "ctx-1", got.ContextID)
		assert.Equal(t, "records-2026-01.csv", got.FileName)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newImportsService(&mockBackend{callFn: noopCallFn})
		got, err := svc.Create(context.Background(), nil)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("input forwarded", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, _, _ string, body, result any) error {
			input, ok := body.(*CreateImportInput)
			require.True(t, ok, "body should be *CreateImportInput")
			assert.Equal(t, "ctx-2", input.ContextID)
			assert.Equal(t, "data.csv", input.FileName)

			return unmarshalInto(testImport, result)
		}}

		svc := newImportsService(mb)
		_, err := svc.Create(context.Background(), &CreateImportInput{
			ContextID: "ctx-2",
			FileName:  "data.csv",
		})
		require.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestImportsGet(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/imports/imp-1", path)
			assert.Nil(t, body)

			return unmarshalInto(testImport, result)
		}}

		svc := newImportsService(mb)
		got, err := svc.Get(context.Background(), "imp-1")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "imp-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newImportsService(&mockBackend{callFn: noopCallFn})
		got, err := svc.Get(context.Background(), "")

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("backend error", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return errors.New("not found")
		}}

		svc := newImportsService(mb)
		got, err := svc.Get(context.Background(), "imp-missing")

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "not found")
	})
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestImportsList(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/imports")

		resp := models.ListResponse[Import]{
			Items: []Import{
				{ID: "imp-1", FileName: "a.csv"},
				{ID: "imp-2", FileName: "b.csv"},
			},
			Pagination: models.Pagination{Total: 2, Limit: 10},
		}

		return unmarshalInto(resp, result)
	}}

	svc := newImportsService(mb)
	iter := svc.List(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "imp-1", items[0].ID)
	assert.Equal(t, "imp-2", items[1].ID)
}

func TestImportsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mb := &mockBackend{callFn: func(_ context.Context, _, path string, _, result any) error {
		receivedPath = path

		resp := models.ListResponse[Import]{
			Items:      []Import{{ID: "imp-1"}},
			Pagination: models.Pagination{Total: 1, Limit: 5},
		}

		return unmarshalInto(resp, result)
	}}

	svc := newImportsService(mb)
	opts := &models.CursorListOptions{Limit: 5}
	iter := svc.List(context.Background(), opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=5")
}

// ---------------------------------------------------------------------------
// Cancel
// ---------------------------------------------------------------------------

func TestImportsCancel(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/imports/imp-1/cancel", path)
			assert.Nil(t, body)

			cancelled := testImport
			cancelled.Status = "cancelled"

			return unmarshalInto(cancelled, result)
		}}

		svc := newImportsService(mb)
		got, err := svc.Cancel(context.Background(), "imp-1")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "imp-1", got.ID)
		assert.Equal(t, "cancelled", got.Status)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newImportsService(&mockBackend{callFn: noopCallFn})
		got, err := svc.Cancel(context.Background(), "")

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("backend error", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return errors.New("cannot cancel completed import")
		}}

		svc := newImportsService(mb)
		got, err := svc.Cancel(context.Background(), "imp-1")

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "cannot cancel")
	})
}

// ---------------------------------------------------------------------------
// GetStatus
// ---------------------------------------------------------------------------

func TestImportsGetStatus(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		status := ImportStatus{
			ID:          "imp-1",
			Status:      "processing",
			Progress:    75,
			RecordCount: 750,
			ErrorCount:  3,
			UpdatedAt:   time.Date(2026, 1, 15, 10, 10, 0, 0, time.UTC),
		}

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/imports/imp-1/status", path)
			assert.Nil(t, body)

			return unmarshalInto(status, result)
		}}

		svc := newImportsService(mb)
		got, err := svc.GetStatus(context.Background(), "imp-1")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "imp-1", got.ID)
		assert.Equal(t, "processing", got.Status)
		assert.Equal(t, 75, got.Progress)
		assert.Equal(t, 750, got.RecordCount)
		assert.Equal(t, 3, got.ErrorCount)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newImportsService(&mockBackend{callFn: noopCallFn})
		got, err := svc.GetStatus(context.Background(), "")

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// Imports — Additional Backend Error Propagation
// ---------------------------------------------------------------------------

func TestImportsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newImportsService(mb)
	got, err := svc.Create(context.Background(), &CreateImportInput{
		ContextID: "ctx-1", FileName: "data.csv",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestImportsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newImportsService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestImportsGetStatusBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newImportsService(mb)
	got, err := svc.GetStatus(context.Background(), "imp-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ importsServiceAPI = (*importsService)(nil)
