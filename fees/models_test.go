package fees

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Package — JSON round-trip
// ---------------------------------------------------------------------------

func TestPackageJSONRoundTrip(t *testing.T) {
	t.Parallel()

	desc := "Standard retail fees"
	segID := "seg-retail"
	route := "ted_out"
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	waived := []string{"acct-waived-001", "acct-waived-002"}

	original := Package{
		ID:               "pkg-001",
		FeeGroupLabel:    "standard_fees",
		Description:      &desc,
		SegmentID:        &segID,
		LedgerID:         "ledger-001",
		TransactionRoute: &route,
		MinimumAmount:    "0.00",
		MaximumAmount:    "1000000.00",
		WaivedAccounts:   &waived,
		Fees: map[string]Fee{
			"ted_fee": {
				FeeLabel: "TED Transfer Fee",
				CalculationModel: &CalculationModel{
					ApplicationRule: "flatFee",
					Calculations: []Calculation{
						{Type: "flat", Value: "5.00"},
					},
				},
				ReferenceAmount:  "originalAmount",
				Priority:         1,
				IsDeductibleFrom: boolPtr(false),
				CreditAccount:    "platform-fee-account",
			},
			"pct_fee": {
				FeeLabel: "Percentage Fee",
				CalculationModel: &CalculationModel{
					ApplicationRule: "percentual",
					Calculations: []Calculation{
						{Type: "percentage", Value: "2.5"},
					},
				},
				ReferenceAmount:  "originalAmount",
				Priority:         2,
				IsDeductibleFrom: boolPtr(true),
				CreditAccount:    "platform-pct-account",
				RouteFrom:        strPtr("route_a"),
				RouteTo:          strPtr("route_b"),
			},
		},
		Enable:    boolPtr(true),
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Package

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.FeeGroupLabel, decoded.FeeGroupLabel)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, *original.SegmentID, *decoded.SegmentID)
	assert.Equal(t, original.LedgerID, decoded.LedgerID)
	assert.Equal(t, *original.TransactionRoute, *decoded.TransactionRoute)
	assert.Equal(t, original.MinimumAmount, decoded.MinimumAmount)
	assert.Equal(t, original.MaximumAmount, decoded.MaximumAmount)
	require.NotNil(t, decoded.WaivedAccounts)
	assert.Len(t, *decoded.WaivedAccounts, 2)
	assert.Len(t, decoded.Fees, 2)

	// Verify flat fee.
	tedFee, ok := decoded.Fees["ted_fee"]
	require.True(t, ok)
	assert.Equal(t, "TED Transfer Fee", tedFee.FeeLabel)
	require.NotNil(t, tedFee.CalculationModel)
	assert.Equal(t, "flatFee", tedFee.CalculationModel.ApplicationRule)
	assert.Len(t, tedFee.CalculationModel.Calculations, 1)
	assert.Equal(t, "flat", tedFee.CalculationModel.Calculations[0].Type)
	assert.Equal(t, "5.00", tedFee.CalculationModel.Calculations[0].Value)
	assert.Equal(t, "platform-fee-account", tedFee.CreditAccount)
	assert.False(t, *tedFee.IsDeductibleFrom)

	// Verify percentage fee.
	pctFee, ok := decoded.Fees["pct_fee"]
	require.True(t, ok)
	assert.Equal(t, "Percentage Fee", pctFee.FeeLabel)
	assert.Equal(t, "percentual", pctFee.CalculationModel.ApplicationRule)
	assert.True(t, *pctFee.IsDeductibleFrom)
	assert.Equal(t, "route_a", *pctFee.RouteFrom)
	assert.Equal(t, "route_b", *pctFee.RouteTo)

	assert.True(t, *decoded.Enable)
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt))
	assert.True(t, original.UpdatedAt.Equal(decoded.UpdatedAt))
}

func TestPackageJSONOmitsNilFields(t *testing.T) {
	t.Parallel()

	pkg := Package{
		ID:            "pkg-002",
		FeeGroupLabel: "minimal",
		LedgerID:      "ledger-001",
		MinimumAmount: "0",
		MaximumAmount: "100",
		Fees:          map[string]Fee{},
		Enable:        boolPtr(false),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	data, err := json.Marshal(pkg)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "description")
	assert.NotContains(t, raw, "segmentId")
	assert.NotContains(t, raw, "transactionRoute")
	assert.NotContains(t, raw, "waivedAccounts")
	assert.NotContains(t, raw, "deletedAt")
}

// ---------------------------------------------------------------------------
// Fee — JSON round-trip
// ---------------------------------------------------------------------------

func TestFeeJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := Fee{
		FeeLabel: "Transfer Fee",
		CalculationModel: &CalculationModel{
			ApplicationRule: "maxBetweenTypes",
			Calculations: []Calculation{
				{Type: "flat", Value: "3.00"},
				{Type: "percentage", Value: "1.5"},
			},
		},
		ReferenceAmount:  "originalAmount",
		Priority:         1,
		IsDeductibleFrom: boolPtr(true),
		CreditAccount:    "fee-collector",
		RouteFrom:        strPtr("route_in"),
		RouteTo:          strPtr("route_out"),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Fee

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.FeeLabel, decoded.FeeLabel)
	require.NotNil(t, decoded.CalculationModel)
	assert.Equal(t, "maxBetweenTypes", decoded.CalculationModel.ApplicationRule)
	assert.Len(t, decoded.CalculationModel.Calculations, 2)
	assert.Equal(t, original.ReferenceAmount, decoded.ReferenceAmount)
	assert.Equal(t, 1, decoded.Priority)
	assert.True(t, *decoded.IsDeductibleFrom)
	assert.Equal(t, "fee-collector", decoded.CreditAccount)
	assert.Equal(t, "route_in", *decoded.RouteFrom)
	assert.Equal(t, "route_out", *decoded.RouteTo)
}

func TestFeeJSONOmitsNilOptionalFields(t *testing.T) {
	t.Parallel()

	fee := Fee{
		FeeLabel:        "Simple Fee",
		ReferenceAmount: "originalAmount",
		CreditAccount:   "fee-acct",
	}

	data, err := json.Marshal(fee)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "routeFrom")
	assert.NotContains(t, raw, "routeTo")
}

// ---------------------------------------------------------------------------
// CalculationModel — JSON round-trip
// ---------------------------------------------------------------------------

func TestCalculationModelJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := CalculationModel{
		ApplicationRule: "maxBetweenTypes",
		Calculations: []Calculation{
			{Type: "flat", Value: "10.00"},
			{Type: "percentage", Value: "3.0"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CalculationModel

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ApplicationRule, decoded.ApplicationRule)
	require.Len(t, decoded.Calculations, 2)
	assert.Equal(t, "flat", decoded.Calculations[0].Type)
	assert.Equal(t, "10.00", decoded.Calculations[0].Value)
	assert.Equal(t, "percentage", decoded.Calculations[1].Type)
	assert.Equal(t, "3.0", decoded.Calculations[1].Value)
}

// ---------------------------------------------------------------------------
// Calculation — JSON round-trip
// ---------------------------------------------------------------------------

func TestCalculationJSONRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		calc Calculation
	}{
		{"flat fee", Calculation{Type: "flat", Value: "250.00"}},
		{"percentage", Calculation{Type: "percentage", Value: "2.5"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.calc)
			require.NoError(t, err)

			var decoded Calculation

			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.calc.Type, decoded.Type)
			assert.Equal(t, tt.calc.Value, decoded.Value)
		})
	}
}

// ---------------------------------------------------------------------------
// CreatePackageInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestCreatePackageInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	desc := "A test package"

	original := CreatePackageInput{
		FeeGroupLabel: "test_fees",
		Description:   &desc,
		LedgerID:      "ledger-001",
		MinimumAmount: "0",
		MaximumAmount: "500000",
		Fees: map[string]Fee{
			"flat_fee": {
				FeeLabel: "Flat Fee",
				CalculationModel: &CalculationModel{
					ApplicationRule: "flatFee",
					Calculations:    []Calculation{{Type: "flat", Value: "1.00"}},
				},
				ReferenceAmount: "originalAmount",
				CreditAccount:   "fee-account",
			},
		},
		Enable: boolPtr(true),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CreatePackageInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.FeeGroupLabel, decoded.FeeGroupLabel)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, original.LedgerID, decoded.LedgerID)
	assert.Equal(t, original.MinimumAmount, decoded.MinimumAmount)
	assert.Equal(t, original.MaximumAmount, decoded.MaximumAmount)
	assert.Contains(t, decoded.Fees, "flat_fee")
	assert.True(t, *decoded.Enable)
}

func TestCreatePackageInputJSONOmitsNilFields(t *testing.T) {
	t.Parallel()

	input := CreatePackageInput{
		FeeGroupLabel: "minimal",
		LedgerID:      "ledger-001",
		MinimumAmount: "0",
		MaximumAmount: "100",
		Fees:          map[string]Fee{},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "description")
	assert.NotContains(t, raw, "segmentId")
	assert.NotContains(t, raw, "transactionRoute")
	assert.NotContains(t, raw, "waivedAccounts")
}

// ---------------------------------------------------------------------------
// UpdatePackageInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestUpdatePackageInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	minAmt := "100.00"
	maxAmt := "999999.00"

	original := UpdatePackageInput{
		FeeGroupLabel: "updated_fees",
		Description:   "Updated description",
		MinimumAmount: &minAmt,
		MaximumAmount: &maxAmt,
		Enable:        boolPtr(false),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded UpdatePackageInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "updated_fees", decoded.FeeGroupLabel)
	assert.Equal(t, "Updated description", decoded.Description)
	require.NotNil(t, decoded.MinimumAmount)
	assert.Equal(t, "100.00", *decoded.MinimumAmount)
	require.NotNil(t, decoded.MaximumAmount)
	assert.Equal(t, "999999.00", *decoded.MaximumAmount)
	require.NotNil(t, decoded.Enable)
	assert.False(t, *decoded.Enable)
}

