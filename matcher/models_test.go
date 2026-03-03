package matcher

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr returns a pointer to the given value — handy for optional fields in tests.
func ptr[T any](v T) *T { return &v }

// ---------------------------------------------------------------------------
// Core entity types — JSON round-trip
// ---------------------------------------------------------------------------

func TestContextJSON(t *testing.T) {
	t.Parallel()

	tol := 0.01
	original := Context{
		ID:          "ctx-001",
		Name:        "Monthly Reconciliation",
		Description: ptr("End-of-month bank reconciliation"),
		Status:      "active",
		Config: &ContextConfig{
			MatchingStrategy: "exact",
			Tolerance:        &tol,
			AutoApprove:      true,
			Settings:         map[string]any{"maxBatch": float64(1000)},
		},
		Metadata:  map[string]any{"env": "production"},
		CreatedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 20, 14, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Verify camelCase keys.
	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	for _, key := range []string{"id", "name", "description", "status", "config", "metadata", "createdAt", "updatedAt"} {
		assert.Contains(t, raw, key, "expected key %q in JSON output", key)
	}

	// Round-trip.
	var decoded Context
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, original.Status, decoded.Status)
	require.NotNil(t, decoded.Config)
	assert.Equal(t, "exact", decoded.Config.MatchingStrategy)
	require.NotNil(t, decoded.Config.Tolerance)
	assert.InDelta(t, 0.01, *decoded.Config.Tolerance, 1e-9)
	assert.True(t, decoded.Config.AutoApprove)
	assert.Equal(t, "production", decoded.Metadata["env"])
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt))
	assert.True(t, original.UpdatedAt.Equal(decoded.UpdatedAt))
}

func TestContextOmitempty(t *testing.T) {
	t.Parallel()

	minimal := Context{
		ID:        "ctx-min",
		Name:      "Minimal",
		Status:    "draft",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(minimal)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.NotContains(t, raw, "description", "nil description should be omitted")
	assert.NotContains(t, raw, "config", "nil config should be omitted")
	assert.NotContains(t, raw, "metadata", "nil metadata should be omitted")
}

func TestRuleJSON(t *testing.T) {
	t.Parallel()

	original := Rule{
		ID:          "rule-001",
		ContextID:   "ctx-001",
		Name:        "Amount Match",
		Description: ptr("Match by exact amount"),
		Priority:    1,
		Expression:  "source.amount == target.amount",
		Status:      "active",
		Metadata:    map[string]any{"version": float64(2)},
		CreatedAt:   time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 2, 5, 11, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Rule
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.ContextID, decoded.ContextID)
	assert.Equal(t, original.Name, decoded.Name)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, 1, decoded.Priority)
	assert.Equal(t, "source.amount == target.amount", decoded.Expression)
	assert.Equal(t, "active", decoded.Status)
	assert.Equal(t, float64(2), decoded.Metadata["version"])
}

func TestExceptionJSON(t *testing.T) {
	t.Parallel()

	original := Exception{
		ID:          "exc-001",
		ContextID:   "ctx-001",
		Type:        "amount_mismatch",
		Status:      "open",
		Priority:    "high",
		Description: ptr("Source and target amounts differ by $5.00"),
		AssignedTo:  ptr("user-42"),
		Metadata:    map[string]any{"delta": float64(500)},
		CreatedAt:   time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 3, 1, 8, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	// Verify key JSON fields.
	for _, key := range []string{"id", "contextId", "type", "status", "priority", "description", "assignedTo"} {
		assert.Contains(t, raw, key)
	}

	var decoded Exception
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, "amount_mismatch", decoded.Type)
	assert.Equal(t, "high", decoded.Priority)
	require.NotNil(t, decoded.AssignedTo)
	assert.Equal(t, "user-42", *decoded.AssignedTo)
}

// ---------------------------------------------------------------------------
// Result types — JSON round-trip
// ---------------------------------------------------------------------------

