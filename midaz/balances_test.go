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
// Balances — backend wiring verification
// ---------------------------------------------------------------------------

func TestBalancesUsesTransactionBackend(t *testing.T) {
	t.Parallel()

	// The key invariant for the Balances service: it must be wired to the
	// transaction backend, NOT the onboarding backend. We verify this by
	// injecting two distinct mocks and confirming the Balances service
	// calls the transaction mock.

	onboardingCalled := false
	transactionCalled := false

	onboardingMock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			onboardingCalled = true
			return nil
		},
	}

	transactionMock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			transactionCalled = true

			return unmarshalInto(Balance{ID: "bal-1"}, result)
		},
	}

	client := NewClient(onboardingMock, transactionMock, Config{})

	_, err := client.Balances.Get(context.Background(), testOrgID, testLedgerID, "bal-1")
	require.NoError(t, err)

	assert.False(t, onboardingCalled, "Balances must NOT use the onboarding backend")
	assert.True(t, transactionCalled, "Balances must use the transaction backend")
}

// ---------------------------------------------------------------------------
// Balances — Create
// ---------------------------------------------------------------------------

func TestBalancesCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/balances", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateBalanceInput)
			require.True(t, ok)
			assert.Equal(t, "acc-1", input.AccountID)
			assert.Equal(t, "BRL", input.AssetCode)
			assert.True(t, input.AllowSending)
			assert.True(t, input.AllowReceiving)

			return unmarshalInto(Balance{
				ID:        "bal-1",
				AccountID: "acc-1",
				AssetCode: "BRL",
			}, result)
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.Create(context.Background(), testOrgID, testLedgerID, &CreateBalanceInput{
		AccountID:      "acc-1",
		AssetCode:      "BRL",
		AllowSending:   true,
		AllowReceiving: true,
	})

	require.NoError(t, err)
	require.NotNil(t, bal)
	assert.Equal(t, "bal-1", bal.ID)
	assert.Equal(t, "acc-1", bal.AccountID)
}

func TestBalancesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.Create(context.Background(), testOrgID, testLedgerID, nil)

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.Create(context.Background(), "", testLedgerID, &CreateBalanceInput{})

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.Create(context.Background(), testOrgID, "", &CreateBalanceInput{})

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — Get
// ---------------------------------------------------------------------------

func TestBalancesGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/balances/bal-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Balance{
				ID:        "bal-1",
				Available: 5000000,
				Scale:     2,
			}, result)
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.Get(context.Background(), testOrgID, testLedgerID, "bal-1")

	require.NoError(t, err)
	require.NotNil(t, bal)
	assert.Equal(t, "bal-1", bal.ID)
	assert.Equal(t, int64(5000000), bal.Available)
}

func TestBalancesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.Get(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.Get(context.Background(), "", testLedgerID, "bal-1")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.Get(context.Background(), testOrgID, "", "bal-1")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — List
// ---------------------------------------------------------------------------

func TestBalancesList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/balances")
			assert.Nil(t, body)

			resp := models.ListResponse[Balance]{
				Items: []Balance{
					{ID: "bal-1", AccountID: "acc-1", AssetCode: "BRL"},
					{ID: "bal-2", AccountID: "acc-2", AssetCode: "USD"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newBalancesService(mock)
	iter := svc.List(context.Background(), testOrgID, testLedgerID, nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "bal-1", items[0].ID)
	assert.Equal(t, "bal-2", items[1].ID)
}

func TestBalancesListValidation(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})

	tests := []struct {
		name     string
		orgID    string
		ledgerID string
	}{
		{name: "empty orgID", orgID: "", ledgerID: testLedgerID},
		{name: "empty ledgerID", orgID: testOrgID, ledgerID: ""},
		{name: "both empty", orgID: "", ledgerID: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			iter := svc.List(context.Background(), tc.orgID, tc.ledgerID, nil)
			require.NotNil(t, iter)

			items, err := iter.Collect(context.Background())
			require.Error(t, err)
			assert.Nil(t, items)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
			assert.Contains(t, err.Error(), "required")
		})
	}
}

// ---------------------------------------------------------------------------
// Balances — Update
// ---------------------------------------------------------------------------

func TestBalancesUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/balances/bal-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Balance{
				ID:           "bal-1",
				AllowSending: false,
			}, result)
		},
	}

	svc := newBalancesService(mock)
	sending := false
	bal, err := svc.Update(context.Background(), testOrgID, testLedgerID, "bal-1", &UpdateBalanceInput{
		AllowSending: &sending,
	})

	require.NoError(t, err)
	require.NotNil(t, bal)
	assert.False(t, bal.AllowSending)
}

func TestBalancesUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.Update(context.Background(), testOrgID, testLedgerID, "", &UpdateBalanceInput{})

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.Update(context.Background(), testOrgID, testLedgerID, "bal-1", nil)

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — Delete
// ---------------------------------------------------------------------------

func TestBalancesDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/balances/bal-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newBalancesService(mock)
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "bal-1")

	require.NoError(t, err)
}

func TestBalancesDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	err := svc.Delete(context.Background(), "", testLedgerID, "bal-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — GetByAlias
// ---------------------------------------------------------------------------

func TestBalancesGetByAlias(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/balances/alias/primary-checking", path)
			assert.Nil(t, body)

			return unmarshalInto(Balance{
				ID:           "bal-1",
				AccountAlias: ptr("primary-checking"),
			}, result)
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.GetByAlias(context.Background(), testOrgID, testLedgerID, "primary-checking")

	require.NoError(t, err)
	require.NotNil(t, bal)
	assert.Equal(t, "bal-1", bal.ID)
	require.NotNil(t, bal.AccountAlias)
	assert.Equal(t, "primary-checking", *bal.AccountAlias)
}

func TestBalancesGetByAliasEmptyAlias(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.GetByAlias(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesGetByAliasEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.GetByAlias(context.Background(), "", testLedgerID, "my-alias")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — GetByExternalCode
// ---------------------------------------------------------------------------

func TestBalancesGetByExternalCode(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/balances/external-code/EXT-BAL-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Balance{
				ID: "bal-1",
			}, result)
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.GetByExternalCode(context.Background(), testOrgID, testLedgerID, "EXT-BAL-1")

	require.NoError(t, err)
	require.NotNil(t, bal)
	assert.Equal(t, "bal-1", bal.ID)
}

func TestBalancesGetByExternalCodeEmpty(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.GetByExternalCode(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesGetByExternalCodeEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.GetByExternalCode(context.Background(), testOrgID, "", "EXT-BAL-1")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — GetByAccountID
// ---------------------------------------------------------------------------

func TestBalancesGetByAccountID(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/balances/account/acc-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Balance{
				ID:        "bal-1",
				AccountID: "acc-1",
			}, result)
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.GetByAccountID(context.Background(), testOrgID, testLedgerID, "acc-1")

	require.NoError(t, err)
	require.NotNil(t, bal)
	assert.Equal(t, "bal-1", bal.ID)
	assert.Equal(t, "acc-1", bal.AccountID)
}

func TestBalancesGetByAccountIDEmpty(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.GetByAccountID(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesGetByAccountIDEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.GetByAccountID(context.Background(), "", testLedgerID, "acc-1")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesGetByAccountIDEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.GetByAccountID(context.Background(), testOrgID, "", "acc-1")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestBalancesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.Create(context.Background(), testOrgID, testLedgerID, &CreateBalanceInput{AccountID: "acc-1", AssetCode: "BRL"})

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.Get(context.Background(), testOrgID, testLedgerID, "bal-1")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	iter := svc.List(context.Background(), testOrgID, testLedgerID, nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	sending := false
	bal, err := svc.Update(context.Background(), testOrgID, testLedgerID, "bal-1", &UpdateBalanceInput{AllowSending: &sending})

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "bal-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesGetByAliasBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.GetByAlias(context.Background(), testOrgID, testLedgerID, "my-alias")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesGetByExternalCodeBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.GetByExternalCode(context.Background(), testOrgID, testLedgerID, "EXT-BAL-1")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesGetByAccountIDBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.GetByAccountID(context.Background(), testOrgID, testLedgerID, "acc-1")

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.Equal(t, expectedErr, err)
}
