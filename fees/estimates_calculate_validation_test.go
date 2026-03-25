package fees

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEstimatesCalculateNilInput(t *testing.T) {
	t.Parallel()

	svc := newEstimatesService(&mockBackend{})

	resp, err := svc.Calculate(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Estimates.Calculate", sdkErr.Operation)
	assert.Equal(t, "Estimate", sdkErr.Resource)
	assert.Contains(t, sdkErr.Message, "input is required")
}

func TestEstimatesCalculateRequiresIdentifiersAndTransactionShape(t *testing.T) {
	t.Parallel()

	svc := newEstimatesService(&mockBackend{})

	tests := []struct {
		name    string
		input   *FeeEstimateInput
		message string
	}{
		{
			name:    "missing package id",
			input:   &FeeEstimateInput{LedgerID: "ledger-001", Transaction: testTransactionDSL()},
			message: "package ID is required",
		},
		{
			name:    "missing ledger id",
			input:   &FeeEstimateInput{PackageID: "pkg-001", Transaction: testTransactionDSL()},
			message: "ledger ID is required",
		},
		{
			name:    "missing asset",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}}}}},
			message: "transaction send asset is required",
		},
		{
			name:    "missing source legs",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}}}}},
			message: "transaction source legs are required",
		},
		{
			name:    "missing value",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}}}}},
			message: "transaction send value is required",
		},
		{
			name:    "blank string value",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: " ", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 100}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction send value is required",
		},
		{
			name:    "missing distribute legs",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}}}}},
			message: "transaction distribute legs are required",
		},
		{
			name:    "missing leg identifier",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{Amount: &TransactionDSLAmount{Value: "100.00"}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction source leg identifier is required at index 0",
		},
		{
			name:    "missing amount and share",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction source leg amount or share is required at index 0",
		},
		{
			name:    "blank amount value",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Amount: &TransactionDSLAmount{Asset: "BRL", Value: ""}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction source leg amount value is required at index 0",
		},
		{
			name:    "zero share",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 0}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction source leg share must be greater than zero at index 0",
		},
		{
			name:    "distribute leg missing identifier",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 100}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{Amount: &TransactionDSLAmount{Value: "100.00"}}}}}}},
			message: "transaction distribute leg identifier is required at index 0",
		},
		{
			name:    "distribute leg missing amount and share",
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 100}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}}}}},
			message: "transaction distribute leg amount or share is required at index 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := svc.Calculate(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, resp)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestEstimatesCalculateAcceptsBalanceKeyAndPercentageOfPercentage(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(testEstimateResponse, result)
		},
	}

	svc := newEstimatesService(backend)
	resp, err := svc.Calculate(context.Background(), &FeeEstimateInput{
		PackageID: "pkg-001",
		LedgerID:  "ledger-001",
		Transaction: TransactionDSL{Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "100.00",
			Source: TransactionDSLSource{From: []TransactionDSLLeg{{
				BalanceKey: "available",
				Share:      &TransactionDSLShare{PercentageOfPercentage: 100},
			}}},
			Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{
				AccountAlias: "recipient",
				Share:        &TransactionDSLShare{Percentage: 100},
			}}},
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestEstimatesCalculateRejectsTypedNilSendValue(t *testing.T) {
	t.Parallel()

	svc := newEstimatesService(&mockBackend{})

	var typedNil *nilStringer

	resp, err := svc.Calculate(context.Background(), &FeeEstimateInput{
		PackageID: "pkg-001",
		LedgerID:  "ledger-001",
		Transaction: TransactionDSL{Send: TransactionDSLSend{
			Asset:      "BRL",
			Value:      typedNil,
			Source:     TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}},
			Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}},
		}},
	})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "transaction send value is required")
}
