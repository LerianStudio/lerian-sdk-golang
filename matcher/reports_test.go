package matcher

import (
	"context"
	"errors"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noopCallFn is a convenience helper for tests that only exercise guard
// clauses and never reach the backend.
func noopCallFn(_ context.Context, _, _ string, _, _ any) error { return nil }

// ---------------------------------------------------------------------------
// Table-driven tests for all single-result report methods
// ---------------------------------------------------------------------------

func TestReportsGetMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		expectedPath string
		call         func(svc ReportsService, ctx context.Context) (any, error)
		response     any
	}{
		{
			name:         "GetSummary",
			expectedPath: "/contexts/ctx-1/reports/summary",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetSummary(ctx, "ctx-1")
			},
			response: ReconciliationSummary{Period: "2026-01", TotalRecords: 1000, MatchRate: 0.95},
		},
		{
			name:         "GetMatchRate",
			expectedPath: "/contexts/ctx-1/reports/match-rate",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetMatchRate(ctx, "ctx-1")
			},
			response: MatchRateReport{Period: "2026-01", Overall: 0.92},
		},
		{
			name:         "GetExceptionTrend",
			expectedPath: "/contexts/ctx-1/reports/exception-trend",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetExceptionTrend(ctx, "ctx-1")
			},
			response: ExceptionTrendReport{Period: "2026-01"},
		},
		{
			name:         "GetAgingAnalysis",
			expectedPath: "/contexts/ctx-1/reports/aging-analysis",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetAgingAnalysis(ctx, "ctx-1")
			},
			response: AgingAnalysisReport{Total: 42, AvgAge: 3.5},
		},
		{
			name:         "GetSourceComparison",
			expectedPath: "/contexts/ctx-1/reports/source-comparison",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetSourceComparison(ctx, "ctx-1")
			},
			response: SourceComparisonReport{Sources: []SourceMetrics{{SourceID: "s-1", Name: "Bank Feed"}}},
		},
		{
			name:         "GetVolumeAnalysis",
			expectedPath: "/contexts/ctx-1/reports/volume-analysis",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetVolumeAnalysis(ctx, "ctx-1")
			},
			response: VolumeAnalysisReport{Period: "2026-01", Total: 5000},
		},
		{
			name:         "GetDisputeMetrics",
			expectedPath: "/contexts/ctx-1/reports/dispute-metrics",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetDisputeMetrics(ctx, "ctx-1")
			},
			response: DisputeMetricsReport{Total: 7, AvgResolutionTime: 24.5},
		},
		{
			name:         "GetFeeAnalysis",
			expectedPath: "/contexts/ctx-1/reports/fee-analysis",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetFeeAnalysis(ctx, "ctx-1")
			},
			response: FeeAnalysisReport{TotalFees: 150000, Scale: 2, Currency: "USD"},
		},
		{
			name:         "GetPerformanceMetrics",
			expectedPath: "/contexts/ctx-1/reports/performance-metrics",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetPerformanceMetrics(ctx, "ctx-1")
			},
			response: PerformanceMetricsReport{TotalRuns: 100, SuccessRate: 0.99},
		},
		{
			name:         "GetDashboard",
			expectedPath: "/contexts/ctx-1/reports/dashboard",
			call: func(svc ReportsService, ctx context.Context) (any, error) {
				return svc.GetDashboard(ctx, "ctx-1")
			},
			response: DashboardReport{MatchRate: 0.95},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			t.Parallel()

			resp := tt.response
			mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, body, result any) error {
				assert.Equal(t, "GET", method)
				assert.Equal(t, tt.expectedPath, path)
				assert.Nil(t, body)
				unmarshalInto(t, resp, result)
				return nil
			}}

			svc := newReportsService(mb)
			got, err := tt.call(svc, context.Background())
			require.NoError(t, err)
			require.NotNil(t, got)
		})

		t.Run(tt.name+"_empty_contextID", func(t *testing.T) {
			t.Parallel()

			svc := newReportsService(&mockBackend{t: t, callFn: noopCallFn})

			// Call with empty contextID — should use "" to trigger guard.
			callEmpty := func() (any, error) {
				switch tt.name {
				case "GetSummary":
					return svc.GetSummary(context.Background(), "")
				case "GetMatchRate":
					return svc.GetMatchRate(context.Background(), "")
				case "GetExceptionTrend":
					return svc.GetExceptionTrend(context.Background(), "")
				case "GetAgingAnalysis":
					return svc.GetAgingAnalysis(context.Background(), "")
				case "GetSourceComparison":
					return svc.GetSourceComparison(context.Background(), "")
				case "GetVolumeAnalysis":
					return svc.GetVolumeAnalysis(context.Background(), "")
				case "GetDisputeMetrics":
					return svc.GetDisputeMetrics(context.Background(), "")
				case "GetFeeAnalysis":
					return svc.GetFeeAnalysis(context.Background(), "")
				case "GetPerformanceMetrics":
					return svc.GetPerformanceMetrics(context.Background(), "")
				case "GetDashboard":
					return svc.GetDashboard(context.Background(), "")
				default:
					t.Fatalf("unhandled method: %s", tt.name)
					return nil, nil
				}
			}

			got, err := callEmpty()
			require.Error(t, err)
			assert.Nil(t, got)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
		})
	}
}

