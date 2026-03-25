package midaz

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testOrgID    = "org-1"
	testLedgerID = "led-1"
)

// ---------------------------------------------------------------------------
// Accounts — Create
// ---------------------------------------------------------------------------

func TestAccountsCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateAccountInput)
			require.True(t, ok)
			assert.Equal(t, "Checking Account", input.Name)
			assert.Equal(t, "deposit", input.Type)
			assert.Equal(t, "BRL", input.AssetCode)

			return unmarshalInto(Account{
				ID:   "acc-1",
				Name: "Checking Account",
			}, result)
		},
	}

	svc := newAccountsService(mock)
	acc, err := svc.Create(context.Background(), testOrgID, testLedgerID, &CreateAccountInput{
		Name:      "Checking Account",
		Type:      "deposit",
		AssetCode: "BRL",
	})

	require.NoError(t, err)
	require.NotNil(t, acc)
	assert.Equal(t, "acc-1", acc.ID)
	assert.Equal(t, "Checking Account", acc.Name)
}

func TestAccountsCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.Create(context.Background(), testOrgID, testLedgerID, nil)

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountsCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.Create(context.Background(), "", testLedgerID, &CreateAccountInput{})

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountsCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.Create(context.Background(), testOrgID, "", &CreateAccountInput{})

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Accounts — Get
// ---------------------------------------------------------------------------

func TestAccountsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/acc-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Account{
				ID:   "acc-1",
				Name: "Checking Account",
			}, result)
		},
	}

	svc := newAccountsService(mock)
	acc, err := svc.Get(context.Background(), testOrgID, testLedgerID, "acc-1")

	require.NoError(t, err)
	require.NotNil(t, acc)
	assert.Equal(t, "acc-1", acc.ID)
}

func TestAccountsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.Get(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountsGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.Get(context.Background(), "", testLedgerID, "acc-1")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountsGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.Get(context.Background(), testOrgID, "", "acc-1")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Accounts — List
// ---------------------------------------------------------------------------

func TestAccountsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/accounts")
			assert.Nil(t, body)

			resp := models.ListResponse[Account]{
				Items: []Account{
					{ID: "acc-1", Name: "Checking"},
					{ID: "acc-2", Name: "Savings"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newAccountsService(mock)
	iter := svc.List(context.Background(), testOrgID, testLedgerID, nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "acc-1", items[0].ID)
	assert.Equal(t, "acc-2", items[1].ID)
}

func TestAccountsListValidation(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})

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
// Accounts — Update
// ---------------------------------------------------------------------------

func TestAccountsUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/acc-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Account{
				ID:   "acc-1",
				Name: "Updated Account",
			}, result)
		},
	}

	svc := newAccountsService(mock)
	newName := "Updated Account"
	acc, err := svc.Update(context.Background(), testOrgID, testLedgerID, "acc-1", &UpdateAccountInput{
		Name: &newName,
	})

	require.NoError(t, err)
	require.NotNil(t, acc)
	assert.Equal(t, "Updated Account", acc.Name)
}

func TestAccountsUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.Update(context.Background(), testOrgID, testLedgerID, "", &UpdateAccountInput{})

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountsUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.Update(context.Background(), testOrgID, testLedgerID, "acc-1", nil)

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Accounts — Delete
// ---------------------------------------------------------------------------

func TestAccountsDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/acc-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newAccountsService(mock)
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "acc-1")

	require.NoError(t, err)
}

func TestAccountsDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountsDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	err := svc.Delete(context.Background(), "", testLedgerID, "acc-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Accounts — GetByAlias
// ---------------------------------------------------------------------------

func TestAccountsGetByAlias(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/alias/primary-checking", path)
			assert.Nil(t, body)

			return unmarshalInto(Account{
				ID:    "acc-1",
				Alias: ptr("primary-checking"),
			}, result)
		},
	}

	svc := newAccountsService(mock)
	acc, err := svc.GetByAlias(context.Background(), testOrgID, testLedgerID, "primary-checking")

	require.NoError(t, err)
	require.NotNil(t, acc)
	assert.Equal(t, "acc-1", acc.ID)
	require.NotNil(t, acc.Alias)
	assert.Equal(t, "primary-checking", *acc.Alias)
}

func TestAccountsGetByAliasEmptyAlias(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.GetByAlias(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountsGetByAliasEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.GetByAlias(context.Background(), "", testLedgerID, "my-alias")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Accounts — GetByExternalCode
// ---------------------------------------------------------------------------

func TestAccountsGetByExternalCode(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/external/EXT-001", path)
			assert.Nil(t, body)

			return unmarshalInto(Account{
				ID:           "acc-1",
				ExternalCode: ptr("EXT-001"),
			}, result)
		},
	}

	svc := newAccountsService(mock)
	acc, err := svc.GetByExternalCode(context.Background(), testOrgID, testLedgerID, "EXT-001")

	require.NoError(t, err)
	require.NotNil(t, acc)
	assert.Equal(t, "acc-1", acc.ID)
	require.NotNil(t, acc.ExternalCode)
	assert.Equal(t, "EXT-001", *acc.ExternalCode)
}

func TestAccountsGetByExternalCodeDoesNotFallbackOnNonCompatibilityErrors(t *testing.T) {
	t.Parallel()

	requestCount := 0
	expectedErr := &sdkerrors.Error{StatusCode: http.StatusInternalServerError, Message: "server error"}
	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			requestCount++

			assert.Equal(t, http.MethodGet, method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/external/EXT-1", path)
			assert.Nil(t, body)

			return expectedErr
		},
	}

	svc := newAccountsService(mock)
	account, err := svc.GetByExternalCode(context.Background(), testOrgID, testLedgerID, "EXT-1")
	require.Error(t, err)
	assert.Nil(t, account)
	assert.Equal(t, 1, requestCount)
	assert.ErrorIs(t, err, expectedErr)
}

func TestAccountsGetByExternalCodeEmptyCode(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.GetByExternalCode(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountsGetByExternalCodeEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAccountsService(&mockBackend{})
	acc, err := svc.GetByExternalCode(context.Background(), testOrgID, "", "EXT-001")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Accounts — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestAccountsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountsService(mock)
	acc, err := svc.Create(context.Background(), testOrgID, testLedgerID, &CreateAccountInput{Name: "X", Type: "deposit", AssetCode: "BRL"})

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.Equal(t, expectedErr, err)
}

func TestAccountsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountsService(mock)
	acc, err := svc.Get(context.Background(), testOrgID, testLedgerID, "acc-1")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.Equal(t, expectedErr, err)
}

func TestAccountsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountsService(mock)
	iter := svc.List(context.Background(), testOrgID, testLedgerID, nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestAccountsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountsService(mock)
	newName := "X"
	acc, err := svc.Update(context.Background(), testOrgID, testLedgerID, "acc-1", &UpdateAccountInput{Name: &newName})

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.Equal(t, expectedErr, err)
}

func TestAccountsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountsService(mock)
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "acc-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestAccountsGetByAliasBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountsService(mock)
	acc, err := svc.GetByAlias(context.Background(), testOrgID, testLedgerID, "my-alias")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.Equal(t, expectedErr, err)
}

func TestAccountsGetByExternalCodeBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountsService(mock)
	acc, err := svc.GetByExternalCode(context.Background(), testOrgID, testLedgerID, "EXT-001")

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.Equal(t, expectedErr, err)
}
