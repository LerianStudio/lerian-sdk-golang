package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test types
// ---------------------------------------------------------------------------

type testOrg struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type testCreateInput struct {
	Name string `json:"name"`
}

type testUpdateInput struct {
	Name string `json:"name"`
}

type testActionInput struct {
	Reason string `json:"reason"`
}

// ---------------------------------------------------------------------------
// Mock Backend
// ---------------------------------------------------------------------------

// mockBackend is a hand-written mock that implements the Backend interface.
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
var _ Backend = (*mockBackend)(nil)

// ---------------------------------------------------------------------------
// Helper: unmarshalInto uses JSON round-trip to populate the result pointer
// from a source value. This simulates what a real backend does.
// ---------------------------------------------------------------------------

func unmarshalInto(src any, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dst)
}

// ---------------------------------------------------------------------------
// Get tests
// ---------------------------------------------------------------------------

func TestGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1", path)
			assert.Nil(t, body)

			return unmarshalInto(testOrg{ID: "org-1", Name: "Acme"}, result)
		},
	}

	svc := &BaseService{Backend: mock}
	org, err := Get[testOrg](context.Background(), svc, "/organizations/org-1")

	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, "org-1", org.ID)
	assert.Equal(t, "Acme", org.Name)
}

func TestGetError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := &BaseService{Backend: mock}
	org, err := Get[testOrg](context.Background(), svc, "/organizations/missing")

	require.Error(t, err)
	assert.Nil(t, org)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations", path)
			assert.NotNil(t, body)

			// Verify the body is the correct input type.
			input, ok := body.(*testCreateInput)
			require.True(t, ok)
			assert.Equal(t, "New Org", input.Name)

			return unmarshalInto(testOrg{ID: "org-new", Name: "New Org"}, result)
		},
	}

	svc := &BaseService{Backend: mock}
	input := &testCreateInput{Name: "New Org"}
	org, err := Create[testOrg, testCreateInput](context.Background(), svc, "/organizations", input)

	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, "org-new", org.ID)
	assert.Equal(t, "New Org", org.Name)
}

func TestCreateError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: validation failed")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := &BaseService{Backend: mock}
	input := &testCreateInput{Name: "Bad"}
	org, err := Create[testOrg, testCreateInput](context.Background(), svc, "/organizations", input)

	require.Error(t, err)
	assert.Nil(t, org)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Update tests
// ---------------------------------------------------------------------------

func TestUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1", path)
			assert.NotNil(t, body)

			input, ok := body.(*testUpdateInput)
			require.True(t, ok)
			assert.Equal(t, "Updated Org", input.Name)

			return unmarshalInto(testOrg{ID: "org-1", Name: "Updated Org"}, result)
		},
	}

	svc := &BaseService{Backend: mock}
	input := &testUpdateInput{Name: "Updated Org"}
	org, err := Update[testOrg, testUpdateInput](context.Background(), svc, "/organizations/org-1", input)

	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, "org-1", org.ID)
	assert.Equal(t, "Updated Org", org.Name)
}

func TestUpdateError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := &BaseService{Backend: mock}
	input := &testUpdateInput{Name: "Conflict"}
	org, err := Update[testOrg, testUpdateInput](context.Background(), svc, "/organizations/org-1", input)

	require.Error(t, err)
	assert.Nil(t, org)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := &BaseService{Backend: mock}
	err := Delete(context.Background(), svc, "/organizations/org-1")

	require.NoError(t, err)
}

func TestDeleteError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := &BaseService{Backend: mock}
	err := Delete(context.Background(), svc, "/organizations/org-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Action tests
// ---------------------------------------------------------------------------

func TestAction(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/transactions/txn-1/commit", path)

			return unmarshalInto(testOrg{ID: "txn-1", Name: "committed"}, result)
		},
	}

	svc := &BaseService{Backend: mock}
	res, err := Action[testOrg](context.Background(), svc, "/transactions/txn-1/commit", nil)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "txn-1", res.ID)
	assert.Equal(t, "committed", res.Name)
}

func TestActionWithInput(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/transactions/txn-2/cancel", path)
			assert.NotNil(t, body)

			// Verify the input body is passed through.
			input, ok := body.(*testActionInput)
			require.True(t, ok)
			assert.Equal(t, "duplicate", input.Reason)

			return unmarshalInto(testOrg{ID: "txn-2", Name: "cancelled"}, result)
		},
	}

	svc := &BaseService{Backend: mock}
	input := &testActionInput{Reason: "duplicate"}
	res, err := Action[testOrg](context.Background(), svc, "/transactions/txn-2/cancel", input)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "txn-2", res.ID)
	assert.Equal(t, "cancelled", res.Name)
}

