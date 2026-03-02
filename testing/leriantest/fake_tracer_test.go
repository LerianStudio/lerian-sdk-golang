package leriantest_test

import (
	"context"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/testing/leriantest"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 1. Rules -- CRUD + Activate/Deactivate/Draft lifecycle
// ---------------------------------------------------------------------------

func TestFakeTracerRulesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Tracer.Rules.Create(ctx, &tracer.CreateRuleInput{
		Name:     "AML Screening",
		Priority: 10,
		Conditions: []tracer.RuleCondition{
			{Field: "amount", Operator: "gt", Value: 50000},
		},
		Actions: []string{"flag", "review"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "AML Screening", created.Name)
	assert.Equal(t, tracer.RuleStatusDraft, created.Status)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Tracer.Rules.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "AML Screening", got.Name)

	// Update
	newName := "AML Screening v2"
	updated, err := client.Tracer.Rules.Update(ctx, created.ID, &tracer.UpdateRuleInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "AML Screening v2", updated.Name)

	// Verify update persisted
	got2, err := client.Tracer.Rules.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "AML Screening v2", got2.Name)

	// List -- should have 1 item
	iter := client.Tracer.Rules.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Tracer.Rules.Delete(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Tracer.Rules.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeTracerRulesLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	rule, err := client.Tracer.Rules.Create(ctx, &tracer.CreateRuleInput{
		Name:     "Sanctions Check",
		Priority: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusDraft, rule.Status)

	// DRAFT -> ACTIVE
	activated, err := client.Tracer.Rules.Activate(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusActive, activated.Status)
	assert.True(t, activated.UpdatedAt.After(rule.UpdatedAt) || activated.UpdatedAt.Equal(rule.UpdatedAt))

	// ACTIVE -> INACTIVE
	deactivated, err := client.Tracer.Rules.Deactivate(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusInactive, deactivated.Status)

	// INACTIVE -> DRAFT
	drafted, err := client.Tracer.Rules.Draft(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusDraft, drafted.Status)

	// Verify state persisted via Get
	got, err := client.Tracer.Rules.Get(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusDraft, got.Status)

	// Clean up
	err = client.Tracer.Rules.Delete(ctx, rule.ID)
	require.NoError(t, err)
}

func TestFakeTracerRulesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-rule-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Tracer.Rules.Get(ctx, ghost); return err }},
		{"Update", func() error {
			n := "x"
			_, err := client.Tracer.Rules.Update(ctx, ghost, &tracer.UpdateRuleInput{Name: &n})
			return err
		}},
		{"Delete", func() error { return client.Tracer.Rules.Delete(ctx, ghost) }},
		{"Activate", func() error { _, err := client.Tracer.Rules.Activate(ctx, ghost); return err }},
		{"Deactivate", func() error { _, err := client.Tracer.Rules.Deactivate(ctx, ghost); return err }},
		{"Draft", func() error { _, err := client.Tracer.Rules.Draft(ctx, ghost); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

// ---------------------------------------------------------------------------
// 2. Limits -- CRUD + Activate/Deactivate/Draft lifecycle
// ---------------------------------------------------------------------------

func TestFakeTracerLimitsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Tracer.Limits.Create(ctx, &tracer.CreateLimitInput{
		Name:      "Daily Transfer Cap",
		Type:      "transaction_amount",
		MaxAmount: 1000000,
		Period:    "daily",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Daily Transfer Cap", created.Name)
	assert.Equal(t, tracer.RuleStatusDraft, created.Status)
	assert.False(t, created.CreatedAt.IsZero())

	// Get
	got, err := client.Tracer.Limits.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Daily Transfer Cap", got.Name)

	// Update
	newName := "Daily Transfer Cap v2"
	updated, err := client.Tracer.Limits.Update(ctx, created.ID, &tracer.UpdateLimitInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Daily Transfer Cap v2", updated.Name)

	// Verify update persisted
	got2, err := client.Tracer.Limits.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Daily Transfer Cap v2", got2.Name)

	// List
	iter := client.Tracer.Limits.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Tracer.Limits.Delete(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Tracer.Limits.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeTracerLimitsLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	limit, err := client.Tracer.Limits.Create(ctx, &tracer.CreateLimitInput{
		Name:      "Monthly Wire Limit",
		Type:      "wire_count",
		MaxAmount: 100,
		Period:    "monthly",
	})
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusDraft, limit.Status)

	// DRAFT -> ACTIVE
	activated, err := client.Tracer.Limits.Activate(ctx, limit.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusActive, activated.Status)

	// ACTIVE -> INACTIVE
	deactivated, err := client.Tracer.Limits.Deactivate(ctx, limit.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusInactive, deactivated.Status)

	// INACTIVE -> DRAFT
	drafted, err := client.Tracer.Limits.Draft(ctx, limit.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusDraft, drafted.Status)

	// Verify persisted
	got, err := client.Tracer.Limits.Get(ctx, limit.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatusDraft, got.Status)
}

func TestFakeTracerLimitsNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-limit-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Tracer.Limits.Get(ctx, ghost); return err }},
		{"Update", func() error {
			n := "x"
			_, err := client.Tracer.Limits.Update(ctx, ghost, &tracer.UpdateLimitInput{Name: &n})
			return err
		}},
		{"Delete", func() error { return client.Tracer.Limits.Delete(ctx, ghost) }},
		{"Activate", func() error { _, err := client.Tracer.Limits.Activate(ctx, ghost); return err }},
		{"Deactivate", func() error { _, err := client.Tracer.Limits.Deactivate(ctx, ghost); return err }},
		{"Draft", func() error { _, err := client.Tracer.Limits.Draft(ctx, ghost); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Validations -- Create (RPC-style) + Get + List
// ---------------------------------------------------------------------------

func TestFakeTracerValidationsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create a validation (submits a transaction for rule evaluation)
	created, err := client.Tracer.Validations.Create(ctx, &tracer.CreateValidationInput{
		Transaction: map[string]any{
			"amount":    75000,
			"currency":  "BRL",
			"accountId": "acct-123",
		},
		RuleIDs: []string{"rule-1", "rule-2"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "PASSED", created.Status)
	assert.False(t, created.CreatedAt.IsZero())

	// Get
	got, err := client.Tracer.Validations.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "PASSED", got.Status)

	// Create another for list
	created2, err := client.Tracer.Validations.Create(ctx, &tracer.CreateValidationInput{
		Transaction: map[string]any{
			"amount":   500,
			"currency": "USD",
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created2.ID)
	assert.NotEqual(t, created.ID, created2.ID)

	// List -- should have 2 items
	iter := client.Tracer.Validations.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestFakeTracerValidationsNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	_, err := client.Tracer.Validations.Get(ctx, "nonexistent-validation-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 4. AuditEvents -- Get + List + Verify (read-only, no Create)
// ---------------------------------------------------------------------------

func TestFakeTracerAuditEventsReadOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Stores start empty -- List should return zero items.
	iter := client.Tracer.AuditEvents.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)

	// Get on empty store -- not found
	_, err = client.Tracer.AuditEvents.Get(ctx, "nonexistent-event-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Verify on empty store -- not found
	_, err = client.Tracer.AuditEvents.Verify(ctx, "nonexistent-event-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestFakeTracerAuditEventsVerify tests the Verify action by first creating
// a validation (which populates the store indirectly if seeded), but since
// AuditEvents has no Create in the fake, we verify Verify returns not-found
// for missing IDs and demonstrate the interface works correctly.
func TestFakeTracerAuditEventsVerifyNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	_, err := client.Tracer.AuditEvents.Verify(ctx, "ghost-event")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// Supplemental: List with multiple items
// ---------------------------------------------------------------------------

func TestFakeTracerListMultipleItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create 3 rules
	for i := 0; i < 3; i++ {
		_, err := client.Tracer.Rules.Create(ctx, &tracer.CreateRuleInput{
			Name:     "Rule",
			Priority: i,
		})
		require.NoError(t, err)
	}

	ruleIter := client.Tracer.Rules.List(ctx, nil)
	rules, err := ruleIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, rules, 3)

	// Create 3 limits
	for i := 0; i < 3; i++ {
		_, err := client.Tracer.Limits.Create(ctx, &tracer.CreateLimitInput{
			Name:      "Limit",
			Type:      "amount",
			MaxAmount: int64(i * 1000),
			Period:    "daily",
		})
		require.NoError(t, err)
	}

	limitIter := client.Tracer.Limits.List(ctx, nil)
	limits, err := limitIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, limits, 3)

	// Create 3 validations
	for i := 0; i < 3; i++ {
		_, err := client.Tracer.Validations.Create(ctx, &tracer.CreateValidationInput{
			Transaction: map[string]any{"amount": i * 100},
		})
		require.NoError(t, err)
	}

	valIter := client.Tracer.Validations.List(ctx, nil)
	vals, err := valIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, vals, 3)
}

// ---------------------------------------------------------------------------
// Supplemental: Error injection
// ---------------------------------------------------------------------------

func TestFakeTracerRulesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("tracer.Rules.Create", injectedErr),
	)

	_, err := client.Tracer.Rules.Create(ctx, &tracer.CreateRuleInput{
		Name:     "Should Fail",
		Priority: 1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)

	// Non-injected operations still work
	_, err = client.Tracer.Limits.Create(ctx, &tracer.CreateLimitInput{
		Name:      "Works Fine",
		Type:      "amount",
		MaxAmount: 500,
		Period:    "daily",
	})
	require.NoError(t, err)
}
