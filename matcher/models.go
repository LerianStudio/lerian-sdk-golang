// models.go defines all Matcher entity types as returned by the Matcher API.
// Fields use camelCase JSON tags to match the API contract. Nullable fields
// are represented as pointer types (*string, *time.Time, *int64).
//
// The Matcher service has the largest model surface across all Lerian products,
// covering reconciliation contexts, rules, schedules, sources, field mappings,
// fee schedules, exceptions, disputes, exports, imports, and a rich set of
// analytics and reporting types.
package matcher

import "time"

// ---------------------------------------------------------------------------
// Core entity types
// ---------------------------------------------------------------------------

// Context represents a reconciliation context in the Matcher service.
// Contexts are the top-level scope for all reconciliation operations,
// rules, sources, and related configuration.
type Context struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Status      string         `json:"status"`
	Config      *ContextConfig `json:"config,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// ContextConfig holds the matching strategy and related settings for a
// reconciliation [Context]. It controls how records are compared and
// whether results require manual approval.
type ContextConfig struct {
	MatchingStrategy string         `json:"matchingStrategy"`
	Tolerance        *float64       `json:"tolerance,omitempty"`
	AutoApprove      bool           `json:"autoApprove"`
	Settings         map[string]any `json:"settings,omitempty"`
}

// Rule represents a matching rule within a reconciliation context.
// Rules define expressions that are evaluated against record pairs to
// determine whether they match, ordered by priority.
type Rule struct {
	ID          string         `json:"id"`
	ContextID   string         `json:"contextId"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Priority    int            `json:"priority"`
	Expression  string         `json:"expression"`
	Status      string         `json:"status"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// Schedule represents a recurring reconciliation schedule attached to a
// context. Schedules use cron expressions to define when automated
// reconciliation runs are triggered.
type Schedule struct {
	ID        string     `json:"id"`
	ContextID string     `json:"contextId"`
	Name      string     `json:"name"`
	CronExpr  string     `json:"cronExpr"`
	Status    string     `json:"status"`
	LastRunAt *time.Time `json:"lastRunAt,omitempty"`
	NextRunAt *time.Time `json:"nextRunAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// Source represents a data source connected to a reconciliation context.
