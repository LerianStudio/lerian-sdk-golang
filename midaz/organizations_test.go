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
// Organizations — Create
// ---------------------------------------------------------------------------

func TestOrganizationsCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateOrganizationInput)
			require.True(t, ok)
			assert.Equal(t, "Acme Corp", input.LegalName)

			return unmarshalInto(Organization{
				ID:        "org-1",
				LegalName: "Acme Corp",
			}, result)
		},
	}

	svc := newOrganizationsService(mock)
	org, err := svc.Create(context.Background(), &CreateOrganizationInput{
		LegalName:     "Acme Corp",
		LegalDocument: "12345678000100",
	})

	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, "org-1", org.ID)
	assert.Equal(t, "Acme Corp", org.LegalName)
}

func TestOrganizationsCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newOrganizationsService(&mockBackend{})
	org, err := svc.Create(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, org)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Organizations — Get
// ---------------------------------------------------------------------------

func TestOrganizationsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Organization{
				ID:        "org-1",
				LegalName: "Acme Corp",
			}, result)
		},
	}

	svc := newOrganizationsService(mock)
	org, err := svc.Get(context.Background(), "org-1")

	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, "org-1", org.ID)
}

func TestOrganizationsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newOrganizationsService(&mockBackend{})
	org, err := svc.Get(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, org)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Organizations — List
// ---------------------------------------------------------------------------

func TestOrganizationsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations")
			assert.Nil(t, body)

			resp := models.ListResponse[Organization]{
				Items: []Organization{
					{ID: "org-1", LegalName: "Acme"},
					{ID: "org-2", LegalName: "Beta"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newOrganizationsService(mock)
	iter := svc.List(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "org-1", items[0].ID)
	assert.Equal(t, "org-2", items[1].ID)
}

// ---------------------------------------------------------------------------
// Organizations — Update
// ---------------------------------------------------------------------------

func TestOrganizationsUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Organization{
				ID:        "org-1",
				LegalName: "Updated Corp",
			}, result)
		},
	}

	svc := newOrganizationsService(mock)
	newName := "Updated Corp"
	org, err := svc.Update(context.Background(), "org-1", &UpdateOrganizationInput{
		LegalName: &newName,
	})

	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, "Updated Corp", org.LegalName)
}

func TestOrganizationsUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newOrganizationsService(&mockBackend{})
	org, err := svc.Update(context.Background(), "", &UpdateOrganizationInput{})

	require.Error(t, err)
	assert.Nil(t, org)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestOrganizationsUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newOrganizationsService(&mockBackend{})
	org, err := svc.Update(context.Background(), "org-1", nil)

	require.Error(t, err)
	assert.Nil(t, org)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Organizations — Delete
// ---------------------------------------------------------------------------

func TestOrganizationsDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newOrganizationsService(mock)
	err := svc.Delete(context.Background(), "org-1")

	require.NoError(t, err)
}

func TestOrganizationsDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newOrganizationsService(&mockBackend{})
	err := svc.Delete(context.Background(), "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Organizations — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestOrganizationsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOrganizationsService(mock)
	org, err := svc.Create(context.Background(), &CreateOrganizationInput{LegalName: "X", LegalDocument: "123"})

	require.Error(t, err)
	assert.Nil(t, org)
	assert.Equal(t, expectedErr, err)
}

func TestOrganizationsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOrganizationsService(mock)
	org, err := svc.Get(context.Background(), "org-1")

	require.Error(t, err)
	assert.Nil(t, org)
	assert.Equal(t, expectedErr, err)
}

func TestOrganizationsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOrganizationsService(mock)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestOrganizationsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOrganizationsService(mock)
	newName := "X"
	org, err := svc.Update(context.Background(), "org-1", &UpdateOrganizationInput{LegalName: &newName})

	require.Error(t, err)
	assert.Nil(t, org)
	assert.Equal(t, expectedErr, err)
}

func TestOrganizationsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newOrganizationsService(mock)
	err := svc.Delete(context.Background(), "org-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
