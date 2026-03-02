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
// DisputesService.Create
// ---------------------------------------------------------------------------

func TestDisputesCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/disputes", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateDisputeInput)
			require.True(t, ok)
			assert.Equal(t, "ctx-1", input.ContextID)
			assert.Equal(t, "incorrect match", input.Reason)

			unmarshalInto(t, Dispute{ID: "dsp-1", ContextID: "ctx-1", Reason: "incorrect match", Status: "open"}, result)

			return nil
		}}

		svc := newDisputesService(mb)
		got, err := svc.Create(context.Background(), &CreateDisputeInput{
			ContextID: "ctx-1",
			Reason:    "incorrect match",
		})

		require.NoError(t, err)
		assert.Equal(t, "dsp-1", got.ID)
		assert.Equal(t, "open", got.Status)
		assert.Equal(t, "incorrect match", got.Reason)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newDisputesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// DisputesService.Get
// ---------------------------------------------------------------------------

func TestDisputesGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/disputes/dsp-1", path)
			unmarshalInto(t, Dispute{ID: "dsp-1", Status: "open"}, result)

			return nil
		}}

		svc := newDisputesService(mb)
		got, err := svc.Get(context.Background(), "dsp-1")
		require.NoError(t, err)
		assert.Equal(t, "dsp-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newDisputesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// DisputesService.List
// ---------------------------------------------------------------------------

func TestDisputesList(t *testing.T) {
	mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/disputes")

		resp := models.ListResponse[Dispute]{
			Items: []Dispute{
				{ID: "dsp-1", Status: "open"},
				{ID: "dsp-2", Status: "resolved"},
			},
			Pagination: models.Pagination{Total: 2, Limit: 10},
		}
		unmarshalInto(t, resp, result)

		return nil
	}}

	svc := newDisputesService(mb)
	iter := svc.List(context.Background(), nil)
	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "dsp-1", items[0].ID)
	assert.Equal(t, "dsp-2", items[1].ID)
}

func TestDisputesListWithOptions(t *testing.T) {
	var receivedPath string

	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, path string, _, result any) error {
		receivedPath = path

		resp := models.ListResponse[Dispute]{
			Items:      []Dispute{{ID: "dsp-1"}},
			Pagination: models.Pagination{Total: 1, Limit: 25},
		}
		unmarshalInto(t, resp, result)

		return nil
	}}

	svc := newDisputesService(mb)
	opts := &models.ListOptions{Limit: 25, SortOrder: "desc"}
	iter := svc.List(context.Background(), opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortOrder=desc")
}

// ---------------------------------------------------------------------------
// DisputesService.Update
// ---------------------------------------------------------------------------

func TestDisputesUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reason := "updated reason"
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/disputes/dsp-1", path)
			assert.NotNil(t, body)

			input, ok := body.(*UpdateDisputeInput)
			require.True(t, ok)
			assert.Equal(t, "updated reason", *input.Reason)

			unmarshalInto(t, Dispute{ID: "dsp-1", Reason: "updated reason"}, result)

			return nil
		}}

		svc := newDisputesService(mb)
		got, err := svc.Update(context.Background(), "dsp-1", &UpdateDisputeInput{Reason: &reason})
		require.NoError(t, err)
		assert.Equal(t, "updated reason", got.Reason)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newDisputesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateDisputeInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newDisputesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "dsp-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// DisputesService.Resolve
// ---------------------------------------------------------------------------

func TestDisputesResolve(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/disputes/dsp-1/resolve", path)
			assert.NotNil(t, body)

			input, ok := body.(*ResolveDisputeInput)
			require.True(t, ok)
			assert.Equal(t, "records confirmed as correct", input.Resolution)

			resolution := "records confirmed as correct"
			unmarshalInto(t, Dispute{ID: "dsp-1", Status: "resolved", Resolution: &resolution}, result)

			return nil
		}}

		svc := newDisputesService(mb)
		got, err := svc.Resolve(context.Background(), "dsp-1", &ResolveDisputeInput{
			Resolution: "records confirmed as correct",
		})
		require.NoError(t, err)
		assert.Equal(t, "resolved", got.Status)
		require.NotNil(t, got.Resolution)
		assert.Equal(t, "records confirmed as correct", *got.Resolution)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newDisputesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Resolve(context.Background(), "", &ResolveDisputeInput{Resolution: "test"})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newDisputesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Resolve(context.Background(), "dsp-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// DisputesService.Escalate
// ---------------------------------------------------------------------------

func TestDisputesEscalate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/disputes/dsp-1/escalate", path)
			assert.Nil(t, body)
			unmarshalInto(t, Dispute{ID: "dsp-1", Status: "escalated"}, result)

			return nil
		}}

		svc := newDisputesService(mb)
		got, err := svc.Escalate(context.Background(), "dsp-1")
		require.NoError(t, err)
		assert.Equal(t, "escalated", got.Status)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newDisputesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Escalate(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// Disputes — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestDisputesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newDisputesService(mb)
	got, err := svc.Create(context.Background(), &CreateDisputeInput{
		ContextID: "ctx-1", Reason: "incorrect match",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestDisputesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newDisputesService(mb)
	got, err := svc.Get(context.Background(), "dsp-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestDisputesListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newDisputesService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestDisputesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newDisputesService(mb)
	reason := "updated"
	got, err := svc.Update(context.Background(), "dsp-1", &UpdateDisputeInput{Reason: &reason})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestDisputesResolveBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newDisputesService(mb)
	got, err := svc.Resolve(context.Background(), "dsp-1", &ResolveDisputeInput{Resolution: "confirmed"})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestDisputesEscalateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newDisputesService(mb)
	got, err := svc.Escalate(context.Background(), "dsp-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}
