package fees

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				From: []TransactionDSLLeg{{
					AccountAlias:    "sender",
					BalanceKey:      "available",
					Amount:          &TransactionDSLAmount{Asset: "BRL", Operation: "debit", TransactionType: "amount", Value: "10000"},
					Route:           "internal",
					Description:     "debit leg",
					ChartOfAccounts: "chart-001",
					Metadata:        map[string]any{"legType": "debit"},
				}},
			},
			Distribute: TransactionDSLDistribute{
				Remaining: "remaining_dist",
				To: []TransactionDSLLeg{{
					AccountAlias: "recipient",
					Share:        &TransactionDSLShare{Percentage: 100},
					Description:  "credit leg",
				}},
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
