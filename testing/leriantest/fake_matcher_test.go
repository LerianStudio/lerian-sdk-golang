package leriantest_test

import (
	"context"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/testing/leriantest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 1. Contexts
// ---------------------------------------------------------------------------

func TestFakeMatcherContextsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.Contexts.Create(ctx, &matcher.CreateContextInput{
		Name: "Recon Alpha",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Recon Alpha", created.Name)

	// Get
	got, err := client.Matcher.Contexts.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Recon Alpha", got.Name)

	// Update
	newName := "Recon Alpha v2"
	updated, err := client.Matcher.Contexts.Update(ctx, created.ID, &matcher.UpdateContextInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Recon Alpha v2", updated.Name)

	// Verify update persisted
	got2, err := client.Matcher.Contexts.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Recon Alpha v2", got2.Name)

	// Clone
	cloned, err := client.Matcher.Contexts.Clone(ctx, created.ID)
	require.NoError(t, err)
	assert.NotEqual(t, created.ID, cloned.ID)
	assert.Equal(t, "Recon Alpha v2", cloned.Name)

	// List -- should have original + clone = 2
	iter := client.Matcher.Contexts.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// Delete
	err = client.Matcher.Contexts.Delete(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Matcher.Contexts.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Clone should still be present
	_, err = client.Matcher.Contexts.Get(ctx, cloned.ID)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// 2. Rules
// ---------------------------------------------------------------------------

func TestFakeMatcherRulesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.Rules.Create(ctx, &matcher.CreateRuleInput{
		Name:       "Amount Match",
		ContextID:  "ctx-1",
		Priority:   1,
		Expression: "source.amount == target.amount",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Amount Match", created.Name)

	// Get
	got, err := client.Matcher.Rules.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Update
	newName := "Amount Match v2"
	updated, err := client.Matcher.Rules.Update(ctx, created.ID, &matcher.UpdateRuleInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Amount Match v2", updated.Name)

	// Reorder -- no-op in fake, just verify no error
	err = client.Matcher.Rules.Reorder(ctx, "ctx-1", &matcher.ReorderRulesInput{
		RuleIDs: []string{created.ID},
	})
	require.NoError(t, err)

	// List
	iter := client.Matcher.Rules.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Delete
	err = client.Matcher.Rules.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = client.Matcher.Rules.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 3. Schedules
// ---------------------------------------------------------------------------

func TestFakeMatcherSchedulesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.Schedules.Create(ctx, &matcher.CreateScheduleInput{
		ContextID: "ctx-1",
		Name:      "Daily Recon",
		CronExpr:  "0 2 * * *",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Daily Recon", created.Name)

	// Get
	got, err := client.Matcher.Schedules.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Update
	updated, err := client.Matcher.Schedules.Update(ctx, created.ID, &matcher.UpdateScheduleInput{})
	require.NoError(t, err)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List
	iter := client.Matcher.Schedules.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Delete
	err = client.Matcher.Schedules.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = client.Matcher.Schedules.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 4. Sources
// ---------------------------------------------------------------------------

func TestFakeMatcherSourcesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.Sources.Create(ctx, &matcher.CreateSourceInput{
		ContextID: "ctx-1",
		Name:      "Bank Feed",
		Type:      "csv",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Bank Feed", created.Name)

	// Get
	got, err := client.Matcher.Sources.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Update
	updated, err := client.Matcher.Sources.Update(ctx, created.ID, &matcher.UpdateSourceInput{})
	require.NoError(t, err)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List
	iter := client.Matcher.Sources.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Delete
	err = client.Matcher.Sources.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = client.Matcher.Sources.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 5. SourceFieldMaps
// ---------------------------------------------------------------------------

func TestFakeMatcherSourceFieldMapsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.SourceFieldMaps.Create(ctx, &matcher.CreateSourceFieldMapInput{
		SourceID:  "src-1",
		FieldName: "amount",
		MappedTo:  "transaction_amount",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	// Get
	got, err := client.Matcher.SourceFieldMaps.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Update
	updated, err := client.Matcher.SourceFieldMaps.Update(ctx, created.ID, &matcher.UpdateSourceFieldMapInput{})
	require.NoError(t, err)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List
	iter := client.Matcher.SourceFieldMaps.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Delete
	err = client.Matcher.SourceFieldMaps.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = client.Matcher.SourceFieldMaps.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 6. FeeSchedules
// ---------------------------------------------------------------------------

func TestFakeMatcherFeeSchedulesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.FeeSchedules.Create(ctx, &matcher.CreateFeeScheduleInput{
		ContextID: "ctx-1",
		Name:      "Standard Fees",
		Rules: []matcher.FeeRule{
			{Type: "percentage", Currency: "BRL"},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Standard Fees", created.Name)

	// Get
	got, err := client.Matcher.FeeSchedules.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Update
	updated, err := client.Matcher.FeeSchedules.Update(ctx, created.ID, &matcher.UpdateFeeScheduleInput{})
	require.NoError(t, err)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// Simulate
	simResult, err := client.Matcher.FeeSchedules.Simulate(ctx, created.ID, &matcher.SimulateFeeScheduleInput{
		Amount:   10000,
		Currency: "BRL",
	})
	require.NoError(t, err)
	assert.NotNil(t, simResult)

	// List
	iter := client.Matcher.FeeSchedules.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Delete
	err = client.Matcher.FeeSchedules.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = client.Matcher.FeeSchedules.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 7. FieldMaps
// ---------------------------------------------------------------------------

func TestFakeMatcherFieldMapsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.FieldMaps.Create(ctx, &matcher.CreateFieldMapInput{
		ContextID:   "ctx-1",
		SourceField: "date",
		TargetField: "transaction_date",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	// Get
	got, err := client.Matcher.FieldMaps.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Update
	updated, err := client.Matcher.FieldMaps.Update(ctx, created.ID, &matcher.UpdateFieldMapInput{})
	require.NoError(t, err)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List
	iter := client.Matcher.FieldMaps.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Delete
	err = client.Matcher.FieldMaps.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = client.Matcher.FieldMaps.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 8. ExportJobs
// ---------------------------------------------------------------------------

func TestFakeMatcherExportJobsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.ExportJobs.Create(ctx, &matcher.CreateExportJobInput{
		ContextID: "ctx-1",
		Format:    "csv",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	// Get
	got, err := client.Matcher.ExportJobs.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Cancel
	cancelled, err := client.Matcher.ExportJobs.Cancel(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, cancelled.ID)

	// Download
	data, err := client.Matcher.ExportJobs.Download(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, []byte("fake-export-data"), data)

	// Download not found
	_, err = client.Matcher.ExportJobs.Download(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Cancel not found
	_, err = client.Matcher.ExportJobs.Cancel(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// List
	// Create a second export job to verify list count
	_, err = client.Matcher.ExportJobs.Create(ctx, &matcher.CreateExportJobInput{
		ContextID: "ctx-1",
		Format:    "xlsx",
	})
	require.NoError(t, err)

	iter := client.Matcher.ExportJobs.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

// ---------------------------------------------------------------------------
// 9. Disputes
// ---------------------------------------------------------------------------

func TestFakeMatcherDisputesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.Disputes.Create(ctx, &matcher.CreateDisputeInput{
		ContextID: "ctx-1",
		Reason:    "Amount mismatch",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	// Get
	got, err := client.Matcher.Disputes.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Update
	updated, err := client.Matcher.Disputes.Update(ctx, created.ID, &matcher.UpdateDisputeInput{})
	require.NoError(t, err)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// Resolve
	resolved, err := client.Matcher.Disputes.Resolve(ctx, created.ID, &matcher.ResolveDisputeInput{
		Resolution: "Adjusted amounts manually",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, resolved.ID)

	// Escalate
	escalated, err := client.Matcher.Disputes.Escalate(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, escalated.ID)

	// Resolve/Escalate not found
	_, err = client.Matcher.Disputes.Resolve(ctx, "nonexistent", &matcher.ResolveDisputeInput{Resolution: "n/a"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = client.Matcher.Disputes.Escalate(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// List
	iter := client.Matcher.Disputes.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

// ---------------------------------------------------------------------------
// 10. Exceptions
// ---------------------------------------------------------------------------

func TestFakeMatcherExceptionsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.Exceptions.Create(ctx, &matcher.CreateExceptionInput{
		ContextID: "ctx-1",
		Type:      "amount_mismatch",
		Priority:  "high",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	// Get
	got, err := client.Matcher.Exceptions.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Update
	updated, err := client.Matcher.Exceptions.Update(ctx, created.ID, &matcher.UpdateExceptionInput{})
	require.NoError(t, err)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// Approve
	approved, err := client.Matcher.Exceptions.Approve(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, approved.ID)

	// Reject
	rejected, err := client.Matcher.Exceptions.Reject(ctx, created.ID, &matcher.RejectExceptionInput{
		Reason: "Invalid data",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, rejected.ID)

	// Reassign
	reassigned, err := client.Matcher.Exceptions.Reassign(ctx, created.ID, &matcher.ReassignExceptionInput{
		AssignTo: "user-42",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, reassigned.ID)

	// Not found variants
	_, err = client.Matcher.Exceptions.Approve(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = client.Matcher.Exceptions.Reject(ctx, "nonexistent", &matcher.RejectExceptionInput{Reason: "n/a"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = client.Matcher.Exceptions.Reassign(ctx, "nonexistent", &matcher.ReassignExceptionInput{AssignTo: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Bulk operations -- these return empty results in the fake
	bulkApproveResult, err := client.Matcher.Exceptions.BulkApprove(ctx, &matcher.BulkExceptionInput{
		IDs: []string{created.ID},
	})
	require.NoError(t, err)
	assert.NotNil(t, bulkApproveResult)

	bulkRejectResult, err := client.Matcher.Exceptions.BulkReject(ctx, &matcher.BulkRejectInput{
		IDs:    []string{created.ID},
		Reason: "bulk rejection",
	})
	require.NoError(t, err)
	assert.NotNil(t, bulkRejectResult)

	bulkReassignResult, err := client.Matcher.Exceptions.BulkReassign(ctx, &matcher.BulkReassignInput{
		IDs:      []string{created.ID},
		AssignTo: "team-lead",
	})
	require.NoError(t, err)
	assert.NotNil(t, bulkReassignResult)

	// ListByContext
	iterByCtx := client.Matcher.Exceptions.ListByContext(ctx, "ctx-1", nil)
	itemsByCtx, err := iterByCtx.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, itemsByCtx, 1)

	// GetStatistics
	stats, err := client.Matcher.Exceptions.GetStatistics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, stats)

	// List
	iter := client.Matcher.Exceptions.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Delete
	err = client.Matcher.Exceptions.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = client.Matcher.Exceptions.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 11. Governance
// ---------------------------------------------------------------------------

func TestFakeMatcherGovernanceCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Governance is read-only with no Create -- stores start empty.
	// ListArchives on empty store
	archiveIter := client.Matcher.Governance.ListArchives(ctx, nil)
	archives, err := archiveIter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, archives)

	// GetArchive not found
	_, err = client.Matcher.Governance.GetArchive(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// ListAuditLogs on empty store
	auditIter := client.Matcher.Governance.ListAuditLogs(ctx, nil)
	auditLogs, err := auditIter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, auditLogs)

	// GetAuditLog not found
	_, err = client.Matcher.Governance.GetAuditLog(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 12. Imports
// ---------------------------------------------------------------------------

func TestFakeMatcherImportsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Matcher.Imports.Create(ctx, &matcher.CreateImportInput{
		ContextID: "ctx-1",
		FileName:  "transactions_2024.csv",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)

	// Get
	got, err := client.Matcher.Imports.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// GetStatus
	status, err := client.Matcher.Imports.GetStatus(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, status.ID)

	// GetStatus not found
	_, err = client.Matcher.Imports.GetStatus(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Cancel
	cancelled, err := client.Matcher.Imports.Cancel(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, cancelled.ID)

	// Cancel not found
	_, err = client.Matcher.Imports.Cancel(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// List
	// Create a second import to verify count
	_, err = client.Matcher.Imports.Create(ctx, &matcher.CreateImportInput{
		ContextID: "ctx-1",
		FileName:  "balances_2024.csv",
	})
	require.NoError(t, err)

	iter := client.Matcher.Imports.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

// ---------------------------------------------------------------------------
// 13. Matching
// ---------------------------------------------------------------------------

func TestFakeMatcherMatchingActions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Run
	result, err := client.Matcher.Matching.Run(ctx, "ctx-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "ctx-1", result.ContextID)

	// Manual
	manualResult, err := client.Matcher.Matching.Manual(ctx, &matcher.ManualMatchInput{
		ContextID:       "ctx-1",
		SourceRecordIDs: []string{"src-rec-1", "src-rec-2"},
		TargetRecordIDs: []string{"tgt-rec-1", "tgt-rec-2"},
	})
	require.NoError(t, err)
	assert.NotNil(t, manualResult)

	// Adjust
	adj, err := client.Matcher.Matching.Adjust(ctx, &matcher.AdjustmentInput{
		ContextID: "ctx-1",
		Type:      "credit",
		Amount:    500,
		Reason:    "Rounding correction",
	})
	require.NoError(t, err)
	assert.NotNil(t, adj)
	assert.NotEmpty(t, adj.ID)
}

// ---------------------------------------------------------------------------
// 14. Reports
// ---------------------------------------------------------------------------

func TestFakeMatcherReportsActions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	contextID := "ctx-1"

	// GetSummary
	summary, err := client.Matcher.Reports.GetSummary(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, summary)

	// GetMatchRate
	matchRate, err := client.Matcher.Reports.GetMatchRate(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, matchRate)

	// GetExceptionTrend
	exTrend, err := client.Matcher.Reports.GetExceptionTrend(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, exTrend)

	// GetAgingAnalysis
	aging, err := client.Matcher.Reports.GetAgingAnalysis(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, aging)

	// GetSourceComparison
	srcComp, err := client.Matcher.Reports.GetSourceComparison(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, srcComp)

	// GetVolumeAnalysis
	volume, err := client.Matcher.Reports.GetVolumeAnalysis(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, volume)

	// GetDisputeMetrics
	dispMetrics, err := client.Matcher.Reports.GetDisputeMetrics(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, dispMetrics)

	// GetFeeAnalysis
	feeAnalysis, err := client.Matcher.Reports.GetFeeAnalysis(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, feeAnalysis)

	// GetReconciliationHistory -- returns empty iterator in fake
	histIter := client.Matcher.Reports.GetReconciliationHistory(ctx, contextID, nil)
	histItems, err := histIter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, histItems)

	// GetPerformanceMetrics
	perfMetrics, err := client.Matcher.Reports.GetPerformanceMetrics(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, perfMetrics)

	// GetDashboard
	dashboard, err := client.Matcher.Reports.GetDashboard(ctx, contextID)
	require.NoError(t, err)
	assert.NotNil(t, dashboard)
}

// ---------------------------------------------------------------------------
// Supplemental: NotFound errors for all store-backed services
// ---------------------------------------------------------------------------

func TestFakeMatcherNotFoundErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-id"

	// Each service's Get should return not-found for a missing ID.
	tests := []struct {
		name string
		fn   func() error
	}{
		{"Contexts.Get", func() error { _, err := client.Matcher.Contexts.Get(ctx, ghost); return err }},
		{"Contexts.Update", func() error {
			_, err := client.Matcher.Contexts.Update(ctx, ghost, &matcher.UpdateContextInput{})
			return err
		}},
		{"Contexts.Delete", func() error { return client.Matcher.Contexts.Delete(ctx, ghost) }},
		{"Contexts.Clone", func() error { _, err := client.Matcher.Contexts.Clone(ctx, ghost); return err }},
		{"Rules.Get", func() error { _, err := client.Matcher.Rules.Get(ctx, ghost); return err }},
		{"Rules.Update", func() error {
			_, err := client.Matcher.Rules.Update(ctx, ghost, &matcher.UpdateRuleInput{})
			return err
		}},
		{"Rules.Delete", func() error { return client.Matcher.Rules.Delete(ctx, ghost) }},
		{"Schedules.Get", func() error { _, err := client.Matcher.Schedules.Get(ctx, ghost); return err }},
		{"Schedules.Update", func() error {
			_, err := client.Matcher.Schedules.Update(ctx, ghost, &matcher.UpdateScheduleInput{})
			return err
		}},
		{"Schedules.Delete", func() error { return client.Matcher.Schedules.Delete(ctx, ghost) }},
		{"Sources.Get", func() error { _, err := client.Matcher.Sources.Get(ctx, ghost); return err }},
		{"Sources.Update", func() error {
			_, err := client.Matcher.Sources.Update(ctx, ghost, &matcher.UpdateSourceInput{})
			return err
		}},
		{"Sources.Delete", func() error { return client.Matcher.Sources.Delete(ctx, ghost) }},
		{"SourceFieldMaps.Get", func() error { _, err := client.Matcher.SourceFieldMaps.Get(ctx, ghost); return err }},
		{"SourceFieldMaps.Update", func() error {
			_, err := client.Matcher.SourceFieldMaps.Update(ctx, ghost, &matcher.UpdateSourceFieldMapInput{})
			return err
		}},
		{"SourceFieldMaps.Delete", func() error { return client.Matcher.SourceFieldMaps.Delete(ctx, ghost) }},
		{"FeeSchedules.Get", func() error { _, err := client.Matcher.FeeSchedules.Get(ctx, ghost); return err }},
		{"FeeSchedules.Update", func() error {
			_, err := client.Matcher.FeeSchedules.Update(ctx, ghost, &matcher.UpdateFeeScheduleInput{})
			return err
		}},
		{"FeeSchedules.Delete", func() error { return client.Matcher.FeeSchedules.Delete(ctx, ghost) }},
		{"FieldMaps.Get", func() error { _, err := client.Matcher.FieldMaps.Get(ctx, ghost); return err }},
		{"FieldMaps.Update", func() error {
			_, err := client.Matcher.FieldMaps.Update(ctx, ghost, &matcher.UpdateFieldMapInput{})
			return err
		}},
		{"FieldMaps.Delete", func() error { return client.Matcher.FieldMaps.Delete(ctx, ghost) }},
		{"ExportJobs.Get", func() error { _, err := client.Matcher.ExportJobs.Get(ctx, ghost); return err }},
		{"ExportJobs.Cancel", func() error { _, err := client.Matcher.ExportJobs.Cancel(ctx, ghost); return err }},
		{"ExportJobs.Download", func() error { _, err := client.Matcher.ExportJobs.Download(ctx, ghost); return err }},
		{"Disputes.Get", func() error { _, err := client.Matcher.Disputes.Get(ctx, ghost); return err }},
		{"Disputes.Update", func() error {
			_, err := client.Matcher.Disputes.Update(ctx, ghost, &matcher.UpdateDisputeInput{})
			return err
		}},
		{"Disputes.Resolve", func() error {
			_, err := client.Matcher.Disputes.Resolve(ctx, ghost, &matcher.ResolveDisputeInput{Resolution: "x"})
			return err
		}},
		{"Disputes.Escalate", func() error { _, err := client.Matcher.Disputes.Escalate(ctx, ghost); return err }},
		{"Exceptions.Get", func() error { _, err := client.Matcher.Exceptions.Get(ctx, ghost); return err }},
		{"Exceptions.Update", func() error {
			_, err := client.Matcher.Exceptions.Update(ctx, ghost, &matcher.UpdateExceptionInput{})
			return err
		}},
		{"Exceptions.Delete", func() error { return client.Matcher.Exceptions.Delete(ctx, ghost) }},
		{"Exceptions.Approve", func() error { _, err := client.Matcher.Exceptions.Approve(ctx, ghost); return err }},
		{"Exceptions.Reject", func() error {
			_, err := client.Matcher.Exceptions.Reject(ctx, ghost, &matcher.RejectExceptionInput{Reason: "x"})
			return err
		}},
		{"Exceptions.Reassign", func() error {
			_, err := client.Matcher.Exceptions.Reassign(ctx, ghost, &matcher.ReassignExceptionInput{AssignTo: "x"})
			return err
		}},
		{"Governance.GetArchive", func() error { _, err := client.Matcher.Governance.GetArchive(ctx, ghost); return err }},
		{"Governance.GetAuditLog", func() error { _, err := client.Matcher.Governance.GetAuditLog(ctx, ghost); return err }},
		{"Imports.Get", func() error { _, err := client.Matcher.Imports.Get(ctx, ghost); return err }},
		{"Imports.Cancel", func() error { _, err := client.Matcher.Imports.Cancel(ctx, ghost); return err }},
		{"Imports.GetStatus", func() error { _, err := client.Matcher.Imports.GetStatus(ctx, ghost); return err }},
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
// Supplemental: List with multiple items for store-backed services
// ---------------------------------------------------------------------------

func TestFakeMatcherListMultipleItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create 3 items for several services, then verify List returns them all.

	// Rules
	for i := 0; i < 3; i++ {
		_, err := client.Matcher.Rules.Create(ctx, &matcher.CreateRuleInput{
			Name:       "Rule",
			ContextID:  "ctx-1",
			Priority:   i,
			Expression: "a == b",
		})
		require.NoError(t, err)
	}

	ruleIter := client.Matcher.Rules.List(ctx, nil)
	rules, err := ruleIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, rules, 3)

	// Schedules
	for i := 0; i < 3; i++ {
		_, err := client.Matcher.Schedules.Create(ctx, &matcher.CreateScheduleInput{
			ContextID: "ctx-1",
			Name:      "Schedule",
			CronExpr:  "0 * * * *",
		})
		require.NoError(t, err)
	}

	schedIter := client.Matcher.Schedules.List(ctx, nil)
	scheds, err := schedIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, scheds, 3)

	// Sources
	for i := 0; i < 3; i++ {
		_, err := client.Matcher.Sources.Create(ctx, &matcher.CreateSourceInput{
			ContextID: "ctx-1",
			Name:      "Source",
			Type:      "api",
		})
		require.NoError(t, err)
	}

	srcIter := client.Matcher.Sources.List(ctx, nil)
	srcs, err := srcIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, srcs, 3)

	// Disputes
	for i := 0; i < 3; i++ {
		_, err := client.Matcher.Disputes.Create(ctx, &matcher.CreateDisputeInput{
			ContextID: "ctx-1",
			Reason:    "Reason",
		})
		require.NoError(t, err)
	}

	dispIter := client.Matcher.Disputes.List(ctx, nil)
	disps, err := dispIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, disps, 3)
}
