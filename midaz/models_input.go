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

// CreateBalanceInput holds the fields needed to create an additional balance.
type CreateBalanceInput struct {
	Key string `json:"key"`

	// AllowSending and AllowReceiving are pointers so omitted fields stay omitted
	// and the server can apply its default behavior.
	AllowSending   *bool           `json:"allowSending,omitempty"`
	AllowReceiving *bool           `json:"allowReceiving,omitempty"`
	Status         *models.Status  `json:"-"`
	Metadata       models.Metadata `json:"-"`
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
// Transactions use the canonical send-based contract via the Send field.
type CreateTransactionInput struct {
	Description              *string          `json:"description,omitempty"`
	Code                     *string          `json:"code,omitempty"`
	ChartOfAccountsGroupName *string          `json:"chartOfAccountsGroupName,omitempty"`
	Pending                  *bool            `json:"pending,omitempty"`
	Route                    *string          `json:"route,omitempty"`
	TransactionDate          *string          `json:"transactionDate,omitempty"`
	Send                     *TransactionSend `json:"send,omitempty"`
	Metadata                 models.Metadata  `json:"metadata,omitempty"`
	ParentTransactionID      *string          `json:"-"`
}

// TransactionSend represents the Midaz send-based transaction payload.
type TransactionSend struct {
	Asset      string                      `json:"asset"`
	Value      string                      `json:"value"`
	Source     TransactionSendSource       `json:"source"`
	Distribute TransactionSendDistribution `json:"distribute"`
}

// TransactionSendSource contains the debit/source legs for a transaction.
type TransactionSendSource struct {
	Remaining string                    `json:"remaining,omitempty"`
	From      []TransactionOperationLeg `json:"from,omitempty"`
}

// TransactionSendDistribution contains the credit/distribution legs.
type TransactionSendDistribution struct {
	Remaining string                    `json:"remaining,omitempty"`
	To        []TransactionOperationLeg `json:"to,omitempty"`
}

// CreateTransactionInflowInput holds the fields needed to create an inflow
// transaction, which only defines distribution targets.
type CreateTransactionInflowInput struct {
	Description              *string               `json:"description,omitempty"`
	Code                     *string               `json:"code,omitempty"`
	ChartOfAccountsGroupName *string               `json:"chartOfAccountsGroupName,omitempty"`
	Route                    *string               `json:"route,omitempty"`
	TransactionDate          *string               `json:"transactionDate,omitempty"`
	Send                     TransactionInflowSend `json:"send"`
	Metadata                 models.Metadata       `json:"metadata,omitempty"`
}

// TransactionInflowSend contains the asset, amount value, and distribution
// legs for an inflow transaction.
type TransactionInflowSend struct {
	Asset      string                        `json:"asset"`
	Value      string                        `json:"value"`
	Distribute TransactionInflowDistribution `json:"distribute"`
}

// TransactionInflowDistribution describes the destination operations of an
// inflow transaction.
type TransactionInflowDistribution struct {
	Remaining string                    `json:"remaining,omitempty"`
	To        []TransactionOperationLeg `json:"to"`
}

// CreateTransactionOutflowInput holds the fields needed to create an outflow
// transaction, which only defines source legs.
type CreateTransactionOutflowInput struct {
	Description              *string                `json:"description,omitempty"`
	Code                     *string                `json:"code,omitempty"`
	ChartOfAccountsGroupName *string                `json:"chartOfAccountsGroupName,omitempty"`
	Route                    *string                `json:"route,omitempty"`
	TransactionDate          *string                `json:"transactionDate,omitempty"`
	Pending                  *bool                  `json:"pending,omitempty"`
	Send                     TransactionOutflowSend `json:"send"`
	Metadata                 models.Metadata        `json:"metadata,omitempty"`
}

// TransactionOutflowSend contains the asset, amount value, and source legs
// for an outflow transaction.
type TransactionOutflowSend struct {
	Asset  string                   `json:"asset"`
	Value  string                   `json:"value"`
	Source TransactionOutflowSource `json:"source"`
}

// TransactionOutflowSource describes the source operations of an outflow
// transaction.
type TransactionOutflowSource struct {
	Remaining string                    `json:"remaining,omitempty"`
	From      []TransactionOperationLeg `json:"from"`
}

// TransactionOperationLeg identifies one operation leg in inflow/outflow
// transaction payloads.
type TransactionOperationLeg struct {
	AccountAlias    string                `json:"accountAlias,omitempty"`
	BalanceKey      string                `json:"balanceKey,omitempty"`
	Amount          *TransactionLegAmount `json:"amount,omitempty"`
	Share           *TransactionLegShare  `json:"share,omitempty"`
	Remaining       string                `json:"remaining,omitempty"`
	Rate            *TransactionLegRate   `json:"rate,omitempty"`
	Metadata        models.Metadata       `json:"metadata,omitempty"`
	ChartOfAccounts *string               `json:"chartOfAccounts,omitempty"`
	Description     *string               `json:"description,omitempty"`
	Route           string                `json:"route,omitempty"`
	IsFrom          bool                  `json:"isFrom,omitempty"`
}

// TransactionLegAmount specifies the asset and value of a transaction leg.
type TransactionLegAmount struct {
	Asset string `json:"asset"`
	Value string `json:"value"`
}

// TransactionLegShare specifies percentage-based distribution details.
type TransactionLegShare struct {
	Percentage             int64  `json:"percentage,omitempty"`
	PercentageOfPercentage *int64 `json:"percentageOfPercentage,omitempty"`
}

// TransactionLegRate specifies conversion-rate metadata for a leg.
type TransactionLegRate struct {
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Value      string `json:"value,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
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
// Operation
// ---------------------------------------------------------------------------

// UpdateOperationInput contains the mutable fields for updating an operation.
// Only non-nil fields are sent in the PATCH request.
type UpdateOperationInput struct {
	Description *string         `json:"description,omitempty"`
	Metadata    models.Metadata `json:"metadata,omitempty"`
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

// ---------------------------------------------------------------------------
// Holder (CRM)
// ---------------------------------------------------------------------------

// CreateHolderInput contains the fields for creating a new holder.
type CreateHolderInput struct {
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Document      string          `json:"document"`
	ExternalID    *string         `json:"externalId,omitempty"`
	Addresses     *HolderAddress  `json:"addresses,omitempty"`
	Contact       *Contact        `json:"contact,omitempty"`
	NaturalPerson *NaturalPerson  `json:"naturalPerson,omitempty"`
	LegalPerson   *LegalPerson    `json:"legalPerson,omitempty"`
	Metadata      models.Metadata `json:"metadata,omitempty"`
}

// UpdateHolderInput contains the mutable fields for updating a holder.
// Only non-nil fields are sent in the PATCH request.
type UpdateHolderInput struct {
	Name          *string         `json:"name,omitempty"`
	ExternalID    *string         `json:"externalId,omitempty"`
	Addresses     *HolderAddress  `json:"addresses,omitempty"`
	Contact       *Contact        `json:"contact,omitempty"`
	NaturalPerson *NaturalPerson  `json:"naturalPerson,omitempty"`
	LegalPerson   *LegalPerson    `json:"legalPerson,omitempty"`
	Metadata      models.Metadata `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Alias (CRM)
// ---------------------------------------------------------------------------

// CreateAliasInput contains the fields for creating a new alias account
// linking a holder to a ledger account.
type CreateAliasInput struct {
	LedgerID         string            `json:"ledgerId"`
	AccountID        string            `json:"accountId"`
	BankingDetails   *BankingDetails   `json:"bankingDetails,omitempty"`
	RegulatoryFields *RegulatoryFields `json:"regulatoryFields,omitempty"`
	RelatedParties   []RelatedParty    `json:"relatedParties,omitempty"`
	Metadata         models.Metadata   `json:"metadata,omitempty"`
}

// UpdateAliasInput contains the mutable fields for updating an alias.
// Only non-nil fields are sent in the PATCH request.
type UpdateAliasInput struct {
	BankingDetails   *BankingDetails   `json:"bankingDetails,omitempty"`
	RegulatoryFields *RegulatoryFields `json:"regulatoryFields,omitempty"`
	RelatedParties   *[]RelatedParty   `json:"relatedParties,omitempty"`
	Metadata         models.Metadata   `json:"metadata,omitempty"`
}
