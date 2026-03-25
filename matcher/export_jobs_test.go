package matcher

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// exportJobsServiceAPI.Create
// ---------------------------------------------------------------------------

func TestExportJobsCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/export-jobs", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateExportJobInput)
			require.True(t, ok)
			assert.Equal(t, "ctx-1", input.ContextID)
			assert.Equal(t, "csv", input.Format)

			return unmarshalInto(ExportJob{ID: "exp-1", ContextID: "ctx-1", Status: "pending", Format: "csv"}, result)
		}}

		svc := newExportJobsService(mb)
		got, err := svc.Create(context.Background(), &CreateExportJobInput{
			ContextID: "ctx-1",
			Format:    "csv",
		})

		require.NoError(t, err)
		assert.Equal(t, "exp-1", got.ID)
		assert.Equal(t, "pending", got.Status)
		assert.Equal(t, "csv", got.Format)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newExportJobsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// exportJobsServiceAPI.Get
// ---------------------------------------------------------------------------

func TestExportJobsGet(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/export-jobs/exp-1", path)

			return unmarshalInto(ExportJob{ID: "exp-1", Status: "completed", Format: "xlsx"}, result)
		}}

		svc := newExportJobsService(mb)
		got, err := svc.Get(context.Background(), "exp-1")
		require.NoError(t, err)
		assert.Equal(t, "exp-1", got.ID)
		assert.Equal(t, "completed", got.Status)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newExportJobsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// exportJobsServiceAPI.List
// ---------------------------------------------------------------------------

func TestExportJobsList(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/export-jobs")

		resp := models.ListResponse[ExportJob]{
			Items: []ExportJob{
				{ID: "exp-1", Status: "completed"},
				{ID: "exp-2", Status: "pending"},
			},
			Pagination: models.Pagination{Total: 2, Limit: 10},
		}

		return unmarshalInto(resp, result)
	}}

	svc := newExportJobsService(mb)
	iter := svc.List(context.Background(), nil)
	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "exp-1", items[0].ID)
	assert.Equal(t, "exp-2", items[1].ID)
}

func TestExportJobsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mb := &mockBackend{callFn: func(_ context.Context, _, path string, _, result any) error {
		receivedPath = path

		resp := models.ListResponse[ExportJob]{
			Items:      []ExportJob{{ID: "exp-1"}},
			Pagination: models.Pagination{Total: 1, Limit: 50},
		}

		return unmarshalInto(resp, result)
	}}

	svc := newExportJobsService(mb)
	opts := &models.CursorListOptions{Limit: 50, SortBy: "createdAt"}
	iter := svc.List(context.Background(), opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=50")
	assert.Contains(t, receivedPath, "sortBy=createdAt")
}

// ---------------------------------------------------------------------------
// exportJobsServiceAPI.Cancel
// ---------------------------------------------------------------------------

func TestExportJobsCancel(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/export-jobs/exp-1/cancel", path)
			assert.Nil(t, body)

			return unmarshalInto(ExportJob{ID: "exp-1", Status: "cancelled"}, result)
		}}

		svc := newExportJobsService(mb)
		got, err := svc.Cancel(context.Background(), "exp-1")
		require.NoError(t, err)
		assert.Equal(t, "cancelled", got.Status)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newExportJobsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Cancel(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// exportJobsServiceAPI.Download — returns raw bytes
// ---------------------------------------------------------------------------

func TestExportJobsDownload(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedContent := []byte("id,amount,status\n1,1000,matched\n2,2500,unmatched\n")

		mb := &mockBackend{
			callFn: func(context.Context, string, string, any, any) error { return nil },
			callRawFn: func(_ context.Context, method, path string, body any) ([]byte, error) {
				assert.Equal(t, "GET", method)
				assert.Equal(t, "/export-jobs/exp-1/download", path)
				assert.Nil(t, body)

				return expectedContent, nil
			},
		}

		svc := newExportJobsService(mb)
		data, err := svc.Download(context.Background(), "exp-1")
		require.NoError(t, err)
		assert.Equal(t, expectedContent, data)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newExportJobsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		data, err := svc.Download(context.Background(), "")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("backend error", func(t *testing.T) {
		t.Parallel()

		expectedErr := fmt.Errorf("export not ready")
		mb := &mockBackend{
			callFn: func(context.Context, string, string, any, any) error { return nil },
			callRawFn: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
				return nil, expectedErr
			},
		}

		svc := newExportJobsService(mb)
		data, err := svc.Download(context.Background(), "exp-1")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("nil backend uses core error", func(t *testing.T) {
		t.Parallel()

		svc := newExportJobsService(nil)
		data, err := svc.Download(context.Background(), "exp-1")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.ErrorIs(t, err, core.ErrNilBackend)
	})

	t.Run("typed nil backend uses core error", func(t *testing.T) {
		t.Parallel()

		var mb *mockBackend

		svc := newExportJobsService(mb)
		data, err := svc.Download(context.Background(), "exp-1")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.ErrorIs(t, err, core.ErrNilBackend)
	})
}

// ---------------------------------------------------------------------------
// ExportJobs — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestExportJobsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExportJobsService(mb)
	got, err := svc.Create(context.Background(), &CreateExportJobInput{
		ContextID: "ctx-1", Format: "csv",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExportJobsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExportJobsService(mb)
	got, err := svc.Get(context.Background(), "exp-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExportJobsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExportJobsService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestExportJobsCancelBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExportJobsService(mb)
	got, err := svc.Cancel(context.Background(), "exp-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}
