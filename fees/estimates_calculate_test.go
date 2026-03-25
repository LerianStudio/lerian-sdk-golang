package fees

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEstimatesCalculate(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/estimates", path)
			assert.NotNil(t, body)

			return jsonInto(testEstimateResponse, result)
		},
	}

	svc := newEstimatesService(backend)
	input := &FeeEstimateInput{
		PackageID:   "pkg-001",
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	resp, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "fees calculated successfully", resp.Message)
	require.NotNil(t, resp.FeesApplied)
	assert.Equal(t, "ledger-001", resp.FeesApplied.LedgerID)
	assert.Equal(t, "15500", resp.FeesApplied.Transaction.Send.Value)
	assert.Len(t, resp.FeesApplied.Transaction.Send.Source.From, 2)
	assert.Len(t, resp.FeesApplied.Transaction.Send.Distribute.To, 2)
	assert.Equal(t, "platform-fee-account", resp.FeesApplied.Transaction.Send.Distribute.To[1].AccountAlias)
	assert.Equal(t, "ted_transfer_fee", resp.FeesApplied.Transaction.Send.Distribute.To[1].Metadata["feeLabel"])
}

func TestEstimatesCalculateBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("service unavailable")
		},
	}

	svc := newEstimatesService(backend)
	input := &FeeEstimateInput{
		PackageID:   "pkg-001",
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	resp, err := svc.Calculate(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "service unavailable")
}

func TestEstimatesCalculateInputForwarded(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, body, result any) error {
			input, ok := body.(*FeeEstimateInput)
			require.True(t, ok, "body should be *FeeEstimateInput")
			assert.Equal(t, "pkg-002", input.PackageID)
			assert.Equal(t, "ledger-002", input.LedgerID)
			assert.Equal(t, "BRL", input.Transaction.Send.Asset)

			return jsonInto(testEstimateResponse, result)
		},
	}

	svc := newEstimatesService(backend)
	input := &FeeEstimateInput{
		PackageID:   "pkg-002",
		LedgerID:    "ledger-002",
		Transaction: testTransactionDSL(),
	}

	_, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
}

func TestEstimatesCalculateNoFeesApplied(t *testing.T) {
	t.Parallel()

	noFeesResp := FeeEstimateResponse{
		Message:     "no matching fee rules found",
		FeesApplied: nil,
	}

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(noFeesResp, result)
		},
	}

	svc := newEstimatesService(backend)
	input := &FeeEstimateInput{
		PackageID:   "pkg-003",
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	resp, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "no matching fee rules found", resp.Message)
	assert.Nil(t, resp.FeesApplied)
}

var _ estimatesServiceAPI = (*estimatesService)(nil)
