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
// Portfolios — Create
// ---------------------------------------------------------------------------

func TestPortfoliosCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/portfolios", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreatePortfolioInput)
			require.True(t, ok)
			assert.Equal(t, "My Portfolio", input.Name)

			return unmarshalInto(Portfolio{
				ID:             "port-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				Name:           "My Portfolio",
			}, result)
		},
	}

	svc := newPortfoliosService(mock)
	port, err := svc.Create(context.Background(), "org-1", "led-1", &CreatePortfolioInput{
		Name: "My Portfolio",
	})

	require.NoError(t, err)
	require.NotNil(t, port)
	assert.Equal(t, "port-1", port.ID)
	assert.Equal(t, "My Portfolio", port.Name)
}

func TestPortfoliosCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Create(context.Background(), "", "led-1", &CreatePortfolioInput{Name: "X"})

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Create(context.Background(), "org-1", "", &CreatePortfolioInput{Name: "X"})

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Create(context.Background(), "org-1", "led-1", nil)

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newPortfoliosService(mock)
	port, err := svc.Create(context.Background(), "org-1", "led-1", &CreatePortfolioInput{Name: "X"})

	require.Error(t, err)
	assert.Nil(t, port)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Portfolios — Get
// ---------------------------------------------------------------------------

func TestPortfoliosGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/portfolios/port-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Portfolio{
				ID:             "port-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				Name:           "My Portfolio",
			}, result)
		},
	}

	svc := newPortfoliosService(mock)
	port, err := svc.Get(context.Background(), "org-1", "led-1", "port-1")

	require.NoError(t, err)
	require.NotNil(t, port)
	assert.Equal(t, "port-1", port.ID)
	assert.Equal(t, "My Portfolio", port.Name)
}

func TestPortfoliosGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Get(context.Background(), "", "led-1", "port-1")

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Get(context.Background(), "org-1", "", "port-1")

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Get(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newPortfoliosService(mock)
	port, err := svc.Get(context.Background(), "org-1", "led-1", "port-1")

	require.Error(t, err)
	assert.Nil(t, port)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Portfolios — List
// ---------------------------------------------------------------------------

func TestPortfoliosList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/portfolios")
			assert.Nil(t, body)

			resp := models.ListResponse[Portfolio]{
				Items: []Portfolio{
					{ID: "port-1", Name: "Alpha"},
					{ID: "port-2", Name: "Beta"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newPortfoliosService(mock)
	iter := svc.List(context.Background(), "org-1", "led-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "port-1", items[0].ID)
	assert.Equal(t, "port-2", items[1].ID)
}

func TestPortfoliosListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)

			receivedPath = path

			resp := models.ListResponse[Portfolio]{
				Items:      []Portfolio{{ID: "port-1", Name: "Found"}},
				Pagination: models.Pagination{Total: 1, Limit: 25},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newPortfoliosService(mock)
	opts := &models.ListOptions{
		Limit:     25,
		SortBy:    "name",
		SortOrder: "asc",
	}

	iter := svc.List(context.Background(), "org-1", "led-1", opts)

	ctx := context.Background()
	require.True(t, iter.Next(ctx))

	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortBy=name")
	assert.Contains(t, receivedPath, "sortOrder=asc")
}

func TestPortfoliosListEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	iter := svc.List(context.Background(), "", "led-1", nil)

	require.NotNil(t, iter)
	assert.False(t, iter.Next(context.Background()))
	require.Error(t, iter.Err())
	assert.True(t, errors.Is(iter.Err(), sdkerrors.ErrValidation))
}

func TestPortfoliosListEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	iter := svc.List(context.Background(), "org-1", "", nil)

	require.NotNil(t, iter)
	assert.False(t, iter.Next(context.Background()))
	require.Error(t, iter.Err())
	assert.True(t, errors.Is(iter.Err(), sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Portfolios — Update
// ---------------------------------------------------------------------------

func TestPortfoliosUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/portfolios/port-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Portfolio{
				ID:   "port-1",
				Name: "Updated Portfolio",
			}, result)
		},
	}

	svc := newPortfoliosService(mock)
	newName := "Updated Portfolio"
	port, err := svc.Update(context.Background(), "org-1", "led-1", "port-1", &UpdatePortfolioInput{
		Name: &newName,
	})

	require.NoError(t, err)
	require.NotNil(t, port)
	assert.Equal(t, "port-1", port.ID)
	assert.Equal(t, "Updated Portfolio", port.Name)
}

func TestPortfoliosUpdateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Update(context.Background(), "", "led-1", "port-1", &UpdatePortfolioInput{})

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosUpdateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Update(context.Background(), "org-1", "", "port-1", &UpdatePortfolioInput{})

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Update(context.Background(), "org-1", "led-1", "", &UpdatePortfolioInput{})

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	port, err := svc.Update(context.Background(), "org-1", "led-1", "port-1", nil)

	require.Error(t, err)
	assert.Nil(t, port)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newPortfoliosService(mock)
	newName := "X"
	port, err := svc.Update(context.Background(), "org-1", "led-1", "port-1", &UpdatePortfolioInput{
		Name: &newName,
	})

	require.Error(t, err)
	assert.Nil(t, port)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Portfolios — Delete
// ---------------------------------------------------------------------------

func TestPortfoliosDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/portfolios/port-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newPortfoliosService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "port-1")

	require.NoError(t, err)
}

func TestPortfoliosDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	err := svc.Delete(context.Background(), "", "led-1", "port-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosDeleteEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "", "port-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newPortfoliosService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPortfoliosDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newPortfoliosService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "port-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Portfolios — compile-time interface check
// ---------------------------------------------------------------------------

func TestPortfoliosServiceInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ PortfoliosService = (*portfoliosService)(nil)

	t.Log("portfoliosService implements PortfoliosService")
}
