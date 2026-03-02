// Package midaz provides the client for the Midaz financial ledger platform.
//
// Midaz manages organizations, ledgers, accounts, assets, portfolios,
// segments, and transactions. It connects to two microservices: onboarding
// (organization/ledger/account management) and transaction (balance,
// transaction, and operation management).
//
// # Usage
//
// Access Midaz services through the umbrella client:
//
//	client, _ := lerian.New(
//	    lerian.WithMidaz(
//	        midaz.WithOnboardingURL("http://localhost:3000/v1"),
//	        midaz.WithTransactionURL("http://localhost:3001/v1"),
//	        midaz.WithAuthToken("my-token"),
//	    ),
//	)
//
//	org, err := client.Midaz.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
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
//
// # Transaction Lifecycle
//
// Transactions follow a state machine: Create -> Commit, Cancel, or Revert.
// Once created, a transaction must be explicitly committed to take effect:
//
//	txn, _ := client.Midaz.Transactions.Create(ctx, orgID, ledgerID, &midaz.CreateTransactionInput{...})
//	committed, _ := client.Midaz.Transactions.Commit(ctx, orgID, ledgerID, txn.ID)
package midaz
