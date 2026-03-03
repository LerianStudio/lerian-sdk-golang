package matcher

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourcesCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/sources", path)
			assert.NotNil(t, body)

			return unmarshalInto(Source{ID: "src-1", Name: "bank-feed"}, result)
		}}
		svc := newSourcesService(mb)
		got, err := svc.Create(context.Background(), &CreateSourceInput{
			ContextID: "ctx-1",
			Name:      "bank-feed",
			Type:      "api",
		})
		require.NoError(t, err)
		assert.Equal(t, "src-1", got.ID)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newSourcesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSourcesGet(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/sources/src-1", path)

			return unmarshalInto(Source{ID: "src-1"}, result)
		}}
		svc := newSourcesService(mb)
		got, err := svc.Get(context.Background(), "src-1")
		require.NoError(t, err)
		assert.Equal(t, "src-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newSourcesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSourcesList(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/sources")

		return nil
	}}
	svc := newSourcesService(mb)
	it := svc.List(context.Background(), nil)
	require.NotNil(t, it)
}

func TestSourcesUpdate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		name := "erp-system"
		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/sources/src-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Source{ID: "src-1", Name: "erp-system"}, result)
		}}
		svc := newSourcesService(mb)
		got, err := svc.Update(context.Background(), "src-1", &UpdateSourceInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "erp-system", got.Name)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newSourcesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateSourceInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newSourcesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "src-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestSourcesDelete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/sources/src-1", path)
			assert.Nil(t, result)

			return nil
		}}
		svc := newSourcesService(mb)
		err := svc.Delete(context.Background(), "src-1")
		require.NoError(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newSourcesService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// Sources — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestSourcesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourcesService(mb)
	got, err := svc.Create(context.Background(), &CreateSourceInput{
		ContextID: "ctx-1", Name: "bank-feed", Type: "api",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSourcesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourcesService(mb)
	got, err := svc.Get(context.Background(), "src-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSourcesListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourcesService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestSourcesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourcesService(mb)
	name := "erp"
	got, err := svc.Update(context.Background(), "src-1", &UpdateSourceInput{Name: &name})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestSourcesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newSourcesService(mb)
	err := svc.Delete(context.Background(), "src-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
