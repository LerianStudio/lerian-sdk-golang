// models_input.go defines Create, Update, and action input types for all
// Matcher entities.
//
// Create inputs use concrete types for required fields and pointer types for
// optional fields. Update inputs use pointer types throughout with
// json:",omitempty" so that only explicitly set fields are serialized in the
// PATCH request body. Action inputs (reject, reassign, resolve, simulate, etc.)
// carry the minimum payload for the corresponding RPC-style endpoint.
package matcher

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

// CreateContextInput holds the fields needed to create a reconciliation context.
type CreateContextInput struct {
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Config      *ContextConfig `json:"config,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateContextInput holds the fields that may be updated on an existing
// reconciliation context. Only non-nil fields are sent in the PATCH request.
type UpdateContextInput struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Config      *ContextConfig `json:"config,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Rule
// ---------------------------------------------------------------------------

// CreateRuleInput holds the fields needed to create a matching rule.
type CreateRuleInput struct {
	ContextID   string         `json:"contextId"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Priority    int            `json:"priority"`
	Expression  string         `json:"expression"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateRuleInput holds the fields that may be updated on an existing
// matching rule. Only non-nil fields are sent in the PATCH request.
type UpdateRuleInput struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Priority    *int           `json:"priority,omitempty"`
	Expression  *string        `json:"expression,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ReorderRulesInput provides an ordered list of rule IDs to reorder
// the evaluation priority of rules within a context.
type ReorderRulesInput struct {
	RuleIDs []string `json:"ruleIds"`
}

// ---------------------------------------------------------------------------
// Schedule
// ---------------------------------------------------------------------------

// CreateScheduleInput holds the fields needed to create a reconciliation
// schedule.
type CreateScheduleInput struct {
	ContextID string `json:"contextId"`
	Name      string `json:"name"`
	CronExpr  string `json:"cronExpr"`
}

// UpdateScheduleInput holds the fields that may be updated on an existing
// schedule. Only non-nil fields are sent in the PATCH request.
type UpdateScheduleInput struct {
	Name     *string `json:"name,omitempty"`
	CronExpr *string `json:"cronExpr,omitempty"`
}

// ---------------------------------------------------------------------------
// Source
// ---------------------------------------------------------------------------

// CreateSourceInput holds the fields needed to create a data source.
type CreateSourceInput struct {
	ContextID string         `json:"contextId"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Config    map[string]any `json:"config,omitempty"`
}

// UpdateSourceInput holds the fields that may be updated on an existing
// data source. Only non-nil fields are sent in the PATCH request.
type UpdateSourceInput struct {
	Name   *string        `json:"name,omitempty"`
	Config map[string]any `json:"config,omitempty"`
}

// ---------------------------------------------------------------------------
// SourceFieldMap
// ---------------------------------------------------------------------------

// CreateSourceFieldMapInput holds the fields needed to create a source
// field mapping.
type CreateSourceFieldMapInput struct {
	SourceID  string  `json:"sourceId"`
	FieldName string  `json:"fieldName"`
	MappedTo  string  `json:"mappedTo"`
	Transform *string `json:"transform,omitempty"`
}

// UpdateSourceFieldMapInput holds the fields that may be updated on an
// existing source field mapping. Only non-nil fields are sent in the
// PATCH request.
type UpdateSourceFieldMapInput struct {
	MappedTo  *string `json:"mappedTo,omitempty"`
	Transform *string `json:"transform,omitempty"`
}

// ---------------------------------------------------------------------------
// FeeSchedule
// ---------------------------------------------------------------------------

// CreateFeeScheduleInput holds the fields needed to create a fee schedule.
type CreateFeeScheduleInput struct {
	ContextID   string         `json:"contextId"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Rules       []FeeRule      `json:"rules"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateFeeScheduleInput holds the fields that may be updated on an existing
// fee schedule. Only non-nil fields are sent in the PATCH request.
type UpdateFeeScheduleInput struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Rules       []FeeRule      `json:"rules,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// SimulateFeeScheduleInput holds the parameters for simulating a fee
// schedule against a given amount and currency.
type SimulateFeeScheduleInput struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// ---------------------------------------------------------------------------
// FieldMap
// ---------------------------------------------------------------------------

// CreateFieldMapInput holds the fields needed to create a field mapping.
type CreateFieldMapInput struct {
	ContextID   string  `json:"contextId"`
	SourceField string  `json:"sourceField"`
	TargetField string  `json:"targetField"`
	Transform   *string `json:"transform,omitempty"`
}

// UpdateFieldMapInput holds the fields that may be updated on an existing
// field mapping. Only non-nil fields are sent in the PATCH request.
type UpdateFieldMapInput struct {
	TargetField *string `json:"targetField,omitempty"`
	Transform   *string `json:"transform,omitempty"`
}

// ---------------------------------------------------------------------------
// Exception
// ---------------------------------------------------------------------------

// CreateExceptionInput holds the fields needed to create a reconciliation
// exception.
type CreateExceptionInput struct {
	ContextID   string         `json:"contextId"`
	Type        string         `json:"type"`
	Priority    string         `json:"priority"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateExceptionInput holds the fields that may be updated on an existing
// exception. Only non-nil fields are sent in the PATCH request.
type UpdateExceptionInput struct {
	Priority    *string        `json:"priority,omitempty"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// RejectExceptionInput holds the reason for rejecting a single exception.
type RejectExceptionInput struct {
	Reason string `json:"reason"`
}

// ReassignExceptionInput holds the target assignee for reassigning a
// single exception.
type ReassignExceptionInput struct {
	AssignTo string `json:"assignTo"`
}

// BulkExceptionInput holds a list of exception IDs for bulk approval.
type BulkExceptionInput struct {
	IDs []string `json:"ids"`
}

// BulkRejectInput holds a list of exception IDs and a shared rejection
// reason for bulk rejection.
type BulkRejectInput struct {
	IDs    []string `json:"ids"`
	Reason string   `json:"reason"`
}

// BulkReassignInput holds a list of exception IDs and a target assignee
// for bulk reassignment.
type BulkReassignInput struct {
	IDs      []string `json:"ids"`
	AssignTo string   `json:"assignTo"`
}

// ---------------------------------------------------------------------------
// Dispute
// ---------------------------------------------------------------------------

// CreateDisputeInput holds the fields needed to create a dispute.
type CreateDisputeInput struct {
	ContextID   string         `json:"contextId"`
	ExceptionID *string        `json:"exceptionId,omitempty"`
	Reason      string         `json:"reason"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateDisputeInput holds the fields that may be updated on an existing
// dispute. Only non-nil fields are sent in the PATCH request.
type UpdateDisputeInput struct {
	Reason   *string        `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ResolveDisputeInput holds the resolution text for closing a dispute.
type ResolveDisputeInput struct {
	Resolution string `json:"resolution"`
}

// ---------------------------------------------------------------------------
// ExportJob
// ---------------------------------------------------------------------------

// CreateExportJobInput holds the fields needed to create a data export job.
type CreateExportJobInput struct {
	ContextID string `json:"contextId"`
	Format    string `json:"format"`
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

// CreateImportInput holds the fields needed to create a data import job.
type CreateImportInput struct {
	ContextID string `json:"contextId"`
	FileName  string `json:"fileName"`
}

// ---------------------------------------------------------------------------
// Matching (RPC-style actions)
// ---------------------------------------------------------------------------

// ManualMatchInput holds the parameters for manually matching source
// records to target records within a reconciliation context.
type ManualMatchInput struct {
	ContextID       string   `json:"contextId"`
	SourceRecordIDs []string `json:"sourceRecordIds"`
	TargetRecordIDs []string `json:"targetRecordIds"`
}

// AdjustmentInput holds the parameters for creating a manual monetary
// adjustment within a reconciliation context.
type AdjustmentInput struct {
	ContextID string `json:"contextId"`
	Type      string `json:"type"`
	Amount    int64  `json:"amount"`
	Reason    string `json:"reason"`
}
