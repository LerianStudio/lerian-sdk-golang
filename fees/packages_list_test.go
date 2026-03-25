package fees

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesList(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/packages")
			assert.Nil(t, body)

			resp := struct {
				Items []Package `json:"items"`
				Page  int       `json:"page"`
				Limit int       `json:"limit"`
				Total int       `json:"total"`
			}{
				Items: []Package{testPackage},
				Page:  1,
				Limit: 10,
				Total: 1,
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)

	resp, err := svc.List(context.Background(), &PackageListOptions{PageSize: 10})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "pkg-001", resp.Items[0].ID)
	assert.Equal(t, 1, resp.TotalItems)
	assert.Equal(t, 10, resp.PageSize)
}

func TestPackagesListNilOptions(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/packages", path)

			resp := struct {
				Items []Package `json:"items"`
				Page  int       `json:"page"`
				Limit int       `json:"limit"`
				Total int       `json:"total"`
			}{
				Items: []Package{},
				Page:  1,
				Limit: 10,
				Total: 0,
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)

	resp, err := svc.List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Items)
}

func TestPackagesListWithFilters(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, _, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "ledgerId=ledger-001")
			assert.Contains(t, path, "segmentId=seg-retail")
			assert.Contains(t, path, "enable=true")

			resp := struct {
				Items []Package `json:"items"`
				Page  int       `json:"page"`
				Limit int       `json:"limit"`
				Total int       `json:"total"`
			}{
				Items: []Package{testPackage},
				Page:  1,
				Limit: 25,
				Total: 1,
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)

	resp, err := svc.List(context.Background(), &PackageListOptions{
		LedgerID:  "ledger-001",
		SegmentID: "seg-retail",
		Enabled:   boolPtr(true),
		PageSize:  25,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Items, 1)
}

func TestPackagesListNilBackendUsesCoreError(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(nil)

	resp, err := svc.List(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, core.ErrNilBackend)
}

func TestPackagesListRejectsIncompleteDateRange(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newPackagesService(&mockBackend{})

	resp, err := svc.List(context.Background(), &PackageListOptions{CreatedFrom: &start})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	assert.Contains(t, err.Error(), "createdFrom and createdTo must both be provided")
}

func TestPackagesListRejectsInvalidRangeAndSortOrder(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := newPackagesService(&mockBackend{})

	resp, err := svc.List(context.Background(), &PackageListOptions{CreatedFrom: &start, CreatedTo: &end})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "createdFrom must be before or equal to createdTo")

	resp, err = svc.List(context.Background(), &PackageListOptions{SortOrder: "sideways"})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "sortOrder must be either asc or desc")
}

func TestPackagesListBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("service unavailable")
		},
	}

	svc := newPackagesService(backend)

	resp, err := svc.List(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "service unavailable")
}
