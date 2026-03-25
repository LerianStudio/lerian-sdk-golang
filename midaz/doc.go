// Package midaz provides the client for the Midaz financial ledger platform.
//
// Midaz manages organizations, ledgers, accounts, assets, portfolios,
// segments, transactions, holders, and aliases. It connects to three
// microservices: onboarding (organization/ledger/account management),
// transaction (balance, transaction, and operation management), and CRM
// (holder and alias management).
//
// # Usage
//
// Access Midaz services through the umbrella client:
//
//	client, _ := lerian.New(lerian.Config{
//	    Midaz: &midaz.Config{
//	        OnboardingURL:  "http://localhost:3000/v1",
//	        TransactionURL: "http://localhost:3001/v1",
//	        CRMURL:         "http://localhost:4003/v1",
//	    },
//	})
//
//	org, err := client.Midaz.Onboarding.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
//	    LegalName:   "Acme Corp",
//	    LegalDocument: "12345678000100",
//	})
//
// # Available Services
//
// The Midaz client exposes the following services:
//
//   - Organizations -- top-level organisational entities
//   - Ledgers -- isolated double-entry ledgers within an organization
//   - Accounts -- financial accounts holding balances
//   - Assets -- tradable instruments and currencies
//   - AssetRates -- exchange rates between asset pairs
//   - Portfolios -- logical groupings of accounts
//   - Segments -- classification and grouping of accounts
//   - Transactions -- atomic financial movements
//   - Operations -- individual debit/credit legs within transactions
//   - Balances -- account balance entries (routed through the transaction backend)
//   - TransactionRoutes -- routing templates for transaction processing
//   - OperationRoutes -- per-operation routing rules within transaction routes
//   - AccountTypes -- account type classifications within a ledger
//   - Holders -- customer/entity records in the CRM system
//   - Aliases -- alias accounts linking holders to ledger accounts
//
// # CRM Services
//
// The CRM backend (Holders and Aliases) uses the X-Organization-Id header
// instead of URL path parameters for organization context. CRM list methods
// follow the SDK iterator pattern and accept `CRMListOptions` or
// `AliasListOptions` for paging and filtering.
//
// # Transaction Lifecycle
//
// Transactions follow a state machine: Create -> Commit, Cancel, or Revert.
// Once created, a transaction must be explicitly committed to take effect.
// Alternative creation formats are exposed directly through `Transactions`
// for annotation, DSL, inflow, and outflow payloads:
//
//	txn, _ := client.Midaz.Transactions.Transactions.Create(ctx, orgID, ledgerID, &midaz.CreateTransactionInput{...})
//	committed, _ := client.Midaz.Transactions.Transactions.Commit(ctx, orgID, ledgerID, txn.ID)
package midaz
