package fees

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesListNormalizesPaginationAndCollections(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			resp := struct {
				Items []Package `json:"items"`
				Page  int       `json:"page"`
				Limit int       `json:"limit"`
				Total int       `json:"total"`
			}{
				Items: []Package{{ID: "pkg-001", FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "100.00", Fees: nil, Enable: boolPtr(true)}},
				Page:  3,
				Limit: 10,
				Total: 21,
			}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)
	resp, err := svc.List(context.Background(), &PackageListOptions{PageSize: 10})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 3, resp.PageNumber)
	assert.Equal(t, 10, resp.PageSize)
	assert.Equal(t, 21, resp.TotalItems)
	assert.Equal(t, 3, resp.TotalPages)
	require.Len(t, resp.Items, 1)
	assert.NotNil(t, resp.Items[0].Fees)
	assert.Empty(t, resp.Items[0].Fees)
}

func TestPackagesListRejectsMissingItemsPayload(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			resp := struct {
				Page  int `json:"page"`
				Limit int `json:"limit"`
				Total int `json:"total"`
			}{Page: 1, Limit: 10, Total: 1}

			return jsonInto(resp, result)
		},
	}

	svc := newPackagesService(backend)
	resp, err := svc.List(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal))
	assert.Contains(t, err.Error(), "response contained no items payload")
}

func TestPackagesListRejectsNegativePaginationValues(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	resp, err := svc.List(context.Background(), &PackageListOptions{PageNumber: -1})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "pageNumber must be greater than or equal to zero")

	resp, err = svc.List(context.Background(), &PackageListOptions{PageSize: -1})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "pageSize must be greater than or equal to zero")
}