func TestActionNilInput(t *testing.T) {
	t.Parallel()

	var receivedBody any

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/accounts/acc-1/activate", path)
			receivedBody = body

			return unmarshalInto(testOrg{ID: "acc-1", Name: "active"}, result)
		},
	}

	svc := &BaseService{Backend: mock}
	res, err := Action[testOrg](context.Background(), svc, "/accounts/acc-1/activate", nil)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Nil(t, receivedBody)
	assert.Equal(t, "acc-1", res.ID)
	assert.Equal(t, "active", res.Name)
}

func TestActionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict on action")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := &BaseService{Backend: mock}
	res, err := Action[testOrg](context.Background(), svc, "/transactions/txn-3/commit", nil)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// List tests
// ---------------------------------------------------------------------------

func TestList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations", path)
			assert.Nil(t, body)

			resp := models.ListResponse[testOrg]{
				Items: []testOrg{
					{ID: "org-1", Name: "Org One"},
					{ID: "org-2", Name: "Org Two"},
				},
				Pagination: models.Pagination{
					Total:      2,
					Limit:      10,
					NextCursor: "",
				},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := &BaseService{Backend: mock}
	iter := List[testOrg](context.Background(), svc, "/organizations", nil)

	require.NotNil(t, iter)

	// Collect all items.
	ctx := context.Background()
	var items []testOrg

	for iter.Next(ctx) {
		items = append(items, iter.Item())
	}

	require.NoError(t, iter.Err())
	assert.Len(t, items, 2)
	assert.Equal(t, "org-1", items[0].ID)
	assert.Equal(t, "org-2", items[1].ID)
}

func TestListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			receivedPath = path
			assert.Nil(t, body)

			resp := models.ListResponse[testOrg]{
				Items: []testOrg{{ID: "org-1", Name: "Found"}},
				Pagination: models.Pagination{
					Total: 1,
					Limit: 25,
				},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := &BaseService{Backend: mock}
	opts := &models.ListOptions{
		Limit:     25,
		SortBy:    "name",
		SortOrder: "asc",
	}

	iter := List[testOrg](context.Background(), svc, "/organizations", opts)

	ctx := context.Background()
	require.True(t, iter.Next(ctx))

	// Verify query parameters appear in the path.
	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortBy=name")
	assert.Contains(t, receivedPath, "sortOrder=asc")
}

func TestListWithCursor(t *testing.T) {
	t.Parallel()

	callCount := 0

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			callCount++
			assert.Equal(t, "GET", method)
			assert.Nil(t, body)

			if callCount == 1 {
				// First page.
				assert.Equal(t, "/organizations", path)

				resp := models.ListResponse[testOrg]{
					Items: []testOrg{{ID: "org-1", Name: "First"}},
					Pagination: models.Pagination{
						Total:      2,
						Limit:      1,
						NextCursor: "cursor-page-2",
					},
				}

				return unmarshalInto(resp, result)
			}

			// Second page -- cursor should be in the path.
			assert.Contains(t, path, "cursor=cursor-page-2")

			resp := models.ListResponse[testOrg]{
				Items: []testOrg{{ID: "org-2", Name: "Second"}},
				Pagination: models.Pagination{
					Total:      2,
					Limit:      1,
					NextCursor: "",
				},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := &BaseService{Backend: mock}
	iter := List[testOrg](context.Background(), svc, "/organizations", nil)

	ctx := context.Background()
	var items []testOrg

	for iter.Next(ctx) {
		items = append(items, iter.Item())
	}

	require.NoError(t, iter.Err())
	assert.Len(t, items, 2)
	assert.Equal(t, "org-1", items[0].ID)
	assert.Equal(t, "org-2", items[1].ID)
	assert.Equal(t, 2, callCount)
}

func TestListError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal server error")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := &BaseService{Backend: mock}
	iter := List[testOrg](context.Background(), svc, "/organizations", nil)

	ctx := context.Background()
	assert.False(t, iter.Next(ctx))
	assert.Equal(t, expectedErr, iter.Err())
}