// Sources define where record data comes from (e.g., bank feed, ERP system)
// and how it is accessed.
type Source struct {
	ID        string         `json:"id"`
	ContextID string         `json:"contextId"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Config    map[string]any `json:"config,omitempty"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// SourceFieldMap defines how a field from a [Source] maps to the
// canonical reconciliation schema. Transforms can be applied to
// normalize data during import.
type SourceFieldMap struct {
	ID        string    `json:"id"`
	SourceID  string    `json:"sourceId"`
	FieldName string    `json:"fieldName"`
	MappedTo  string    `json:"mappedTo"`
	Transform *string   `json:"transform,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FeeSchedule represents a fee configuration within a reconciliation
// context. Fee schedules contain one or more [FeeRule] entries that
// define how fees are calculated for matched records.
type FeeSchedule struct {
	ID          string         `json:"id"`
	ContextID   string         `json:"contextId"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Rules       []FeeRule      `json:"rules"`
	Status      string         `json:"status"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// FeeRule defines a single fee calculation rule within a [FeeSchedule].
// A rule can be percentage-based, fixed-amount, or a combination, and
// is denominated in a specific currency.
type FeeRule struct {
	Type       string  `json:"type"`
	Amount     *int64  `json:"amount,omitempty"`
	Percentage *string `json:"percentage,omitempty"`
	Currency   string  `json:"currency"`
}

// FieldMap defines a mapping between source and target fields within a
// reconciliation context. Field maps control how data is aligned across
// different sources for comparison.
type FieldMap struct {
	ID          string    `json:"id"`
	ContextID   string    `json:"contextId"`
	SourceField string    `json:"sourceField"`
	TargetField string    `json:"targetField"`
	Transform   *string   `json:"transform,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ExportJob represents an asynchronous data export job. Exports produce
// downloadable files in the requested format (CSV, XLSX, etc.) containing
// reconciliation data.
type ExportJob struct {
	ID          string     `json:"id"`
	ContextID   string     `json:"contextId"`
	Status      string     `json:"status"`
	Format      string     `json:"format"`
	FileSize    *int64     `json:"fileSize,omitempty"`
	DownloadURL *string    `json:"downloadUrl,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// Dispute represents a formal dispute raised against a reconciliation
// result or exception. Disputes track the resolution workflow from
// creation through final resolution.
type Dispute struct {
	ID          string         `json:"id"`
	ContextID   string         `json:"contextId"`
	ExceptionID *string        `json:"exceptionId,omitempty"`
	Reason      string         `json:"reason"`
	Status      string         `json:"status"`
	Resolution  *string        `json:"resolution,omitempty"`
	ResolvedBy  *string        `json:"resolvedBy,omitempty"`
	ResolvedAt  *time.Time     `json:"resolvedAt,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// Exception represents an anomaly or discrepancy detected during
// reconciliation that requires human review. Exceptions are triaged
// by type, priority, and assignment.
type Exception struct {
	ID          string         `json:"id"`
	ContextID   string         `json:"contextId"`
	Type        string         `json:"type"`
	Status      string         `json:"status"`
	Priority    string         `json:"priority"`
	Description *string        `json:"description,omitempty"`
	AssignedTo  *string        `json:"assignedTo,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// Archive represents an archived batch of reconciliation records.
// Archives are created when historical data is moved out of the
// active working set.
type Archive struct {
	ID          string    `json:"id"`
	ContextID   string    `json:"contextId"`
	Type        string    `json:"type"`
	RecordCount int       `json:"recordCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

// AuditLog represents a single audit trail entry within the Matcher
// service. Every significant action (create, update, approve, reject)
// is logged with the actor, resource, and details.
type AuditLog struct {
	ID         string         `json:"id"`
	ContextID  string         `json:"contextId"`
	Action     string         `json:"action"`
	Actor      string         `json:"actor"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resourceId"`
	Details    map[string]any `json:"details,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// Import represents a data import job that loads records from an
// external file into the Matcher service for reconciliation.
type Import struct {
	ID          string    `json:"id"`
	ContextID   string    `json:"contextId"`
	Status      string    `json:"status"`
	FileName    string    `json:"fileName"`
	RecordCount *int      `json:"recordCount,omitempty"`
	ErrorCount  *int      `json:"errorCount,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ImportStatus provides progress information for an in-flight [Import]
// job, including record and error counts and completion percentage.
type ImportStatus struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	RecordCount int       `json:"recordCount"`
	ErrorCount  int       `json:"errorCount"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ---------------------------------------------------------------------------
// Result and analytics types
// ---------------------------------------------------------------------------

// MatchResult captures the outcome of a single reconciliation run,
// including counts of matched, unmatched, and exception records along
// with the run duration in milliseconds.
type MatchResult struct {
	ID             string         `json:"id"`
	ContextID      string         `json:"contextId"`
	Status         string         `json:"status"`
	MatchedCount   int            `json:"matchedCount"`
	UnmatchedCount int            `json:"unmatchedCount"`
	ExceptionCount int            `json:"exceptionCount"`
	Duration       int64          `json:"duration"` // milliseconds
	Details        map[string]any `json:"details,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
}

// BulkExceptionResult reports the outcome of a bulk exception operation
// (approve, reject, reassign), indicating how many items succeeded or
// failed.
type BulkExceptionResult struct {
	Processed int      `json:"processed"`
	Succeeded int      `json:"succeeded"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

// FeeSimulationResult is the response from simulating a fee schedule
// against a given amount. It includes the total computed fee and a
// breakdown by individual [FeeResult] entries.
type FeeSimulationResult struct {
	TotalFee  int64       `json:"totalFee"`
	Scale     int         `json:"scale"`
	Currency  string      `json:"currency"`
	Breakdown []FeeResult `json:"breakdown"`
}

// FeeResult represents a single fee computation within a
// [FeeSimulationResult] breakdown.
type FeeResult struct {
	RuleType string `json:"ruleType"`
	Amount   int64  `json:"amount"`
	Scale    int    `json:"scale"`
	Currency string `json:"currency"`
}

// Adjustment represents a manual monetary adjustment applied within a
// reconciliation context to correct discrepancies.
type Adjustment struct {
	ID        string         `json:"id"`
	ContextID string         `json:"contextId"`
	Type      string         `json:"type"`
	Amount    int64          `json:"amount"`
	Reason    string         `json:"reason"`
	Status    string         `json:"status"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

// ExceptionStatistics provides aggregate statistics about exceptions
// within a reconciliation context, broken down by status, priority,
// and type.
type ExceptionStatistics struct {
	Total      int            `json:"total"`
	ByStatus   map[string]int `json:"byStatus"`
	ByPriority map[string]int `json:"byPriority"`
	ByType     map[string]int `json:"byType"`
	AvgAge     float64        `json:"avgAge"` // hours
}

// CursorPagination holds next/previous cursor tokens for paginated
// list responses in the Matcher API.
type CursorPagination struct {
	Next *string `json:"next,omitempty"`
	Prev *string `json:"prev,omitempty"`
}

// ---------------------------------------------------------------------------
// Analytics report types
// ---------------------------------------------------------------------------

// ReconciliationSummary provides a high-level overview of reconciliation
// activity for a given period, including total record counts and the
// computed match rate.
type ReconciliationSummary struct {
	Period           string  `json:"period"`
	TotalRecords     int     `json:"totalRecords"`
	MatchedRecords   int     `json:"matchedRecords"`
	UnmatchedRecords int     `json:"unmatchedRecords"`
	ExceptionRecords int     `json:"exceptionRecords"`
	MatchRate        float64 `json:"matchRate"`
}

// MatchRateReport breaks down the match rate by source and rule for a
// given period, with a historical trend.
type MatchRateReport struct {
	Period   string             `json:"period"`
	Overall  float64            `json:"overall"`
	BySource map[string]float64 `json:"bySource"`
	ByRule   map[string]float64 `json:"byRule"`
	Trend    []TrendPoint       `json:"trend"`
}

// TrendPoint represents a single data point in a time-series trend,
// associating a date string with a numeric value.
type TrendPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// ExceptionTrendReport tracks exception counts over time, with an
// overall trend and per-type breakdowns.
type ExceptionTrendReport struct {
	Period string                  `json:"period"`
	Trend  []TrendPoint            `json:"trend"`
	ByType map[string][]TrendPoint `json:"byType"`
}

// AgingAnalysisReport categorizes unresolved exceptions into age
// buckets, providing visibility into the backlog distribution.
type AgingAnalysisReport struct {
	Buckets []AgingBucket `json:"buckets"`
	Total   int           `json:"total"`
	AvgAge  float64       `json:"avgAge"`
}

// AgingBucket represents a single bucket in an [AgingAnalysisReport],
// grouping exceptions by age range.
type AgingBucket struct {
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// SourceComparisonReport compares key metrics across all data sources
// within a reconciliation context.
type SourceComparisonReport struct {
	Sources []SourceMetrics `json:"sources"`
}

// SourceMetrics holds per-source performance indicators used in a
// [SourceComparisonReport].
type SourceMetrics struct {
	SourceID    string  `json:"sourceId"`
	Name        string  `json:"name"`
	RecordCount int     `json:"recordCount"`
	MatchRate   float64 `json:"matchRate"`
	ErrorRate   float64 `json:"errorRate"`
}

// VolumeAnalysisReport tracks record volumes over time, broken down
// by source.
type VolumeAnalysisReport struct {
	Period   string         `json:"period"`
	Total    int            `json:"total"`
	Trend    []TrendPoint   `json:"trend"`
	BySource map[string]int `json:"bySource"`
}

// DisputeMetricsReport provides aggregate dispute statistics including
// resolution times and status breakdowns.
type DisputeMetricsReport struct {
	Total             int            `json:"total"`
	ByStatus          map[string]int `json:"byStatus"`
	AvgResolutionTime float64        `json:"avgResolutionTime"`
	Trend             []TrendPoint   `json:"trend"`
}

// FeeAnalysisReport summarizes fee activity for a reconciliation
// context, with breakdowns by fee type and a historical trend.
type FeeAnalysisReport struct {
	TotalFees int64            `json:"totalFees"`
	Scale     int              `json:"scale"`
	Currency  string           `json:"currency"`
	ByType    map[string]int64 `json:"byType"`
	Trend     []TrendPoint     `json:"trend"`
}

// ReconciliationHistoryEntry represents a single historical
// reconciliation run with its outcome statistics.
type ReconciliationHistoryEntry struct {
	ID             string    `json:"id"`
	ContextID      string    `json:"contextId"`
	RunAt          time.Time `json:"runAt"`
	Duration       int64     `json:"duration"`
	MatchedCount   int       `json:"matchedCount"`
	UnmatchedCount int       `json:"unmatchedCount"`
	ExceptionCount int       `json:"exceptionCount"`
	Status         string    `json:"status"`
}

// PerformanceMetricsReport captures latency percentiles, throughput,
// and success rates across reconciliation runs.
type PerformanceMetricsReport struct {
	AvgDuration      float64 `json:"avgDuration"`
	P95Duration      float64 `json:"p95Duration"`
	P99Duration      float64 `json:"p99Duration"`
	TotalRuns        int     `json:"totalRuns"`
	SuccessRate      float64 `json:"successRate"`
	AvgRecordsPerRun int     `json:"avgRecordsPerRun"`
}

// DashboardReport is a composite view combining the reconciliation
// summary, overall match rate, exception statistics, and recent
// reconciliation history for a single-screen dashboard.
type DashboardReport struct {
	Summary    ReconciliationSummary      `json:"summary"`
	MatchRate  float64                    `json:"matchRate"`
	Exceptions ExceptionStatistics        `json:"exceptions"`
	RecentRuns []ReconciliationHistoryEntry `json:"recentRuns"`
}
