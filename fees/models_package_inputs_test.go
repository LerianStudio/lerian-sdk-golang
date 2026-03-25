package fees

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