func TestListEmptyResult(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			resp := models.ListResponse[testOrg]{
				Items: []testOrg{},
				Pagination: models.Pagination{
					Total: 0,
					Limit: 10,
				},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := &BaseService{Backend: mock}
	iter := List[testOrg](context.Background(), svc, "/organizations", nil)

	ctx := context.Background()
	assert.False(t, iter.Next(ctx))
	require.NoError(t, iter.Err())
}

// ---------------------------------------------------------------------------
// buildListPath tests
// ---------------------------------------------------------------------------

func TestBuildListPathEmpty(t *testing.T) {
	t.Parallel()

	// No options, no cursor -- path should be unchanged.
	result := buildListPath("/organizations", nil, "")
	assert.Equal(t, "/organizations", result)
}

func TestBuildListPathEmptyOpts(t *testing.T) {
	t.Parallel()

	// Empty ListOptions struct (all zero values) -- path should be unchanged.
	opts := &models.ListOptions{}
	result := buildListPath("/organizations", opts, "")
	assert.Equal(t, "/organizations", result)
}

func TestBuildListPathWithCursorOnly(t *testing.T) {
	t.Parallel()

	result := buildListPath("/organizations", nil, "abc-cursor")
	assert.Contains(t, result, "cursor=abc-cursor")
	assert.True(t, len(result) > len("/organizations"))
}

func TestBuildListPathWithLimit(t *testing.T) {
	t.Parallel()

	opts := &models.ListOptions{Limit: 50}
	result := buildListPath("/organizations", opts, "")
	assert.Contains(t, result, "limit=50")
}

func TestBuildListPathWithPage(t *testing.T) {
	t.Parallel()

	opts := &models.ListOptions{Page: 3}
	result := buildListPath("/organizations", opts, "")
	assert.Contains(t, result, "page=3")
}

func TestBuildListPathWithSorting(t *testing.T) {
	t.Parallel()

	opts := &models.ListOptions{
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	result := buildListPath("/ledgers", opts, "")
	assert.Contains(t, result, "sortBy=created_at")
	assert.Contains(t, result, "sortOrder=desc")
}

func TestBuildListPathWithDates(t *testing.T) {
	t.Parallel()

	startDate := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	endDate := time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC)

	opts := &models.ListOptions{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	result := buildListPath("/transactions", opts, "")
	assert.Contains(t, result, "startDate=2025-01-15T10%3A30%3A00Z")
	assert.Contains(t, result, "endDate=2025-06-30T23%3A59%3A59Z")
}

func TestBuildListPathWithFilters(t *testing.T) {
	t.Parallel()

	opts := &models.ListOptions{
		Filters: map[string]string{
			"status": "active",
		},
	}
	result := buildListPath("/accounts", opts, "")
	assert.Contains(t, result, "filter%5Bstatus%5D=active")
}

func TestBuildListPathCursorPrecedence(t *testing.T) {
	t.Parallel()

	// When both cursor arg and opts.Cursor are present, the cursor arg wins
	// because it comes from pagination (the next page cursor).
	opts := &models.ListOptions{
		Cursor: "initial-cursor",
	}
	result := buildListPath("/organizations", opts, "pagination-cursor")
	assert.Contains(t, result, "cursor=pagination-cursor")
	// The initial cursor from opts should NOT appear.
	assert.NotContains(t, result, "initial-cursor")
}

func TestBuildListPathInitialCursor(t *testing.T) {
	t.Parallel()

	// When cursor arg is empty, opts.Cursor should be used (initial page).
	opts := &models.ListOptions{
		Cursor: "start-here",
	}
	result := buildListPath("/organizations", opts, "")
	assert.Contains(t, result, "cursor=start-here")
}

func TestBuildListPathAllOptions(t *testing.T) {
	t.Parallel()

	startDate := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)

	opts := &models.ListOptions{
		Limit:     20,
		Page:      2,
		Cursor:    "initial",
		SortBy:    "name",
		SortOrder: "asc",
		StartDate: &startDate,
		EndDate:   &endDate,
		Filters: map[string]string{
			"type": "savings",
		},
	}

	result := buildListPath("/accounts", opts, "")

	assert.Contains(t, result, "limit=20")
	assert.Contains(t, result, "page=2")
	assert.Contains(t, result, "cursor=initial")
	assert.Contains(t, result, "sortBy=name")
	assert.Contains(t, result, "sortOrder=asc")
	assert.Contains(t, result, "startDate=")
	assert.Contains(t, result, "endDate=")
	assert.Contains(t, result, "filter%5Btype%5D=savings")
}

// ---------------------------------------------------------------------------
// Additional edge cases
// ---------------------------------------------------------------------------

func TestGetPreservesContext(t *testing.T) {
	t.Parallel()

	type ctxKey string

	ctx := context.WithValue(context.Background(), ctxKey("trace"), "trace-123")

	mock := &mockBackend{
		callFn: func(receivedCtx context.Context, _, _ string, _, result any) error {
			// Verify the context is passed through.
			val, ok := receivedCtx.Value(ctxKey("trace")).(string)
			assert.True(t, ok)
			assert.Equal(t, "trace-123", val)

			return unmarshalInto(testOrg{ID: "ctx-1", Name: "Context"}, result)
		},
	}

	svc := &BaseService{Backend: mock}
	org, err := Get[testOrg](ctx, svc, "/organizations/ctx-1")

	require.NoError(t, err)
	assert.Equal(t, "ctx-1", org.ID)
}

func TestCreateWithNilInputPointer(t *testing.T) {
	t.Parallel()

	// Even though it's unusual, verify that a nil *Input pointer is passed
	// through without issue (the Backend will decide how to handle it).
	mock := &mockBackend{
		callFn: func(_ context.Context, method, _ string, body, result any) error {
			assert.Equal(t, "POST", method)
			// body will be (*testCreateInput)(nil) -- not the same as nil interface
			return unmarshalInto(testOrg{ID: "nil-input", Name: "Created"}, result)
		},
	}

	svc := &BaseService{Backend: mock}
	org, err := Create[testOrg, testCreateInput](context.Background(), svc, "/organizations", nil)

	require.NoError(t, err)
	assert.Equal(t, "nil-input", org.ID)
}

func TestBaseServiceEmbedding(t *testing.T) {
	t.Parallel()

	// Simulate how a product service would embed BaseService.
	type orgService struct {
		BaseService
	}

	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return unmarshalInto(testOrg{ID: "embed-1", Name: "Embedded"}, result)
		},
	}

	service := &orgService{
		BaseService: BaseService{Backend: mock},
	}

	org, err := Get[testOrg](context.Background(), &service.BaseService, "/organizations/embed-1")
	require.NoError(t, err)
	assert.Equal(t, "embed-1", org.ID)
	assert.Equal(t, "Embedded", org.Name)
}

