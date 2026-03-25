package fees

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeesCalculateAllowsDynamicTransactionValues(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(testFeeCalculateResponse, result)
		},
	}

	svc := newFeesCalcService(backend)
	resp, err := svc.Calculate(context.Background(), &FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: TransactionDSL{Send: TransactionDSLSend{
			Asset:      "BRL",
			Value:      map[string]any{"$gt": "0"},
			Source:     TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 100}}}},
			Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}},
		}},
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestFeesCalculateRequiresLegAmountAssetAndMirroredDistributeValidation(t *testing.T) {
	t.Parallel()

	svc := newFeesCalcService(&mockBackend{})

	resp, err := svc.Calculate(context.Background(), &FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: TransactionDSL{Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "100.00",
			Source: TransactionDSLSource{From: []TransactionDSLLeg{{
				AccountAlias: "sender",
				Amount:       &TransactionDSLAmount{Value: "100.00"},
			}}},
			Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{
				AccountAlias: "recipient",
				Share:        &TransactionDSLShare{Percentage: 100},
			}}},
		}},
	})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "transaction source leg amount asset is required at index 0")

	var typedNil *nilStringer

	resp, err = svc.Calculate(context.Background(), &FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: TransactionDSL{Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "100.00",
			Source: TransactionDSLSource{From: []TransactionDSLLeg{{
				AccountAlias: "sender",
				Share:        &TransactionDSLShare{Percentage: 100},
			}}},
			Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{
				AccountAlias: "recipient",
				Amount:       &TransactionDSLAmount{Asset: "BRL", Value: typedNil},
			}}},
		}},
	})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "transaction distribute leg amount value is required at index 0")

	resp, err = svc.Calculate(context.Background(), &FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: TransactionDSL{Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "100.00",
			Source: TransactionDSLSource{From: []TransactionDSLLeg{{
				AccountAlias: "sender",
				Share:        &TransactionDSLShare{Percentage: 100},
			}}},
			Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{
				AccountAlias: "recipient",
				Share:        &TransactionDSLShare{Percentage: 0},
			}}},
		}},
	})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "transaction distribute leg share must be greater than zero at index 0")
}