func TestUpdatePackageInputJSONOmitsEmptyFields(t *testing.T) {
	t.Parallel()

	input := UpdatePackageInput{}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "feeGroupLabel")
	assert.NotContains(t, raw, "description")
	assert.NotContains(t, raw, "minimumAmount")
	assert.NotContains(t, raw, "maximumAmount")
	assert.NotContains(t, raw, "waivedAccounts")
	assert.NotContains(t, raw, "fees")
	assert.NotContains(t, raw, "enable")
}

// ---------------------------------------------------------------------------
// FeeCalculate — JSON round-trip
// ---------------------------------------------------------------------------

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
					From: []TransactionDSLLeg{
						{AccountAlias: "sender", Amount: &TransactionDSLAmount{Asset: "BRL", Value: "15000"}},
					},
				},
				Distribute: TransactionDSLDistribute{
					To: []TransactionDSLLeg{
						{AccountAlias: "recipient", Amount: &TransactionDSLAmount{Asset: "BRL", Value: "15000"}},
					},
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

// ---------------------------------------------------------------------------
// FeeEstimateInput — JSON round-trip
// ---------------------------------------------------------------------------

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
					From: []TransactionDSLLeg{
						{AccountAlias: "sender", Amount: &TransactionDSLAmount{Asset: "BRL", Value: "50000"}},
					},
				},
				Distribute: TransactionDSLDistribute{
					To: []TransactionDSLLeg{
						{AccountAlias: "recipient", Amount: &TransactionDSLAmount{Asset: "BRL", Value: "50000"}},
					},
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

// ---------------------------------------------------------------------------
// FeeEstimateResponse — JSON round-trip
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// PackagePage — JSON round-trip
// ---------------------------------------------------------------------------

func TestPackagePageJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	original := PackagePage{
		Items: []Package{
			{
				ID:            "pkg-001",
				FeeGroupLabel: "standard_fees",
				LedgerID:      "ledger-001",
				MinimumAmount: "0",
				MaximumAmount: "1000000",
				Fees:          map[string]Fee{},
				Enable:        boolPtr(true),
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		},
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

// ---------------------------------------------------------------------------
// TransactionDSL — JSON round-trip
// ---------------------------------------------------------------------------

func TestTransactionDSLJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := TransactionDSL{
		ChartOfAccountsGroupName: "default",
		Description:              "Full DSL test",
		Code:                     "TXN-001",
		Pending:                  true,
		Route:                    "ted_out",
		Metadata: map[string]any{
			"transferType": "TED_OUT",
		},
		Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "10000",
			Source: TransactionDSLSource{
				Remaining: "remaining_source",
				From: []TransactionDSLLeg{
					{
						AccountAlias:    "sender",
						BalanceKey:      "available",
						Amount:          &TransactionDSLAmount{Asset: "BRL", Operation: "debit", TransactionType: "amount", Value: "10000"},
						Route:           "internal",
						Description:     "debit leg",
						ChartOfAccounts: "chart-001",
						Metadata:        map[string]any{"legType": "debit"},
					},
				},
			},
			Distribute: TransactionDSLDistribute{
				Remaining: "remaining_dist",
				To: []TransactionDSLLeg{
					{
						AccountAlias: "recipient",
						Share:        &TransactionDSLShare{Percentage: 100},
						Description:  "credit leg",
					},
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded TransactionDSL

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "default", decoded.ChartOfAccountsGroupName)
	assert.Equal(t, "Full DSL test", decoded.Description)
	assert.Equal(t, "TXN-001", decoded.Code)
	assert.True(t, decoded.Pending)
	assert.Equal(t, "ted_out", decoded.Route)
	assert.Equal(t, "TED_OUT", decoded.Metadata["transferType"])
	assert.Equal(t, "BRL", decoded.Send.Asset)
	assert.Equal(t, "remaining_source", decoded.Send.Source.Remaining)
	require.Len(t, decoded.Send.Source.From, 1)
	assert.Equal(t, "sender", decoded.Send.Source.From[0].AccountAlias)
	assert.Equal(t, "available", decoded.Send.Source.From[0].BalanceKey)
	require.NotNil(t, decoded.Send.Source.From[0].Amount)
	assert.Equal(t, "debit", decoded.Send.Source.From[0].Amount.Operation)
	assert.Equal(t, "amount", decoded.Send.Source.From[0].Amount.TransactionType)
	assert.Equal(t, "remaining_dist", decoded.Send.Distribute.Remaining)
	require.Len(t, decoded.Send.Distribute.To, 1)
	assert.Equal(t, "recipient", decoded.Send.Distribute.To[0].AccountAlias)
	require.NotNil(t, decoded.Send.Distribute.To[0].Share)
	assert.Equal(t, int64(100), decoded.Send.Distribute.To[0].Share.Percentage)
}
