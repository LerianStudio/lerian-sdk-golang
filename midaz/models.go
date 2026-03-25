// models.go defines all Midaz entity types as returned by the Midaz API.
// Fields use camelCase JSON tags to match the API contract. Nullable fields
// are represented as pointer types (*string, *time.Time, *int64).
package midaz

import (
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
)

// Organization represents a top-level organizational entity in Midaz.
// Organizations are the root scope for ledgers, accounts, and all other
// financial domain objects.
type Organization struct {
	ID                   string          `json:"id"`
	LegalName            string          `json:"legalName"`
	LegalDocument        string          `json:"legalDocument"`
	ParentOrganizationID *string         `json:"parentOrganizationId,omitempty"`
	Status               models.Status   `json:"status"`
	Address              *models.Address `json:"address,omitempty"`
	Metadata             models.Metadata `json:"metadata,omitempty"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
	DeletedAt            *time.Time      `json:"deletedAt,omitempty"`
}

// Ledger represents an isolated double-entry ledger within an organization.
// All accounts, transactions, and balances belong to exactly one ledger.
type Ledger struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	Name           string          `json:"name"`
	Status         models.Status   `json:"status"`
	Metadata       models.Metadata `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	DeletedAt      *time.Time      `json:"deletedAt,omitempty"`
}

// Account represents a financial account within a ledger.
// Accounts hold balances denominated in a specific asset and participate
// in transactions via operations.
type Account struct {
	ID              string          `json:"id"`
	OrganizationID  string          `json:"organizationId"`
	LedgerID        string          `json:"ledgerId"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	AssetCode       string          `json:"assetCode"`
	Alias           *string         `json:"alias,omitempty"`
	ExternalCode    *string         `json:"externalCode,omitempty"`
	PortfolioID     *string         `json:"portfolioId,omitempty"`
	SegmentID       *string         `json:"segmentId,omitempty"`
	ParentAccountID *string         `json:"parentAccountId,omitempty"`
	EntityID        *string         `json:"entityId,omitempty"`
	Status          models.Status   `json:"status"`
	Metadata        models.Metadata `json:"metadata,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	DeletedAt       *time.Time      `json:"deletedAt,omitempty"`
}

// AccountType defines a classification for accounts within a ledger
// (e.g., "deposit", "savings", "external").
type AccountType struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	LedgerID       string    `json:"ledgerId"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// Asset represents a tradable instrument or currency within a ledger
// (e.g., "BRL", "USD", "BTC").
type Asset struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	LedgerID       string          `json:"ledgerId"`
	Name           string          `json:"name"`
	Code           string          `json:"code"`
	Type           string          `json:"type"`
	Status         models.Status   `json:"status"`
	Metadata       models.Metadata `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// AssetRate represents an exchange rate between two assets at a point in time.
type AssetRate struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organizationId"`
	LedgerID         string    `json:"ledgerId"`
	BaseAssetCode    string    `json:"baseAssetCode"`
	CounterAssetCode string    `json:"counterAssetCode"`
	Amount           int64     `json:"amount"`
	Scale            int       `json:"scale"`
	Source           *string   `json:"source,omitempty"`
	ExternalID       *string   `json:"externalId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Balance represents the current balance state of an account for a
// specific asset. Available and OnHold are stored as integers in the
// smallest denomination (e.g., cents), with Scale indicating the
// decimal precision.
type Balance struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	LedgerID       string          `json:"ledgerId"`
	AccountID      string          `json:"accountId"`
	AssetCode      string          `json:"assetCode"`
	Available      int64           `json:"available"`
	OnHold         int64           `json:"onHold"`
	Scale          int             `json:"scale"`
	AccountAlias   *string         `json:"accountAlias,omitempty"`
	AllowSending   bool            `json:"allowSending"`
	AllowReceiving bool            `json:"allowReceiving"`
	Status         models.Status   `json:"status"`
	Metadata       models.Metadata `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// Transaction represents an atomic financial movement composed of one or
