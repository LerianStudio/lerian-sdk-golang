package midaz

import (
	"context"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cursorListTestCase struct {
	name string
	run  func(*testing.T)
}

func runCursorListTest(t *testing.T, tc cursorListTestCase) {
	t.Helper()

	t.Run(tc.name, func(t *testing.T) {
		t.Parallel()
		tc.run(t)
	})
}

func TestAdditionalListMethodsPropagateCursorOptions(t *testing.T) {
	t.Parallel()

	tests := []cursorListTestCase{
		{
			name: "organizations",
			run: func(t *testing.T) {
				t.Helper()

				var receivedPath string

				mock := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
					assert.Equal(t, "GET", method)

					receivedPath = path

					assert.Nil(t, body)

					return unmarshalInto(models.ListResponse[Organization]{
						Items:      []Organization{{ID: "org-1"}},
						Pagination: models.Pagination{Total: 1, Limit: 10},
					}, result)
				}}

				iter := newOrganizationsService(mock).List(context.Background(), &models.CursorListOptions{Limit: 10, Cursor: "cursor-org", SortBy: "createdAt", SortOrder: "desc"})

				require.True(t, iter.Next(context.Background()))
				assert.Contains(t, receivedPath, "cursor=cursor-org")
				assert.Contains(t, receivedPath, "limit=10")
				assert.Contains(t, receivedPath, "sortBy=createdAt")
				assert.Contains(t, receivedPath, "sortOrder=desc")
			},
		},
		{
			name: "ledgers",
			run: func(t *testing.T) {
				t.Helper()

				var receivedPath string

				mock := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
					assert.Equal(t, "GET", method)

					receivedPath = path

					assert.Nil(t, body)

					return unmarshalInto(models.ListResponse[Ledger]{
						Items:      []Ledger{{ID: "led-1"}},
						Pagination: models.Pagination{Total: 1, Limit: 10},
					}, result)
				}}

				iter := newLedgersService(mock).List(context.Background(), "org-1", &models.CursorListOptions{Limit: 10, Cursor: "cursor-ledger", SortBy: "name"})

				require.True(t, iter.Next(context.Background()))
				assert.Contains(t, receivedPath, "cursor=cursor-ledger")
				assert.Contains(t, receivedPath, "limit=10")
				assert.Contains(t, receivedPath, "sortBy=name")
			},
		},
		{
			name: "account types",
			run: func(t *testing.T) {
				t.Helper()

				var receivedPath string

				mock := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
					assert.Equal(t, "GET", method)

					receivedPath = path

					assert.Nil(t, body)

					return unmarshalInto(models.ListResponse[AccountType]{
						Items:      []AccountType{{ID: "type-1"}},
						Pagination: models.Pagination{Total: 1, Limit: 10},
					}, result)
				}}

				iter := newAccountTypesService(mock).List(context.Background(), "org-1", "led-1", &models.CursorListOptions{Limit: 10, Cursor: "cursor-type", SortOrder: "asc"})

				require.True(t, iter.Next(context.Background()))
				assert.Contains(t, receivedPath, "cursor=cursor-type")
				assert.Contains(t, receivedPath, "limit=10")
				assert.Contains(t, receivedPath, "sortOrder=asc")
			},
		},
		{
			name: "accounts",
			run: func(t *testing.T) {
				t.Helper()

				var receivedPath string

				mock := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
					assert.Equal(t, "GET", method)

					receivedPath = path

					assert.Nil(t, body)

					return unmarshalInto(models.ListResponse[Account]{
						Items:      []Account{{ID: "acc-1"}},
						Pagination: models.Pagination{Total: 1, Limit: 10},
					}, result)
				}}

				iter := newAccountsService(mock).List(context.Background(), "org-1", "led-1", &models.CursorListOptions{Limit: 10, Cursor: "cursor-account", SortBy: "name"})

				require.True(t, iter.Next(context.Background()))
				assert.Contains(t, receivedPath, "cursor=cursor-account")
				assert.Contains(t, receivedPath, "limit=10")
				assert.Contains(t, receivedPath, "sortBy=name")
			},
		},
		{
			name: "assets",
			run: func(t *testing.T) {
				t.Helper()

				var receivedPath string

				mock := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
					assert.Equal(t, "GET", method)

					receivedPath = path

					assert.Nil(t, body)

					return unmarshalInto(models.ListResponse[Asset]{
						Items:      []Asset{{ID: "asset-1"}},
						Pagination: models.Pagination{Total: 1, Limit: 10},
					}, result)
				}}

				iter := newAssetsService(mock).List(context.Background(), "org-1", "led-1", &models.CursorListOptions{Limit: 10, Cursor: "cursor-asset", SortOrder: "desc"})

				require.True(t, iter.Next(context.Background()))
				assert.Contains(t, receivedPath, "cursor=cursor-asset")
				assert.Contains(t, receivedPath, "limit=10")
				assert.Contains(t, receivedPath, "sortOrder=desc")
			},
		},
	}

	for _, tc := range tests {
		runCursorListTest(t, tc)
	}
}
