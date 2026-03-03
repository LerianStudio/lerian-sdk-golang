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
// TransactionRoutes — Create
// ---------------------------------------------------------------------------

func TestTransactionRoutesCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transaction-routes", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateTransactionRouteInput)
			require.True(t, ok)
			assert.Equal(t, "PIX", input.TransactionType)

			return unmarshalInto(TransactionRoute{
				ID:              "tr-1",
				OrganizationID:  "org-1",
				LedgerID:        "led-1",
				TransactionType: "PIX",
			}, result)
		},
	}

	svc := newTransactionRoutesService(mock)
	tr, err := svc.Create(context.Background(), "org-1", "led-1", &CreateTransactionRouteInput{
		TransactionType: "PIX",
	})

	require.NoError(t, err)
	require.NotNil(t, tr)
	assert.Equal(t, "tr-1", tr.ID)
	assert.Equal(t, "PIX", tr.TransactionType)
}

func TestTransactionRoutesCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Create(context.Background(), "", "led-1", &CreateTransactionRouteInput{TransactionType: "PIX"})

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Create(context.Background(), "org-1", "", &CreateTransactionRouteInput{TransactionType: "PIX"})

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Create(context.Background(), "org-1", "led-1", nil)

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: validation failed")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTransactionRoutesService(mock)
	tr, err := svc.Create(context.Background(), "org-1", "led-1", &CreateTransactionRouteInput{TransactionType: "PIX"})

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// TransactionRoutes — Get
// ---------------------------------------------------------------------------

func TestTransactionRoutesGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transaction-routes/tr-1", path)
			assert.Nil(t, body)

			return unmarshalInto(TransactionRoute{
				ID:              "tr-1",
				OrganizationID:  "org-1",
				LedgerID:        "led-1",
				TransactionType: "PIX",
			}, result)
		},
	}

	svc := newTransactionRoutesService(mock)
	tr, err := svc.Get(context.Background(), "org-1", "led-1", "tr-1")

	require.NoError(t, err)
	require.NotNil(t, tr)
	assert.Equal(t, "tr-1", tr.ID)
	assert.Equal(t, "PIX", tr.TransactionType)
}

func TestTransactionRoutesGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Get(context.Background(), "", "led-1", "tr-1")

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Get(context.Background(), "org-1", "", "tr-1")

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Get(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// TransactionRoutes — List
// ---------------------------------------------------------------------------

func TestTransactionRoutesList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/transaction-routes")
			assert.Nil(t, body)

			resp := models.ListResponse[TransactionRoute]{
				Items: []TransactionRoute{
					{ID: "tr-1", TransactionType: "PIX"},
					{ID: "tr-2", TransactionType: "TED"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newTransactionRoutesService(mock)
	iter := svc.List(context.Background(), "org-1", "led-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "tr-1", items[0].ID)
	assert.Equal(t, "tr-2", items[1].ID)
}

func TestTransactionRoutesListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)

			receivedPath = path

			assert.Nil(t, body)

			resp := models.ListResponse[TransactionRoute]{
				Items:      []TransactionRoute{{ID: "tr-1"}},
				Pagination: models.Pagination{Total: 1, Limit: 50},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newTransactionRoutesService(mock)
	opts := &models.ListOptions{Limit: 50, SortBy: "transactionType"}
	iter := svc.List(context.Background(), "org-1", "led-1", opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=50")
	assert.Contains(t, receivedPath, "sortBy=transactionType")
}

// ---------------------------------------------------------------------------
// TransactionRoutes — Update
// ---------------------------------------------------------------------------

func TestTransactionRoutesUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transaction-routes/tr-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(TransactionRoute{
				ID:              "tr-1",
				TransactionType: "PIX",
			}, result)
		},
	}

	svc := newTransactionRoutesService(mock)
	desc := "updated description"
	tr, err := svc.Update(context.Background(), "org-1", "led-1", "tr-1", &UpdateTransactionRouteInput{
		Description: &desc,
	})

	require.NoError(t, err)
	require.NotNil(t, tr)
	assert.Equal(t, "tr-1", tr.ID)
}

func TestTransactionRoutesUpdateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Update(context.Background(), "", "led-1", "tr-1", &UpdateTransactionRouteInput{})

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesUpdateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Update(context.Background(), "org-1", "", "tr-1", &UpdateTransactionRouteInput{})

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Update(context.Background(), "org-1", "led-1", "", &UpdateTransactionRouteInput{})

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	tr, err := svc.Update(context.Background(), "org-1", "led-1", "tr-1", nil)

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTransactionRoutesService(mock)
	tr, err := svc.Update(context.Background(), "org-1", "led-1", "tr-1", &UpdateTransactionRouteInput{})

	require.Error(t, err)
	assert.Nil(t, tr)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// TransactionRoutes — Delete
// ---------------------------------------------------------------------------

func TestTransactionRoutesDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transaction-routes/tr-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newTransactionRoutesService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "tr-1")

	require.NoError(t, err)
}

func TestTransactionRoutesDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	err := svc.Delete(context.Background(), "", "led-1", "tr-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesDeleteEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "", "tr-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTransactionRoutesService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionRoutesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTransactionRoutesService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "tr-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// TransactionRoutes — Interface compliance
// ---------------------------------------------------------------------------

func TestTransactionRoutesServiceInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ TransactionRoutesService = (*transactionRoutesService)(nil)

	t.Log("transactionRoutesService implements TransactionRoutesService")
}