// ---------------------------------------------------------------------------
// Nil Backend / nil BaseService guard tests
// ---------------------------------------------------------------------------

func TestGet_NilBackend(t *testing.T) {
	t.Parallel()

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()

		org, err := Get[testOrg](context.Background(), nil, "/organizations/org-1")
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ErrNilService)
	})

	t.Run("nil backend", func(t *testing.T) {
		t.Parallel()

		svc := &BaseService{Backend: nil}
		org, err := Get[testOrg](context.Background(), svc, "/organizations/org-1")
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ErrNilBackend)
	})
}

func TestCreate_NilBackend(t *testing.T) {
	t.Parallel()

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()

		input := &testCreateInput{Name: "Test"}
		org, err := Create[testOrg, testCreateInput](context.Background(), nil, "/organizations", input)
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ErrNilService)
	})

	t.Run("nil backend", func(t *testing.T) {
		t.Parallel()

		svc := &BaseService{Backend: nil}
		input := &testCreateInput{Name: "Test"}
		org, err := Create[testOrg, testCreateInput](context.Background(), svc, "/organizations", input)
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ErrNilBackend)
	})
}

func TestUpdate_NilBackend(t *testing.T) {
	t.Parallel()

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()

		input := &testUpdateInput{Name: "Updated"}
		org, err := Update[testOrg, testUpdateInput](context.Background(), nil, "/organizations/org-1", input)
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ErrNilService)
	})

	t.Run("nil backend", func(t *testing.T) {
		t.Parallel()

		svc := &BaseService{Backend: nil}
		input := &testUpdateInput{Name: "Updated"}
		org, err := Update[testOrg, testUpdateInput](context.Background(), svc, "/organizations/org-1", input)
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ErrNilBackend)
	})
}

func TestDelete_NilBackend(t *testing.T) {
	t.Parallel()

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()

		err := Delete(context.Background(), nil, "/organizations/org-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNilService)
	})

	t.Run("nil backend", func(t *testing.T) {
		t.Parallel()

		svc := &BaseService{Backend: nil}
		err := Delete(context.Background(), svc, "/organizations/org-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNilBackend)
	})
}

func TestList_NilBackend(t *testing.T) {
	t.Parallel()

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()

		iter := List[testOrg](context.Background(), nil, "/organizations", nil)
		require.NotNil(t, iter, "List must never return a nil iterator")

		ctx := context.Background()
		assert.False(t, iter.Next(ctx), "poisoned iterator must return false on first Next()")
		assert.ErrorIs(t, iter.Err(), ErrNilService)
	})

	t.Run("nil backend", func(t *testing.T) {
		t.Parallel()

		svc := &BaseService{Backend: nil}
		iter := List[testOrg](context.Background(), svc, "/organizations", nil)
		require.NotNil(t, iter, "List must never return a nil iterator")

		ctx := context.Background()
		assert.False(t, iter.Next(ctx), "poisoned iterator must return false on first Next()")
		assert.ErrorIs(t, iter.Err(), ErrNilBackend)
	})
}

func TestAction_NilBackend(t *testing.T) {
	t.Parallel()

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()

		res, err := Action[testOrg](context.Background(), nil, "/transactions/txn-1/commit", nil)
		require.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrNilService)
	})

	t.Run("nil backend", func(t *testing.T) {
		t.Parallel()

		svc := &BaseService{Backend: nil}
		res, err := Action[testOrg](context.Background(), svc, "/transactions/txn-1/commit", nil)
		require.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrNilBackend)
	})
}
