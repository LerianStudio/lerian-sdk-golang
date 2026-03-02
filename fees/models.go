package fees

import "time"

// ---------------------------------------------------------------------------
// Entity types
// ---------------------------------------------------------------------------

// Package represents a fee package configuration. A package groups one or
// more [FeeRule] definitions that are evaluated together when calculating
// fees for a transaction.
type Package struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Status      string         `json:"status"`
	Rules       []FeeRule      `json:"rules"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// FeeRule defines a single fee calculation rule within a [Package].
// The Type field determines which value fields are relevant:
//   - "flat": uses Amount
//   - "percentage": uses Percentage
//   - "tiered": uses both Amount and Percentage with MinAmount/MaxAmount caps
type FeeRule struct {
	Type       string  `json:"type"`                 // "flat", "percentage", "tiered"
	Amount     *int64  `json:"amount,omitempty"`      // for flat fees (in smallest unit)
	Percentage *string `json:"percentage,omitempty"`  // for percentage fees (e.g., "2.5")
	MinAmount  *int64  `json:"minAmount,omitempty"`   // minimum fee cap
	MaxAmount  *int64  `json:"maxAmount,omitempty"`   // maximum fee cap
	Currency   string  `json:"currency"`
	AssetCode  *string `json:"assetCode,omitempty"`
}

// Estimate represents a fee estimation result. Estimates are computed
// without being associated with a real transaction and are useful for
// previewing fee charges before committing.
type Estimate struct {
	ID            string      `json:"id"`
	PackageID     string      `json:"packageId"`
	Amount        int64       `json:"amount"`
	Scale         int         `json:"scale"`
	Currency      string      `json:"currency"`
	FeeResults    []FeeResult `json:"feeResults"`
	TotalFee      int64       `json:"totalFee"`
	TotalFeeScale int         `json:"totalFeeScale"`
	CreatedAt     time.Time   `json:"createdAt"`
}

// Fee represents a calculated fee result that may be linked to an actual
// transaction. Unlike [Estimate], a Fee tracks its lifecycle via a Status
// field and optionally carries a TransactionID.
type Fee struct {
	ID            string      `json:"id"`
	PackageID     string      `json:"packageId"`
	TransactionID *string     `json:"transactionId,omitempty"`
	Amount        int64       `json:"amount"`
	Scale         int         `json:"scale"`
	Currency      string      `json:"currency"`
	FeeResults    []FeeResult `json:"feeResults"`
	TotalFee      int64       `json:"totalFee"`
	TotalFeeScale int         `json:"totalFeeScale"`
	Status        string      `json:"status"`
	CreatedAt     time.Time   `json:"createdAt"`
}

// FeeResult represents the result of applying a single fee rule during
// fee calculation. Each result corresponds to one [FeeRule] in the
// package and indicates whether the rule was applied along with the
// computed amount.
type FeeResult struct {
	RuleType string `json:"ruleType"`
	Amount   int64  `json:"amount"`
	Scale    int    `json:"scale"`
	Currency string `json:"currency"`
	Applied  bool   `json:"applied"`
	Reason   string `json:"reason,omitempty"`
}

// ---------------------------------------------------------------------------
// Input types
// ---------------------------------------------------------------------------

// CreatePackageInput is the request payload for creating a new fee package.
type CreatePackageInput struct {
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Rules       []FeeRule      `json:"rules"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdatePackageInput is the request payload for partially updating an
// existing fee package. Only non-nil/non-empty fields are patched.
type UpdatePackageInput struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Rules       []FeeRule      `json:"rules,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CalculateEstimateInput is the request payload for an RPC-style fee
// estimation. The caller provides the package, amount, and currency to
// receive a preview of the fees that would be charged.
type CalculateEstimateInput struct {
	PackageID string `json:"packageId"`
	Amount    int64  `json:"amount"`
	Scale     int    `json:"scale"`
	Currency  string `json:"currency"`
	AssetCode string `json:"assetCode,omitempty"`
}

// CalculateFeeInput is the request payload for an RPC-style fee
// calculation. Similar to [CalculateEstimateInput] but optionally links
// the resulting fee to a transaction via TransactionID.
type CalculateFeeInput struct {
	PackageID     string  `json:"packageId"`
	TransactionID *string `json:"transactionId,omitempty"`
	Amount        int64   `json:"amount"`
	Scale         int     `json:"scale"`
	Currency      string  `json:"currency"`
	AssetCode     string  `json:"assetCode,omitempty"`
}