// more operations. Amounts are stored as integers with an explicit scale
// factor (e.g., amount=15050, amountScale=2 represents 150.50).
type Transaction struct {
	ID                       string          `json:"id"`
	OrganizationID           string          `json:"organizationId"`
	LedgerID                 string          `json:"ledgerId"`
	Description              *string         `json:"description,omitempty"`
	Status                   models.Status   `json:"status"`
	Amount                   int64           `json:"amount"`
	AmountScale              int             `json:"amountScale"`
	AssetCode                string          `json:"assetCode"`
	ChartOfAccountsGroupName *string         `json:"chartOfAccountsGroupName,omitempty"`
	ParentTransactionID      *string         `json:"parentTransactionId,omitempty"`
	Metadata                 models.Metadata `json:"metadata,omitempty"`
	CreatedAt                time.Time       `json:"createdAt"`
	UpdatedAt                time.Time       `json:"updatedAt"`
	DeletedAt                *time.Time      `json:"deletedAt,omitempty"`
}

// Operation represents a single debit or credit leg within a transaction.
// Each operation affects exactly one account.
type Operation struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	LedgerID       string          `json:"ledgerId"`
	TransactionID  string          `json:"transactionId"`
	AccountID      string          `json:"accountId"`
	AccountAlias   *string         `json:"accountAlias,omitempty"`
	Type           string          `json:"type"`
	AssetCode      string          `json:"assetCode"`
	Amount         int64           `json:"amount"`
	AmountScale    int             `json:"amountScale"`
	Status         models.Status   `json:"status"`
	BalanceAfter   *int64          `json:"balanceAfter,omitempty"`
	Description    *string         `json:"description,omitempty"`
	Metadata       models.Metadata `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// Portfolio groups related accounts under a single logical unit.
type Portfolio struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	LedgerID       string          `json:"ledgerId"`
	Name           string          `json:"name"`
	EntityID       *string         `json:"entityId,omitempty"`
	Status         models.Status   `json:"status"`
	Metadata       models.Metadata `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	DeletedAt      *time.Time      `json:"deletedAt,omitempty"`
}

// Segment is a classification or grouping within a ledger, used to
// organize accounts into logical segments for reporting and access control.
type Segment struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	LedgerID       string          `json:"ledgerId"`
	Name           string          `json:"name"`
	Status         models.Status   `json:"status"`
	Metadata       models.Metadata `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	DeletedAt      *time.Time      `json:"deletedAt,omitempty"`
}

// TransactionRoute defines a routing template that governs how
// transactions of a given type are processed.
type TransactionRoute struct {
	ID              string          `json:"id"`
	OrganizationID  string          `json:"organizationId"`
	LedgerID        string          `json:"ledgerId"`
	TransactionType string          `json:"transactionType"`
	Description     *string         `json:"description,omitempty"`
	Code            *string         `json:"code,omitempty"`
	Metadata        models.Metadata `json:"metadata,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// OperationRoute defines a routing rule within a transaction route,
// mapping a specific operation leg to an account.
type OperationRoute struct {
	ID                 string          `json:"id"`
	OrganizationID     string          `json:"organizationId"`
	LedgerID           string          `json:"ledgerId"`
	TransactionRouteID string          `json:"transactionRouteId"`
	AccountID          string          `json:"accountId"`
	AccountAlias       *string         `json:"accountAlias,omitempty"`
	Type               string          `json:"type"`
	Description        *string         `json:"description,omitempty"`
	Metadata           models.Metadata `json:"metadata,omitempty"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}

// ---------------------------------------------------------------------------
// CRM Models
// ---------------------------------------------------------------------------

// Holder represents a customer or entity in the CRM system. Holders are the
// root objects for managing customer relationships and are linked to ledger
// accounts through aliases.
type Holder struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Document      string          `json:"document"`
	ExternalID    string          `json:"externalId,omitempty"`
	Addresses     *HolderAddress  `json:"addresses,omitempty"`
	Contact       *Contact        `json:"contact,omitempty"`
	NaturalPerson *NaturalPerson  `json:"naturalPerson,omitempty"`
	LegalPerson   *LegalPerson    `json:"legalPerson,omitempty"`
	Metadata      models.Metadata `json:"metadata,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	DeletedAt     *time.Time      `json:"deletedAt,omitempty"`
}

