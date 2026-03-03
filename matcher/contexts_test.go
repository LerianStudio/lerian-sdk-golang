package matcher

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextsCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/contexts", path)
			assert.NotNil(t, body)

			return unmarshalInto(Context{ID: "ctx-1", Name: "test"}, result)
		}}
		svc := newContextsService(mb)
		got, err := svc.Create(context.Background(), &CreateContextInput{Name: "test"})
		require.NoError(t, err)
		assert.Equal(t, "ctx-1", got.ID)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newContextsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestContextsGet(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/contexts/ctx-1", path)

			return unmarshalInto(Context{ID: "ctx-1"}, result)
		}}
		svc := newContextsService(mb)
		got, err := svc.Get(context.Background(), "ctx-1")
		require.NoError(t, err)
		assert.Equal(t, "ctx-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newContextsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestContextsList(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/contexts")

		return nil
	}}
	svc := newContextsService(mb)
	it := svc.List(context.Background(), nil)
	require.NotNil(t, it)
}

func TestContextsUpdate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		name := "updated"
		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/contexts/ctx-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Context{ID: "ctx-1", Name: "updated"}, result)
		}}
		svc := newContextsService(mb)
		got, err := svc.Update(context.Background(), "ctx-1", &UpdateContextInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "updated", got.Name)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newContextsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateContextInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newContextsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "ctx-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestContextsDelete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/contexts/ctx-1", path)
			assert.Nil(t, result)

			return nil
		}}
		svc := newContextsService(mb)
		err := svc.Delete(context.Background(), "ctx-1")
		require.NoError(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newContextsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestContextsClone(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/contexts/ctx-1/clone", path)
			assert.Nil(t, body)

			return unmarshalInto(Context{ID: "ctx-2", Name: "cloned"}, result)
		}}
		svc := newContextsService(mb)
		got, err := svc.Clone(context.Background(), "ctx-1")
		require.NoError(t, err)
		assert.Equal(t, "ctx-2", got.ID)
		assert.Equal(t, "cloned", got.Name)
	})

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		svc := newContextsService(&mockBackend{callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Clone(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// Contexts — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestContextsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newContextsService(mb)
	got, err := svc.Create(context.Background(), &CreateContextInput{Name: "test"})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestContextsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newContextsService(mb)
	got, err := svc.Get(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestContextsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newContextsService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestContextsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newContextsService(mb)
	name := "updated"
	got, err := svc.Update(context.Background(), "ctx-1", &UpdateContextInput{Name: &name})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestContextsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newContextsService(mb)
	err := svc.Delete(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestContextsCloneBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newContextsService(mb)
	got, err := svc.Clone(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}
