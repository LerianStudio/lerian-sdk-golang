package fees

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Package — JSON round-trip with FeeRule
// ---------------------------------------------------------------------------

func TestPackageJSONRoundTrip(t *testing.T) {
	t.Parallel()

	desc := "Standard retail fees"
	flatAmount := int64(500)
	pct := "2.5"
	minAmt := int64(100)
	maxAmt := int64(50000)
	assetCode := "BRL"
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	original := Package{
		ID:          "pkg-001",
		Name:        "Retail Package",
		Description: &desc,
		Status:      "active",
		Rules: []FeeRule{
			{
				Type:     "flat",
				Amount:   &flatAmount,
				Currency: "BRL",
			},
			{
				Type:       "percentage",
				Percentage: &pct,
				MinAmount:  &minAmt,
				MaxAmount:  &maxAmt,
				Currency:   "USD",
				AssetCode:  &assetCode,
			},
		},
		Metadata: map[string]any{
			"tier": "standard",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Package

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Len(t, decoded.Rules, 2)

	// Verify first rule (flat).
	assert.Equal(t, "flat", decoded.Rules[0].Type)
	assert.Equal(t, int64(500), *decoded.Rules[0].Amount)
	assert.Nil(t, decoded.Rules[0].Percentage)
	assert.Equal(t, "BRL", decoded.Rules[0].Currency)

	// Verify second rule (percentage).
	assert.Equal(t, "percentage", decoded.Rules[1].Type)
	assert.Equal(t, "2.5", *decoded.Rules[1].Percentage)
	assert.Equal(t, int64(100), *decoded.Rules[1].MinAmount)
	assert.Equal(t, int64(50000), *decoded.Rules[1].MaxAmount)
	assert.Equal(t, "BRL", *decoded.Rules[1].AssetCode)

	assert.Equal(t, "standard", decoded.Metadata["tier"])
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt))
	assert.True(t, original.UpdatedAt.Equal(decoded.UpdatedAt))
}

func TestPackageJSONOmitsNilFields(t *testing.T) {
	t.Parallel()

	pkg := Package{
		ID:     "pkg-002",
		Name:   "Minimal",
		Status: "draft",
		Rules:  []FeeRule{},
	}

	data, err := json.Marshal(pkg)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "description")
	assert.NotContains(t, raw, "metadata")
}

// ---------------------------------------------------------------------------
// Estimate — JSON round-trip
// ---------------------------------------------------------------------------

func TestEstimateJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC)

	original := Estimate{
		ID:        "est-001",
		PackageID: "pkg-001",
		Amount:    100000,
		Scale:     2,
		Currency:  "BRL",
		FeeResults: []FeeResult{
			{
				RuleType: "flat",
				Amount:   500,
				Scale:    2,
				Currency: "BRL",
				Applied:  true,
			},
			{
				RuleType: "percentage",
				Amount:   2500,
				Scale:    2,
				Currency: "BRL",
				Applied:  false,
				Reason:   "below minimum threshold",
			},
		},
		TotalFee:      3000,
		TotalFeeScale: 2,
		CreatedAt:     now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Estimate

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.PackageID, decoded.PackageID)
	assert.Equal(t, original.Amount, decoded.Amount)
	assert.Equal(t, original.Scale, decoded.Scale)
	assert.Equal(t, original.Currency, decoded.Currency)
	assert.Len(t, decoded.FeeResults, 2)
	assert.Equal(t, original.TotalFee, decoded.TotalFee)
	assert.Equal(t, original.TotalFeeScale, decoded.TotalFeeScale)
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt))

	// Verify FeeResult fields.
	assert.True(t, decoded.FeeResults[0].Applied)
	assert.Empty(t, decoded.FeeResults[0].Reason)
	assert.False(t, decoded.FeeResults[1].Applied)
	assert.Equal(t, "below minimum threshold", decoded.FeeResults[1].Reason)
}

// ---------------------------------------------------------------------------
// Fee — JSON round-trip
// ---------------------------------------------------------------------------

func TestFeeJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 1, 15, 0, 0, 0, time.UTC)
	txnID := "txn-abc-123"

	original := Fee{
		ID:            "fee-001",
		PackageID:     "pkg-001",
		TransactionID: &txnID,
		Amount:        50000,
		Scale:         2,
		Currency:      "USD",
		FeeResults: []FeeResult{
			{
				RuleType: "tiered",
				Amount:   750,
				Scale:    2,
				Currency: "USD",
				Applied:  true,
			},
		},
		TotalFee:      750,
		TotalFeeScale: 2,
		Status:        "calculated",
		CreatedAt:     now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Fee

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.PackageID, decoded.PackageID)
	require.NotNil(t, decoded.TransactionID)
	assert.Equal(t, txnID, *decoded.TransactionID)
	assert.Equal(t, original.Amount, decoded.Amount)
	assert.Equal(t, original.Scale, decoded.Scale)
	assert.Equal(t, original.Currency, decoded.Currency)
	assert.Len(t, decoded.FeeResults, 1)
	assert.Equal(t, original.TotalFee, decoded.TotalFee)
	assert.Equal(t, original.TotalFeeScale, decoded.TotalFeeScale)
	assert.Equal(t, "calculated", decoded.Status)
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt))
}

