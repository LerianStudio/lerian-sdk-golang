// models_input.go defines Create and Update input types for all Midaz entities.
//
// Create inputs use concrete types for required fields and pointer types for
// optional fields. Update inputs use pointer types throughout with
// json:",omitempty" so that only explicitly set fields are serialized in the
// PATCH request body.
package midaz

import "github.com/LerianStudio/lerian-sdk-golang/models"

// ---------------------------------------------------------------------------
// Organization
// ---------------------------------------------------------------------------

// CreateOrganizationInput holds the fields needed to create an organization.
type CreateOrganizationInput struct {
	LegalName            string          `json:"legalName"`
	LegalDocument        string          `json:"legalDocument"`
	ParentOrganizationID *string         `json:"parentOrganizationId,omitempty"`
	Address              *models.Address `json:"address,omitempty"`
	Status               *models.Status  `json:"status,omitempty"`
	Metadata             models.Metadata `json:"metadata,omitempty"`
}

// UpdateOrganizationInput holds the fields that may be updated on an
// existing organization. Only non-nil fields are sent in the PATCH request.
type UpdateOrganizationInput struct {
	LegalName            *string         `json:"legalName,omitempty"`
	LegalDocument        *string         `json:"legalDocument,omitempty"`
	ParentOrganizationID *string         `json:"parentOrganizationId,omitempty"`
	Address              *models.Address `json:"address,omitempty"`
	Status               *models.Status  `json:"status,omitempty"`
	Metadata             models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Ledger
// ---------------------------------------------------------------------------

// CreateLedgerInput holds the fields needed to create a ledger.
type CreateLedgerInput struct {
	Name     string          `json:"name"`
	Status   *models.Status  `json:"status,omitempty"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

// UpdateLedgerInput holds the fields that may be updated on an existing ledger.
type UpdateLedgerInput struct {
	Name     *string         `json:"name,omitempty"`
	Status   *models.Status  `json:"status,omitempty"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Account
// ---------------------------------------------------------------------------

// CreateAccountInput holds the fields needed to create an account.
type CreateAccountInput struct {
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	AssetCode       string          `json:"assetCode"`
	Alias           *string         `json:"alias,omitempty"`
	ExternalCode    *string         `json:"externalCode,omitempty"`
	PortfolioID     *string         `json:"portfolioId,omitempty"`
	SegmentID       *string         `json:"segmentId,omitempty"`
	ParentAccountID *string         `json:"parentAccountId,omitempty"`
	EntityID        *string         `json:"entityId,omitempty"`
	Status          *models.Status  `json:"status,omitempty"`
	Metadata        models.Metadata `json:"metadata,omitempty"`
}

// UpdateAccountInput holds the fields that may be updated on an existing account.
type UpdateAccountInput struct {
	Name            *string         `json:"name,omitempty"`
	Alias           *string         `json:"alias,omitempty"`
	ExternalCode    *string         `json:"externalCode,omitempty"`
	PortfolioID     *string         `json:"portfolioId,omitempty"`
	SegmentID       *string         `json:"segmentId,omitempty"`
	ParentAccountID *string         `json:"parentAccountId,omitempty"`
	EntityID        *string         `json:"entityId,omitempty"`
	Status          *models.Status  `json:"status,omitempty"`
	Metadata        models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// AccountType
// ---------------------------------------------------------------------------

// CreateAccountTypeInput holds the fields needed to create an account type.
type CreateAccountTypeInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// UpdateAccountTypeInput holds the fields that may be updated on an
// existing account type.
type UpdateAccountTypeInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ---------------------------------------------------------------------------
// Asset
// ---------------------------------------------------------------------------

// CreateAssetInput holds the fields needed to create an asset.
type CreateAssetInput struct {
	Name     string          `json:"name"`
	Code     string          `json:"code"`
	Type     string          `json:"type"`
	Status   *models.Status  `json:"status,omitempty"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

// UpdateAssetInput holds the fields that may be updated on an existing asset.
type UpdateAssetInput struct {
	Name     *string         `json:"name,omitempty"`
	Status   *models.Status  `json:"status,omitempty"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// AssetRate
// ---------------------------------------------------------------------------

// CreateAssetRateInput holds the fields needed to create an asset rate.
type CreateAssetRateInput struct {
	BaseAssetCode    string  `json:"baseAssetCode"`
	CounterAssetCode string  `json:"counterAssetCode"`
	Amount           int64   `json:"amount"`
	Scale            int     `json:"scale"`
	Source           *string `json:"source,omitempty"`
	ExternalID       *string `json:"externalId,omitempty"`
}

// UpdateAssetRateInput holds the fields that may be updated on an
// existing asset rate.
type UpdateAssetRateInput struct {
	Amount     *int64  `json:"amount,omitempty"`
	Scale      *int    `json:"scale,omitempty"`
	Source     *string `json:"source,omitempty"`
	ExternalID *string `json:"externalId,omitempty"`
}

// ---------------------------------------------------------------------------
// Balance
// ---------------------------------------------------------------------------

// CreateBalanceInput holds the fields needed to create a balance entry.
type CreateBalanceInput struct {
	AccountID      string          `json:"accountId"`
	AssetCode      string          `json:"assetCode"`
	AllowSending   bool            `json:"allowSending"`
	AllowReceiving bool            `json:"allowReceiving"`
	Status         *models.Status  `json:"status,omitempty"`
	Metadata       models.Metadata `json:"metadata,omitempty"`
}

// UpdateBalanceInput holds the fields that may be updated on an
// existing balance.
type UpdateBalanceInput struct {
	AllowSending   *bool           `json:"allowSending,omitempty"`
	AllowReceiving *bool           `json:"allowReceiving,omitempty"`
	Status         *models.Status  `json:"status,omitempty"`
	Metadata       models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Transaction
// ---------------------------------------------------------------------------

// CreateTransactionInput holds the fields needed to create a transaction.
// The Send/Distribute arrays define the money movement using the Midaz DSL.
type CreateTransactionInput struct {
	Description              *string             `json:"description,omitempty"`
	ChartOfAccountsGroupName *string             `json:"chartOfAccountsGroupName,omitempty"`
	AssetCode                string              `json:"assetCode"`
	Amount                   int64               `json:"amount"`
	Scale                    int                 `json:"scale"`
	Source                   []TransactionSource `json:"source"`
	ParentTransactionID      *string             `json:"parentTransactionId,omitempty"`
	Metadata                 models.Metadata     `json:"metadata,omitempty"`
}

// UpdateTransactionInput holds the fields that may be updated on an
// existing transaction.
type UpdateTransactionInput struct {
	Description *string         `json:"description,omitempty"`
	Status      *models.Status  `json:"status,omitempty"`
	Metadata    models.Metadata `json:"metadata,omitempty"`
}

// TransactionSource describes a single source-to-destination leg of a
// transaction, linking a sender account to a receiver account with an
// amount or percentage share.
type TransactionSource struct {
	From   TransactionFromTo  `json:"from"`
	To     TransactionFromTo  `json:"to"`
	Amount *TransactionAmount `json:"amount,omitempty"`
	Share  *TransactionShare  `json:"share,omitempty"`
}

// TransactionFromTo identifies an account in a transaction source leg.
type TransactionFromTo struct {
	Account         string  `json:"account"`
	AccountType     *string `json:"accountType,omitempty"`
	ChartOfAccounts *string `json:"chartOfAccounts,omitempty"`
}

// TransactionAmount specifies a fixed monetary amount with explicit scale
// and asset denomination.
type TransactionAmount struct {
	Amount int64  `json:"amount"`
	Scale  int    `json:"scale"`
	Asset  string `json:"asset"`
}

// TransactionShare specifies a percentage-based share for distributing
// a transaction amount.
type TransactionShare struct {
	Percentage   int     `json:"percentage"`
	PercentageOf *string `json:"percentageOf,omitempty"`
}

// ---------------------------------------------------------------------------
// Portfolio
// ---------------------------------------------------------------------------

// CreatePortfolioInput holds the fields needed to create a portfolio.
type CreatePortfolioInput struct {
	Name     string          `json:"name"`
	EntityID *string         `json:"entityId,omitempty"`
	Status   *models.Status  `json:"status,omitempty"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

// UpdatePortfolioInput holds the fields that may be updated on an
// existing portfolio.
type UpdatePortfolioInput struct {
	Name     *string         `json:"name,omitempty"`
	EntityID *string         `json:"entityId,omitempty"`
	Status   *models.Status  `json:"status,omitempty"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Segment
// ---------------------------------------------------------------------------

// CreateSegmentInput holds the fields needed to create a segment.
type CreateSegmentInput struct {
	Name     string          `json:"name"`
	Status   *models.Status  `json:"status,omitempty"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

// UpdateSegmentInput holds the fields that may be updated on an
// existing segment.
type UpdateSegmentInput struct {
	Name     *string         `json:"name,omitempty"`
	Status   *models.Status  `json:"status,omitempty"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// TransactionRoute
// ---------------------------------------------------------------------------

// CreateTransactionRouteInput holds the fields needed to create a
// transaction route.
type CreateTransactionRouteInput struct {
	TransactionType string          `json:"transactionType"`
	Description     *string         `json:"description,omitempty"`
	Code            *string         `json:"code,omitempty"`
	Metadata        models.Metadata `json:"metadata,omitempty"`
}

// UpdateTransactionRouteInput holds the fields that may be updated on
// an existing transaction route.
type UpdateTransactionRouteInput struct {
	Description *string         `json:"description,omitempty"`
	Code        *string         `json:"code,omitempty"`
	Metadata    models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// OperationRoute
// ---------------------------------------------------------------------------

// CreateOperationRouteInput holds the fields needed to create an
// operation route within a transaction route.
type CreateOperationRouteInput struct {
	AccountID    string          `json:"accountId"`
	AccountAlias *string         `json:"accountAlias,omitempty"`
	Type         string          `json:"type"`
	Description  *string         `json:"description,omitempty"`
	Metadata     models.Metadata `json:"metadata,omitempty"`
}

// UpdateOperationRouteInput holds the fields that may be updated on
// an existing operation route.
type UpdateOperationRouteInput struct {
	AccountID    *string         `json:"accountId,omitempty"`
	AccountAlias *string         `json:"accountAlias,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Metadata     models.Metadata `json:"metadata,omitempty"`
}