// ---------------------------------------------------------------------------
// GetReconciliationHistory — paginated iterator
// ---------------------------------------------------------------------------

func TestReportsGetReconciliationHistory(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{t: t, callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/contexts/ctx-1/reports/reconciliation-history")

		resp := models.ListResponse[ReconciliationHistoryEntry]{
			Items: []ReconciliationHistoryEntry{
				{ID: "run-1", ContextID: "ctx-1", Status: "completed"},
				{ID: "run-2", ContextID: "ctx-1", Status: "completed"},
			},
			Pagination: models.Pagination{Total: 2, Limit: 10},
		}
		unmarshalInto(t, resp, result)

		return nil
	}}

	svc := newReportsService(mb)
	iter := svc.GetReconciliationHistory(context.Background(), "ctx-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "run-1", items[0].ID)
	assert.Equal(t, "run-2", items[1].ID)
}

func TestReportsGetReconciliationHistoryWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, path string, _, result any) error {
		receivedPath = path

		resp := models.ListResponse[ReconciliationHistoryEntry]{
			Items:      []ReconciliationHistoryEntry{{ID: "run-1"}},
			Pagination: models.Pagination{Total: 1, Limit: 5},
		}
		unmarshalInto(t, resp, result)

		return nil
	}}

	svc := newReportsService(mb)
	opts := &models.ListOptions{Limit: 5}
	iter := svc.GetReconciliationHistory(context.Background(), "ctx-1", opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=5")
}

// ---------------------------------------------------------------------------
// Backend error propagation
// ---------------------------------------------------------------------------

func TestReportsGetSummaryBackendError(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return errors.New("network timeout")
	}}

	svc := newReportsService(mb)
	got, err := svc.GetSummary(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "network timeout")
}

func TestReportsGetMatchRateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetMatchRate(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetExceptionTrendBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetExceptionTrend(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetAgingAnalysisBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetAgingAnalysis(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetSourceComparisonBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetSourceComparison(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetVolumeAnalysisBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetVolumeAnalysis(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetDisputeMetricsBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetDisputeMetrics(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetFeeAnalysisBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetFeeAnalysis(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetPerformanceMetricsBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetPerformanceMetrics(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetDashboardBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	got, err := svc.GetDashboard(context.Background(), "ctx-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func TestReportsGetReconciliationHistoryBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{t: t, callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newReportsService(mb)
	iter := svc.GetReconciliationHistory(context.Background(), "ctx-1", nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ ReportsService = (*reportsService)(nil)
