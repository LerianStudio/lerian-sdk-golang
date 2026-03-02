package fees

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// mockBackend — configurable core.Backend stub for service tests
// ---------------------------------------------------------------------------

// mockBackend delegates Call to a user-supplied callFn so tests can
// verify HTTP method, path, and body while injecting canned responses.
type mockBackend struct {
	callFn func(ctx context.Context, method, path string, body, result any) error
}

func (m *mockBackend) Call(ctx context.Context, method, path string, body, result any) error {
	if m.callFn != nil {
		return m.callFn(ctx, method, path, body, result)
	}

	return nil
}

func (m *mockBackend) CallWithHeaders(_ context.Context, _, _ string,
	_ map[string]string, _, _ any) error {
	return nil
}

func (m *mockBackend) CallRaw(_ context.Context, _, _ string, _ any) ([]byte, error) {
	return nil, nil
}

// Compile-time check.
var _ core.Backend = (*mockBackend)(nil)

// jsonInto marshals src to JSON and unmarshals into dst. This is how the
// mock backend simulates the real backend populating result pointers.
func jsonInto(src, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("mock marshal: %w", err)
	}

	return json.Unmarshal(b, dst)
}

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

var testPackage = Package{
	ID:          "pkg-001",
	Name:        "Standard Fees",
	Description: strPtr("Default fee package"),
	Status:      "active",
	Rules: []FeeRule{
		{Type: "flat", Amount: int64Ptr(100), Currency: "USD"},
	},
	CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
}

func strPtr(s string) *string   { return &s }
func int64Ptr(i int64) *int64   { return &i }

// ---------------------------------------------------------------------------
// PackagesService.Create
// ---------------------------------------------------------------------------

func TestPackagesCreate(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/packages", path)
			assert.NotNil(t, body)

			return jsonInto(testPackage, result)
		},
	}

	svc := newPackagesService(backend)
	input := &CreatePackageInput{
		Name:  "Standard Fees",
		Rules: []FeeRule{{Type: "flat", Amount: int64Ptr(100), Currency: "USD"}},
	}

	pkg, err := svc.Create(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "pkg-001", pkg.ID)
	assert.Equal(t, "Standard Fees", pkg.Name)
	assert.Len(t, pkg.Rules, 1)
}

func TestPackagesCreateNilInput(t *testing.T) {
	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Create(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesCreateBackendError(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("network failure")
		},
	}

	svc := newPackagesService(backend)
	input := &CreatePackageInput{Name: "Test"}

	pkg, err := svc.Create(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "network failure")
}

// ---------------------------------------------------------------------------
// PackagesService.Get
// ---------------------------------------------------------------------------

func TestPackagesGet(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/packages/pkg-001", path)
			assert.Nil(t, body)

			return jsonInto(testPackage, result)
		},
	}

	svc := newPackagesService(backend)

	pkg, err := svc.Get(context.Background(), "pkg-001")
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "pkg-001", pkg.ID)
	assert.Equal(t, "active", pkg.Status)
}

func TestPackagesGetEmptyID(t *testing.T) {
	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Get(context.Background(), "")
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesGetBackendError(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("not found")
		},
	}

	svc := newPackagesService(backend)

	pkg, err := svc.Get(context.Background(), "pkg-999")
	require.Error(t, err)
	assert.Nil(t, pkg)
}

// ---------------------------------------------------------------------------
// PackagesService.List
// ---------------------------------------------------------------------------

func TestPackagesList(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/packages")
			assert.Nil(t, body)

			resp := models.ListResponse[Package]{
				Items: []Package{testPackage},
				Pagination: models.Pagination{
					Total: 1,
					Limit: 10,
				},
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)

	it := svc.List(context.Background(), &models.ListOptions{Limit: 10})
	require.NotNil(t, it)

	items, err := it.Collect(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "pkg-001", items[0].ID)
}

func TestPackagesListNilOptions(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/packages", path) // no query params

			resp := models.ListResponse[Package]{
				Items: []Package{},
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)

	it := svc.List(context.Background(), nil)
	require.NotNil(t, it)

	items, err := it.Collect(context.Background())
	require.NoError(t, err)
	assert.Empty(t, items)
}

// ---------------------------------------------------------------------------
// PackagesService.Update
// ---------------------------------------------------------------------------

func TestPackagesUpdate(t *testing.T) {
	updated := testPackage
	updated.Name = "Premium Fees"

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/packages/pkg-001", path)
			assert.NotNil(t, body)

			return jsonInto(updated, result)
		},
	}

	svc := newPackagesService(backend)
	name := "Premium Fees"
	input := &UpdatePackageInput{Name: &name}

	pkg, err := svc.Update(context.Background(), "pkg-001", input)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "Premium Fees", pkg.Name)
}

func TestPackagesUpdateEmptyID(t *testing.T) {
	svc := newPackagesService(&mockBackend{})
	name := "Test"
	input := &UpdatePackageInput{Name: &name}

	pkg, err := svc.Update(context.Background(), "", input)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesUpdateNilInput(t *testing.T) {
	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Update(context.Background(), "pkg-001", nil)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesUpdateBackendError(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("server error")
		},
	}

	svc := newPackagesService(backend)
	name := "Test"
	input := &UpdatePackageInput{Name: &name}

	pkg, err := svc.Update(context.Background(), "pkg-001", input)
	require.Error(t, err)
	assert.Nil(t, pkg)
}

// ---------------------------------------------------------------------------
// PackagesService.Delete
// ---------------------------------------------------------------------------

func TestPackagesDelete(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/packages/pkg-001", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newPackagesService(backend)

	err := svc.Delete(context.Background(), "pkg-001")
	require.NoError(t, err)
}

func TestPackagesDeleteEmptyID(t *testing.T) {
	svc := newPackagesService(&mockBackend{})

	err := svc.Delete(context.Background(), "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesDeleteBackendError(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("delete failed")
		},
	}

	svc := newPackagesService(backend)

	err := svc.Delete(context.Background(), "pkg-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ PackagesService = (*packagesService)(nil)