func TestFeeJSONOmitsNilTransactionID(t *testing.T) {
	t.Parallel()

	fee := Fee{
		ID:        "fee-002",
		PackageID: "pkg-001",
		Amount:    1000,
		Currency:  "BRL",
		Status:    "pending",
	}

	data, err := json.Marshal(fee)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "transactionId")
}

// ---------------------------------------------------------------------------
// FeeResult — JSON round-trip
// ---------------------------------------------------------------------------

func TestFeeResultJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := FeeResult{
		RuleType: "flat",
		Amount:   250,
		Scale:    2,
		Currency: "BRL",
		Applied:  true,
		Reason:   "standard flat fee",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FeeResult

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.RuleType, decoded.RuleType)
	assert.Equal(t, original.Amount, decoded.Amount)
	assert.Equal(t, original.Scale, decoded.Scale)
	assert.Equal(t, original.Currency, decoded.Currency)
	assert.Equal(t, original.Applied, decoded.Applied)
	assert.Equal(t, original.Reason, decoded.Reason)
}

func TestFeeResultJSONOmitsEmptyReason(t *testing.T) {
	t.Parallel()

	result := FeeResult{
		RuleType: "percentage",
		Amount:   100,
		Scale:    2,
		Currency: "USD",
		Applied:  true,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "reason")
}

// ---------------------------------------------------------------------------
// CalculateEstimateInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestCalculateEstimateInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := CalculateEstimateInput{
		PackageID: "pkg-001",
		Amount:    100000,
		Scale:     2,
		Currency:  "BRL",
		AssetCode: "BRL",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CalculateEstimateInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.PackageID, decoded.PackageID)
	assert.Equal(t, original.Amount, decoded.Amount)
	assert.Equal(t, original.Scale, decoded.Scale)
	assert.Equal(t, original.Currency, decoded.Currency)
	assert.Equal(t, original.AssetCode, decoded.AssetCode)
}

// ---------------------------------------------------------------------------
// CalculateFeeInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestCalculateFeeInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	txnID := "txn-xyz-789"

	original := CalculateFeeInput{
		PackageID:     "pkg-001",
		TransactionID: &txnID,
		Amount:        50000,
		Scale:         2,
		Currency:      "USD",
		AssetCode:     "USD",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CalculateFeeInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.PackageID, decoded.PackageID)
	require.NotNil(t, decoded.TransactionID)
	assert.Equal(t, txnID, *decoded.TransactionID)
	assert.Equal(t, original.Amount, decoded.Amount)
	assert.Equal(t, original.Scale, decoded.Scale)
	assert.Equal(t, original.Currency, decoded.Currency)
	assert.Equal(t, original.AssetCode, decoded.AssetCode)
}

func TestCalculateFeeInputJSONOmitsNilTransactionID(t *testing.T) {
	t.Parallel()

	input := CalculateFeeInput{
		PackageID: "pkg-001",
		Amount:    1000,
		Currency:  "BRL",
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "transactionId")
}

// ---------------------------------------------------------------------------
// CreatePackageInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestCreatePackageInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	desc := "A test package"
	flatAmount := int64(100)

	original := CreatePackageInput{
		Name:        "Test Package",
		Description: &desc,
		Rules: []FeeRule{
			{
				Type:     "flat",
				Amount:   &flatAmount,
				Currency: "BRL",
			},
		},
		Metadata: map[string]any{
			"env": "test",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CreatePackageInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Len(t, decoded.Rules, 1)
	assert.Equal(t, "test", decoded.Metadata["env"])
}

// ---------------------------------------------------------------------------
// UpdatePackageInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestUpdatePackageInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	name := "Updated Name"
	desc := "Updated description"

	original := UpdatePackageInput{
		Name:        &name,
		Description: &desc,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded UpdatePackageInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	require.NotNil(t, decoded.Name)
	assert.Equal(t, name, *decoded.Name)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, desc, *decoded.Description)
	assert.Nil(t, decoded.Rules)
	assert.Nil(t, decoded.Metadata)
}

func TestUpdatePackageInputJSONOmitsNilFields(t *testing.T) {
	t.Parallel()

	input := UpdatePackageInput{}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "name")
	assert.NotContains(t, raw, "description")
	assert.NotContains(t, raw, "rules")
	assert.NotContains(t, raw, "metadata")
}
