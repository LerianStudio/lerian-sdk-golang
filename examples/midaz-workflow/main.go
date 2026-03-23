// Example: Midaz Workflow
//
// Demonstrates the complete Midaz financial ledger lifecycle using the
// Lerian SDK v3. The example walks through organization setup, ledger
// creation, asset definition, account provisioning, and transaction
// execution (create -> commit), finishing with a pagination demo and
// structured error handling.
//
// Configure via environment variables:
//
//	LERIAN_MIDAZ_ONBOARDING_URL   (default: http://localhost:3000/v1)
//	LERIAN_MIDAZ_TRANSACTION_URL  (default: http://localhost:3001/v1)
//	LERIAN_MIDAZ_CLIENT_ID        (required with secret + token URL for OAuth2)
//	LERIAN_MIDAZ_CLIENT_SECRET    (required with client ID + token URL for OAuth2)
//	LERIAN_MIDAZ_TOKEN_URL        (required with client ID + secret for OAuth2)
//	LERIAN_DEBUG                  (optional, set "true" for verbose logging)
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	lerian "github.com/LerianStudio/lerian-sdk-golang"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
)

func main() {
	// -----------------------------------------------------------------------
	// Step 1: Create the SDK client configured for the Midaz product.
	//
	// WithMidaz() requires at minimum the two service URLs (onboarding and
	// transaction). Authentication is configured with OAuth2 client credentials.
	// -----------------------------------------------------------------------
	// NOTE: Use HTTPS URLs in production. HTTP is only for local development.
	client, err := lerian.New(
		lerian.WithMidaz(
			midaz.WithOnboardingURL(envOr("LERIAN_MIDAZ_ONBOARDING_URL", "http://localhost:3000/v1")),
			midaz.WithTransactionURL(envOr("LERIAN_MIDAZ_TRANSACTION_URL", "http://localhost:3001/v1")),
		),
		lerian.WithDebug(os.Getenv("LERIAN_DEBUG") == "true"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Always shut down to flush buffered telemetry.
	defer client.Shutdown(context.Background())

	ctx := context.Background()

	// -----------------------------------------------------------------------
	// Step 2: Create an organization -- the top-level domain scope.
	//
	// Organizations own ledgers, accounts, assets, and transactions. The
	// LegalName and LegalDocument fields are required by the API.
	// -----------------------------------------------------------------------
	org, err := client.Midaz.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
		LegalName:     "Acme Corp",
		LegalDocument: "12345678000100",
	})
	if err != nil {
		log.Fatalf("Failed to create organization: %v", err)
	}

	fmt.Printf("Created organization: %s (ID: %s)\n", org.LegalName, org.ID)

	// -----------------------------------------------------------------------
	// Step 3: Create a ledger within the organization.
	//
	// A ledger is an isolated double-entry bookkeeping container. All
	// accounts, assets, and transactions belong to exactly one ledger.
	// -----------------------------------------------------------------------
	ledger, err := client.Midaz.Ledgers.Create(ctx, org.ID, &midaz.CreateLedgerInput{
		Name: "Main Ledger",
	})
	if err != nil {
		log.Fatalf("Failed to create ledger: %v", err)
	}

	fmt.Printf("Created ledger: %s (ID: %s)\n", ledger.Name, ledger.ID)

	// -----------------------------------------------------------------------
	// Step 4: Define an asset (currency) within the ledger.
	//
	// Assets specify the denomination for account balances and transaction
	// amounts. Here we register BRL (Brazilian Real) as a currency type.
	// -----------------------------------------------------------------------
	asset, err := client.Midaz.Assets.Create(ctx, org.ID, ledger.ID, &midaz.CreateAssetInput{
		Name: "Brazilian Real",
		Code: "BRL",
		Type: "currency",
	})
	if err != nil {
		log.Fatalf("Failed to create asset: %v", err)
	}

	fmt.Printf("Created asset: %s (%s)\n", asset.Name, asset.Code)

	// -----------------------------------------------------------------------
	// Step 5: Provision two accounts -- a sender and a receiver.
	//
	// Accounts hold balances denominated in a specific asset. The Type field
	// categorizes the account (e.g. deposit, savings, external). AssetCode
	// must match an existing asset within the same ledger.
	// -----------------------------------------------------------------------
	sender, err := client.Midaz.Accounts.Create(ctx, org.ID, ledger.ID, &midaz.CreateAccountInput{
		Name:      "Sender Account",
		AssetCode: "BRL",
		Type:      "deposit",
	})
	if err != nil {
		log.Fatalf("Failed to create sender account: %v", err)
	}

	fmt.Printf("Created sender account: %s (ID: %s)\n", sender.Name, sender.ID)

	receiver, err := client.Midaz.Accounts.Create(ctx, org.ID, ledger.ID, &midaz.CreateAccountInput{
		Name:      "Receiver Account",
		AssetCode: "BRL",
		Type:      "deposit",
	})
	if err != nil {
		log.Fatalf("Failed to create receiver account: %v", err)
	}

	fmt.Printf("Created receiver account: %s (ID: %s)\n", receiver.Name, receiver.ID)

	// -----------------------------------------------------------------------
	// Step 6: Create a transaction with explicit source -> destination routing.
	//
	// Transactions in Midaz use a DSL-style Source array to describe money
	// movement. Each TransactionSource links a From account to a To account
	// with an explicit amount. The Amount/Scale at the top level represent
	// the total amount (10000 with scale 2 = R$ 100.00).
	// -----------------------------------------------------------------------
	txn, err := client.Midaz.Transactions.Create(ctx, org.ID, ledger.ID, &midaz.CreateTransactionInput{
		AssetCode: "BRL",
		Amount:    10000, // R$ 100.00 in cents
		Scale:     2,
		Source: []midaz.TransactionSource{
			{
				From: midaz.TransactionFromTo{Account: sender.ID},
				To:   midaz.TransactionFromTo{Account: receiver.ID},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create transaction: %v", err)
	}

	fmt.Printf("Created transaction: %s (status: %s)\n", txn.ID, txn.Status.Code)

	// -----------------------------------------------------------------------
	// Step 7: Commit the transaction.
	//
	// Transactions follow a state machine: CREATED -> COMMITTED / CANCELLED.
	// Committing finalizes all operations and applies balance changes. Once
	// committed, a transaction can be reverted but not cancelled.
	// -----------------------------------------------------------------------
	committed, err := client.Midaz.Transactions.Commit(ctx, org.ID, ledger.ID, txn.ID)
	if err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Printf("Committed transaction: %s (status: %s)\n", committed.ID, committed.Status.Code)

	// -----------------------------------------------------------------------
	// Step 8: Demonstrate paginated listing.
	//
	// The SDK uses an Iterator pattern for paginated endpoints. Call Next()
	// to advance the cursor and Item() to retrieve the current record.
	// Pages are fetched lazily -- only when the current buffer is exhausted.
	// -----------------------------------------------------------------------
	fmt.Println("\n--- Listing Organizations ---")

	iter := client.Midaz.Organizations.List(ctx, nil)
	for iter.Next(ctx) {
		o := iter.Item()
		fmt.Printf("  - %s (ID: %s)\n", o.LegalName, o.ID)
	}

	if err := iter.Err(); err != nil {
		log.Fatalf("Failed to list organizations: %v", err)
	}

	// -----------------------------------------------------------------------
	// Step 9: Structured error handling with sentinel errors.
	//
	// The SDK provides category-level sentinel errors (ErrNotFound,
	// ErrValidation, ErrConflict, etc.) that work with errors.Is() across
	// all products. This enables product-agnostic error handling.
	// -----------------------------------------------------------------------
	_, err = client.Midaz.Organizations.Get(ctx, "non-existent-id")
	if err != nil {
		if errors.Is(err, lerian.ErrNotFound) {
			fmt.Println("\nCorrectly caught ErrNotFound for non-existent organization")
		} else {
			fmt.Printf("\nUnexpected error type: %v\n", err)
		}
	}

	fmt.Println("\nMidaz workflow complete!")
}

// envOr returns the value of the environment variable named by key, or
// the fallback value if the variable is empty or unset.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
