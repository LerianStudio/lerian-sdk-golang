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
// AssetRates — Create
// ---------------------------------------------------------------------------

func TestAssetRatesCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/asset-rates", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateAssetRateInput)
			require.True(t, ok)
			assert.Equal(t, "BRL", input.BaseAssetCode)
			assert.Equal(t, "USD", input.CounterAssetCode)
			assert.Equal(t, int64(520), input.Amount)
			assert.Equal(t, 2, input.Scale)

			return unmarshalInto(AssetRate{
				ID:               "rate-1",
				BaseAssetCode:    "BRL",
				CounterAssetCode: "USD",
				Amount:           520,
				Scale:            2,
			}, result)
		},
	}

	svc := newAssetRatesService(mock)
	rate, err := svc.Create(context.Background(), testOrgID, testLedgerID, &CreateAssetRateInput{
		BaseAssetCode:    "BRL",
		CounterAssetCode: "USD",
		Amount:           520,
		Scale:            2,
	})

	require.NoError(t, err)
	require.NotNil(t, rate)
	assert.Equal(t, "rate-1", rate.ID)
	assert.Equal(t, "BRL", rate.BaseAssetCode)
	assert.Equal(t, "USD", rate.CounterAssetCode)
}

func TestAssetRatesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.Create(context.Background(), testOrgID, testLedgerID, nil)

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetRatesCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.Create(context.Background(), "", testLedgerID, &CreateAssetRateInput{})

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetRatesCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.Create(context.Background(), testOrgID, "", &CreateAssetRateInput{})

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AssetRates — Get
// ---------------------------------------------------------------------------

func TestAssetRatesGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/asset-rates/rate-1", path)
			assert.Nil(t, body)

			return unmarshalInto(AssetRate{
				ID:               "rate-1",
				BaseAssetCode:    "BRL",
				CounterAssetCode: "USD",
			}, result)
		},
	}

	svc := newAssetRatesService(mock)
	rate, err := svc.Get(context.Background(), testOrgID, testLedgerID, "rate-1")

	require.NoError(t, err)
	require.NotNil(t, rate)
	assert.Equal(t, "rate-1", rate.ID)
}

func TestAssetRatesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.Get(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetRatesGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.Get(context.Background(), "", testLedgerID, "rate-1")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AssetRates — List
// ---------------------------------------------------------------------------

func TestAssetRatesList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/asset-rates")
			assert.Nil(t, body)

			resp := models.ListResponse[AssetRate]{
				Items: []AssetRate{
					{ID: "rate-1", BaseAssetCode: "BRL", CounterAssetCode: "USD"},
					{ID: "rate-2", BaseAssetCode: "EUR", CounterAssetCode: "USD"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newAssetRatesService(mock)
	iter := svc.List(context.Background(), testOrgID, testLedgerID, nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "rate-1", items[0].ID)
	assert.Equal(t, "rate-2", items[1].ID)
}

func TestAssetRatesListValidation(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})

	tests := []struct {
		name     string
		orgID    string
		ledgerID string
	}{
		{name: "empty orgID", orgID: "", ledgerID: testLedgerID},
		{name: "empty ledgerID", orgID: testOrgID, ledgerID: ""},
		{name: "both empty", orgID: "", ledgerID: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			iter := svc.List(context.Background(), tc.orgID, tc.ledgerID, nil)
			require.NotNil(t, iter)

			items, err := iter.Collect(context.Background())
			require.Error(t, err)
			assert.Nil(t, items)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
			assert.Contains(t, err.Error(), "required")
		})
	}
}

// ---------------------------------------------------------------------------
// AssetRates — Update
// ---------------------------------------------------------------------------

func TestAssetRatesUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/asset-rates/rate-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(AssetRate{
				ID:     "rate-1",
				Amount: 550,
				Scale:  2,
			}, result)
		},
	}

	svc := newAssetRatesService(mock)
	newAmount := int64(550)
	rate, err := svc.Update(context.Background(), testOrgID, testLedgerID, "rate-1", &UpdateAssetRateInput{
		Amount: &newAmount,
	})

	require.NoError(t, err)
	require.NotNil(t, rate)
	assert.Equal(t, int64(550), rate.Amount)
}

func TestAssetRatesUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.Update(context.Background(), testOrgID, testLedgerID, "", &UpdateAssetRateInput{})

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetRatesUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.Update(context.Background(), testOrgID, testLedgerID, "rate-1", nil)

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AssetRates — Delete
// ---------------------------------------------------------------------------

func TestAssetRatesDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/asset-rates/rate-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newAssetRatesService(mock)
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "rate-1")

	require.NoError(t, err)
}

func TestAssetRatesDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetRatesDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	err := svc.Delete(context.Background(), "", testLedgerID, "rate-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AssetRates — GetByExternalID
// ---------------------------------------------------------------------------

func TestAssetRatesGetByExternalID(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/asset-rates/external-id/ext-rate-1", path)
			assert.Nil(t, body)

			return unmarshalInto(AssetRate{
				ID:         "rate-1",
				ExternalID: ptr("ext-rate-1"),
			}, result)
		},
	}

	svc := newAssetRatesService(mock)
	rate, err := svc.GetByExternalID(context.Background(), testOrgID, testLedgerID, "ext-rate-1")

	require.NoError(t, err)
	require.NotNil(t, rate)
	assert.Equal(t, "rate-1", rate.ID)
	require.NotNil(t, rate.ExternalID)
	assert.Equal(t, "ext-rate-1", *rate.ExternalID)
}

func TestAssetRatesGetByExternalIDEmpty(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.GetByExternalID(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetRatesGetByExternalIDEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.GetByExternalID(context.Background(), "", testLedgerID, "ext-rate-1")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AssetRates — GetFromAssetCode
// ---------------------------------------------------------------------------

func TestAssetRatesGetFromAssetCode(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/asset-rates/from/BRL", path)
			assert.Nil(t, body)

			return unmarshalInto(AssetRate{
				ID:            "rate-1",
				BaseAssetCode: "BRL",
			}, result)
		},
	}

	svc := newAssetRatesService(mock)
	rate, err := svc.GetFromAssetCode(context.Background(), testOrgID, testLedgerID, "BRL")

	require.NoError(t, err)
	require.NotNil(t, rate)
	assert.Equal(t, "rate-1", rate.ID)
	assert.Equal(t, "BRL", rate.BaseAssetCode)
}

func TestAssetRatesGetFromAssetCodeEmpty(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.GetFromAssetCode(context.Background(), testOrgID, testLedgerID, "")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAssetRatesGetFromAssetCodeEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newAssetRatesService(&mockBackend{})
	rate, err := svc.GetFromAssetCode(context.Background(), testOrgID, "", "BRL")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// AssetRates — Backend Error Propagation
// ---------------------------------------------------------------------------

func TestAssetRatesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetRatesService(mock)
	rate, err := svc.Create(context.Background(), testOrgID, testLedgerID, &CreateAssetRateInput{BaseAssetCode: "BRL", CounterAssetCode: "USD", Amount: 520, Scale: 2})

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.Equal(t, expectedErr, err)
}

func TestAssetRatesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetRatesService(mock)
	rate, err := svc.Get(context.Background(), testOrgID, testLedgerID, "rate-1")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.Equal(t, expectedErr, err)
}

func TestAssetRatesListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetRatesService(mock)
	iter := svc.List(context.Background(), testOrgID, testLedgerID, nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestAssetRatesUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetRatesService(mock)
	newAmount := int64(550)
	rate, err := svc.Update(context.Background(), testOrgID, testLedgerID, "rate-1", &UpdateAssetRateInput{Amount: &newAmount})

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.Equal(t, expectedErr, err)
}

func TestAssetRatesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetRatesService(mock)
	err := svc.Delete(context.Background(), testOrgID, testLedgerID, "rate-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestAssetRatesGetByExternalIDBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetRatesService(mock)
	rate, err := svc.GetByExternalID(context.Background(), testOrgID, testLedgerID, "ext-rate-1")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.Equal(t, expectedErr, err)
}

func TestAssetRatesGetFromAssetCodeBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAssetRatesService(mock)
	rate, err := svc.GetFromAssetCode(context.Background(), testOrgID, testLedgerID, "BRL")

	require.Error(t, err)
	assert.Nil(t, rate)
	assert.Equal(t, expectedErr, err)
}
