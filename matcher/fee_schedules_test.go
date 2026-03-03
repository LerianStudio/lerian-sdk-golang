package matcher

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeeSchedulesCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fee-schedules", path)
			assert.NotNil(t, body)

			return unmarshalInto(FeeSchedule{ID: "fs-1", Name: "standard"}, result)
		}}
		svc := newFeeSchedulesService(mb)
		got, err := svc.Create(context.Background(), &CreateFeeScheduleInput{
			ContextID: "ctx-1",
			Name:      "standard",
			Rules:     []FeeRule{{Type: "percentage", Currency: "USD"}},
		})
		require.NoError(t, err)
		assert.Equal(t, "fs-1", got.ID)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newFeeSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestFeeSchedulesGet(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/fee-schedules/fs-1", path)

			return unmarshalInto(FeeSchedule{ID: "fs-1"}, result)
		}}
		svc := newFeeSchedulesService(mb)
		got, err := svc.Get(context.Background(), "fs-1")
		require.NoError(t, err)
		assert.Equal(t, "fs-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newFeeSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestFeeSchedulesList(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/fee-schedules")

		return nil
	}}
	svc := newFeeSchedulesService(mb)
	it := svc.List(context.Background(), nil)
	require.NotNil(t, it)
}

func TestFeeSchedulesUpdate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		name := "premium"
		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/fee-schedules/fs-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(FeeSchedule{ID: "fs-1", Name: "premium"}, result)
		}}
		svc := newFeeSchedulesService(mb)
		got, err := svc.Update(context.Background(), "fs-1", &UpdateFeeScheduleInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "premium", got.Name)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newFeeSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateFeeScheduleInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newFeeSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "fs-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestFeeSchedulesDelete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/fee-schedules/fs-1", path)
			assert.Nil(t, result)

			return nil
		}}
		svc := newFeeSchedulesService(mb)
		err := svc.Delete(context.Background(), "fs-1")
		require.NoError(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newFeeSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestFeeSchedulesSimulate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fee-schedules/fs-1/simulate", path)
			assert.NotNil(t, body)

			return unmarshalInto(FeeSimulationResult{
				TotalFee: 150,
				Scale:    2,
				Currency: "USD",
				Breakdown: []FeeResult{
					{RuleType: "percentage", Amount: 100, Scale: 2, Currency: "USD"},
					{RuleType: "fixed", Amount: 50, Scale: 2, Currency: "USD"},
				},
			}, result)
		}}
		svc := newFeeSchedulesService(mb)
		got, err := svc.Simulate(context.Background(), "fs-1", &SimulateFeeScheduleInput{
			Amount:   10000,
			Currency: "USD",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(150), got.TotalFee)
		assert.Equal(t, "USD", got.Currency)
		assert.Len(t, got.Breakdown, 2)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newFeeSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Simulate(context.Background(), "", &SimulateFeeScheduleInput{Amount: 100, Currency: "USD"})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newFeeSchedulesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Simulate(context.Background(), "fs-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// FeeSchedules — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestFeeSchedulesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFeeSchedulesService(mb)
	got, err := svc.Create(context.Background(), &CreateFeeScheduleInput{
		ContextID: "ctx-1", Name: "standard", Rules: []FeeRule{{Type: "percentage", Currency: "USD"}},
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestFeeSchedulesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFeeSchedulesService(mb)
	got, err := svc.Get(context.Background(), "fs-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestFeeSchedulesListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFeeSchedulesService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestFeeSchedulesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFeeSchedulesService(mb)
	name := "premium"
	got, err := svc.Update(context.Background(), "fs-1", &UpdateFeeScheduleInput{Name: &name})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestFeeSchedulesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFeeSchedulesService(mb)
	err := svc.Delete(context.Background(), "fs-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestFeeSchedulesSimulateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFeeSchedulesService(mb)
	got, err := svc.Simulate(context.Background(), "fs-1", &SimulateFeeScheduleInput{
		Amount: 10000, Currency: "USD",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}
