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
// OperationRoutes — Create
// ---------------------------------------------------------------------------

func TestOperationRoutesCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/operation-routes", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateOperationRouteInput)
			require.True(t, ok)
			assert.Equal(t, "acc-1", input.AccountID)
			assert.Equal(t, "debit", input.Type)

			return unmarshalInto(OperationRoute{
				ID:             "or-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				AccountID:      "acc-1",
				Type:           "debit",
			}, result)
		},
	}

	svc := newOperationRoutesService(mock)
	or, err := svc.Create(context.Background(), "org-1", "led-1", &CreateOperationRouteInput{
		AccountID: "acc-1",
		Type:      "debit",
	})

	require.NoError(t, err)
	require.NotNil(t, or)
	assert.Equal(t, "or-1", or.ID)
	assert.Equal(t, "acc-1", or.AccountID)
	assert.Equal(t, "debit", or.Type)
}

func TestOperationRoutesCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Create(context.Background(), "", "led-1", &CreateOperationRouteInput{AccountID: "acc-1", Type: "debit"})

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Create(context.Background(), "org-1", "", &CreateOperationRouteInput{AccountID: "acc-1", Type: "debit"})

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Create(context.Background(), "org-1", "led-1", nil)

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: validation failed")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOperationRoutesService(mock)
	or, err := svc.Create(context.Background(), "org-1", "led-1", &CreateOperationRouteInput{AccountID: "acc-1", Type: "debit"})

	require.Error(t, err)
	assert.Nil(t, or)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// OperationRoutes — Get
// ---------------------------------------------------------------------------

func TestOperationRoutesGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/operation-routes/or-1", path)
			assert.Nil(t, body)

			return unmarshalInto(OperationRoute{
				ID:             "or-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				AccountID:      "acc-1",
				Type:           "debit",
			}, result)
		},
	}

	svc := newOperationRoutesService(mock)
	or, err := svc.Get(context.Background(), "org-1", "led-1", "or-1")

	require.NoError(t, err)
	require.NotNil(t, or)
	assert.Equal(t, "or-1", or.ID)
	assert.Equal(t, "debit", or.Type)
}

func TestOperationRoutesGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Get(context.Background(), "", "led-1", "or-1")

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Get(context.Background(), "org-1", "", "or-1")

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Get(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOperationRoutesService(mock)
	or, err := svc.Get(context.Background(), "org-1", "led-1", "or-1")

	require.Error(t, err)
	assert.Nil(t, or)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// OperationRoutes — List
// ---------------------------------------------------------------------------

func TestOperationRoutesList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/operation-routes")
			assert.Nil(t, body)

			resp := models.ListResponse[OperationRoute]{
				Items: []OperationRoute{
					{ID: "or-1", Type: "debit", AccountID: "acc-1"},
					{ID: "or-2", Type: "credit", AccountID: "acc-2"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOperationRoutesService(mock)
	iter := svc.List(context.Background(), "org-1", "led-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "or-1", items[0].ID)
	assert.Equal(t, "or-2", items[1].ID)
}

func TestOperationRoutesListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)

			receivedPath = path

			assert.Nil(t, body)

			resp := models.ListResponse[OperationRoute]{
				Items:      []OperationRoute{{ID: "or-1"}},
				Pagination: models.Pagination{Total: 1, Limit: 50},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOperationRoutesService(mock)
	opts := &models.CursorListOptions{Limit: 50, Cursor: "cursor-1", SortBy: "type"}
	iter := svc.List(context.Background(), "org-1", "led-1", opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "cursor=cursor-1")
	assert.Contains(t, receivedPath, "limit=50")
	assert.Contains(t, receivedPath, "sortBy=type")
}

// ---------------------------------------------------------------------------
// OperationRoutes — Update
// ---------------------------------------------------------------------------

func TestOperationRoutesUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/operation-routes/or-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(OperationRoute{
				ID:        "or-1",
				AccountID: "acc-2",
				Type:      "debit",
			}, result)
		},
	}

	svc := newOperationRoutesService(mock)
	newAccountID := "acc-2"
	or, err := svc.Update(context.Background(), "org-1", "led-1", "or-1", &UpdateOperationRouteInput{
		AccountID: &newAccountID,
	})

	require.NoError(t, err)
	require.NotNil(t, or)
	assert.Equal(t, "or-1", or.ID)
	assert.Equal(t, "acc-2", or.AccountID)
}

func TestOperationRoutesUpdateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Update(context.Background(), "", "led-1", "or-1", &UpdateOperationRouteInput{})

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesUpdateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Update(context.Background(), "org-1", "", "or-1", &UpdateOperationRouteInput{})

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Update(context.Background(), "org-1", "led-1", "", &UpdateOperationRouteInput{})

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	or, err := svc.Update(context.Background(), "org-1", "led-1", "or-1", nil)

	require.Error(t, err)
	assert.Nil(t, or)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOperationRoutesService(mock)
	or, err := svc.Update(context.Background(), "org-1", "led-1", "or-1", &UpdateOperationRouteInput{})

	require.Error(t, err)
	assert.Nil(t, or)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// OperationRoutes — Delete
// ---------------------------------------------------------------------------

func TestOperationRoutesDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/operation-routes/or-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newOperationRoutesService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "or-1")

	require.NoError(t, err)
}

func TestOperationRoutesDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	err := svc.Delete(context.Background(), "", "led-1", "or-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesDeleteEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "", "or-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newOperationRoutesService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOperationRoutesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOperationRoutesService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "or-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// OperationRoutes — Interface compliance
// ---------------------------------------------------------------------------

func TestOperationRoutesServiceInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ operationRoutesServiceAPI = (*operationRoutesService)(nil)

	t.Log("operationRoutesService implements operationRoutesServiceAPI")
}
