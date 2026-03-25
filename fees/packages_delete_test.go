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

func TestPackagesDelete(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	err := svc.Delete(context.Background(), "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesDeleteBackendError(t *testing.T) {
	t.Parallel()

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
