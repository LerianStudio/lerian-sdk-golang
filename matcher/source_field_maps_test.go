package matcher

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceFieldMapsCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/source-field-maps", path)
			assert.NotNil(t, body)
			unmarshalInto(t, SourceFieldMap{ID: "sfm-1", FieldName: "amount"}, result)
			return nil
		}}
		svc := newSourceFieldMapsService(mb)
		got, err := svc.Create(context.Background(), &CreateSourceFieldMapInput{
			SourceID:  "src-1",
			FieldName: "amount",
			MappedTo:  "canonical_amount",
		})
		require.NoError(t, err)
		assert.Equal(t, "sfm-1", got.ID)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newSourceFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSourceFieldMapsGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/source-field-maps/sfm-1", path)
			unmarshalInto(t, SourceFieldMap{ID: "sfm-1"}, result)
			return nil
		}}
		svc := newSourceFieldMapsService(mb)
		got, err := svc.Get(context.Background(), "sfm-1")
		require.NoError(t, err)
		assert.Equal(t, "sfm-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newSourceFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSourceFieldMapsList(t *testing.T) {
	mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/source-field-maps")
		return nil
	}}
	svc := newSourceFieldMapsService(mb)
	it := svc.List(context.Background(), nil)
	require.NotNil(t, it)
}

func TestSourceFieldMapsUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mappedTo := "canonical_currency"
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/source-field-maps/sfm-1", path)
			assert.NotNil(t, body)
			unmarshalInto(t, SourceFieldMap{ID: "sfm-1", MappedTo: "canonical_currency"}, result)
			return nil
		}}
		svc := newSourceFieldMapsService(mb)
		got, err := svc.Update(context.Background(), "sfm-1", &UpdateSourceFieldMapInput{MappedTo: &mappedTo})
		require.NoError(t, err)
		assert.Equal(t, "canonical_currency", got.MappedTo)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newSourceFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateSourceFieldMapInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newSourceFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "sfm-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSourceFieldMapsDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/source-field-maps/sfm-1", path)
			assert.Nil(t, result)
			return nil
		}}
		svc := newSourceFieldMapsService(mb)
		err := svc.Delete(context.Background(), "sfm-1")
		require.NoError(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newSourceFieldMapsService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// SourceFieldMaps — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestSourceFieldMapsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourceFieldMapsService(mb)
	got, err := svc.Create(context.Background(), &CreateSourceFieldMapInput{
		SourceID: "src-1", FieldName: "amount", MappedTo: "canonical_amount",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSourceFieldMapsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourceFieldMapsService(mb)
	got, err := svc.Get(context.Background(), "sfm-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSourceFieldMapsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourceFieldMapsService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestSourceFieldMapsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourceFieldMapsService(mb)
	mappedTo := "canonical_currency"
	got, err := svc.Update(context.Background(), "sfm-1", &UpdateSourceFieldMapInput{MappedTo: &mappedTo})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSourceFieldMapsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourceFieldMapsService(mb)
	err := svc.Delete(context.Background(), "sfm-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
