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
// Ledgers — Create
// ---------------------------------------------------------------------------

func TestLedgersCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateLedgerInput)
			require.True(t, ok)
			assert.Equal(t, "My Ledger", input.Name)

			return unmarshalInto(Ledger{
				ID:             "led-1",
				OrganizationID: "org-1",
				Name:           "My Ledger",
			}, result)
		},
	}

	svc := newLedgersService(mock)
	led, err := svc.Create(context.Background(), "org-1", &CreateLedgerInput{
		Name: "My Ledger",
	})

	require.NoError(t, err)
	require.NotNil(t, led)
	assert.Equal(t, "led-1", led.ID)
	assert.Equal(t, "My Ledger", led.Name)
}

func TestLedgersCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	led, err := svc.Create(context.Background(), "", &CreateLedgerInput{Name: "X"})

	require.Error(t, err)
	assert.Nil(t, led)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLedgersCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	led, err := svc.Create(context.Background(), "org-1", nil)

	require.Error(t, err)
	assert.Nil(t, led)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Ledgers — Get
// ---------------------------------------------------------------------------

func TestLedgersGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Ledger{
				ID:             "led-1",
				OrganizationID: "org-1",
				Name:           "My Ledger",
			}, result)
		},
	}

	svc := newLedgersService(mock)
	led, err := svc.Get(context.Background(), "org-1", "led-1")

	require.NoError(t, err)
	require.NotNil(t, led)
	assert.Equal(t, "led-1", led.ID)
}

func TestLedgersGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	led, err := svc.Get(context.Background(), "", "led-1")

	require.Error(t, err)
	assert.Nil(t, led)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLedgersGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	led, err := svc.Get(context.Background(), "org-1", "")

	require.Error(t, err)
	assert.Nil(t, led)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Ledgers — List
// ---------------------------------------------------------------------------

func TestLedgersList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers")
			assert.Nil(t, body)

			resp := models.ListResponse[Ledger]{
				Items: []Ledger{
					{ID: "led-1", Name: "Alpha"},
					{ID: "led-2", Name: "Beta"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newLedgersService(mock)
	iter := svc.List(context.Background(), "org-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "led-1", items[0].ID)
	assert.Equal(t, "led-2", items[1].ID)
}

func TestLedgersListValidation(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})

	// Empty orgID must produce a poisoned iterator.
	iter := svc.List(context.Background(), "", nil)
	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	assert.Contains(t, err.Error(), "required")
}

// ---------------------------------------------------------------------------
// Ledgers — Update
// ---------------------------------------------------------------------------

func TestLedgersUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Ledger{
				ID:   "led-1",
				Name: "Updated Ledger",
			}, result)
		},
	}

	svc := newLedgersService(mock)
	newName := "Updated Ledger"
	led, err := svc.Update(context.Background(), "org-1", "led-1", &UpdateLedgerInput{
		Name: &newName,
	})

	require.NoError(t, err)
	require.NotNil(t, led)
	assert.Equal(t, "Updated Ledger", led.Name)
}

func TestLedgersUpdateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	led, err := svc.Update(context.Background(), "", "led-1", &UpdateLedgerInput{})

	require.Error(t, err)
	assert.Nil(t, led)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLedgersUpdateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	led, err := svc.Update(context.Background(), "org-1", "", &UpdateLedgerInput{})

	require.Error(t, err)
	assert.Nil(t, led)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLedgersUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	led, err := svc.Update(context.Background(), "org-1", "led-1", nil)

	require.Error(t, err)
	assert.Nil(t, led)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Ledgers — Delete
// ---------------------------------------------------------------------------

func TestLedgersDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newLedgersService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1")

	require.NoError(t, err)
}

func TestLedgersDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	err := svc.Delete(context.Background(), "", "led-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestLedgersDeleteEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newLedgersService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Ledgers — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestLedgersCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLedgersService(mock)
	led, err := svc.Create(context.Background(), "org-1", &CreateLedgerInput{Name: "X"})

	require.Error(t, err)
	assert.Nil(t, led)
	assert.Equal(t, expectedErr, err)
}

func TestLedgersGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLedgersService(mock)
	led, err := svc.Get(context.Background(), "org-1", "led-1")

	require.Error(t, err)
	assert.Nil(t, led)
	assert.Equal(t, expectedErr, err)
}

func TestLedgersListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLedgersService(mock)
	iter := svc.List(context.Background(), "org-1", nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestLedgersUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLedgersService(mock)
	newName := "X"
	led, err := svc.Update(context.Background(), "org-1", "led-1", &UpdateLedgerInput{Name: &newName})

	require.Error(t, err)
	assert.Nil(t, led)
	assert.Equal(t, expectedErr, err)
}

func TestLedgersDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newLedgersService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
