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
// AccountTypes — Create
// ---------------------------------------------------------------------------

func TestAccountTypesCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/account-types", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateAccountTypeInput)
			require.True(t, ok)
			assert.Equal(t, "deposit", input.Name)

			return unmarshalInto(AccountType{
				ID:             "at-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				Name:           "deposit",
			}, result)
		},
	}

	svc := newAccountTypesService(mock)
	at, err := svc.Create(context.Background(), "org-1", "led-1", &CreateAccountTypeInput{
		Name: "deposit",
	})

	require.NoError(t, err)
	require.NotNil(t, at)
	assert.Equal(t, "at-1", at.ID)
	assert.Equal(t, "deposit", at.Name)
}

func TestAccountTypesCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Create(context.Background(), "", "led-1", &CreateAccountTypeInput{Name: "x"})

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Create(context.Background(), "org-1", "", &CreateAccountTypeInput{Name: "x"})

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Create(context.Background(), "org-1", "led-1", nil)

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AccountTypes — Get
// ---------------------------------------------------------------------------

func TestAccountTypesGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/account-types/at-1", path)
			assert.Nil(t, body)

			return unmarshalInto(AccountType{
				ID:             "at-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				Name:           "deposit",
			}, result)
		},
	}

	svc := newAccountTypesService(mock)
	at, err := svc.Get(context.Background(), "org-1", "led-1", "at-1")

	require.NoError(t, err)
	require.NotNil(t, at)
	assert.Equal(t, "at-1", at.ID)
	assert.Equal(t, "deposit", at.Name)
}

func TestAccountTypesGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Get(context.Background(), "", "led-1", "at-1")

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Get(context.Background(), "org-1", "", "at-1")

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Get(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AccountTypes — List
// ---------------------------------------------------------------------------

func TestAccountTypesList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/account-types")
			assert.Nil(t, body)

			resp := models.ListResponse[AccountType]{
				Items: []AccountType{
					{ID: "at-1", Name: "deposit"},
					{ID: "at-2", Name: "savings"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newAccountTypesService(mock)
	iter := svc.List(context.Background(), "org-1", "led-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "at-1", items[0].ID)
	assert.Equal(t, "at-2", items[1].ID)
}

func TestAccountTypesListValidation(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})

	tests := []struct {
		name     string
		orgID    string
		ledgerID string
	}{
		{name: "empty orgID", orgID: "", ledgerID: "led-1"},
		{name: "empty ledgerID", orgID: "org-1", ledgerID: ""},
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
// AccountTypes — Update
// ---------------------------------------------------------------------------

func TestAccountTypesUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/account-types/at-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(AccountType{
				ID:   "at-1",
				Name: "updated-type",
			}, result)
		},
	}

	svc := newAccountTypesService(mock)
	newName := "updated-type"
	at, err := svc.Update(context.Background(), "org-1", "led-1", "at-1", &UpdateAccountTypeInput{
		Name: &newName,
	})

	require.NoError(t, err)
	require.NotNil(t, at)
	assert.Equal(t, "updated-type", at.Name)
}

func TestAccountTypesUpdateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Update(context.Background(), "", "led-1", "at-1", &UpdateAccountTypeInput{})

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesUpdateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Update(context.Background(), "org-1", "", "at-1", &UpdateAccountTypeInput{})

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Update(context.Background(), "org-1", "led-1", "", &UpdateAccountTypeInput{})

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	at, err := svc.Update(context.Background(), "org-1", "led-1", "at-1", nil)

	require.Error(t, err)
	assert.Nil(t, at)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AccountTypes — Delete
// ---------------------------------------------------------------------------

func TestAccountTypesDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/account-types/at-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newAccountTypesService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "at-1")

	require.NoError(t, err)
}

func TestAccountTypesDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	err := svc.Delete(context.Background(), "", "led-1", "at-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesDeleteEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "", "at-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAccountTypesDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAccountTypesService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AccountTypes — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestAccountTypesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountTypesService(mock)
	at, err := svc.Create(context.Background(), "org-1", "led-1", &CreateAccountTypeInput{Name: "deposit"})

	require.Error(t, err)
	assert.Nil(t, at)
	assert.Equal(t, expectedErr, err)
}

func TestAccountTypesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountTypesService(mock)
	at, err := svc.Get(context.Background(), "org-1", "led-1", "at-1")

	require.Error(t, err)
	assert.Nil(t, at)
	assert.Equal(t, expectedErr, err)
}

func TestAccountTypesListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountTypesService(mock)
	iter := svc.List(context.Background(), "org-1", "led-1", nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestAccountTypesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountTypesService(mock)
	newName := "X"
	at, err := svc.Update(context.Background(), "org-1", "led-1", "at-1", &UpdateAccountTypeInput{Name: &newName})

	require.Error(t, err)
	assert.Nil(t, at)
	assert.Equal(t, expectedErr, err)
}

func TestAccountTypesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAccountTypesService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "at-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
