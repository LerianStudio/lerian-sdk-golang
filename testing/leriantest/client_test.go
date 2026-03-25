package leriantest_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	lerian "github.com/LerianStudio/lerian-sdk-golang"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/testing/leriantest"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TestNewFakeClient -- verify all product clients are populated
// ---------------------------------------------------------------------------

func TestNewFakeClient(t *testing.T) {
	t.Parallel()

	client := leriantest.NewFakeClient()

	assert.NotNil(t, client.Midaz, "Midaz should be non-nil")
	assert.NotNil(t, client.Matcher, "Matcher should be non-nil")
	assert.NotNil(t, client.Tracer, "Tracer should be non-nil")
	assert.NotNil(t, client.Reporter, "Reporter should be non-nil")
	assert.NotNil(t, client.Fees, "Fees should be non-nil")

	// Verify individual Midaz services are wired.
	assert.NotNil(t, client.Midaz.Onboarding.Organizations)
	assert.NotNil(t, client.Midaz.Onboarding.Ledgers)
	assert.NotNil(t, client.Midaz.Onboarding.Accounts)
	assert.NotNil(t, client.Midaz.Onboarding.AccountTypes)
	assert.NotNil(t, client.Midaz.Onboarding.Assets)
	assert.NotNil(t, client.Midaz.Transactions.AssetRates)
	assert.NotNil(t, client.Midaz.Onboarding.Portfolios)
	assert.NotNil(t, client.Midaz.Onboarding.Segments)
	assert.NotNil(t, client.Midaz.Transactions.Balances)
	assert.NotNil(t, client.Midaz.Transactions.Balances)
	assert.NotNil(t, client.Midaz.Transactions.Transactions)
	assert.NotNil(t, client.Midaz.Transactions.Transactions)
	assert.NotNil(t, client.Midaz.Transactions.TransactionRoutes)
	assert.NotNil(t, client.Midaz.Transactions.Operations)
	assert.NotNil(t, client.Midaz.Transactions.Operations)
	assert.NotNil(t, client.Midaz.Transactions.OperationRoutes)
	assert.NotNil(t, client.Midaz.Onboarding.Organizations)
	assert.NotNil(t, client.Midaz.Onboarding.Ledgers)
	assert.NotNil(t, client.Midaz.Onboarding.Accounts)
	assert.NotNil(t, client.Midaz.Onboarding.Assets)
	assert.NotNil(t, client.Midaz.Onboarding.Portfolios)
	assert.NotNil(t, client.Midaz.Onboarding.Segments)
	assert.NotNil(t, client.Midaz.CRM.Holders)
	assert.NotNil(t, client.Midaz.CRM.Aliases)

	// Verify Matcher services are wired.
	assert.NotNil(t, client.Matcher.Contexts)
	assert.NotNil(t, client.Matcher.Rules)
	assert.NotNil(t, client.Matcher.Schedules)
	assert.NotNil(t, client.Matcher.Sources)
	assert.NotNil(t, client.Matcher.SourceFieldMaps)
	assert.NotNil(t, client.Matcher.FeeSchedules)
	assert.NotNil(t, client.Matcher.FieldMaps)
	assert.NotNil(t, client.Matcher.ExportJobs)
	assert.NotNil(t, client.Matcher.Disputes)
	assert.NotNil(t, client.Matcher.Exceptions)
	assert.NotNil(t, client.Matcher.Governance)
	assert.NotNil(t, client.Matcher.Imports)
	assert.NotNil(t, client.Matcher.Matching)
	assert.NotNil(t, client.Matcher.Reports)

	// Verify Tracer services.
	assert.NotNil(t, client.Tracer.Rules)
	assert.NotNil(t, client.Tracer.Limits)
	assert.NotNil(t, client.Tracer.Validations)
	assert.NotNil(t, client.Tracer.AuditEvents)

	// Verify Reporter services.
	assert.NotNil(t, client.Reporter.DataSources)
	assert.NotNil(t, client.Reporter.Reports)
	assert.NotNil(t, client.Reporter.Templates)

	// Verify Fees services.
	assert.NotNil(t, client.Fees.Packages)
	assert.NotNil(t, client.Fees.Estimates)
	assert.NotNil(t, client.Fees.Fees)
}

