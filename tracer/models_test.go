package tracer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// RuleStatus constants
// ---------------------------------------------------------------------------

func TestRuleStatusConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, RuleStatus("DRAFT"), RuleStatusDraft)
	assert.Equal(t, RuleStatus("ACTIVE"), RuleStatusActive)
	assert.Equal(t, RuleStatus("INACTIVE"), RuleStatusInactive)
}

func TestRuleStatusJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status RuleStatus
		want   string
	}{
		{"draft", RuleStatusDraft, `"DRAFT"`},
		{"active", RuleStatusActive, `"ACTIVE"`},
		{"inactive", RuleStatusInactive, `"INACTIVE"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.status)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))

			var got RuleStatus

			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.status, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Rule round-trip
// ---------------------------------------------------------------------------

func TestRuleJSONRoundTrip(t *testing.T) {
	t.Parallel()

	desc := "AML screening rule"
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	rule := Rule{
		ID:          "rule-001",
		Name:        "AML Check",
		Description: &desc,
		Status:      RuleStatusActive,
		Priority:    10,
		Conditions: []RuleCondition{
			{Field: "amount", Operator: "gt", Value: float64(10000)},
			{Field: "currency", Operator: "eq", Value: "USD"},
		},
		Actions:   []string{"flag", "notify"},
		Metadata:  map[string]any{"region": "US"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(rule)
	require.NoError(t, err)

	var got Rule

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, rule.ID, got.ID)
	assert.Equal(t, rule.Name, got.Name)
	require.NotNil(t, got.Description)
	assert.Equal(t, *rule.Description, *got.Description)
	assert.Equal(t, rule.Status, got.Status)
	assert.Equal(t, rule.Priority, got.Priority)
	assert.Len(t, got.Conditions, 2)
	assert.Equal(t, "amount", got.Conditions[0].Field)
	assert.Equal(t, "gt", got.Conditions[0].Operator)
	assert.Equal(t, float64(10000), got.Conditions[0].Value)
	assert.Equal(t, "currency", got.Conditions[1].Field)
	assert.Equal(t, rule.Actions, got.Actions)
	assert.Equal(t, "US", got.Metadata["region"])
	assert.True(t, rule.CreatedAt.Equal(got.CreatedAt))
	assert.True(t, rule.UpdatedAt.Equal(got.UpdatedAt))
}

func TestRuleOmitsEmptyOptionalFields(t *testing.T) {
	t.Parallel()

	rule := Rule{
		ID:     "rule-002",
		Name:   "Minimal Rule",
		Status: RuleStatusDraft,
		Conditions: []RuleCondition{
			{Field: "type", Operator: "eq", Value: "wire"},
		},
	}

	data, err := json.Marshal(rule)
	require.NoError(t, err)

	// description, actions, metadata should be absent
	assert.NotContains(t, string(data), `"description"`)
	assert.NotContains(t, string(data), `"actions"`)
	assert.NotContains(t, string(data), `"metadata"`)
}

// ---------------------------------------------------------------------------
// Limit round-trip
// ---------------------------------------------------------------------------

func TestLimitJSONRoundTrip(t *testing.T) {
	t.Parallel()

	desc := "Daily wire transfer cap"
	currency := "USD"
	now := time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC)

	limit := Limit{
		ID:          "lim-001",
		Name:        "Daily Wire Cap",
		Description: &desc,
		Status:      RuleStatusActive,
		Type:        "amount",
		MaxAmount:   50000,
		Period:      "daily",
		Currency:    &currency,
		Metadata:    map[string]any{"tier": "premium"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(limit)
	require.NoError(t, err)

	var got Limit

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, limit.ID, got.ID)
	assert.Equal(t, limit.Name, got.Name)
	require.NotNil(t, got.Description)
	assert.Equal(t, *limit.Description, *got.Description)
	assert.Equal(t, limit.Status, got.Status)
	assert.Equal(t, limit.Type, got.Type)
	assert.Equal(t, limit.MaxAmount, got.MaxAmount)
	assert.Equal(t, limit.Period, got.Period)
	require.NotNil(t, got.Currency)
	assert.Equal(t, "USD", *got.Currency)
	assert.Equal(t, "premium", got.Metadata["tier"])
	assert.True(t, limit.CreatedAt.Equal(got.CreatedAt))
}

func TestLimitOmitsEmptyOptionalFields(t *testing.T) {
	t.Parallel()

	limit := Limit{
		ID:        "lim-002",
		Name:      "Bare Limit",
		Status:    RuleStatusDraft,
		Type:      "count",
		MaxAmount: 100,
		Period:    "monthly",
	}

	data, err := json.Marshal(limit)
	require.NoError(t, err)

	assert.NotContains(t, string(data), `"description"`)
	assert.NotContains(t, string(data), `"currency"`)
	assert.NotContains(t, string(data), `"metadata"`)
}

// ---------------------------------------------------------------------------
// Validation round-trip
// ---------------------------------------------------------------------------

func TestValidationJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC)

	val := Validation{
		ID:           "val-001",
		Status:       "completed",
		Result:       "approved",
		RulesApplied: []string{"rule-001", "rule-002"},
		Violations:   []string{},
		Transaction:  map[string]any{"amount": float64(5000), "currency": "BRL"},
		Metadata:     map[string]any{"source": "api"},
		CreatedAt:    now,
	}

	data, err := json.Marshal(val)
	require.NoError(t, err)

	var got Validation

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, val.ID, got.ID)
	assert.Equal(t, val.Status, got.Status)
	assert.Equal(t, val.Result, got.Result)
	assert.Equal(t, val.RulesApplied, got.RulesApplied)
	assert.Equal(t, float64(5000), got.Transaction["amount"])
	assert.Equal(t, "api", got.Metadata["source"])
	assert.True(t, val.CreatedAt.Equal(got.CreatedAt))
}

func TestValidationOmitsEmptyOptionalFields(t *testing.T) {
	t.Parallel()

	val := Validation{
		ID:        "val-002",
		Status:    "pending",
		Result:    "pending",
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(val)
	require.NoError(t, err)

	assert.NotContains(t, string(data), `"rulesApplied"`)
	assert.NotContains(t, string(data), `"violations"`)
	assert.NotContains(t, string(data), `"transaction"`)
	assert.NotContains(t, string(data), `"metadata"`)
}

// ---------------------------------------------------------------------------
// AuditEvent round-trip
// ---------------------------------------------------------------------------

func TestAuditEventJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 2, 10, 15, 0, 0, time.UTC)

	event := AuditEvent{
		ID:         "evt-001",
		Type:       "transaction",
		Action:     "create",
		Actor:      "user-42",
		Resource:   "Transaction",
		ResourceID: "txn-abc-123",
		Details:    map[string]any{"amount": float64(1500), "status": "committed"},
		Timestamp:  now,
		CreatedAt:  now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var got AuditEvent

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, event.ID, got.ID)
	assert.Equal(t, event.Type, got.Type)
	assert.Equal(t, event.Action, got.Action)
	assert.Equal(t, event.Actor, got.Actor)
	assert.Equal(t, event.Resource, got.Resource)
	assert.Equal(t, event.ResourceID, got.ResourceID)
	assert.Equal(t, float64(1500), got.Details["amount"])
	assert.Equal(t, "committed", got.Details["status"])
	assert.True(t, event.Timestamp.Equal(got.Timestamp))
	assert.True(t, event.CreatedAt.Equal(got.CreatedAt))
}

func TestAuditEventOmitsEmptyDetails(t *testing.T) {
	t.Parallel()

	event := AuditEvent{
		ID:         "evt-002",
		Type:       "account",
		Action:     "delete",
		Actor:      "admin-1",
		Resource:   "Account",
		ResourceID: "acc-xyz",
		Timestamp:  time.Now(),
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	assert.NotContains(t, string(data), `"details"`)
}

// ---------------------------------------------------------------------------
// AuditVerification round-trip
// ---------------------------------------------------------------------------

func TestAuditVerificationJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 2, 11, 0, 0, 0, time.UTC)

	verification := AuditVerification{
		ID:         "ver-001",
		EventID:    "evt-001",
		Valid:      true,
		Hash:       "sha256:abcdef1234567890",
		VerifiedAt: now,
	}

	data, err := json.Marshal(verification)
	require.NoError(t, err)

	var got AuditVerification

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, verification.ID, got.ID)
	assert.Equal(t, verification.EventID, got.EventID)
	assert.True(t, got.Valid)
	assert.Equal(t, verification.Hash, got.Hash)
	assert.True(t, verification.VerifiedAt.Equal(got.VerifiedAt))
}

func TestAuditVerificationInvalidHash(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 2, 11, 5, 0, 0, time.UTC)

	verification := AuditVerification{
		ID:         "ver-002",
		EventID:    "evt-001",
		Valid:      false,
		Hash:       "sha256:tampered000000",
		VerifiedAt: now,
	}

	data, err := json.Marshal(verification)
	require.NoError(t, err)

	var got AuditVerification

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.False(t, got.Valid)
}

// ---------------------------------------------------------------------------
// Input types — CreateRuleInput
// ---------------------------------------------------------------------------

func TestCreateRuleInputJSON(t *testing.T) {
	t.Parallel()

	desc := "Checks high-value transfers"
	input := CreateRuleInput{
		Name:        "High Value Check",
		Description: &desc,
		Priority:    5,
		Conditions: []RuleCondition{
			{Field: "amount", Operator: "gte", Value: float64(100000)},
		},
		Actions:  []string{"hold", "review"},
		Metadata: map[string]any{"owner": "compliance-team"},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var got CreateRuleInput

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, input.Name, got.Name)
	require.NotNil(t, got.Description)
	assert.Equal(t, *input.Description, *got.Description)
	assert.Equal(t, input.Priority, got.Priority)
	assert.Len(t, got.Conditions, 1)
	assert.Equal(t, input.Actions, got.Actions)
	assert.Equal(t, "compliance-team", got.Metadata["owner"])
}

// ---------------------------------------------------------------------------
// Input types — UpdateRuleInput (partial update)
// ---------------------------------------------------------------------------

func TestUpdateRuleInputPartialJSON(t *testing.T) {
	t.Parallel()

	name := "Updated Name"
	priority := 3

	input := UpdateRuleInput{
		Name:     &name,
		Priority: &priority,
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	// Only name and priority should be present
	assert.Contains(t, string(data), `"name"`)
	assert.Contains(t, string(data), `"priority"`)
	assert.NotContains(t, string(data), `"description"`)
	assert.NotContains(t, string(data), `"conditions"`)
	assert.NotContains(t, string(data), `"actions"`)
	assert.NotContains(t, string(data), `"metadata"`)
}

// ---------------------------------------------------------------------------
// Input types — CreateLimitInput
// ---------------------------------------------------------------------------

func TestCreateLimitInputJSON(t *testing.T) {
	t.Parallel()

	currency := "EUR"
	input := CreateLimitInput{
		Name:      "Monthly Cap",
		Type:      "amount",
		MaxAmount: 250000,
		Period:    "monthly",
		Currency:  &currency,
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var got CreateLimitInput

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, input.Name, got.Name)
	assert.Equal(t, input.Type, got.Type)
	assert.Equal(t, input.MaxAmount, got.MaxAmount)
	assert.Equal(t, input.Period, got.Period)
	require.NotNil(t, got.Currency)
	assert.Equal(t, "EUR", *got.Currency)
}

// ---------------------------------------------------------------------------
// Input types — UpdateLimitInput (partial update)
// ---------------------------------------------------------------------------

func TestUpdateLimitInputPartialJSON(t *testing.T) {
	t.Parallel()

	newMax := int64(500000)
	input := UpdateLimitInput{
		MaxAmount: &newMax,
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"maxAmount"`)
	assert.NotContains(t, string(data), `"name"`)
	assert.NotContains(t, string(data), `"description"`)
	assert.NotContains(t, string(data), `"period"`)
	assert.NotContains(t, string(data), `"currency"`)
	assert.NotContains(t, string(data), `"metadata"`)
}

// ---------------------------------------------------------------------------
// Input types — CreateValidationInput
// ---------------------------------------------------------------------------

func TestCreateValidationInputJSON(t *testing.T) {
	t.Parallel()

	input := CreateValidationInput{
		Transaction: map[string]any{
			"amount":   float64(7500),
			"currency": "BRL",
			"type":     "wire",
		},
		RuleIDs:  []string{"rule-001", "rule-003"},
		Metadata: map[string]any{"channel": "mobile"},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var got CreateValidationInput

	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	assert.Equal(t, float64(7500), got.Transaction["amount"])
	assert.Equal(t, "BRL", got.Transaction["currency"])
	assert.Equal(t, input.RuleIDs, got.RuleIDs)
	assert.Equal(t, "mobile", got.Metadata["channel"])
}
