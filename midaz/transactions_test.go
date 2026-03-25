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
// Transactions — Create
// ---------------------------------------------------------------------------

func TestTransactionsCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions", path)
			assert.NotNil(t, body)

			input, ok := body.(*createTransactionRequest)
			require.True(t, ok)
			require.NotNil(t, input.Send)
			assert.Equal(t, "BRL", input.Send.Asset)
			assert.Equal(t, "150.00", input.Send.Value)
			require.Len(t, input.Send.Source.From, 1)
			require.Len(t, input.Send.Distribute.To, 1)
			assert.Equal(t, "acc-1", input.Send.Source.From[0].AccountAlias)
			assert.Equal(t, "acc-2", input.Send.Distribute.To[0].AccountAlias)

			return unmarshalInto(Transaction{
				ID:             "txn-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				AssetCode:      "BRL",
				Amount:         15000,
				AmountScale:    2,
			}, result)
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Create(context.Background(), "org-1", "led-1", &CreateTransactionInput{
		Send: &TransactionSend{
			Asset: "BRL",
			Value: "150.00",
			Source: TransactionSendSource{From: []TransactionOperationLeg{{
				AccountAlias: "acc-1",
			}}},
			Distribute: TransactionSendDistribution{To: []TransactionOperationLeg{{
				AccountAlias: "acc-2",
			}}},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-1", txn.ID)
	assert.Equal(t, "BRL", txn.AssetCode)
	assert.Equal(t, int64(15000), txn.Amount)
}

func TestTransactionsCreateWithModernSendPayload(t *testing.T) {
	t.Parallel()

	description := "modern payload"
	code := "TX-001"
	pending := true
	route := "route-1"
	transactionDate := "2026-03-24T10:00:00Z"
	parentTransactionID := "txn-parent"

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions", path)

			input, ok := body.(*createTransactionRequest)
			require.True(t, ok)
			require.NotNil(t, input.Send)
			require.NotNil(t, input.Pending)
			require.NotNil(t, input.Route)
			require.NotNil(t, input.TransactionDate)
			assert.Equal(t, description, *input.Description)
			assert.Equal(t, code, *input.Code)
			require.NotNil(t, input.ParentTransactionID)
			assert.Equal(t, parentTransactionID, *input.ParentTransactionID)
			assert.True(t, *input.Pending)
			assert.Equal(t, route, *input.Route)
			assert.Equal(t, transactionDate, *input.TransactionDate)
			assert.Equal(t, "BRL", input.Send.Asset)
			assert.Equal(t, "150.00", input.Send.Value)
			assert.Equal(t, "freeze", input.Send.Source.From[0].BalanceKey)
			assert.Equal(t, "route-from", input.Send.Source.From[0].Route)
			assert.Equal(t, "remaining", input.Send.Distribute.To[0].Remaining)

			return unmarshalInto(Transaction{ID: "txn-modern"}, result)
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Create(context.Background(), "org-1", "led-1", &CreateTransactionInput{
		Description:         &description,
		Code:                &code,
		ParentTransactionID: &parentTransactionID,
		Pending:             &pending,
		Route:               &route,
		TransactionDate:     &transactionDate,
		Send: &TransactionSend{
			Asset: "BRL",
			Value: "150.00",
			Source: TransactionSendSource{From: []TransactionOperationLeg{{
				AccountAlias: "acc-1",
				BalanceKey:   "freeze",
				Route:        "route-from",
			}}},
			Distribute: TransactionSendDistribution{To: []TransactionOperationLeg{{
				AccountAlias: "acc-2",
				Remaining:    "remaining",
			}}},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-modern", txn.ID)
}

func TestTransactionsCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Create(context.Background(), "", "led-1", &CreateTransactionInput{Send: &TransactionSend{Asset: "BRL", Value: "1.00"}})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Create(context.Background(), "org-1", "", &CreateTransactionInput{Send: &TransactionSend{Asset: "BRL", Value: "1.00"}})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Create(context.Background(), "org-1", "led-1", nil)

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Create(context.Background(), "org-1", "led-1", &CreateTransactionInput{
		Send: &TransactionSend{
			Asset:      "BRL",
			Value:      "1.00",
			Source:     TransactionSendSource{From: []TransactionOperationLeg{{AccountAlias: "acc-1"}}},
			Distribute: TransactionSendDistribution{To: []TransactionOperationLeg{{AccountAlias: "acc-2"}}},
		},
	})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.Equal(t, expectedErr, err)
}

func TestTransactionsCreateWithoutSend(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Create(context.Background(), "org-1", "led-1", &CreateTransactionInput{})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	assert.Contains(t, err.Error(), "send is required")
}

// ---------------------------------------------------------------------------
// Transactions — Get
// ---------------------------------------------------------------------------

func TestTransactionsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/txn-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Transaction{
				ID:             "txn-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
			}, result)
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Get(context.Background(), "org-1", "led-1", "txn-1")

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-1", txn.ID)
}

func TestTransactionsGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Get(context.Background(), "", "led-1", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Get(context.Background(), "org-1", "", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Get(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Transactions — List
// ---------------------------------------------------------------------------

func TestTransactionsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/transactions")
			assert.Nil(t, body)

			resp := models.ListResponse[Transaction]{
				Items: []Transaction{
					{ID: "txn-1", AssetCode: "BRL"},
					{ID: "txn-2", AssetCode: "USD"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newTransactionsService(mock)
	iter := svc.List(context.Background(), "org-1", "led-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "txn-1", items[0].ID)
	assert.Equal(t, "txn-2", items[1].ID)
}

func TestTransactionsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)

			receivedPath = path

			assert.Nil(t, body)

			resp := models.ListResponse[Transaction]{
				Items:      []Transaction{{ID: "txn-1"}},
				Pagination: models.Pagination{Total: 1, Limit: 25},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newTransactionsService(mock)
	opts := &models.CursorListOptions{Limit: 25, SortBy: "createdAt", SortOrder: "desc"}
	iter := svc.List(context.Background(), "org-1", "led-1", opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortBy=createdAt")
	assert.Contains(t, receivedPath, "sortOrder=desc")
}

// ---------------------------------------------------------------------------
// Transactions — Update
// ---------------------------------------------------------------------------

func TestTransactionsUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/txn-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Transaction{
				ID:        "txn-1",
				AssetCode: "BRL",
			}, result)
		},
	}

	svc := newTransactionsService(mock)
	desc := "updated description"
	txn, err := svc.Update(context.Background(), "org-1", "led-1", "txn-1", &UpdateTransactionInput{
		Description: &desc,
	})

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-1", txn.ID)
}

func TestTransactionsUpdateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Update(context.Background(), "", "led-1", "txn-1", &UpdateTransactionInput{})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsUpdateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Update(context.Background(), "org-1", "", "txn-1", &UpdateTransactionInput{})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Update(context.Background(), "org-1", "led-1", "", &UpdateTransactionInput{})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Update(context.Background(), "org-1", "led-1", "txn-1", nil)

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Transactions — Commit (state machine action)
// ---------------------------------------------------------------------------

func TestTransactionsCommit(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/txn-1/commit", path)
			assert.Nil(t, body)

			return unmarshalInto(Transaction{
				ID:        "txn-1",
				Status:    models.Status{Code: "COMMITTED"},
				AssetCode: "BRL",
			}, result)
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Commit(context.Background(), "org-1", "led-1", "txn-1")

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-1", txn.ID)
	assert.Equal(t, "COMMITTED", txn.Status.Code)
}

func TestTransactionsCommitEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Commit(context.Background(), "", "led-1", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCommitEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Commit(context.Background(), "org-1", "", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCommitEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Commit(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCommitBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict on commit")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Commit(context.Background(), "org-1", "led-1", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Transactions — Cancel (state machine action)
// ---------------------------------------------------------------------------

func TestTransactionsCancel(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/txn-2/cancel", path)
			assert.Nil(t, body)

			return unmarshalInto(Transaction{
				ID:        "txn-2",
				Status:    models.Status{Code: "CANCELLED"},
				AssetCode: "USD",
			}, result)
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Cancel(context.Background(), "org-1", "led-1", "txn-2")

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-2", txn.ID)
	assert.Equal(t, "CANCELLED", txn.Status.Code)
}

func TestTransactionsCancelEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Cancel(context.Background(), "", "led-1", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCancelEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Cancel(context.Background(), "org-1", "", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCancelEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Cancel(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Transactions — Revert (state machine action)
// ---------------------------------------------------------------------------

func TestTransactionsRevert(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/txn-3/revert", path)
			assert.Nil(t, body)

			return unmarshalInto(Transaction{
				ID:        "txn-3",
				Status:    models.Status{Code: "REVERTED"},
				AssetCode: "BRL",
			}, result)
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Revert(context.Background(), "org-1", "led-1", "txn-3")

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-3", txn.ID)
	assert.Equal(t, "REVERTED", txn.Status.Code)
}

func TestTransactionsRevertEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Revert(context.Background(), "", "led-1", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsRevertEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Revert(context.Background(), "org-1", "", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsRevertEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.Revert(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsRevertBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: cannot revert")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.Revert(context.Background(), "org-1", "led-1", "txn-1")

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Transactions — Interface compliance
// ---------------------------------------------------------------------------

func TestTransactionsServiceInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ transactionsServiceAPI = (*transactionsService)(nil)

	t.Log("transactionsService implements transactionsServiceAPI")
}