// ---------------------------------------------------------------------------
// TestFakeMidazCRUD -- full round-trip for Organizations
// ---------------------------------------------------------------------------

func TestFakeMidazCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	org, err := client.Midaz.Onboarding.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
		LegalName:     "Acme Corp",
		LegalDocument: "12345678000100",
	})
	require.NoError(t, err)
	require.NotEmpty(t, org.ID)
	assert.Equal(t, "Acme Corp", org.LegalName)
	assert.Equal(t, "12345678000100", org.LegalDocument)
	assert.Equal(t, "active", org.Status.Code)

	// Get
	got, err := client.Midaz.Onboarding.Organizations.Get(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, org.ID, got.ID)
	assert.Equal(t, "Acme Corp", got.LegalName)

	// Update
	newName := "Acme Corp v2"
	updated, err := client.Midaz.Onboarding.Organizations.Update(ctx, org.ID, &midaz.UpdateOrganizationInput{
		LegalName: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp v2", updated.LegalName)

	// Verify update persisted
	got2, err := client.Midaz.Onboarding.Organizations.Get(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp v2", got2.LegalName)

	// List
	iter := client.Midaz.Onboarding.Organizations.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, org.ID, items[0].ID)

	// Delete
	err = client.Midaz.Onboarding.Organizations.Delete(ctx, org.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Onboarding.Organizations.Get(ctx, org.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// TestFakeMidazLedgersCRUD -- round-trip for Ledgers
// ---------------------------------------------------------------------------

func TestFakeMidazLedgersCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-test"

	ledger, err := client.Midaz.Onboarding.Ledgers.Create(ctx, orgID, &midaz.CreateLedgerInput{
		Name: "Primary Ledger",
	})
	require.NoError(t, err)
	assert.Equal(t, "Primary Ledger", ledger.Name)
	assert.Equal(t, orgID, ledger.OrganizationID)

	got, err := client.Midaz.Onboarding.Ledgers.Get(ctx, orgID, ledger.ID)
	require.NoError(t, err)
	assert.Equal(t, ledger.ID, got.ID)

	err = client.Midaz.Onboarding.Ledgers.Delete(ctx, orgID, ledger.ID)
	require.NoError(t, err)

	_, err = client.Midaz.Onboarding.Ledgers.Get(ctx, orgID, ledger.ID)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// TestFakeTransactionLifecycle -- create, commit, cancel, revert
// ---------------------------------------------------------------------------

func TestFakeTransactionLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	tx, err := client.Midaz.Transactions.Transactions.Create(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInput{
		Send: &midaz.TransactionSend{
			Asset:      "BRL",
			Value:      "100.00",
			Source:     midaz.TransactionSendSource{From: []midaz.TransactionOperationLeg{{AccountAlias: "acct-a"}}},
			Distribute: midaz.TransactionSendDistribution{To: []midaz.TransactionOperationLeg{{AccountAlias: "acct-b"}}},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "pending", tx.Status.Code)

	// Commit
	committed, err := client.Midaz.Transactions.Transactions.Commit(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)
	assert.Equal(t, "committed", committed.Status.Code)

	// Create another to test cancel.
	tx2, err := client.Midaz.Transactions.Transactions.Create(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInput{
		Send: &midaz.TransactionSend{
			Asset: "USD",
			Value: "5.00",
		},
	})
	require.NoError(t, err)

	cancelled, err := client.Midaz.Transactions.Transactions.Cancel(ctx, "org-1", "ledger-1", tx2.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", cancelled.Status.Code)

	// Revert the committed one.
	reverted, err := client.Midaz.Transactions.Transactions.Revert(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)
	assert.Equal(t, "reverted", reverted.Status.Code)
}

// ---------------------------------------------------------------------------
// TestFakeSeedData -- verify seed options work
// ---------------------------------------------------------------------------

func TestFakeSeedData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient(
		leriantest.WithSeedOrganizations(
			midaz.Organization{ID: "seed-org-1", LegalName: "Seed Corp"},
			midaz.Organization{ID: "seed-org-2", LegalName: "Another Corp"},
		),
		leriantest.WithSeedLedgers(
			midaz.Ledger{ID: "seed-ledger-1", Name: "Main Ledger"},
		),
	)

	// Verify seeded organizations are retrievable.
	org, err := client.Midaz.Onboarding.Organizations.Get(ctx, "seed-org-1")
	require.NoError(t, err)
	assert.Equal(t, "Seed Corp", org.LegalName)

	org2, err := client.Midaz.Onboarding.Organizations.Get(ctx, "seed-org-2")
	require.NoError(t, err)
	assert.Equal(t, "Another Corp", org2.LegalName)

	// Verify list returns seeded items.
	iter := client.Midaz.Onboarding.Organizations.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// Verify seeded ledger.
	ledger, err := client.Midaz.Onboarding.Ledgers.Get(ctx, "", "seed-ledger-1")
	require.NoError(t, err)
	assert.Equal(t, "Main Ledger", ledger.Name)
}

// ---------------------------------------------------------------------------
// TestFakeErrorInjection -- verify WithErrorOn works
// ---------------------------------------------------------------------------

func TestFakeErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := errors.New("injected: service unavailable")

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("midaz.Organizations.Create", injectedErr),
	)

	_, err := client.Midaz.Onboarding.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
		LegalName:     "Should Fail",
		LegalDocument: "000",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)

	// Non-injected operations should still work.
	_, err = client.Midaz.Onboarding.Ledgers.Create(ctx, "org-1", &midaz.CreateLedgerInput{
		Name: "Works Fine",
	})
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// TestFakeConcurrentAccess -- parallel create/get operations
// ---------------------------------------------------------------------------

func TestFakeConcurrentAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	const n = 100
	ids := make([]string, n)

	var wg sync.WaitGroup

	// Parallel creates.
	wg.Add(n)

	for i := range n {
		go func(idx int) {
			defer wg.Done()

			org, err := client.Midaz.Onboarding.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
				LegalName:     "Concurrent Org",
				LegalDocument: "doc",
			})
			require.NoError(t, err)

			ids[idx] = org.ID
		}(i)
	}

	wg.Wait()

	// Parallel gets.
	wg.Add(n)

	for i := range n {
		go func(idx int) {
			defer wg.Done()

			org, err := client.Midaz.Onboarding.Organizations.Get(ctx, ids[idx])
			require.NoError(t, err)
			assert.Equal(t, ids[idx], org.ID)
		}(i)
	}

	wg.Wait()

	// Verify total count.
	iter := client.Midaz.Onboarding.Organizations.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, n)
}

