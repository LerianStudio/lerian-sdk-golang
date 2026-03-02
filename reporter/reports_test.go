package reporter

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ReportsService.Create tests
// ---------------------------------------------------------------------------

func TestReportsCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/reports", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateReportInput)
			require.True(t, ok)
			assert.Equal(t, "Monthly Sales", input.Name)
			assert.Equal(t, "pdf", input.Format)

			return unmarshalInto(Report{ID: "rpt-1", Name: "Monthly Sales", Format: "pdf"}, result)
		},
	}

	svc := newReportsService(mock)
	rpt, err := svc.Create(context.Background(), &CreateReportInput{
		Name:   "Monthly Sales",
		Format: "pdf",
	})

	require.NoError(t, err)
	require.NotNil(t, rpt)
	assert.Equal(t, "rpt-1", rpt.ID)
	assert.Equal(t, "Monthly Sales", rpt.Name)
	assert.Equal(t, "pdf", rpt.Format)
}

func TestReportsCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newReportsService(&mockBackend{})
	rpt, err := svc.Create(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, rpt)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestReportsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("backend create error")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newReportsService(mock)
	rpt, err := svc.Create(context.Background(), &CreateReportInput{Name: "test", Format: "csv"})

	require.Error(t, err)
	assert.Nil(t, rpt)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// ReportsService.Get tests
// ---------------------------------------------------------------------------

func TestReportsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/reports/rpt-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Report{ID: "rpt-1", Name: "Q4 Report", Status: "completed"}, result)
		},
	}

	svc := newReportsService(mock)
	rpt, err := svc.Get(context.Background(), "rpt-1")

	require.NoError(t, err)
	require.NotNil(t, rpt)
	assert.Equal(t, "rpt-1", rpt.ID)
	assert.Equal(t, "Q4 Report", rpt.Name)
	assert.Equal(t, "completed", rpt.Status)
}

func TestReportsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newReportsService(&mockBackend{})
	rpt, err := svc.Get(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, rpt)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestReportsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newReportsService(mock)
	rpt, err := svc.Get(context.Background(), "rpt-missing")

	require.Error(t, err)
	assert.Nil(t, rpt)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// ReportsService.List tests
// ---------------------------------------------------------------------------

func TestReportsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, _ string, _, result any) error {
			assert.Equal(t, "GET", method)

			resp := models.ListResponse[Report]{
				Items: []Report{
					{ID: "rpt-1", Name: "Report A"},
					{ID: "rpt-2", Name: "Report B"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newReportsService(mock)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "rpt-1", items[0].ID)
	assert.Equal(t, "rpt-2", items[1].ID)
}

func TestReportsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, _, path string, _, result any) error {
			receivedPath = path

			resp := models.ListResponse[Report]{
				Items:      []Report{{ID: "rpt-1"}},
				Pagination: models.Pagination{Total: 1, Limit: 25},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newReportsService(mock)
	opts := &models.ListOptions{Limit: 25, SortOrder: "desc"}
	iter := svc.List(context.Background(), opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortOrder=desc")
}

// ---------------------------------------------------------------------------
// ReportsService.Update tests
// ---------------------------------------------------------------------------

func TestReportsUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/reports/rpt-1", path)
			assert.NotNil(t, body)

			input, ok := body.(*UpdateReportInput)
			require.True(t, ok)
			require.NotNil(t, input.Name)
			assert.Equal(t, "Updated Report", *input.Name)

			return unmarshalInto(Report{ID: "rpt-1", Name: "Updated Report"}, result)
		},
	}

	svc := newReportsService(mock)
	name := "Updated Report"
	rpt, err := svc.Update(context.Background(), "rpt-1", &UpdateReportInput{Name: &name})

	require.NoError(t, err)
	require.NotNil(t, rpt)
	assert.Equal(t, "rpt-1", rpt.ID)
	assert.Equal(t, "Updated Report", rpt.Name)
}

func TestReportsUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newReportsService(&mockBackend{})
	rpt, err := svc.Update(context.Background(), "", &UpdateReportInput{})

	require.Error(t, err)
	assert.Nil(t, rpt)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestReportsUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newReportsService(&mockBackend{})
	rpt, err := svc.Update(context.Background(), "rpt-1", nil)

	require.Error(t, err)
	assert.Nil(t, rpt)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestReportsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newReportsService(mock)
	name := "Bad"
	rpt, err := svc.Update(context.Background(), "rpt-1", &UpdateReportInput{Name: &name})

	require.Error(t, err)
	assert.Nil(t, rpt)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// ReportsService.Delete tests
// ---------------------------------------------------------------------------

func TestReportsDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/reports/rpt-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newReportsService(mock)
	err := svc.Delete(context.Background(), "rpt-1")

	require.NoError(t, err)
}

func TestReportsDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newReportsService(&mockBackend{})
	err := svc.Delete(context.Background(), "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestReportsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newReportsService(mock)
	err := svc.Delete(context.Background(), "rpt-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// ReportsService.Download tests
// ---------------------------------------------------------------------------

func TestReportsDownload(t *testing.T) {
	t.Parallel()

	expectedContent := []byte("%PDF-1.4 fake pdf content")

	mock := &mockBackend{
		callRawFn: func(_ context.Context, method, path string, body any) ([]byte, error) {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/reports/rpt-1/download", path)
			assert.Nil(t, body)

			return expectedContent, nil
		},
	}

	svc := newReportsService(mock)
	data, err := svc.Download(context.Background(), "rpt-1")

	require.NoError(t, err)
	assert.Equal(t, expectedContent, data)
}

func TestReportsDownloadEmptyID(t *testing.T) {
	t.Parallel()

	svc := newReportsService(&mockBackend{})
	data, err := svc.Download(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, data)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestReportsDownloadBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("report not ready")
	mock := &mockBackend{
		callRawFn: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return nil, expectedErr
		},
	}

	svc := newReportsService(mock)
	data, err := svc.Download(context.Background(), "rpt-1")

	require.Error(t, err)
	assert.Nil(t, data)
	assert.Equal(t, expectedErr, err)
}
