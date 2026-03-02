package matcher

import (
	"context"
	"errors"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ExceptionsService.Create
// ---------------------------------------------------------------------------

func TestExceptionsCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/exceptions", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateExceptionInput)
			require.True(t, ok)
			assert.Equal(t, "ctx-1", input.ContextID)
			assert.Equal(t, "amount_mismatch", input.Type)

			unmarshalInto(t, Exception{ID: "exc-1", ContextID: "ctx-1", Type: "amount_mismatch", Status: "open", Priority: "high"}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.Create(context.Background(), &CreateExceptionInput{
			ContextID: "ctx-1",
			Type:      "amount_mismatch",
			Priority:  "high",
		})

		require.NoError(t, err)
		assert.Equal(t, "exc-1", got.ID)
		assert.Equal(t, "open", got.Status)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.Get
// ---------------------------------------------------------------------------

func TestExceptionsGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/exceptions/exc-1", path)
			unmarshalInto(t, Exception{ID: "exc-1", Status: "open"}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.Get(context.Background(), "exc-1")
		require.NoError(t, err)
		assert.Equal(t, "exc-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.List
// ---------------------------------------------------------------------------

func TestExceptionsList(t *testing.T) {
	mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/exceptions")

		resp := models.ListResponse[Exception]{
			Items: []Exception{
				{ID: "exc-1"},
				{ID: "exc-2"},
			},
			Pagination: models.Pagination{Total: 2, Limit: 10},
		}
		unmarshalInto(t, resp, result)

		return nil
	}}

	svc := newExceptionsService(mb)
	iter := svc.List(context.Background(), nil)
	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "exc-1", items[0].ID)
	assert.Equal(t, "exc-2", items[1].ID)
}

// ---------------------------------------------------------------------------
// ExceptionsService.Update
// ---------------------------------------------------------------------------

func TestExceptionsUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		priority := "critical"
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/exceptions/exc-1", path)
			assert.NotNil(t, body)

			input, ok := body.(*UpdateExceptionInput)
			require.True(t, ok)
			assert.Equal(t, "critical", *input.Priority)

			unmarshalInto(t, Exception{ID: "exc-1", Priority: "critical"}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.Update(context.Background(), "exc-1", &UpdateExceptionInput{Priority: &priority})
		require.NoError(t, err)
		assert.Equal(t, "critical", got.Priority)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateExceptionInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "exc-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.Delete
// ---------------------------------------------------------------------------

func TestExceptionsDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/exceptions/exc-1", path)
			assert.Nil(t, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		err := svc.Delete(context.Background(), "exc-1")
		require.NoError(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.Approve
// ---------------------------------------------------------------------------

func TestExceptionsApprove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/exceptions/exc-1/approve", path)
			assert.Nil(t, body)
			unmarshalInto(t, Exception{ID: "exc-1", Status: "approved"}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.Approve(context.Background(), "exc-1")
		require.NoError(t, err)
		assert.Equal(t, "approved", got.Status)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Approve(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.Reject
// ---------------------------------------------------------------------------

func TestExceptionsReject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/exceptions/exc-1/reject", path)
			assert.NotNil(t, body)

			input, ok := body.(*RejectExceptionInput)
			require.True(t, ok)
			assert.Equal(t, "duplicate entry", input.Reason)

			unmarshalInto(t, Exception{ID: "exc-1", Status: "rejected"}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.Reject(context.Background(), "exc-1", &RejectExceptionInput{Reason: "duplicate entry"})
		require.NoError(t, err)
		assert.Equal(t, "rejected", got.Status)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Reject(context.Background(), "", &RejectExceptionInput{Reason: "test"})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Reject(context.Background(), "exc-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.Reassign
// ---------------------------------------------------------------------------

func TestExceptionsReassign(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/exceptions/exc-1/reassign", path)
			assert.NotNil(t, body)

			input, ok := body.(*ReassignExceptionInput)
			require.True(t, ok)
			assert.Equal(t, "user-42", input.AssignTo)

			unmarshalInto(t, Exception{ID: "exc-1", AssignedTo: strPtr("user-42")}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.Reassign(context.Background(), "exc-1", &ReassignExceptionInput{AssignTo: "user-42"})
		require.NoError(t, err)
		require.NotNil(t, got.AssignedTo)
		assert.Equal(t, "user-42", *got.AssignedTo)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Reassign(context.Background(), "", &ReassignExceptionInput{AssignTo: "user-42"})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Reassign(context.Background(), "exc-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.BulkApprove
// ---------------------------------------------------------------------------

func TestExceptionsBulkApprove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/exceptions/bulk/approve", path)
			assert.NotNil(t, body)

			input, ok := body.(*BulkExceptionInput)
			require.True(t, ok)
			assert.Equal(t, []string{"exc-1", "exc-2", "exc-3"}, input.IDs)

			unmarshalInto(t, BulkExceptionResult{Processed: 3, Succeeded: 3, Failed: 0}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.BulkApprove(context.Background(), &BulkExceptionInput{
			IDs: []string{"exc-1", "exc-2", "exc-3"},
		})
		require.NoError(t, err)
		assert.Equal(t, 3, got.Processed)
		assert.Equal(t, 3, got.Succeeded)
		assert.Equal(t, 0, got.Failed)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.BulkApprove(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.BulkReject
// ---------------------------------------------------------------------------

func TestExceptionsBulkReject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/exceptions/bulk/reject", path)
			assert.NotNil(t, body)

			input, ok := body.(*BulkRejectInput)
			require.True(t, ok)
			assert.Equal(t, []string{"exc-1", "exc-2"}, input.IDs)
			assert.Equal(t, "not applicable", input.Reason)

			unmarshalInto(t, BulkExceptionResult{Processed: 2, Succeeded: 1, Failed: 1, Errors: []string{"exc-2: already rejected"}}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.BulkReject(context.Background(), &BulkRejectInput{
			IDs:    []string{"exc-1", "exc-2"},
			Reason: "not applicable",
		})
		require.NoError(t, err)
		assert.Equal(t, 2, got.Processed)
		assert.Equal(t, 1, got.Succeeded)
		assert.Equal(t, 1, got.Failed)
		assert.Len(t, got.Errors, 1)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.BulkReject(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.BulkReassign
// ---------------------------------------------------------------------------

func TestExceptionsBulkReassign(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/exceptions/bulk/reassign", path)
			assert.NotNil(t, body)

			input, ok := body.(*BulkReassignInput)
			require.True(t, ok)
			assert.Equal(t, []string{"exc-1", "exc-2"}, input.IDs)
			assert.Equal(t, "user-99", input.AssignTo)

			unmarshalInto(t, BulkExceptionResult{Processed: 2, Succeeded: 2, Failed: 0}, result)

			return nil
		}}

		svc := newExceptionsService(mb)
		got, err := svc.BulkReassign(context.Background(), &BulkReassignInput{
			IDs:      []string{"exc-1", "exc-2"},
			AssignTo: "user-99",
		})
		require.NoError(t, err)
		assert.Equal(t, 2, got.Processed)
		assert.Equal(t, 2, got.Succeeded)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newExceptionsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.BulkReassign(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// ExceptionsService.ListByContext
// ---------------------------------------------------------------------------

func TestExceptionsListByContext(t *testing.T) {
	mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/contexts/ctx-1/exceptions")

		resp := models.ListResponse[Exception]{
			Items: []Exception{
				{ID: "exc-1", ContextID: "ctx-1"},
			},
			Pagination: models.Pagination{Total: 1, Limit: 10},
		}
		unmarshalInto(t, resp, result)

		return nil
	}}

	svc := newExceptionsService(mb)
	iter := svc.ListByContext(context.Background(), "ctx-1", nil)
	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "ctx-1", items[0].ContextID)
}

// ---------------------------------------------------------------------------
// ExceptionsService.GetStatistics
// ---------------------------------------------------------------------------

func TestExceptionsGetStatistics(t *testing.T) {
	mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Equal(t, "/exceptions/statistics", path)

		unmarshalInto(t, ExceptionStatistics{
			Total:      42,
			ByStatus:   map[string]int{"open": 30, "approved": 10, "rejected": 2},
			ByPriority: map[string]int{"high": 15, "medium": 20, "low": 7},
			ByType:     map[string]int{"amount_mismatch": 25, "missing_record": 17},
			AvgAge:     48.5,
		}, result)

		return nil
	}}

	svc := newExceptionsService(mb)
	got, err := svc.GetStatistics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 42, got.Total)
	assert.Equal(t, 30, got.ByStatus["open"])
	assert.Equal(t, 15, got.ByPriority["high"])
	assert.Equal(t, 25, got.ByType["amount_mismatch"])
	assert.InDelta(t, 48.5, got.AvgAge, 0.01)
}

// ---------------------------------------------------------------------------
// Exceptions — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestExceptionsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.Create(context.Background(), &CreateExceptionInput{
		ContextID: "ctx-1", Type: "amount_mismatch", Priority: "high",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.Get(context.Background(), "exc-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	priority := "critical"
	got, err := svc.Update(context.Background(), "exc-1", &UpdateExceptionInput{Priority: &priority})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	err := svc.Delete(context.Background(), "exc-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsApproveBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.Approve(context.Background(), "exc-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsRejectBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.Reject(context.Background(), "exc-1", &RejectExceptionInput{Reason: "test"})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsReassignBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.Reassign(context.Background(), "exc-1", &ReassignExceptionInput{AssignTo: "user-42"})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsBulkApproveBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.BulkApprove(context.Background(), &BulkExceptionInput{IDs: []string{"exc-1"}})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsBulkRejectBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.BulkReject(context.Background(), &BulkRejectInput{IDs: []string{"exc-1"}, Reason: "test"})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsBulkReassignBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.BulkReassign(context.Background(), &BulkReassignInput{IDs: []string{"exc-1"}, AssignTo: "user-99"})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsListByContextBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	iter := svc.ListByContext(context.Background(), "ctx-1", nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestExceptionsGetStatisticsBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newExceptionsService(mb)
	got, err := svc.GetStatistics(context.Background())

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// strPtr returns a pointer to the given string.
func strPtr(s string) *string { return &s }