// Alias represents an alias account linking a holder to a ledger account
// in the CRM system. Aliases carry banking details, regulatory fields, and
// related-party information.
type Alias struct {
	ID               string            `json:"id"`
	HolderID         string            `json:"holderId"`
	LedgerID         string            `json:"ledgerId"`
	AccountID        string            `json:"accountId"`
	Type             string            `json:"type,omitempty"`
	Document         string            `json:"document,omitempty"`
	BankingDetails   *BankingDetails   `json:"bankingDetails,omitempty"`
	RegulatoryFields *RegulatoryFields `json:"regulatoryFields,omitempty"`
	RelatedParties   []RelatedParty    `json:"relatedParties,omitempty"`
	Metadata         models.Metadata   `json:"metadata,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	DeletedAt        *time.Time        `json:"deletedAt,omitempty"`
}

// HolderAddress contains primary and additional address information for a
// holder. The CRM address structure differs from the shared [models.Address]
// type used by organizations.
type HolderAddress struct {
	Primary     *CRMAddress `json:"primary,omitempty"`
	Additional1 *CRMAddress `json:"additional1,omitempty"`
	Additional2 *CRMAddress `json:"additional2,omitempty"`
}

// CRMAddress represents a physical address within the CRM system.
type CRMAddress struct {
	Line1       string `json:"line1,omitempty"`
	Line2       string `json:"line2,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	ZipCode     string `json:"zipCode,omitempty"`
	Country     string `json:"country,omitempty"`
	Description string `json:"description,omitempty"`
}

// Contact holds contact information for a holder.
type Contact struct {
	PrimaryEmail   string `json:"primaryEmail,omitempty"`
	SecondaryEmail string `json:"secondaryEmail,omitempty"`
	MobilePhone    string `json:"mobilePhone,omitempty"`
	OtherPhone     string `json:"otherPhone,omitempty"`
}

// NaturalPerson holds information specific to individual persons.
type NaturalPerson struct {
	BirthDate    string `json:"birthDate,omitempty"`
	Gender       string `json:"gender,omitempty"`
	Nationality  string `json:"nationality,omitempty"`
	MotherName   string `json:"motherName,omitempty"`
	FatherName   string `json:"fatherName,omitempty"`
	SocialName   string `json:"socialName,omitempty"`
	FavoriteName string `json:"favoriteName,omitempty"`
	CivilStatus  string `json:"civilStatus,omitempty"`
	Status       string `json:"status,omitempty"`
}

// LegalPerson holds information specific to legal entities (companies,
// institutions, etc.).
type LegalPerson struct {
	TradeName      string          `json:"tradeName,omitempty"`
	FoundingDate   string          `json:"foundingDate,omitempty"`
	Activity       string          `json:"activity,omitempty"`
	Size           string          `json:"size,omitempty"`
	Type           string          `json:"type,omitempty"`
	Status         string          `json:"status,omitempty"`
	Representative *Representative `json:"representative,omitempty"`
}

// Representative holds information about a legal entity's representative.
type Representative struct {
	Name     string `json:"name,omitempty"`
	Document string `json:"document,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
}

// BankingDetails holds banking information for an alias.
type BankingDetails struct {
	Branch      string `json:"branch,omitempty"`
	Account     string `json:"account,omitempty"`
	BankID      string `json:"bankId,omitempty"`
	Type        string `json:"type,omitempty"`
	IBAN        string `json:"iban,omitempty"`
	CountryCode string `json:"countryCode,omitempty"`
	OpeningDate string `json:"openingDate,omitempty"`
	ClosingDate string `json:"closingDate,omitempty"`
}

// RegulatoryFields holds regulatory information for an alias.
type RegulatoryFields struct {
	ParticipantDocument string `json:"participantDocument,omitempty"`
}

// RelatedParty represents a related party on an alias.
type RelatedParty struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Document  string `json:"document"`
	Role      string `json:"role"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate,omitempty"`
}