// ---------------------------------------------------------------------------
// TestFakeAccountLookups -- GetByAlias and GetByExternalCode
// ---------------------------------------------------------------------------

func TestFakeAccountLookups(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	alias := "checking-main"
	extCode := "EXT-001"

	acct, err := client.Midaz.Onboarding.Accounts.Create(ctx, "org-1", "ledger-1", &midaz.CreateAccountInput{
		Name:         "Checking Account",
		Type:         "deposit",
		AssetCode:    "BRL",
		Alias:        &alias,
		ExternalCode: &extCode,
	})
	require.NoError(t, err)

	// GetByAlias
	got, err := client.Midaz.Onboarding.Accounts.GetByAlias(ctx, "org-1", "ledger-1", "checking-main")
	require.NoError(t, err)
	assert.Equal(t, acct.ID, got.ID)

	// GetByExternalCode
	got2, err := client.Midaz.Onboarding.Accounts.GetByExternalCode(ctx, "org-1", "ledger-1", "EXT-001")
	require.NoError(t, err)
	assert.Equal(t, acct.ID, got2.ID)

	// Not found
	_, err = client.Midaz.Onboarding.Accounts.GetByAlias(ctx, "org-1", "ledger-1", "nonexistent")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// TestFakeNotFoundErrors -- verify proper error on missing resources
// ---------------------------------------------------------------------------

func TestFakeNotFoundErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	_, err := client.Midaz.Onboarding.Organizations.Get(ctx, "nonexistent-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = client.Midaz.Onboarding.Ledgers.Get(ctx, "org-1", "nonexistent-ledger")
	require.Error(t, err)

	err = client.Midaz.Onboarding.Organizations.Delete(ctx, "nonexistent-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// TestFakeMatcherRoundTrip -- basic matcher create/get/list
// ---------------------------------------------------------------------------

func TestFakeMatcherRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	mctx, err := client.Matcher.Contexts.Create(ctx, &matcher.CreateContextInput{
		Name: "Test Recon Context",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, mctx.ID)
	assert.Equal(t, "Test Recon Context", mctx.Name)

	got, err := client.Matcher.Contexts.Get(ctx, mctx.ID)
	require.NoError(t, err)
	assert.Equal(t, mctx.ID, got.ID)

	// Clone
	cloned, err := client.Matcher.Contexts.Clone(ctx, mctx.ID)
	require.NoError(t, err)
	assert.NotEqual(t, mctx.ID, cloned.ID)
	assert.Equal(t, mctx.Name, cloned.Name)

	// List
	iter := client.Matcher.Contexts.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// Delete
	err = client.Matcher.Contexts.Delete(ctx, mctx.ID)
	require.NoError(t, err)

	_, err = client.Matcher.Contexts.Get(ctx, mctx.ID)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// TestFakeTracerRuleLifecycle -- create, activate, deactivate, draft
// ---------------------------------------------------------------------------

func TestFakeTracerRuleLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	rule, err := client.Tracer.Rules.Create(ctx, &tracer.CreateRuleInput{
		Name:     "AML Check",
		Priority: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatus("DRAFT"), rule.Status)

	activated, err := client.Tracer.Rules.Activate(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatus("ACTIVE"), activated.Status)

	deactivated, err := client.Tracer.Rules.Deactivate(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatus("INACTIVE"), deactivated.Status)

	drafted, err := client.Tracer.Rules.Draft(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, tracer.RuleStatus("DRAFT"), drafted.Status)

	err = client.Tracer.Rules.Delete(ctx, rule.ID)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Transaction State Machine Tests
// ---------------------------------------------------------------------------

// createPendingTx is a helper that creates a pending transaction and returns
// the fake client, context, and the newly created transaction.
func createPendingTx(t *testing.T) (*lerian.Client, context.Context, *midaz.Transaction) {
	t.Helper()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	tx, err := client.Midaz.Transactions.Transactions.Create(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInput{
		Send: &midaz.TransactionSend{
			Asset:      "BRL",
			Value:      "100.00",
			Source:     midaz.TransactionSendSource{From: []midaz.TransactionOperationLeg{{AccountAlias: "acct-a"}}},
			Distribute: midaz.TransactionSendDistribution{To: []midaz.TransactionOperationLeg{{AccountAlias: "acct-b"}}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "pending", tx.Status.Code)

	return client, ctx, tx
}

func TestTransactionStateMachine_CommitFromPending(t *testing.T) {
	t.Parallel()

	client, ctx, tx := createPendingTx(t)

	committed, err := client.Midaz.Transactions.Transactions.Commit(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)
	assert.Equal(t, "committed", committed.Status.Code)
}

func TestTransactionStateMachine_CancelFromPending(t *testing.T) {
	t.Parallel()

	client, ctx, tx := createPendingTx(t)

	cancelled, err := client.Midaz.Transactions.Transactions.Cancel(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", cancelled.Status.Code)
}

func TestTransactionStateMachine_RevertFromCommitted(t *testing.T) {
	t.Parallel()

	client, ctx, tx := createPendingTx(t)

	// First commit so the transaction reaches "committed" state.
	_, err := client.Midaz.Transactions.Transactions.Commit(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)

	reverted, err := client.Midaz.Transactions.Transactions.Revert(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)
	assert.Equal(t, "reverted", reverted.Status.Code)
}

func TestTransactionStateMachine_CommitFromCancelled(t *testing.T) {
	t.Parallel()

	client, ctx, tx := createPendingTx(t)

	// Cancel first.
	_, err := client.Midaz.Transactions.Transactions.Cancel(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)

	// Attempt to commit a cancelled transaction -- should fail.
	_, err = client.Midaz.Transactions.Transactions.Commit(ctx, "org-1", "ledger-1", tx.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrConflict), "expected conflict error, got: %v", err)
	assert.Contains(t, err.Error(), "cancelled")
}

func TestTransactionStateMachine_CancelFromCommitted(t *testing.T) {
	t.Parallel()

	client, ctx, tx := createPendingTx(t)

	// Commit first.
	_, err := client.Midaz.Transactions.Transactions.Commit(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)

	// Attempt to cancel a committed transaction -- should fail.
	_, err = client.Midaz.Transactions.Transactions.Cancel(ctx, "org-1", "ledger-1", tx.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrConflict), "expected conflict error, got: %v", err)
	assert.Contains(t, err.Error(), "committed")
}

func TestTransactionStateMachine_RevertFromPending(t *testing.T) {
	t.Parallel()

	client, ctx, tx := createPendingTx(t)

	// Attempt to revert a pending transaction -- should fail.
	_, err := client.Midaz.Transactions.Transactions.Revert(ctx, "org-1", "ledger-1", tx.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrConflict), "expected conflict error, got: %v", err)
	assert.Contains(t, err.Error(), "pending")
}

// ---------------------------------------------------------------------------
// Pagination Tests
// ---------------------------------------------------------------------------

// createNOrgs is a helper that creates n organizations and returns the client.
func createNOrgs(t *testing.T, n int) (*lerian.Client, context.Context) {
	t.Helper()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	for i := range n {
		_, err := client.Midaz.Onboarding.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
			LegalName:     fmt.Sprintf("Org-%d", i),
			LegalDocument: fmt.Sprintf("doc-%d", i),
		})
		require.NoError(t, err)
	}

	return client, ctx
}

func TestFakePaginationDefaultPageSize(t *testing.T) {
	t.Parallel()

	// With 25 items and nil opts (default page size = 10), Collect should
	// still return all items -- the iterator transparently fetches all pages.
	client, ctx := createNOrgs(t, 25)

	iter := client.Midaz.Onboarding.Organizations.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 25, "Collect should drain all pages transparently")

	// Verify ordering is preserved (insertion order).
	for i, org := range items {
		assert.Equal(t, fmt.Sprintf("Org-%d", i), org.LegalName)
	}
}

func TestFakePaginationWithLimit(t *testing.T) {
	t.Parallel()

	// Create 15 orgs, list with Limit=5. Collect should return all 15
	// across 3 pages.
	client, ctx := createNOrgs(t, 15)

	iter := client.Midaz.Onboarding.Organizations.List(ctx, &models.CursorListOptions{Limit: 5})
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 15, "Collect should return all 15 items across 3 pages")
}

func TestFakePaginationCursorProgression(t *testing.T) {
	t.Parallel()

	// Create 10 orgs, list with Limit=3. Use CollectN to manually walk
	// through pages: expect 3 + 3 + 3 + 1 = 10.
	client, ctx := createNOrgs(t, 10)

	iter := client.Midaz.Onboarding.Organizations.List(ctx, &models.CursorListOptions{Limit: 3})

	// Page 1: items 0, 1, 2
	page1, err := iter.CollectN(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, page1, 3, "first page should have 3 items")
	assert.Equal(t, "Org-0", page1[0].LegalName)
	assert.Equal(t, "Org-1", page1[1].LegalName)
	assert.Equal(t, "Org-2", page1[2].LegalName)

	// Page 2: items 3, 4, 5
	page2, err := iter.CollectN(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, page2, 3, "second page should have 3 items")
	assert.Equal(t, "Org-3", page2[0].LegalName)
	assert.Equal(t, "Org-4", page2[1].LegalName)
	assert.Equal(t, "Org-5", page2[2].LegalName)

	// Page 3: items 6, 7, 8
	page3, err := iter.CollectN(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, page3, 3, "third page should have 3 items")
	assert.Equal(t, "Org-6", page3[0].LegalName)

	// Page 4 (partial): item 9
	page4, err := iter.CollectN(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, page4, 1, "fourth page should have 1 remaining item")
	assert.Equal(t, "Org-9", page4[0].LegalName)

	// No more items.
	page5, err := iter.CollectN(ctx, 3)
	require.NoError(t, err)
	assert.Empty(t, page5, "no more items should be available")
}

func TestFakePaginationEmptyStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// List on empty store with explicit opts.
	iter := client.Midaz.Onboarding.Organizations.List(ctx, &models.CursorListOptions{Limit: 5})
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items, "empty store should yield no items")

	// List on empty store with nil opts.
	iter2 := client.Midaz.Onboarding.Organizations.List(ctx, nil)
	items2, err := iter2.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items2, "empty store with nil opts should yield no items")
}
