package fees

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeesCalculate(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fees", path)
			assert.NotNil(t, body)

			return jsonInto(testFeeCalculateResponse, result)
		},
	}

	svc := newFeesCalcService(backend)
	input := &FeeCalculate{
		SegmentID:   strPtr("seg-retail"),
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	resp, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "ledger-001", resp.LedgerID)
	assert.NotNil(t, resp.SegmentID)
	assert.Equal(t, "seg-retail", *resp.SegmentID)
	assert.Equal(t, "TED transfer", resp.Transaction.Description)
	assert.Equal(t, "15500", resp.Transaction.Send.Value)
	assert.Len(t, resp.Transaction.Send.Source.From, 2, "should have original + fee leg")
	assert.Len(t, resp.Transaction.Send.Distribute.To, 2, "should have original + fee leg")
	assert.Equal(t, "platform-fee-account", resp.Transaction.Send.Distribute.To[1].AccountAlias)
	assert.Equal(t, "ted_transfer_fee", resp.Transaction.Send.Distribute.To[1].Metadata["feeLabel"])
}

func TestFeesCalculateBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("internal server error")
		},
	}

	svc := newFeesCalcService(backend)
	input := &FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	resp, err := svc.Calculate(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "internal server error")
}

func TestFeesCalculateInputForwarded(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, body, result any) error {
			input, ok := body.(*FeeCalculate)
			require.True(t, ok, "body should be *FeeCalculate")
			assert.Equal(t, "ledger-002", input.LedgerID)
			assert.NotNil(t, input.SegmentID)
			assert.Equal(t, "seg-premium", *input.SegmentID)
			assert.Equal(t, "BRL", input.Transaction.Send.Asset)

			return jsonInto(testFeeCalculateResponse, result)
		},
	}

	svc := newFeesCalcService(backend)
	input := &FeeCalculate{
		SegmentID:   strPtr("seg-premium"),
		LedgerID:    "ledger-002",
		Transaction: testTransactionDSL(),
	}

	_, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
}

func TestFeesCalculateWithoutSegmentID(t *testing.T) {
	t.Parallel()

	resp := FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(resp, result)
		},
	}

	svc := newFeesCalcService(backend)
	input := &FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	got, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.SegmentID)
	assert.Equal(t, "ledger-001", got.LedgerID)
}

var _ feesServiceAPI = (*feesCalcService)(nil)
