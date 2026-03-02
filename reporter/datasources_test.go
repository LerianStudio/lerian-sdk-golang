package reporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// mockBackend for reporter service tests
// ---------------------------------------------------------------------------

// mockBackend is a hand-written mock that implements core.Backend.
// Each function field can be overridden per-test to control behaviour.
type mockBackend struct {
	callFn         func(ctx context.Context, method, path string, body, result any) error
	callWithHdrsFn func(ctx context.Context, method, path string, headers map[string]string, body, result any) error
	callRawFn      func(ctx context.Context, method, path string, body any) ([]byte, error)
}

func (m *mockBackend) Call(ctx context.Context, method, path string, body, result any) error {
	if m.callFn != nil {
		return m.callFn(ctx, method, path, body, result)
	}

	return fmt.Errorf("mockBackend.Call not configured")
}

func (m *mockBackend) CallWithHeaders(ctx context.Context, method, path string,
	headers map[string]string, body, result any) error {
	if m.callWithHdrsFn != nil {
		return m.callWithHdrsFn(ctx, method, path, headers, body, result)
	}

	return fmt.Errorf("mockBackend.CallWithHeaders not configured")
}

func (m *mockBackend) CallRaw(ctx context.Context, method, path string, body any) ([]byte, error) {
	if m.callRawFn != nil {
		return m.callRawFn(ctx, method, path, body)
	}

	return nil, fmt.Errorf("mockBackend.CallRaw not configured")
}

// Compile-time interface compliance check.
var _ core.Backend = (*mockBackend)(nil)

// unmarshalInto uses JSON round-trip to populate the result pointer from
// a source value, simulating what a real backend does.
func unmarshalInto(src any, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dst)
}

// ---------------------------------------------------------------------------
// DataSourcesService.Get tests
// ---------------------------------------------------------------------------

func TestDataSourcesGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/datasources/ds-1", path)
			assert.Nil(t, body)

			return unmarshalInto(DataSource{ID: "ds-1", Name: "Primary DB", Type: "postgres"}, result)
		},
	}

	svc := newDataSourcesService(mock)
	ds, err := svc.Get(context.Background(), "ds-1")

	require.NoError(t, err)
	require.NotNil(t, ds)
	assert.Equal(t, "ds-1", ds.ID)
	assert.Equal(t, "Primary DB", ds.Name)
	assert.Equal(t, "postgres", ds.Type)
}

func TestDataSourcesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newDataSourcesService(&mockBackend{})
	ds, err := svc.Get(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, ds)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestDataSourcesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("backend failure")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newDataSourcesService(mock)
	ds, err := svc.Get(context.Background(), "ds-missing")

	require.Error(t, err)
	assert.Nil(t, ds)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// DataSourcesService.List tests
// ---------------------------------------------------------------------------

func TestDataSourcesList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/datasources")
			assert.Nil(t, body)

			resp := models.ListResponse[DataSource]{
				Items: []DataSource{
					{ID: "ds-1", Name: "Primary DB"},
					{ID: "ds-2", Name: "Analytics Stream"},
				},
				Pagination: models.Pagination{
					Total: 2,
					Limit: 10,
				},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newDataSourcesService(mock)
	iter := svc.List(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "ds-1", items[0].ID)
	assert.Equal(t, "ds-2", items[1].ID)
}

func TestDataSourcesListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, _, path string, _, result any) error {
			receivedPath = path

			resp := models.ListResponse[DataSource]{
				Items:      []DataSource{{ID: "ds-1", Name: "Found"}},
				Pagination: models.Pagination{Total: 1, Limit: 5},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newDataSourcesService(mock)
	opts := &models.ListOptions{Limit: 5, SortBy: "name"}
	iter := svc.List(context.Background(), opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=5")
	assert.Contains(t, receivedPath, "sortBy=name")
}

func TestDataSourcesListEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			resp := models.ListResponse[DataSource]{
				Items:      []DataSource{},
				Pagination: models.Pagination{Total: 0, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newDataSourcesService(mock)
	iter := svc.List(context.Background(), nil)

	assert.False(t, iter.Next(context.Background()))
	require.NoError(t, iter.Err())
}

func TestDataSourcesListError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("backend list error")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newDataSourcesService(mock)
	iter := svc.List(context.Background(), nil)

	assert.False(t, iter.Next(context.Background()))
	assert.Equal(t, expectedErr, iter.Err())
}
