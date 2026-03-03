package matcher

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedulesCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/schedules", path)
			assert.NotNil(t, body)

			return unmarshalInto(Schedule{ID: "sched-1", Name: "daily"}, result)
		}}
		svc := newSchedulesService(mb)
		got, err := svc.Create(context.Background(), &CreateScheduleInput{
			ContextID: "ctx-1",
			Name:      "daily",
			CronExpr:  "0 0 * * *",
		})
		require.NoError(t, err)
		assert.Equal(t, "sched-1", got.ID)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSchedulesGet(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/schedules/sched-1", path)

			return unmarshalInto(Schedule{ID: "sched-1"}, result)
		}}
		svc := newSchedulesService(mb)
		got, err := svc.Get(context.Background(), "sched-1")
		require.NoError(t, err)
		assert.Equal(t, "sched-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSchedulesList(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/schedules")

		return nil
	}}
	svc := newSchedulesService(mb)
	it := svc.List(context.Background(), nil)
	require.NotNil(t, it)
}

func TestSchedulesUpdate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		name := "weekly"
		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/schedules/sched-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Schedule{ID: "sched-1", Name: "weekly"}, result)
		}}
		svc := newSchedulesService(mb)
		got, err := svc.Update(context.Background(), "sched-1", &UpdateScheduleInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "weekly", got.Name)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateScheduleInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "sched-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSchedulesDelete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/schedules/sched-1", path)
			assert.Nil(t, result)

			return nil
		}}
		svc := newSchedulesService(mb)
		err := svc.Delete(context.Background(), "sched-1")
		require.NoError(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// Schedules — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestSchedulesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSchedulesService(mb)
	got, err := svc.Create(context.Background(), &CreateScheduleInput{
		ContextID: "ctx-1", Name: "daily", CronExpr: "0 0 * * *",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSchedulesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSchedulesService(mb)
	got, err := svc.Get(context.Background(), "sched-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSchedulesListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSchedulesService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestSchedulesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSchedulesService(mb)
	name := "weekly"
	got, err := svc.Update(context.Background(), "sched-1", &UpdateScheduleInput{Name: &name})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSchedulesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSchedulesService(mb)
	err := svc.Delete(context.Background(), "sched-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
