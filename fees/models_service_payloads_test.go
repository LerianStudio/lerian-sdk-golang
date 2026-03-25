package fees

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeeCalculateJSONRoundTrip(t *testing.T) {
	t.Parallel()

	segID := "seg-retail"

	original := FeeCalculate{
		SegmentID: &segID,
		LedgerID:  "ledger-001",
		Transaction: TransactionDSL{
			Description: "TED transfer",
			Route:       "ted_out",
			Pending:     true,
			Send: TransactionDSLSend{
				Asset: "BRL",
				Value: "15000",
				Source: TransactionDSLSource{
					From: []TransactionDSLLeg{{
						AccountAlias: "sender",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					}},
				},
				Distribute: TransactionDSLDistribute{
					To: []TransactionDSLLeg{{
						AccountAlias: "recipient",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					}},
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FeeCalculate

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, *original.SegmentID, *decoded.SegmentID)
	assert.Equal(t, original.LedgerID, decoded.LedgerID)
	assert.Equal(t, "TED transfer", decoded.Transaction.Description)
	assert.Equal(t, "BRL", decoded.Transaction.Send.Asset)
	assert.Len(t, decoded.Transaction.Send.Source.From, 1)
	assert.Len(t, decoded.Transaction.Send.Distribute.To, 1)
}

func TestFeeCalculateJSONOmitsNilSegmentID(t *testing.T) {
	t.Parallel()

	input := FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: TransactionDSL{
			Send: TransactionDSLSend{Asset: "BRL", Value: "100"},
		},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "segmentId")
}

func TestFeeEstimateInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := FeeEstimateInput{
		PackageID: "pkg-001",
		LedgerID:  "ledger-001",
		Transaction: TransactionDSL{
			Description: "Estimate preview",
			Send: TransactionDSLSend{
				Asset: "BRL",
				Value: "50000",
				Source: TransactionDSLSource{
					From: []TransactionDSLLeg{{
						AccountAlias: "sender",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "50000"},
					}},
				},
				Distribute: TransactionDSLDistribute{
					To: []TransactionDSLLeg{{
						AccountAlias: "recipient",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "50000"},
					}},
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FeeEstimateInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.PackageID, decoded.PackageID)
	assert.Equal(t, original.LedgerID, decoded.LedgerID)
	assert.Equal(t, "Estimate preview", decoded.Transaction.Description)
	assert.Equal(t, "BRL", decoded.Transaction.Send.Asset)
}

func TestFeeEstimateResponseJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := FeeEstimateResponse{
		Message: "fees calculated successfully",
		FeesApplied: &FeeCalculate{
			LedgerID: "ledger-001",
			Transaction: TransactionDSL{
				Description: "estimate result",
				Send: TransactionDSLSend{
					Asset: "BRL",
					Value: "15500",
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FeeEstimateResponse

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Message, decoded.Message)
	require.NotNil(t, decoded.FeesApplied)
	assert.Equal(t, "ledger-001", decoded.FeesApplied.LedgerID)
	assert.Equal(t, "15500", decoded.FeesApplied.Transaction.Send.Value)
}

func TestFeeEstimateResponseJSONNilFeesApplied(t *testing.T) {
	t.Parallel()

	original := FeeEstimateResponse{
		Message:     "no matching fee rules found",
		FeesApplied: nil,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FeeEstimateResponse

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "no matching fee rules found", decoded.Message)
	assert.Nil(t, decoded.FeesApplied)
}

func TestPackagePageJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	original := PackagePage{
		Items: []Package{{
			ID:            "pkg-001",
			FeeGroupLabel: "standard_fees",
			LedgerID:      "ledger-001",
			MinimumAmount: "0",
			MaximumAmount: "1000000",
			Fees:          map[string]Fee{},
			Enable:        boolPtr(true),
			CreatedAt:     now,
			UpdatedAt:     now,
		}},
		PageNumber: 1,
		PageSize:   10,
		TotalItems: 1,
		TotalPages: 1,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded PackagePage

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Items, 1)
	assert.Equal(t, "pkg-001", decoded.Items[0].ID)
	assert.Equal(t, 1, decoded.PageNumber)
	assert.Equal(t, 10, decoded.PageSize)
	assert.Equal(t, 1, decoded.TotalItems)
}
