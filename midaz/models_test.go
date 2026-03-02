package midaz

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr returns a pointer to the given value — handy for optional fields in tests.
func ptr[T any](v T) *T { return &v }

// ---------------------------------------------------------------------------
// Organization JSON round-trip
// ---------------------------------------------------------------------------

func TestOrganizationJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		org  Organization
		keys []string // expected camelCase keys in JSON output
	}{
		{
			name: "full organization",
			org: Organization{
				ID:                   "org-001",
				LegalName:            "Acme Corp",
				LegalDocument:        "12345678000100",
				ParentOrganizationID: ptr("org-parent"),
				Status:               models.Status{Code: "ACTIVE"},
				Address: &models.Address{
					Line1:   "123 Main St",
					ZipCode: "10001",
					City:    "New York",
					State:   "NY",
					Country: "US",
				},
				Metadata:  models.Metadata{"env": "production"},
				CreatedAt: now,
				UpdatedAt: now,
				DeletedAt: nil,
			},
			keys: []string{"id", "legalName", "legalDocument", "parentOrganizationId",
				"status", "address", "metadata", "createdAt", "updatedAt"},
		},
		{
			name: "minimal organization",
			org: Organization{
				ID:            "org-002",
				LegalName:     "Small Co",
				LegalDocument: "99999999",
				Status:        models.Status{Code: "ACTIVE"},
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			keys: []string{"id", "legalName", "legalDocument", "status",
				"createdAt", "updatedAt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.org)
			require.NoError(t, err)

			// Verify expected keys exist.
			raw := make(map[string]json.RawMessage)
			require.NoError(t, json.Unmarshal(data, &raw))
			for _, key := range tc.keys {
				assert.Contains(t, raw, key, "expected key %q in JSON", key)
			}

			// Round-trip back.
			var decoded Organization
			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, tc.org.ID, decoded.ID)
			assert.Equal(t, tc.org.LegalName, decoded.LegalName)
			assert.Equal(t, tc.org.Status.Code, decoded.Status.Code)

			if tc.org.ParentOrganizationID != nil {
				require.NotNil(t, decoded.ParentOrganizationID)
				assert.Equal(t, *tc.org.ParentOrganizationID, *decoded.ParentOrganizationID)
			} else {
				assert.Nil(t, decoded.ParentOrganizationID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Transaction JSON round-trip
// ---------------------------------------------------------------------------

func TestTransactionJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		tx   Transaction
		keys []string
	}{
		{
			name: "full transaction",
			tx: Transaction{
				ID:                       "txn-001",
				OrganizationID:           "org-001",
				LedgerID:                 "led-001",
				Description:              ptr("Monthly salary payment"),
				Status:                   models.Status{Code: "COMMITTED"},
				Amount:                   1500000,
				AmountScale:              2,
				AssetCode:                "BRL",
				ChartOfAccountsGroupName: ptr("payroll"),
				ParentTransactionID:      ptr("txn-parent"),
				Metadata:                 models.Metadata{"batch": "2026-03"},
				CreatedAt:                now,
				UpdatedAt:                now,
				DeletedAt:                nil,
			},
			keys: []string{"id", "organizationId", "ledgerId", "description",
				"status", "amount", "amountScale", "assetCode",
				"chartOfAccountsGroupName", "parentTransactionId",
				"metadata", "createdAt", "updatedAt"},
		},
		{
			name: "minimal transaction",
			tx: Transaction{
				ID:             "txn-002",
				OrganizationID: "org-001",
				LedgerID:       "led-001",
				Status:         models.Status{Code: "PENDING"},
				Amount:         50000,
				AmountScale:    2,
				AssetCode:      "USD",
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			keys: []string{"id", "organizationId", "ledgerId", "status",
				"amount", "amountScale", "assetCode", "createdAt", "updatedAt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.tx)
			require.NoError(t, err)

			raw := make(map[string]json.RawMessage)
			require.NoError(t, json.Unmarshal(data, &raw))
			for _, key := range tc.keys {
				assert.Contains(t, raw, key, "expected key %q in JSON", key)
			}

			var decoded Transaction
			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, tc.tx.ID, decoded.ID)
			assert.Equal(t, tc.tx.Amount, decoded.Amount)
			assert.Equal(t, tc.tx.AmountScale, decoded.AmountScale)
			assert.Equal(t, tc.tx.AssetCode, decoded.AssetCode)

			if tc.tx.Description != nil {
				require.NotNil(t, decoded.Description)
				assert.Equal(t, *tc.tx.Description, *decoded.Description)
			} else {
				assert.Nil(t, decoded.Description)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TransactionSource JSON round-trip
// ---------------------------------------------------------------------------

func TestTransactionSourceJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		source TransactionSource
		keys   []string
	}{
		{
			name: "with amount",
			source: TransactionSource{
				From: TransactionFromTo{
					Account:     "acc-sender",
					AccountType: ptr("deposit"),
				},
				To: TransactionFromTo{
					Account:         "acc-receiver",
					ChartOfAccounts: ptr("revenue"),
				},
				Amount: &TransactionAmount{
					Amount: 100000,
					Scale:  2,
					Asset:  "BRL",
				},
			},
			keys: []string{"from", "to", "amount"},
		},
		{
			name: "with share",
			source: TransactionSource{
				From: TransactionFromTo{
					Account: "acc-source",
				},
				To: TransactionFromTo{
					Account: "acc-dest",
				},
				Share: &TransactionShare{
					Percentage:   50,
					PercentageOf: ptr("remaining"),
				},
			},
			keys: []string{"from", "to", "share"},
		},
		{
			name: "minimal source",
			source: TransactionSource{
				From: TransactionFromTo{Account: "acc-a"},
				To:   TransactionFromTo{Account: "acc-b"},
			},
			keys: []string{"from", "to"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.source)
			require.NoError(t, err)

			raw := make(map[string]json.RawMessage)
			require.NoError(t, json.Unmarshal(data, &raw))
			for _, key := range tc.keys {
				assert.Contains(t, raw, key, "expected key %q in JSON", key)
			}

			var decoded TransactionSource
			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, tc.source.From.Account, decoded.From.Account)
			assert.Equal(t, tc.source.To.Account, decoded.To.Account)

			if tc.source.Amount != nil {
				require.NotNil(t, decoded.Amount)
				assert.Equal(t, tc.source.Amount.Amount, decoded.Amount.Amount)
				assert.Equal(t, tc.source.Amount.Scale, decoded.Amount.Scale)
				assert.Equal(t, tc.source.Amount.Asset, decoded.Amount.Asset)
			} else {
				assert.Nil(t, decoded.Amount)
			}

			if tc.source.Share != nil {
				require.NotNil(t, decoded.Share)
				assert.Equal(t, tc.source.Share.Percentage, decoded.Share.Percentage)
			} else {
				assert.Nil(t, decoded.Share)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Balance JSON round-trip
// ---------------------------------------------------------------------------

func TestBalanceJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	bal := Balance{
		ID:             "bal-001",
		OrganizationID: "org-001",
		LedgerID:       "led-001",
		AccountID:      "acc-001",
		AssetCode:      "BRL",
		Available:      5000000,
		OnHold:         100000,
		Scale:          2,
		AccountAlias:   ptr("primary-checking"),
		AllowSending:   true,
		AllowReceiving: true,
		Status:         models.Status{Code: "ACTIVE"},
		Metadata:       models.Metadata{"tier": "gold"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(bal)
	require.NoError(t, err)

	var decoded Balance
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, bal.ID, decoded.ID)
	assert.Equal(t, bal.Available, decoded.Available)
	assert.Equal(t, bal.OnHold, decoded.OnHold)
	assert.Equal(t, bal.Scale, decoded.Scale)
	assert.Equal(t, bal.AllowSending, decoded.AllowSending)
	assert.Equal(t, bal.AllowReceiving, decoded.AllowReceiving)
	require.NotNil(t, decoded.AccountAlias)
	assert.Equal(t, *bal.AccountAlias, *decoded.AccountAlias)
}

// ---------------------------------------------------------------------------
// Operation JSON round-trip
// ---------------------------------------------------------------------------

func TestOperationJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 3, 1, 15, 0, 0, 0, time.UTC)

	op := Operation{
		ID:             "op-001",
		OrganizationID: "org-001",
		LedgerID:       "led-001",
		TransactionID:  "txn-001",
		AccountID:      "acc-001",
		AccountAlias:   ptr("savings"),
		Type:           "DEBIT",
		AssetCode:      "USD",
		Amount:         250000,
		AmountScale:    2,
		Status:         models.Status{Code: "COMPLETED"},
		BalanceAfter:   ptr(int64(4750000)),
		Description:    ptr("Withdrawal"),
		Metadata:       models.Metadata{"ref": "WD-123"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(op)
	require.NoError(t, err)

	var decoded Operation
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, op.ID, decoded.ID)
	assert.Equal(t, op.Type, decoded.Type)
	assert.Equal(t, op.Amount, decoded.Amount)
	require.NotNil(t, decoded.BalanceAfter)
	assert.Equal(t, *op.BalanceAfter, *decoded.BalanceAfter)
}

// ---------------------------------------------------------------------------
// Input types — Create omits zero optional fields, Update omitempty
// ---------------------------------------------------------------------------

func TestCreateOrganizationInputJSON(t *testing.T) {
	input := CreateOrganizationInput{
		LegalName:     "NewCo",
		LegalDocument: "111222333",
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "legalName")
	assert.Contains(t, raw, "legalDocument")
	assert.NotContains(t, raw, "parentOrganizationId",
		"nil optional fields must be omitted")
	assert.NotContains(t, raw, "address")
}

func TestUpdateOrganizationInputOmitsNilFields(t *testing.T) {
	// Only update legalName, leave everything else nil.
	input := UpdateOrganizationInput{
		LegalName: ptr("Updated Name"),
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "legalName")
	assert.NotContains(t, raw, "legalDocument",
		"nil pointer fields must be omitted via omitempty")
	assert.NotContains(t, raw, "status")
	assert.NotContains(t, raw, "address")
}

// ---------------------------------------------------------------------------
// Ledger, Asset, Portfolio — spot-check JSON tags
// ---------------------------------------------------------------------------

func TestLedgerJSONKeys(t *testing.T) {
	now := time.Now().UTC()
	l := Ledger{
		ID:             "led-001",
		OrganizationID: "org-001",
		Name:           "Primary Ledger",
		Status:         models.Status{Code: "ACTIVE"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(l)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	for _, key := range []string{"id", "organizationId", "name", "status", "createdAt", "updatedAt"} {
		assert.Contains(t, raw, key)
	}
}

func TestAssetJSONKeys(t *testing.T) {
	now := time.Now().UTC()
	a := Asset{
		ID:             "ast-001",
		OrganizationID: "org-001",
		LedgerID:       "led-001",
		Name:           "Brazilian Real",
		Code:           "BRL",
		Type:           "currency",
		Status:         models.Status{Code: "ACTIVE"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(a)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	for _, key := range []string{"id", "organizationId", "ledgerId", "name", "code", "type", "status"} {
		assert.Contains(t, raw, key)
	}
}

func TestPortfolioJSONRoundTrip(t *testing.T) {
	now := time.Now().UTC()
	p := Portfolio{
		ID:             "pf-001",
		OrganizationID: "org-001",
		LedgerID:       "led-001",
		Name:           "Investment Portfolio",
		EntityID:       ptr("entity-001"),
		Status:         models.Status{Code: "ACTIVE"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var decoded Portfolio
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, p.ID, decoded.ID)
	assert.Equal(t, p.Name, decoded.Name)
	require.NotNil(t, decoded.EntityID)
	assert.Equal(t, *p.EntityID, *decoded.EntityID)
}
