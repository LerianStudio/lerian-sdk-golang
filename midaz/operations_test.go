package midaz

import (
	"context"
	"errors"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Operations — Get
// ---------------------------------------------------------------------------

func TestOperationsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/operations/op-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Operation{
				ID:            "op-1",
				TransactionID: "txn-1",
				AccountID:     "acc-1",
				Type:          "debit",
				AssetCode:     "BRL",
				Amount:        10000,
				AmountScale:   2,
			}, result)
		},
	}

	svc := newOperationsService(mock)
	op, err := svc.Get(context.Background(), "org-1", "led-1", "op-1")

	require.NoError(t, err)
	require.NotNil(t, op)
	assert.Equal(t, "op-1", op.ID)
	assert.Equal(t, "debit", op.Type)
	assert.Equal(t, int64(10000), op.Amount)
}

func TestOperationsGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newOperationsService(&mockBackend{})
	op, err := svc.Get(context.Background(), "", "led-1", "op-1")

	require.Error(t, err)
	assert.Nil(t, op)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationsGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newOperationsService(&mockBackend{})
	op, err := svc.Get(context.Background(), "org-1", "", "op-1")

	require.Error(t, err)
	assert.Nil(t, op)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newOperationsService(&mockBackend{})
	op, err := svc.Get(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, op)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOperationsService(mock)
	op, err := svc.Get(context.Background(), "org-1", "led-1", "op-1")

	require.Error(t, err)
	assert.Nil(t, op)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Operations — List
// ---------------------------------------------------------------------------

func TestOperationsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/operations")
			assert.Nil(t, body)

			resp := models.ListResponse[Operation]{
				Items: []Operation{
					{ID: "op-1", Type: "debit", AssetCode: "BRL"},
					{ID: "op-2", Type: "credit", AssetCode: "BRL"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOperationsService(mock)
	iter := svc.List(context.Background(), "org-1", "led-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "op-1", items[0].ID)
	assert.Equal(t, "op-2", items[1].ID)
}

func TestOperationsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)

			receivedPath = path

			assert.Nil(t, body)

			resp := models.ListResponse[Operation]{
				Items:      []Operation{{ID: "op-1"}},
				Pagination: models.Pagination{Total: 1, Limit: 25},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOperationsService(mock)
	opts := &models.ListOptions{Limit: 25, SortBy: "type", SortOrder: "asc"}
	iter := svc.List(context.Background(), "org-1", "led-1", opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortBy=type")
	assert.Contains(t, receivedPath, "sortOrder=asc")
}

// ---------------------------------------------------------------------------
// Operations — ListByTransaction
// ---------------------------------------------------------------------------

func TestOperationsListByTransaction(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/transactions/txn-1/operations")
			assert.Nil(t, body)

			resp := models.ListResponse[Operation]{
				Items: []Operation{
					{ID: "op-1", TransactionID: "txn-1", Type: "debit"},
					{ID: "op-2", TransactionID: "txn-1", Type: "credit"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOperationsService(mock)
	iter := svc.ListByTransaction(context.Background(), "org-1", "led-1", "txn-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "op-1", items[0].ID)
	assert.Equal(t, "txn-1", items[0].TransactionID)
	assert.Equal(t, "op-2", items[1].ID)
}

func TestOperationsListByTransactionWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)

			receivedPath = path

			assert.Nil(t, body)

			resp := models.ListResponse[Operation]{
				Items:      []Operation{{ID: "op-1", TransactionID: "txn-1"}},
				Pagination: models.Pagination{Total: 1, Limit: 50},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOperationsService(mock)
	opts := &models.ListOptions{Limit: 50}
	iter := svc.ListByTransaction(context.Background(), "org-1", "led-1", "txn-1", opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "/organizations/org-1/ledgers/led-1/transactions/txn-1/operations")
	assert.Contains(t, receivedPath, "limit=50")
}

func TestOperationsListByTransactionBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal server error")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOperationsService(mock)
	iter := svc.ListByTransaction(context.Background(), "org-1", "led-1", "txn-1", nil)

	assert.False(t, iter.Next(context.Background()))
	assert.Equal(t, expectedErr, iter.Err())
}

// ---------------------------------------------------------------------------
// Operations — ListByAccount
// ---------------------------------------------------------------------------

func TestOperationsListByAccount(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/accounts/acc-1/operations")
			assert.Nil(t, body)

			resp := models.ListResponse[Operation]{
				Items: []Operation{
					{ID: "op-3", AccountID: "acc-1", Type: "debit"},
					{ID: "op-4", AccountID: "acc-1", Type: "credit"},
					{ID: "op-5", AccountID: "acc-1", Type: "debit"},
				},
				Pagination: models.Pagination{Total: 3, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOperationsService(mock)
	iter := svc.ListByAccount(context.Background(), "org-1", "led-1", "acc-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, "op-3", items[0].ID)
	assert.Equal(t, "acc-1", items[0].AccountID)
}

func TestOperationsListByAccountWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)

			receivedPath = path

			assert.Nil(t, body)

			resp := models.ListResponse[Operation]{
				Items:      []Operation{{ID: "op-3", AccountID: "acc-1"}},
				Pagination: models.Pagination{Total: 1, Limit: 30},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOperationsService(mock)
	opts := &models.ListOptions{Limit: 30, SortOrder: "desc"}
	iter := svc.ListByAccount(context.Background(), "org-1", "led-1", "acc-1", opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "/organizations/org-1/ledgers/led-1/accounts/acc-1/operations")
	assert.Contains(t, receivedPath, "limit=30")
	assert.Contains(t, receivedPath, "sortOrder=desc")
}

func TestOperationsListByAccountBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOperationsService(mock)
	iter := svc.ListByAccount(context.Background(), "org-1", "led-1", "acc-1", nil)

	assert.False(t, iter.Next(context.Background()))
	assert.Equal(t, expectedErr, iter.Err())
}

// ---------------------------------------------------------------------------
// Operations — Interface compliance
// ---------------------------------------------------------------------------

func TestOperationsServiceInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ OperationsService = (*operationsService)(nil)

	t.Log("operationsService implements OperationsService")
}
