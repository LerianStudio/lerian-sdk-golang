package matcher

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRulesCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/rules", path)
			assert.NotNil(t, body)
			unmarshalInto(t, Rule{ID: "rule-1", Name: "amount-match"}, result)
			return nil
		}}
		svc := newRulesService(mb)
		got, err := svc.Create(context.Background(), &CreateRuleInput{
			ContextID:  "ctx-1",
			Name:       "amount-match",
			Priority:   1,
			Expression: "amount == amount",
		})
		require.NoError(t, err)
		assert.Equal(t, "rule-1", got.ID)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newRulesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Create(context.Background(), nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestRulesGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/rules/rule-1", path)
			unmarshalInto(t, Rule{ID: "rule-1"}, result)
			return nil
		}}
		svc := newRulesService(mb)
		got, err := svc.Get(context.Background(), "rule-1")
		require.NoError(t, err)
		assert.Equal(t, "rule-1", got.ID)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newRulesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Get(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestRulesList(t *testing.T) {
	mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/rules")
		return nil
	}}
	svc := newRulesService(mb)
	it := svc.List(context.Background(), nil)
	require.NotNil(t, it)
}

func TestRulesUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		name := "updated-rule"
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/rules/rule-1", path)
			assert.NotNil(t, body)
			unmarshalInto(t, Rule{ID: "rule-1", Name: "updated-rule"}, result)
			return nil
		}}
		svc := newRulesService(mb)
		got, err := svc.Update(context.Background(), "rule-1", &UpdateRuleInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "updated-rule", got.Name)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newRulesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "", &UpdateRuleInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newRulesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		_, err := svc.Update(context.Background(), "rule-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestRulesDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/rules/rule-1", path)
			assert.Nil(t, result)
			return nil
		}}
		svc := newRulesService(mb)
		err := svc.Delete(context.Background(), "rule-1")
		require.NoError(t, err)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := newRulesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Delete(context.Background(), "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

func TestRulesReorder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/contexts/ctx-1/rules/reorder", path)
			assert.NotNil(t, body)
			assert.Nil(t, result)
			return nil
		}}
		svc := newRulesService(mb)
		err := svc.Reorder(context.Background(), "ctx-1", &ReorderRulesInput{
			RuleIDs: []string{"rule-2", "rule-1", "rule-3"},
		})
		require.NoError(t, err)
	})

	t.Run("empty context id", func(t *testing.T) {
		svc := newRulesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Reorder(context.Background(), "", &ReorderRulesInput{RuleIDs: []string{"a"}})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})

	t.Run("nil input", func(t *testing.T) {
		svc := newRulesService(&mockBackend{t: t, callFn: func(context.Context, string, string, any, any) error { return nil }})
		err := svc.Reorder(context.Background(), "ctx-1", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	})
}

// ---------------------------------------------------------------------------
// Rules — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestRulesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newRulesService(mb)
	got, err := svc.Create(context.Background(), &CreateRuleInput{
		ContextID: "ctx-1", Name: "r", Priority: 1, Expression: "a == b",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestRulesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newRulesService(mb)
	got, err := svc.Get(context.Background(), "rule-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestRulesListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newRulesService(mb)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestRulesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newRulesService(mb)
	name := "updated"
	got, err := svc.Update(context.Background(), "rule-1", &UpdateRuleInput{Name: &name})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestRulesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newRulesService(mb)
	err := svc.Delete(context.Background(), "rule-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestRulesReorderBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newRulesService(mb)
	err := svc.Reorder(context.Background(), "ctx-1", &ReorderRulesInput{RuleIDs: []string{"a", "b"}})

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
