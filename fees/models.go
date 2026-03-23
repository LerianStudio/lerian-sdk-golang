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
	Amount     *int64  `json:"amount,omitempty"`     // for flat fees (in smallest unit)
	Percentage *string `json:"percentage,omitempty"` // for percentage fees (e.g., "2.5")
	MinAmount  *int64  `json:"minAmount,omitempty"`  // minimum fee cap
	MaxAmount  *int64  `json:"maxAmount,omitempty"`  // maximum fee cap
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

// TransformTransactionInput is the request payload for DSL-based fee
// transformation. The fees service injects fee legs into the transaction DSL
// and returns the mutated structure.
type TransformTransactionInput struct {
	SegmentID   *string        `json:"segmentId,omitempty"`
	LedgerID    string         `json:"ledgerId"`
	Transaction TransactionDSL `json:"transaction"`
}

// TransformTransactionOutput is the response from a DSL fee transformation.
type TransformTransactionOutput struct {
	Transaction TransactionDSL `json:"transaction"`
}

// TransactionDSL represents the transaction DSL shape used by the fees
// transformation endpoint. The service may mutate this structure by adding
// fee legs and metadata to the source and distribute arrays.
type TransactionDSL struct {
	ChartOfAccountsGroupName string             `json:"chartOfAccountsGroupName,omitempty"`
	Description              string             `json:"description,omitempty"`
	Code                     string             `json:"code,omitempty"`
	Pending                  bool               `json:"pending"`
	Metadata                 map[string]any     `json:"metadata,omitempty"`
	Route                    string             `json:"route,omitempty"`
	RouteID                  *string            `json:"routeId,omitempty"`
	TransactionDate          any                `json:"transactionDate,omitempty"`
	Send                     TransactionDSLSend `json:"send"`
}

// TransactionDSLSend describes the send portion of the DSL.
type TransactionDSLSend struct {
	Asset      string                   `json:"asset"`
	Value      any                      `json:"value"`
	Source     TransactionDSLSource     `json:"source"`
	Distribute TransactionDSLDistribute `json:"distribute"`
}

// TransactionDSLSource contains the debit legs.
type TransactionDSLSource struct {
	Remaining string              `json:"remaining,omitempty"`
	From      []TransactionDSLLeg `json:"from"`
}

// TransactionDSLDistribute contains the credit legs.
type TransactionDSLDistribute struct {
	Remaining string              `json:"remaining,omitempty"`
	To        []TransactionDSLLeg `json:"to"`
}

// TransactionDSLLeg is a single source or distribute entry.
type TransactionDSLLeg struct {
	Account         string                `json:"account,omitempty"`
	AccountAlias    string                `json:"accountAlias,omitempty"`
	BalanceKey      string                `json:"balanceKey,omitempty"`
	Amount          *TransactionDSLAmount `json:"amount,omitempty"`
	Share           *TransactionDSLShare  `json:"share,omitempty"`
	Remaining       string                `json:"remaining,omitempty"`
	Rate            *TransactionDSLRate   `json:"rate,omitempty"`
	Route           string                `json:"route,omitempty"`
	RouteID         *string               `json:"routeId,omitempty"`
	Description     string                `json:"description,omitempty"`
	ChartOfAccounts string                `json:"chartOfAccounts,omitempty"`
	Metadata        map[string]any        `json:"metadata,omitempty"`
	IsFrom          bool                  `json:"isFrom,omitempty"`
}

// TransactionDSLAmount specifies the asset and value for a leg.
type TransactionDSLAmount struct {
	Asset string `json:"asset"`
	Value any    `json:"value"`
}

// TransactionDSLShare specifies percentage-based distribution details for a leg.
type TransactionDSLShare struct {
	Percentage             int64 `json:"percentage,omitempty"`
	PercentageOfPercentage int64 `json:"percentageOfPercentage,omitempty"`
}

// TransactionDSLRate represents the rate metadata associated with a leg.
type TransactionDSLRate struct {
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Value      any    `json:"value,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
}
