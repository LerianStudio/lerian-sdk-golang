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
// Assets — Create
// ---------------------------------------------------------------------------

func TestAssetsCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/assets", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateAssetInput)
			require.True(t, ok)
			assert.Equal(t, "Brazilian Real", input.Name)
			assert.Equal(t, "BRL", input.Code)
			assert.Equal(t, "currency", input.Type)

			return unmarshalInto(Asset{
				ID:   "ast-1",
				Name: "Brazilian Real",
				Code: "BRL",
				Type: "currency",
			}, result)
		},
	}

	svc := newAssetsService(mock)
	asset, err := svc.Create(context.Background(), testOrgID, testLedgerID, &CreateAssetInput{
		Name: "Brazilian Real",
		Code: "BRL",
		Type: "currency",
	})

	require.NoError(t, err)
	require.NotNil(t, asset)
	assert.Equal(t, "ast-1", asset.ID)
	assert.Equal(t, "BRL", asset.Code)
}

func TestAssetsCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	asset, err := svc.Create(context.Background(), testOrgID, testLedgerID, nil)

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetsCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	asset, err := svc.Create(context.Background(), "", testLedgerID, &CreateAssetInput{})

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetsCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	asset, err := svc.Create(context.Background(), testOrgID, "", &CreateAssetInput{})

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Assets — Get
// ---------------------------------------------------------------------------

func TestAssetsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/assets/ast-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Asset{
				ID:   "ast-1",
				Code: "BRL",
			}, result)
		},
	}

	svc := newAssetsService(mock)
	asset, err := svc.Get(context.Background(), testOrgID, testLedgerID, "ast-1")

	require.NoError(t, err)
	require.NotNil(t, asset)
	assert.Equal(t, "ast-1", asset.ID)
}

func TestAssetsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	asset, err := svc.Get(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetsGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	asset, err := svc.Get(context.Background(), "", testLedgerID, "ast-1")

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Assets — List
// ---------------------------------------------------------------------------

func TestAssetsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/assets")
			assert.Nil(t, body)

			resp := models.ListResponse[Asset]{
				Items: []Asset{
					{ID: "ast-1", Code: "BRL"},
					{ID: "ast-2", Code: "USD"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newAssetsService(mock)
	iter := svc.List(context.Background(), testOrgID, testLedgerID, nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "BRL", items[0].Code)
	assert.Equal(t, "USD", items[1].Code)
}

func TestAssetsListValidation(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})

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
// Assets — Update
// ---------------------------------------------------------------------------

func TestAssetsUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/assets/ast-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Asset{
				ID:   "ast-1",
				Name: "Updated Asset",
			}, result)
		},
	}

	svc := newAssetsService(mock)
	newName := "Updated Asset"
	asset, err := svc.Update(context.Background(), testOrgID, testLedgerID, "ast-1", &UpdateAssetInput{
		Name: &newName,
	})

	require.NoError(t, err)
	require.NotNil(t, asset)
	assert.Equal(t, "Updated Asset", asset.Name)
}

func TestAssetsUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	asset, err := svc.Update(context.Background(), testOrgID, testLedgerID, "", &UpdateAssetInput{})

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetsUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	asset, err := svc.Update(context.Background(), testOrgID, testLedgerID, "ast-1", nil)

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Assets — Delete
// ---------------------------------------------------------------------------

func TestAssetsDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/assets/ast-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newAssetsService(mock)
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "ast-1")

	require.NoError(t, err)
}

func TestAssetsDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetsDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAssetsService(&mockBackend{})
	err := svc.Delete(context.Background(), "", testLedgerID, "ast-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Assets — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestAssetsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetsService(mock)
	asset, err := svc.Create(context.Background(), testOrgID, testLedgerID, &CreateAssetInput{Name: "X", Code: "BRL", Type: "currency"})

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.Equal(t, expectedErr, err)
}

func TestAssetsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetsService(mock)
	asset, err := svc.Get(context.Background(), testOrgID, testLedgerID, "ast-1")

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.Equal(t, expectedErr, err)
}

func TestAssetsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetsService(mock)
	iter := svc.List(context.Background(), testOrgID, testLedgerID, nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestAssetsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetsService(mock)
	newName := "X"
	asset, err := svc.Update(context.Background(), testOrgID, testLedgerID, "ast-1", &UpdateAssetInput{Name: &newName})

	require.Error(t, err)
	assert.Nil(t, asset)
	assert.Equal(t, expectedErr, err)
}

func TestAssetsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetsService(mock)
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "ast-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
