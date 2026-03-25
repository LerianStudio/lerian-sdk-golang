package fees

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesGet(t *testing.T) {
	t.Parallel()

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
	assert.Equal(t, "ledger-001", pkg.LedgerID)
	assert.True(t, *pkg.Enable)
}

func TestPackagesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Get(context.Background(), "")
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesGetBackendError(t *testing.T) {
	t.Parallel()

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