func TestMatchResultJSON(t *testing.T) {
	t.Parallel()

	original := MatchResult{
		ID:             "mr-001",
		ContextID:      "ctx-001",
		Status:         "completed",
		MatchedCount:   950,
		UnmatchedCount: 30,
		ExceptionCount: 20,
		Duration:       4500,
		Details:        map[string]any{"ruleHits": float64(12)},
		CreatedAt:      time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded MatchResult
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "mr-001", decoded.ID)
	assert.Equal(t, 950, decoded.MatchedCount)
	assert.Equal(t, 30, decoded.UnmatchedCount)
	assert.Equal(t, 20, decoded.ExceptionCount)
	assert.Equal(t, int64(4500), decoded.Duration)
	assert.Equal(t, float64(12), decoded.Details["ruleHits"])
}

func TestBulkExceptionResultJSON(t *testing.T) {
	t.Parallel()

	original := BulkExceptionResult{
		Processed: 10,
		Succeeded: 8,
		Failed:    2,
		Errors:    []string{"exc-005: already resolved", "exc-009: not found"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded BulkExceptionResult
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, 10, decoded.Processed)
	assert.Equal(t, 8, decoded.Succeeded)
	assert.Equal(t, 2, decoded.Failed)
	require.Len(t, decoded.Errors, 2)
	assert.Equal(t, "exc-005: already resolved", decoded.Errors[0])
}

func TestBulkExceptionResultOmitErrors(t *testing.T) {
	t.Parallel()

	result := BulkExceptionResult{
		Processed: 5,
		Succeeded: 5,
		Failed:    0,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.NotContains(t, raw, "errors", "nil errors should be omitted")
}

// ---------------------------------------------------------------------------
// Analytics report types — JSON round-trip
// ---------------------------------------------------------------------------

func TestReconciliationSummaryJSON(t *testing.T) {
	t.Parallel()

	original := ReconciliationSummary{
		Period:           "2026-02",
		TotalRecords:     1000,
		MatchedRecords:   920,
		UnmatchedRecords: 50,
		ExceptionRecords: 30,
		MatchRate:        0.92,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	for _, key := range []string{"period", "totalRecords", "matchedRecords", "unmatchedRecords", "exceptionRecords", "matchRate"} {
		assert.Contains(t, raw, key)
	}

	var decoded ReconciliationSummary
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "2026-02", decoded.Period)
	assert.Equal(t, 1000, decoded.TotalRecords)
	assert.InDelta(t, 0.92, decoded.MatchRate, 1e-9)
}

func TestDashboardReportJSON(t *testing.T) {
	t.Parallel()

	original := DashboardReport{
		Summary: ReconciliationSummary{
			Period:           "2026-02",
			TotalRecords:     500,
			MatchedRecords:   480,
			UnmatchedRecords: 10,
			ExceptionRecords: 10,
			MatchRate:        0.96,
		},
		MatchRate: 0.96,
		Exceptions: ExceptionStatistics{
			Total:      10,
			ByStatus:   map[string]int{"open": 5, "resolved": 5},
			ByPriority: map[string]int{"high": 3, "medium": 4, "low": 3},
			ByType:     map[string]int{"amount_mismatch": 7, "missing_record": 3},
			AvgAge:     24.5,
		},
		RecentRuns: []ReconciliationHistoryEntry{
			{
				ID:             "rh-001",
				ContextID:      "ctx-001",
				RunAt:          time.Date(2026, 2, 28, 23, 0, 0, 0, time.UTC),
				Duration:       3200,
				MatchedCount:   480,
				UnmatchedCount: 10,
				ExceptionCount: 10,
				Status:         "completed",
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded DashboardReport
	require.NoError(t, json.Unmarshal(data, &decoded))

	// Verify composite structure.
	assert.Equal(t, "2026-02", decoded.Summary.Period)
	assert.Equal(t, 500, decoded.Summary.TotalRecords)
	assert.InDelta(t, 0.96, decoded.MatchRate, 1e-9)
	assert.Equal(t, 10, decoded.Exceptions.Total)
	assert.Equal(t, 5, decoded.Exceptions.ByStatus["open"])
	assert.InDelta(t, 24.5, decoded.Exceptions.AvgAge, 1e-9)
	require.Len(t, decoded.RecentRuns, 1)
	assert.Equal(t, "rh-001", decoded.RecentRuns[0].ID)
	assert.Equal(t, int64(3200), decoded.RecentRuns[0].Duration)
}

// ---------------------------------------------------------------------------
// Additional entity types — selective round-trip coverage
// ---------------------------------------------------------------------------

func TestScheduleJSON(t *testing.T) {
	t.Parallel()

	lastRun := time.Date(2026, 2, 28, 23, 0, 0, 0, time.UTC)
	nextRun := time.Date(2026, 3, 31, 23, 0, 0, 0, time.UTC)

	original := Schedule{
		ID:        "sch-001",
		ContextID: "ctx-001",
		Name:      "End of Month",
		CronExpr:  "0 23 L * *",
		Status:    "active",
		LastRunAt: &lastRun,
		NextRunAt: &nextRun,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 2, 28, 23, 5, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Schedule
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "sch-001", decoded.ID)
	assert.Equal(t, "0 23 L * *", decoded.CronExpr)
	require.NotNil(t, decoded.LastRunAt)
	assert.True(t, lastRun.Equal(*decoded.LastRunAt))
	require.NotNil(t, decoded.NextRunAt)
	assert.True(t, nextRun.Equal(*decoded.NextRunAt))
}

func TestFeeScheduleJSON(t *testing.T) {
	t.Parallel()

	amt := int64(100)
	pct := "0.5"

	original := FeeSchedule{
		ID:          "fs-001",
		ContextID:   "ctx-001",
		Name:        "Standard Fee",
		Description: ptr("Standard processing fee"),
		Rules: []FeeRule{
			{Type: "fixed", Amount: &amt, Currency: "USD"},
			{Type: "percentage", Percentage: &pct, Currency: "USD"},
		},
		Status:    "active",
		Metadata:  map[string]any{"tier": "standard"},
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FeeSchedule
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "fs-001", decoded.ID)
	require.Len(t, decoded.Rules, 2)
	assert.Equal(t, "fixed", decoded.Rules[0].Type)
	require.NotNil(t, decoded.Rules[0].Amount)
	assert.Equal(t, int64(100), *decoded.Rules[0].Amount)
	assert.Equal(t, "percentage", decoded.Rules[1].Type)
	require.NotNil(t, decoded.Rules[1].Percentage)
	assert.Equal(t, "0.5", *decoded.Rules[1].Percentage)
}

func TestFeeSimulationResultJSON(t *testing.T) {
	t.Parallel()

	original := FeeSimulationResult{
		TotalFee: 350,
		Scale:    2,
		Currency: "USD",
		Breakdown: []FeeResult{
			{RuleType: "fixed", Amount: 100, Scale: 2, Currency: "USD"},
			{RuleType: "percentage", Amount: 250, Scale: 2, Currency: "USD"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FeeSimulationResult
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, int64(350), decoded.TotalFee)
	assert.Equal(t, 2, decoded.Scale)
	assert.Equal(t, "USD", decoded.Currency)
	require.Len(t, decoded.Breakdown, 2)
	assert.Equal(t, "fixed", decoded.Breakdown[0].RuleType)
	assert.Equal(t, int64(100), decoded.Breakdown[0].Amount)
}

func TestDisputeJSON(t *testing.T) {
	t.Parallel()

	resolvedAt := time.Date(2026, 3, 5, 16, 0, 0, 0, time.UTC)

	original := Dispute{
		ID:          "dsp-001",
		ContextID:   "ctx-001",
		ExceptionID: ptr("exc-001"),
		Reason:      "Amount should have matched with tolerance",
		Status:      "resolved",
		Resolution:  ptr("Approved after manual review"),
		ResolvedBy:  ptr("admin-01"),
		ResolvedAt:  &resolvedAt,
		Metadata:    map[string]any{"source": "internal"},
		CreatedAt:   time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 3, 5, 16, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Dispute
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "dsp-001", decoded.ID)
	require.NotNil(t, decoded.ExceptionID)
	assert.Equal(t, "exc-001", *decoded.ExceptionID)
	assert.Equal(t, "resolved", decoded.Status)
	require.NotNil(t, decoded.Resolution)
	assert.Equal(t, "Approved after manual review", *decoded.Resolution)
	require.NotNil(t, decoded.ResolvedAt)
	assert.True(t, resolvedAt.Equal(*decoded.ResolvedAt))
}

func TestExportJobJSON(t *testing.T) {
	t.Parallel()

	expires := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	original := ExportJob{
		ID:          "exp-001",
		ContextID:   "ctx-001",
		Status:      "completed",
		Format:      "csv",
		FileSize:    ptr(int64(1024000)),
		DownloadURL: ptr("https://storage.example.com/exports/exp-001.csv"),
		ExpiresAt:   &expires,
		CreatedAt:   time.Date(2026, 3, 15, 8, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 3, 15, 8, 5, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ExportJob
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "exp-001", decoded.ID)
	assert.Equal(t, "csv", decoded.Format)
	require.NotNil(t, decoded.FileSize)
	assert.Equal(t, int64(1024000), *decoded.FileSize)
	require.NotNil(t, decoded.DownloadURL)
	assert.Contains(t, *decoded.DownloadURL, "exp-001.csv")
}

func TestImportAndImportStatusJSON(t *testing.T) {
	t.Parallel()

	t.Run("Import round-trip", func(t *testing.T) {
		t.Parallel()

		original := Import{
			ID:          "imp-001",
			ContextID:   "ctx-001",
			Status:      "completed",
			FileName:    "transactions_202602.csv",
			RecordCount: ptr(5000),
			ErrorCount:  ptr(12),
			CreatedAt:   time.Date(2026, 3, 1, 6, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 3, 1, 6, 10, 0, 0, time.UTC),
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded Import
		require.NoError(t, json.Unmarshal(data, &decoded))

		assert.Equal(t, "imp-001", decoded.ID)
		assert.Equal(t, "transactions_202602.csv", decoded.FileName)
		require.NotNil(t, decoded.RecordCount)
		assert.Equal(t, 5000, *decoded.RecordCount)
		require.NotNil(t, decoded.ErrorCount)
		assert.Equal(t, 12, *decoded.ErrorCount)
	})

	t.Run("ImportStatus round-trip", func(t *testing.T) {
		t.Parallel()

		original := ImportStatus{
			ID:          "imp-001",
			Status:      "processing",
			Progress:    65,
			RecordCount: 3250,
			ErrorCount:  3,
			UpdatedAt:   time.Date(2026, 3, 1, 6, 5, 0, 0, time.UTC),
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded ImportStatus
		require.NoError(t, json.Unmarshal(data, &decoded))

		assert.Equal(t, 65, decoded.Progress)
		assert.Equal(t, 3250, decoded.RecordCount)
	})
}

// ---------------------------------------------------------------------------
// Analytics report types — additional coverage
// ---------------------------------------------------------------------------

func TestMatchRateReportJSON(t *testing.T) {
	t.Parallel()

	original := MatchRateReport{
		Period:   "2026-02",
		Overall:  0.94,
		BySource: map[string]float64{"bank-feed": 0.97, "erp": 0.91},
		ByRule:   map[string]float64{"exact-amount": 0.88, "fuzzy-date": 0.96},
		Trend: []TrendPoint{
			{Date: "2026-02-01", Value: 0.93},
			{Date: "2026-02-15", Value: 0.95},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded MatchRateReport
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.InDelta(t, 0.94, decoded.Overall, 1e-9)
	assert.InDelta(t, 0.97, decoded.BySource["bank-feed"], 1e-9)
	require.Len(t, decoded.Trend, 2)
	assert.Equal(t, "2026-02-01", decoded.Trend[0].Date)
}

func TestPerformanceMetricsReportJSON(t *testing.T) {
	t.Parallel()

	original := PerformanceMetricsReport{
		AvgDuration:      3200.5,
		P95Duration:      5400.0,
		P99Duration:      8100.0,
		TotalRuns:        150,
		SuccessRate:      0.993,
		AvgRecordsPerRun: 2500,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded PerformanceMetricsReport
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.InDelta(t, 3200.5, decoded.AvgDuration, 1e-9)
	assert.InDelta(t, 5400.0, decoded.P95Duration, 1e-9)
	assert.InDelta(t, 8100.0, decoded.P99Duration, 1e-9)
	assert.Equal(t, 150, decoded.TotalRuns)
	assert.InDelta(t, 0.993, decoded.SuccessRate, 1e-9)
	assert.Equal(t, 2500, decoded.AvgRecordsPerRun)
}

// ---------------------------------------------------------------------------
// Input types — representative round-trip coverage
// ---------------------------------------------------------------------------

func TestCreateContextInputJSON(t *testing.T) {
	t.Parallel()

	input := CreateContextInput{
		Name:        "Q1 Reconciliation",
		Description: ptr("Quarterly bank recon"),
		Config: &ContextConfig{
			MatchingStrategy: "fuzzy",
			AutoApprove:      false,
		},
		Metadata: map[string]any{"quarter": "Q1"},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded CreateContextInput
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "Q1 Reconciliation", decoded.Name)
	require.NotNil(t, decoded.Config)
	assert.Equal(t, "fuzzy", decoded.Config.MatchingStrategy)
	assert.False(t, decoded.Config.AutoApprove)
}

func TestUpdateContextInputOmitempty(t *testing.T) {
	t.Parallel()

	input := UpdateContextInput{}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	// Empty update should serialize to minimal JSON.
	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.NotContains(t, raw, "name")
	assert.NotContains(t, raw, "description")
	assert.NotContains(t, raw, "config")
	assert.NotContains(t, raw, "metadata")
}

func TestBulkReassignInputJSON(t *testing.T) {
	t.Parallel()

	input := BulkReassignInput{
		IDs:      []string{"exc-001", "exc-002", "exc-003"},
		AssignTo: "user-99",
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded BulkReassignInput
	require.NoError(t, json.Unmarshal(data, &decoded))

	require.Len(t, decoded.IDs, 3)
	assert.Equal(t, "exc-001", decoded.IDs[0])
	assert.Equal(t, "user-99", decoded.AssignTo)
}

func TestManualMatchInputJSON(t *testing.T) {
	t.Parallel()

	input := ManualMatchInput{
		ContextID:       "ctx-001",
		SourceRecordIDs: []string{"src-001", "src-002"},
		TargetRecordIDs: []string{"tgt-001", "tgt-002"},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "contextId")
	assert.Contains(t, raw, "sourceRecordIds")
	assert.Contains(t, raw, "targetRecordIds")

	var decoded ManualMatchInput
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "ctx-001", decoded.ContextID)
	require.Len(t, decoded.SourceRecordIDs, 2)
	require.Len(t, decoded.TargetRecordIDs, 2)
}

func TestSimulateFeeScheduleInputJSON(t *testing.T) {
	t.Parallel()

	input := SimulateFeeScheduleInput{
		Amount:   50000,
		Currency: "BRL",
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded SimulateFeeScheduleInput
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, int64(50000), decoded.Amount)
	assert.Equal(t, "BRL", decoded.Currency)
}
