package fees

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
