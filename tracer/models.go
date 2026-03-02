package tracer

import "time"

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

// RuleStatus represents the lifecycle state of a compliance rule.
type RuleStatus string

const (
	// RuleStatusDraft indicates the rule is being authored and is not yet enforced.
	RuleStatusDraft RuleStatus = "DRAFT"

	// RuleStatusActive indicates the rule is live and actively evaluated.
	RuleStatusActive RuleStatus = "ACTIVE"

	// RuleStatusInactive indicates the rule has been disabled.
	RuleStatusInactive RuleStatus = "INACTIVE"
)

// ---------------------------------------------------------------------------
// Entity types
// ---------------------------------------------------------------------------

// Rule represents a compliance or business rule managed by the Tracer service.
// Rules contain a set of conditions that are evaluated against transactions
// and other resources to determine whether actions should be triggered.
type Rule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Status      RuleStatus        `json:"status"`
	Priority    int               `json:"priority"`
	Conditions  []RuleCondition   `json:"conditions"`
	Actions     []string          `json:"actions,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// RuleCondition defines a single evaluation criterion within a [Rule].
// Each condition compares a resource field against a value using the
// specified operator (e.g., "eq", "gt", "contains").
type RuleCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// Limit represents a rate or amount limit managed by the Tracer service.
// Limits enforce caps on operations such as maximum transaction amounts
// per period, daily transfer counts, and similar constraints.
type Limit struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Status      RuleStatus     `json:"status"` // reuses same status enum
	Type        string         `json:"type"`
	MaxAmount   int64          `json:"maxAmount"`
	Period      string         `json:"period"` // e.g., "daily", "monthly"
	Currency    *string        `json:"currency,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// Validation represents the result of evaluating a transaction (or other
// resource) against the set of active rules and limits. It captures which
// rules were applied and any violations detected.
type Validation struct {
	ID           string         `json:"id"`
	Status       string         `json:"status"`
	Result       string         `json:"result"` // "approved", "rejected", "pending"
	RulesApplied []string       `json:"rulesApplied,omitempty"`
	Violations   []string       `json:"violations,omitempty"`
	Transaction  map[string]any `json:"transaction,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
}

// AuditEvent represents a single auditable action recorded by the Tracer
// service. Every mutation in the system is tracked as an audit event with
// a timestamped record of who did what to which resource.
type AuditEvent struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Action     string         `json:"action"`
	Actor      string         `json:"actor"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resourceId"`
	Details    map[string]any `json:"details,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// AuditVerification represents the integrity verification of an [AuditEvent].
// It confirms whether the event's content hash matches the expected value,
// providing tamper-evidence for the audit trail.
type AuditVerification struct {
	ID         string    `json:"id"`
	EventID    string    `json:"eventId"`
	Valid      bool      `json:"valid"`
	Hash       string    `json:"hash"`
	VerifiedAt time.Time `json:"verifiedAt"`
}

// ---------------------------------------------------------------------------
// Input types
// ---------------------------------------------------------------------------

// CreateValidationInput is the request payload for creating a new validation.
// It submits a transaction (as a freeform map) along with optional rule IDs
// to scope which rules are evaluated.
type CreateValidationInput struct {
	Transaction map[string]any `json:"transaction"`
	RuleIDs     []string       `json:"ruleIds,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CreateRuleInput is the request payload for creating a new compliance rule.
type CreateRuleInput struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Priority    int             `json:"priority"`
	Conditions  []RuleCondition `json:"conditions"`
	Actions     []string        `json:"actions,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// UpdateRuleInput is the request payload for partially updating an existing
// compliance rule. Only non-nil/non-empty fields are patched.
type UpdateRuleInput struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Priority    *int            `json:"priority,omitempty"`
	Conditions  []RuleCondition `json:"conditions,omitempty"`
	Actions     []string        `json:"actions,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// CreateLimitInput is the request payload for creating a new limit.
type CreateLimitInput struct {
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Type        string         `json:"type"`
	MaxAmount   int64          `json:"maxAmount"`
	Period      string         `json:"period"`
	Currency    *string        `json:"currency,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateLimitInput is the request payload for partially updating an existing
// limit. Only non-nil/non-empty fields are patched.
type UpdateLimitInput struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	MaxAmount   *int64         `json:"maxAmount,omitempty"`
	Period      *string        `json:"period,omitempty"`
	Currency    *string        `json:"currency,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
