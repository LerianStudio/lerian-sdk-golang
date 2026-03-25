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

// ---------------------------------------------------------------------------
// Shared fixture
// ---------------------------------------------------------------------------

var testEstimateResponse = FeeEstimateResponse{
	Message: "fees calculated successfully",
	FeesApplied: &FeeCalculate{
		SegmentID: strPtr("seg-retail"),
		LedgerID:  "ledger-001",
		Transaction: TransactionDSL{
			Description: "TED transfer",
			Route:       "ted_out",
			Pending:     true,
			Send: TransactionDSLSend{
				Asset: "BRL",
				Value: "15500",
				Source: TransactionDSLSource{
					From: []TransactionDSLLeg{
						{
							AccountAlias: "sender",
							Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
						},
						{
							AccountAlias: "sender",
							Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "500"},
							Metadata:     map[string]any{"feeLabel": "ted_transfer_fee"},
						},
					},
				},
				Distribute: TransactionDSLDistribute{
					To: []TransactionDSLLeg{
						{
							AccountAlias: "recipient",
							Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
						},
						{
							AccountAlias: "platform-fee-account",
							Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "500"},
							Metadata:     map[string]any{"feeLabel": "ted_transfer_fee"},
						},
					},
				},
			},
		},
	},
}

// ---------------------------------------------------------------------------
// estimatesServiceAPI.Calculate — success
// ---------------------------------------------------------------------------

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

	// Verify that fee legs were injected into the transaction.
	assert.Len(t, resp.FeesApplied.Transaction.Send.Source.From, 2)
	assert.Len(t, resp.FeesApplied.Transaction.Send.Distribute.To, 2)
	assert.Equal(t, "platform-fee-account", resp.FeesApplied.Transaction.Send.Distribute.To[1].AccountAlias)
	assert.Equal(t, "ted_transfer_fee", resp.FeesApplied.Transaction.Send.Distribute.To[1].Metadata["feeLabel"])
}

// ---------------------------------------------------------------------------
// estimatesServiceAPI.Calculate — nil input guard clause
// ---------------------------------------------------------------------------

func TestEstimatesCalculateNilInput(t *testing.T) {
	t.Parallel()

	svc := newEstimatesService(&mockBackend{})

	resp, err := svc.Calculate(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	// Verify the error has the right operation and resource.
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
			input:   &FeeEstimateInput{PackageID: "pkg-001", LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Amount: &TransactionDSLAmount{Value: ""}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
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

// ---------------------------------------------------------------------------
// estimatesServiceAPI.Calculate — backend error propagation
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// estimatesServiceAPI.Calculate — verifies input body is forwarded
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// estimatesServiceAPI.Calculate — response with nil FeesApplied (no matching
// rules)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ estimatesServiceAPI = (*estimatesService)(nil)
