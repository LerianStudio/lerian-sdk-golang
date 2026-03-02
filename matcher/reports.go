// reports.go implements the ReportsService for accessing analytics and
// reporting endpoints within a reconciliation context. All methods are
// read-only and scoped to a specific context ID.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// ReportsService provides analytics and reporting endpoints for
// reconciliation dashboards, match rates, exception trends, and more.
// All methods are read-only and scoped to a reconciliation context.
type ReportsService interface {
	// GetSummary returns a high-level reconciliation summary for the context.
	GetSummary(ctx context.Context, contextID string) (*ReconciliationSummary, error)

	// GetMatchRate returns match rate analytics broken down by source and rule.
	GetMatchRate(ctx context.Context, contextID string) (*MatchRateReport, error)

	// GetExceptionTrend returns exception count trends over time.
	GetExceptionTrend(ctx context.Context, contextID string) (*ExceptionTrendReport, error)

	// GetAgingAnalysis returns an aging analysis of unresolved exceptions.
	GetAgingAnalysis(ctx context.Context, contextID string) (*AgingAnalysisReport, error)

	// GetSourceComparison compares key metrics across data sources.
	GetSourceComparison(ctx context.Context, contextID string) (*SourceComparisonReport, error)

	// GetVolumeAnalysis returns record volume trends over time.
	GetVolumeAnalysis(ctx context.Context, contextID string) (*VolumeAnalysisReport, error)

	// GetDisputeMetrics returns aggregate dispute statistics.
	GetDisputeMetrics(ctx context.Context, contextID string) (*DisputeMetricsReport, error)

	// GetFeeAnalysis returns fee activity summaries and trends.
	GetFeeAnalysis(ctx context.Context, contextID string) (*FeeAnalysisReport, error)

	// GetReconciliationHistory returns a paginated iterator over historical
	// reconciliation runs.
	GetReconciliationHistory(ctx context.Context, contextID string, opts *models.ListOptions) *pagination.Iterator[ReconciliationHistoryEntry]

	// GetPerformanceMetrics returns latency percentiles and throughput metrics.
	GetPerformanceMetrics(ctx context.Context, contextID string) (*PerformanceMetricsReport, error)

	// GetDashboard returns a composite dashboard view combining summary,
	// match rate, exception stats, and recent runs.
	GetDashboard(ctx context.Context, contextID string) (*DashboardReport, error)
}

// reportsService is the concrete implementation of [ReportsService].
type reportsService struct {
	core.BaseService
}

// newReportsService creates a new [ReportsService] backed by the given
// [core.Backend].
func newReportsService(backend core.Backend) ReportsService {
	return &reportsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ ReportsService = (*reportsService)(nil)

// GetSummary returns a high-level reconciliation summary for the context.
func (s *reportsService) GetSummary(ctx context.Context, contextID string) (*ReconciliationSummary, error) {
	const operation = "Reports.GetSummary"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[ReconciliationSummary](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/summary")
}

// GetMatchRate returns match rate analytics broken down by source and rule.
func (s *reportsService) GetMatchRate(ctx context.Context, contextID string) (*MatchRateReport, error) {
	const operation = "Reports.GetMatchRate"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[MatchRateReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/match-rate")
}

// GetExceptionTrend returns exception count trends over time.
func (s *reportsService) GetExceptionTrend(ctx context.Context, contextID string) (*ExceptionTrendReport, error) {
	const operation = "Reports.GetExceptionTrend"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[ExceptionTrendReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/exception-trend")
}

// GetAgingAnalysis returns an aging analysis of unresolved exceptions.
func (s *reportsService) GetAgingAnalysis(ctx context.Context, contextID string) (*AgingAnalysisReport, error) {
	const operation = "Reports.GetAgingAnalysis"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[AgingAnalysisReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/aging-analysis")
}

// GetSourceComparison compares key metrics across data sources.
func (s *reportsService) GetSourceComparison(ctx context.Context, contextID string) (*SourceComparisonReport, error) {
	const operation = "Reports.GetSourceComparison"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[SourceComparisonReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/source-comparison")
}

// GetVolumeAnalysis returns record volume trends over time.
func (s *reportsService) GetVolumeAnalysis(ctx context.Context, contextID string) (*VolumeAnalysisReport, error) {
	const operation = "Reports.GetVolumeAnalysis"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[VolumeAnalysisReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/volume-analysis")
}

// GetDisputeMetrics returns aggregate dispute statistics.
func (s *reportsService) GetDisputeMetrics(ctx context.Context, contextID string) (*DisputeMetricsReport, error) {
	const operation = "Reports.GetDisputeMetrics"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[DisputeMetricsReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/dispute-metrics")
}

// GetFeeAnalysis returns fee activity summaries and trends.
func (s *reportsService) GetFeeAnalysis(ctx context.Context, contextID string) (*FeeAnalysisReport, error) {
	const operation = "Reports.GetFeeAnalysis"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[FeeAnalysisReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/fee-analysis")
}

// GetReconciliationHistory returns a paginated iterator over historical
// reconciliation runs.
func (s *reportsService) GetReconciliationHistory(ctx context.Context, contextID string, opts *models.ListOptions) *pagination.Iterator[ReconciliationHistoryEntry] {
	if contextID == "" {
		return pagination.NewIterator[ReconciliationHistoryEntry](func(_ context.Context, _ string) ([]ReconciliationHistoryEntry, string, error) {
			return nil, "", sdkerrors.NewValidation("Reports.GetReconciliationHistory", "Report", "context ID is required")
		})
	}

	return core.List[ReconciliationHistoryEntry](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/reconciliation-history", opts)
}

// GetPerformanceMetrics returns latency percentiles and throughput metrics.
func (s *reportsService) GetPerformanceMetrics(ctx context.Context, contextID string) (*PerformanceMetricsReport, error) {
	const operation = "Reports.GetPerformanceMetrics"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[PerformanceMetricsReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/performance-metrics")
}

// GetDashboard returns a composite dashboard view combining summary,
// match rate, exception stats, and recent runs.
func (s *reportsService) GetDashboard(ctx context.Context, contextID string) (*DashboardReport, error) {
	const operation = "Reports.GetDashboard"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "Report", "contextID is required")
	}

	return core.Get[DashboardReport](ctx, &s.BaseService, "/contexts/"+url.PathEscape(contextID)+"/reports/dashboard")
}
