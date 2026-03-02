package matcher

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldMapsCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/field-maps", path)
			assert.NotNil(t, body)
			unmarshalInto(t, FieldMap{ID: "fm-1", SourceField: "amount"}, result)
			return nil
		}}
		svc := newFieldMapsService(mb)
		got, err := svc.Create(context.Background(), &CreateFieldMapInput{
			ContextID:   "ctx-1",
			SourceField: "amount",
			TargetField: "target_amount",
		})
		require.NoError(t, err)
		assert.Equal(t, "fm-1", got.ID)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestFieldMapsGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/field-maps/fm-1", path)
			unmarshalInto(t, FieldMap{ID: "fm-1"}, result)
			return nil
		}}
		svc := newFieldMapsService(mb)
		got, err := svc.Get(context.Background(), "fm-1")
		require.NoError(t, err)
		assert.Equal(t, "fm-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestFieldMapsList(t *testing.T) {
	mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/field-maps")
		return nil
	}}
	svc := newFieldMapsService(mb)
	it := svc.List(context.Background(), nil)
	require.NotNil(t, it)
}

func TestFieldMapsUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		targetField := "new_target"
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/field-maps/fm-1", path)
			assert.NotNil(t, body)
			unmarshalInto(t, FieldMap{ID: "fm-1", TargetField: "new_target"}, result)
			return nil
		}}
		svc := newFieldMapsService(mb)
		got, err := svc.Update(context.Background(), "fm-1", &UpdateFieldMapInput{TargetField: &targetField})
		require.NoError(t, err)
		assert.Equal(t, "new_target", got.TargetField)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateFieldMapInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "fm-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestFieldMapsDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/field-maps/fm-1", path)
			assert.Nil(t, result)
			return nil
		}}
		svc := newFieldMapsService(mb)
		err := svc.Delete(context.Background(), "fm-1")
		require.NoError(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// FieldMaps — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestFieldMapsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFieldMapsService(mb)
	got, err := svc.Create(context.Background(), &CreateFieldMapInput{
		ContextID: "ctx-1", SourceField: "amount", TargetField: "target_amount",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestFieldMapsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFieldMapsService(mb)
	got, err := svc.Get(context.Background(), "fm-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestFieldMapsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFieldMapsService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestFieldMapsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFieldMapsService(mb)
	targetField := "new_target"
	got, err := svc.Update(context.Background(), "fm-1", &UpdateFieldMapInput{TargetField: &targetField})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestFieldMapsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newFieldMapsService(mb)
	err := svc.Delete(context.Background(), "fm-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
