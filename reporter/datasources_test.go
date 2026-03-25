package reporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

func (m *mockBackend) Do(ctx context.Context, req core.Request) (*core.Response, error) {
	headers := reqHeaders(req)
	if len(headers) > 0 && m.callWithHdrsFn != nil {
		return m.doJSONWithHeaders(ctx, req, headers)
	}

	return m.doWithoutHeaders(ctx, req)
}

func (m *mockBackend) doWithoutHeaders(ctx context.Context, req core.Request) (*core.Response, error) {
	if m.callRawFn != nil && strings.Contains(req.Path, "/download") {
		return m.doRaw(ctx, req)
	}

	if m.callFn != nil {
		return m.doJSON(ctx, req)
	}

	if m.callRawFn != nil {
		return m.doRaw(ctx, req)
	}

	return nil, fmt.Errorf("mockBackend.Do not configured")
}

func (m *mockBackend) doJSONWithHeaders(ctx context.Context, req core.Request, headers map[string]string) (*core.Response, error) {
	var result any

	resultArg := any(&result)
	if req.ExpectNoResponse {
		resultArg = nil
	}

	if err := m.callWithHdrsFn(ctx, req.Method, req.Path, headers, reqBody(req), resultArg); err != nil {
		return nil, err
	}

	return responseForRequest(req, result)
}

func (m *mockBackend) doJSON(ctx context.Context, req core.Request) (*core.Response, error) {
	var result any

	resultArg := any(&result)
	if req.ExpectNoResponse {
		resultArg = nil
	}

	if err := m.callFn(ctx, req.Method, req.Path, reqBody(req), resultArg); err != nil {
		return nil, err
	}

	return responseForRequest(req, result)
}

func (m *mockBackend) doRaw(ctx context.Context, req core.Request) (*core.Response, error) {
	body, err := m.callRawFn(ctx, req.Method, req.Path, reqBody(req))
	if err != nil {
		return nil, err
	}

	return &core.Response{Body: body}, nil
}

func responseForRequest(req core.Request, result any) (*core.Response, error) {
	if req.ExpectNoResponse {
		return &core.Response{}, nil
	}

	return jsonResponse(result)
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

func jsonResponse(result any) (*core.Response, error) {
	if result == nil {
		return &core.Response{}, nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &core.Response{Body: data}, nil
}

func reqBody(req core.Request) any {
	if len(req.BodyBytes) > 0 {
		return req.BodyBytes
	}

	return req.Body
}

func reqHeaders(req core.Request) map[string]string {
	if len(req.Headers) == 0 && req.ContentType == "" {
		return nil
	}

	headers := map[string]string{}
	for k, v := range req.Headers {
		headers[k] = v
	}

	if req.ContentType != "" {
		headers["Content-Type"] = req.ContentType
	}

	return headers
}

// ---------------------------------------------------------------------------
// dataSourcesServiceAPI.Get tests
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
// dataSourcesServiceAPI.List tests
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
	opts := &models.CursorListOptions{Limit: 5, SortBy: "name"}
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
