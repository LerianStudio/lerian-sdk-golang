package fees

import "time"

// ---------------------------------------------------------------------------
// Entity types
// ---------------------------------------------------------------------------

// Package represents a fee package configuration. A package groups one or
// more [Fee] definitions that are evaluated together when calculating
// fees for a transaction. Packages are scoped to an organization and
// optionally to a specific ledger and segment.
type Package struct {
	ID               string         `json:"id"`
	FeeGroupLabel    string         `json:"feeGroupLabel"`
	Description      *string        `json:"description,omitempty"`
	SegmentID        *string        `json:"segmentId,omitempty"`
	LedgerID         string         `json:"ledgerId"`
	TransactionRoute *string        `json:"transactionRoute,omitempty"`
	MinimumAmount    string         `json:"minimumAmount"`
	MaximumAmount    string         `json:"maximumAmount"`
	WaivedAccounts   *[]string      `json:"waivedAccounts,omitempty"`
	Fees             map[string]Fee `json:"fees"`
	Enable           *bool          `json:"enable"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        *time.Time     `json:"deletedAt,omitempty"`
}

// Fee defines a single fee rule within a [Package]. The CalculationModel
// determines how the fee amount is computed (flat, percentage, or
// max-between-types). Priority controls evaluation order, and
// IsDeductibleFrom indicates whether the fee is subtracted from the
// original transaction amount.
type Fee struct {
	FeeLabel         string            `json:"feeLabel"`
	CalculationModel *CalculationModel `json:"calculationModel"`
	ReferenceAmount  string            `json:"referenceAmount"`
	Priority         int               `json:"priority,omitempty"`
	IsDeductibleFrom *bool             `json:"isDeductibleFrom"`
	CreditAccount    string            `json:"creditAccount"`
	RouteFrom        *string           `json:"routeFrom,omitempty"`
	RouteTo          *string           `json:"routeTo,omitempty"`
}

// CalculationModel defines the rule engine for computing a fee amount.
// The ApplicationRule field selects the strategy:
//   - "flatFee": uses a single Calculation of type "flat"
//   - "percentual": uses a single Calculation of type "percentage"
//   - "maxBetweenTypes": evaluates multiple Calculations and takes the maximum
type CalculationModel struct {
	ApplicationRule string        `json:"applicationRule"`
	Calculations    []Calculation `json:"calculations"`
}

// Calculation holds the type and value for a single fee computation step.
// Type is either "flat" (fixed amount) or "percentage" (proportion of
// the transaction amount). Value is a decimal string (e.g. "100.00" or "2.5").
type Calculation struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ---------------------------------------------------------------------------
// Input types
// ---------------------------------------------------------------------------

// CreatePackageInput is the request payload for creating a new fee package.
type CreatePackageInput struct {
	FeeGroupLabel    string         `json:"feeGroupLabel"`
	Description      *string        `json:"description,omitempty"`
	SegmentID        *string        `json:"segmentId,omitempty"`
	LedgerID         string         `json:"ledgerId"`
	TransactionRoute *string        `json:"transactionRoute,omitempty"`
	MinimumAmount    string         `json:"minimumAmount"`
	MaximumAmount    string         `json:"maximumAmount"`
	WaivedAccounts   *[]string      `json:"waivedAccounts,omitempty"`
	Fees             map[string]Fee `json:"fees"`
	Enable           *bool          `json:"enable"`
}

// UpdatePackageInput is the request payload for partially updating an
// existing fee package. Only non-nil/non-empty fields are patched.
type UpdatePackageInput struct {
	FeeGroupLabel  string         `json:"feeGroupLabel,omitempty"`
	Description    string         `json:"description,omitempty"`
	MinimumAmount  *string        `json:"minimumAmount,omitempty"`
	MaximumAmount  *string        `json:"maximumAmount,omitempty"`
	WaivedAccounts *[]string      `json:"waivedAccounts,omitempty"`
	Fees           map[string]Fee `json:"fees,omitempty"`
	Enable         *bool          `json:"enable,omitempty"`
}

// FeeCalculate is the request and response payload for the fee calculation
// endpoint (POST /fees). The service evaluates matching fee packages and
// mutates the transaction DSL by injecting fee legs into the source and
// distribute arrays.
type FeeCalculate struct {
	SegmentID   *string        `json:"segmentId,omitempty"`
	LedgerID    string         `json:"ledgerId"`
	Transaction TransactionDSL `json:"transaction"`
}

// FeeEstimateInput is the request payload for the fee estimation endpoint
// (POST /estimates). Unlike [FeeCalculate], the caller specifies a
// PackageID explicitly to preview fees for a given transaction.
type FeeEstimateInput struct {
	PackageID   string         `json:"packageId"`
	LedgerID    string         `json:"ledgerId"`
	Transaction TransactionDSL `json:"transaction"`
}

// FeeEstimateResponse is the response payload from the fee estimation
// endpoint. When no matching fee rules are found, Message describes the
// situation and FeesApplied is nil.
type FeeEstimateResponse struct {
	Message     string        `json:"message"`
	FeesApplied *FeeCalculate `json:"feesApplied"`
}

// ---------------------------------------------------------------------------
// List types
// ---------------------------------------------------------------------------

// PackageListOptions configures filtering and page-based listing for fee
// packages. The SDK exposes package semantics here and maps them internally
// to whatever query parameter format the backend expects.
type PackageListOptions struct {
	SegmentID        string     `json:"segmentId,omitempty"`
	LedgerID         string     `json:"ledgerId,omitempty"`
	TransactionRoute string     `json:"transactionRoute,omitempty"`
	Enabled          *bool      `json:"enabled,omitempty"`
	PageNumber       int        `json:"pageNumber,omitempty"`
	PageSize         int        `json:"pageSize,omitempty"`
	SortOrder        string     `json:"sortOrder,omitempty"`
	CreatedFrom      *time.Time `json:"createdFrom,omitempty"`
	CreatedTo        *time.Time `json:"createdTo,omitempty"`
}

// PackagePage is the normalized paginated response for fee packages.
type PackagePage struct {
	Items      []Package `json:"items"`
	PageNumber int       `json:"pageNumber"`
	PageSize   int       `json:"pageSize"`
	TotalItems int       `json:"totalItems"`
	TotalPages int       `json:"totalPages"`
}

// ---------------------------------------------------------------------------
// Transaction DSL types
// ---------------------------------------------------------------------------

// TransactionDSL represents the transaction DSL shape used by the fees
// service. The service may mutate this structure by adding fee legs and
// metadata to the source and distribute arrays.
type TransactionDSL struct {
	ChartOfAccountsGroupName string             `json:"chartOfAccountsGroupName,omitempty"`
	Description              string             `json:"description,omitempty"`
	Code                     string             `json:"code,omitempty"`
	Pending                  bool               `json:"pending,omitempty"`
	Metadata                 map[string]any     `json:"metadata,omitempty"`
	Route                    string             `json:"route,omitempty"`
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
	AccountAlias    string                `json:"accountAlias,omitempty"`
	BalanceKey      string                `json:"balanceKey,omitempty"`
	Amount          *TransactionDSLAmount `json:"amount,omitempty"`
	Share           *TransactionDSLShare  `json:"share,omitempty"`
	Remaining       string                `json:"remaining,omitempty"`
	Rate            *TransactionDSLRate   `json:"rate,omitempty"`
	Route           string                `json:"route,omitempty"`
	Description     string                `json:"description,omitempty"`
	ChartOfAccounts string                `json:"chartOfAccounts,omitempty"`
	Metadata        map[string]any        `json:"metadata,omitempty"`
	IsFrom          bool                  `json:"isFrom,omitempty"`
}

// TransactionDSLAmount specifies the asset and value for a leg.
type TransactionDSLAmount struct {
	Asset           string `json:"asset"`
	Operation       string `json:"operation,omitempty"`
	TransactionType string `json:"transactionType,omitempty"`
	Value           any    `json:"value"`
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
