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

	_, err := client.Transactions.Balances.Get(context.Background(), testOrgID, testLedgerID, "bal-1")
	require.NoError(t, err)

	assert.False(t, onboardingCalled, "Balances must NOT use the onboarding backend")
	assert.True(t, transactionCalled, "Balances must use the transaction backend")
}

// ---------------------------------------------------------------------------
// Balances — Create
// ---------------------------------------------------------------------------

func TestBalancesCreateForAccount(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/acc-1/balances", path)
			assert.NotNil(t, body)

			input, ok := body.(*createAdditionalBalanceRequest)
			require.True(t, ok)
			assert.Equal(t, "asset-freeze", input.Key)
			require.NotNil(t, input.AllowSending)
			require.NotNil(t, input.AllowReceiving)
			assert.True(t, *input.AllowSending)
			assert.True(t, *input.AllowReceiving)

			return unmarshalInto(Balance{
				ID:        "bal-1",
				AccountID: "acc-1",
				AssetCode: "BRL",
			}, result)
		},
	}

	svc := newBalancesService(mock)
	allowSending := true
	allowReceiving := true
	bal, err := svc.CreateForAccount(context.Background(), testOrgID, testLedgerID, "acc-1", &CreateBalanceInput{
		Key:            "asset-freeze",
		AllowSending:   &allowSending,
		AllowReceiving: &allowReceiving,
	})

	require.NoError(t, err)
	require.NotNil(t, bal)
	assert.Equal(t, "bal-1", bal.ID)
	assert.Equal(t, "acc-1", bal.AccountID)
}

func TestBalancesCreateForAccountNilInput(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.CreateForAccount(context.Background(), testOrgID, testLedgerID, "acc-1", nil)

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesCreateForAccountOmitsPermissionFlagsWhenUnset(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/acc-1/balances", path)

			input, ok := body.(*createAdditionalBalanceRequest)
			require.True(t, ok)
			assert.Equal(t, "reserve", input.Key)
			assert.Nil(t, input.AllowSending)
			assert.Nil(t, input.AllowReceiving)

			return unmarshalInto(Balance{ID: "bal-optional"}, result)
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.CreateForAccount(context.Background(), testOrgID, testLedgerID, "acc-1", &CreateBalanceInput{Key: "reserve"})

	require.NoError(t, err)
	require.NotNil(t, bal)
	assert.Equal(t, "bal-optional", bal.ID)
}

func TestBalancesCreateForAccountEmptyAccountID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.CreateForAccount(context.Background(), testOrgID, testLedgerID, "", &CreateBalanceInput{Key: "asset-freeze"})

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesCreateForAccountMissingKey(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.CreateForAccount(context.Background(), testOrgID, testLedgerID, "acc-1", &CreateBalanceInput{})

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	assert.Contains(t, err.Error(), "key is required")
}

func TestBalancesCreateForAccountEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.CreateForAccount(context.Background(), "", testLedgerID, "acc-1", &CreateBalanceInput{Key: "asset-freeze"})

	require.Error(t, err)
	assert.Nil(t, bal)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesCreateForAccountEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	bal, err := svc.CreateForAccount(context.Background(), testOrgID, "", "acc-1", &CreateBalanceInput{Key: "asset-freeze"})

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
// Balances — ListByAlias
// ---------------------------------------------------------------------------

func TestBalancesListByAlias(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, path string, _, result any) error {
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/alias/primary-checking/balances", path)
			return unmarshalInto(balancesLookupResponse{Items: []Balance{{ID: "bal-1"}, {ID: "bal-2"}}}, result)
		},
	}

	svc := newBalancesService(mock)
	items, err := svc.ListByAlias(context.Background(), testOrgID, testLedgerID, "primary-checking")

	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "bal-1", items[0].ID)
	assert.Equal(t, "bal-2", items[1].ID)
}

func TestBalancesListByAliasNilItemsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, path string, _, result any) error {
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/alias/primary-checking/balances", path)
			return unmarshalInto(balancesLookupResponse{}, result)
		},
	}

	svc := newBalancesService(mock)
	items, err := svc.ListByAlias(context.Background(), testOrgID, testLedgerID, "primary-checking")

	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestBalancesListByAliasEmptyAlias(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	items, err := svc.ListByAlias(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesListByAliasEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	items, err := svc.ListByAlias(context.Background(), "", testLedgerID, "my-alias")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — ListByExternalCode
// ---------------------------------------------------------------------------

func TestBalancesListByExternalCode(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, path string, _, result any) error {
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/external/EXT-BAL-1/balances", path)
			return unmarshalInto(balancesLookupResponse{Items: []Balance{{ID: "bal-1"}, {ID: "bal-2"}}}, result)
		},
	}

	svc := newBalancesService(mock)
	items, err := svc.ListByExternalCode(context.Background(), testOrgID, testLedgerID, "EXT-BAL-1")

	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "bal-1", items[0].ID)
	assert.Equal(t, "bal-2", items[1].ID)
}

func TestBalancesListByExternalCodeEmpty(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	items, err := svc.ListByExternalCode(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesListByExternalCodeEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	items, err := svc.ListByExternalCode(context.Background(), testOrgID, "", "EXT-BAL-1")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — ListByAccountID
// ---------------------------------------------------------------------------

func TestBalancesListByAccountID(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, path string, _, result any) error {
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/acc-1/balances", path)
			return unmarshalInto(balancesLookupResponse{Items: []Balance{{ID: "bal-1", AccountID: "acc-1"}, {ID: "bal-2", AccountID: "acc-1"}}}, result)
		},
	}

	svc := newBalancesService(mock)
	items, err := svc.ListByAccountID(context.Background(), testOrgID, testLedgerID, "acc-1")

	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "bal-1", items[0].ID)
	assert.Equal(t, "bal-2", items[1].ID)
}

func TestBalancesListByAccountIDEmpty(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	items, err := svc.ListByAccountID(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesListByAccountIDEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	items, err := svc.ListByAccountID(context.Background(), "", testLedgerID, "acc-1")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBalancesListByAccountIDEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newBalancesService(&mockBackend{})
	items, err := svc.ListByAccountID(context.Background(), testOrgID, "", "acc-1")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Balances — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestBalancesCreateForAccountBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	bal, err := svc.CreateForAccount(context.Background(), testOrgID, testLedgerID, "acc-1", &CreateBalanceInput{Key: "asset-freeze"})

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

func TestBalancesListByAliasBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	items, err := svc.ListByAlias(context.Background(), testOrgID, testLedgerID, "my-alias")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesListByExternalCodeBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	items, err := svc.ListByExternalCode(context.Background(), testOrgID, testLedgerID, "EXT-BAL-1")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestBalancesListByAccountIDBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newBalancesService(mock)
	items, err := svc.ListByAccountID(context.Background(), testOrgID, testLedgerID, "acc-1")

	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}
