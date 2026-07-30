package fees

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeesCalculateRejectsMalformedResponse(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(struct{}{}, result)
		},
	}

	svc := newFeesCalcService(backend)
	resp, err := svc.Calculate(context.Background(), &FeeCalculate{LedgerID: "ledger-001", Transaction: testTransactionDSL()})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal))
	assert.Contains(t, err.Error(), "response contained no ledger ID")
}

func TestEstimatesCalculateRejectsMalformedResponse(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(FeeEstimateResponse{Message: "fees calculated successfully"}, result)
		},
	}

	svc := newEstimatesService(backend)
	resp, err := svc.Calculate(context.Background(), &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: testTransactionDSL()})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal))
	assert.Contains(t, err.Error(), "response contained no fees payload")
}
