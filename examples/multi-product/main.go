// Example: Multi-Product Workflow
//
// Demonstrates a cross-product workflow using the Lerian SDK v3 with
// three products initialized in a single client: Midaz (ledger),
// Tracer (compliance), and Matcher (reconciliation). The example
// shows how to create a financial transaction in Midaz, validate it
// against compliance rules in Tracer, and then reconcile the records
// using Matcher -- all sharing the same observability and debug config.
//
// Configure via environment variables:
//
//	LERIAN_MIDAZ_ONBOARDING_URL  (default: http://localhost:3000/v1)
//	LERIAN_MIDAZ_TRANSACTION_URL (default: http://localhost:3001/v1)
//	LERIAN_MIDAZ_AUTH_TOKEN       (optional)
//	LERIAN_TRACER_URL             (default: http://localhost:3003/v1)
//	LERIAN_TRACER_API_KEY         (optional)
//	LERIAN_MATCHER_URL            (default: http://localhost:3002/v1)
//	LERIAN_MATCHER_API_KEY        (optional)
//	LERIAN_DEBUG                  (optional, set "true" for verbose logging)
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	lerian "github.com/LerianStudio/lerian-sdk-golang"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
)

func main() {
	// -----------------------------------------------------------------------
	// Step 1: Create a single client with multiple products.
	//
	// The lerian.New() function accepts multiple With<Product>() options,
	// each configuring a different product. Shared options like WithDebug()
	// and WithObservability() apply to ALL products, providing a unified
	// configuration surface.
	// -----------------------------------------------------------------------
	// NOTE: Use HTTPS URLs in production. HTTP is only for local development.
	client, err := lerian.New(
		// Midaz: financial ledger
		lerian.WithMidaz(
			midaz.WithOnboardingURL(envOr("LERIAN_MIDAZ_ONBOARDING_URL", "http://localhost:3000/v1")),
			midaz.WithTransactionURL(envOr("LERIAN_MIDAZ_TRANSACTION_URL", "http://localhost:3001/v1")),
			midaz.WithAuthToken(os.Getenv("LERIAN_MIDAZ_AUTH_TOKEN")),
		),

		// Tracer: compliance validation
		lerian.WithTracer(
			tracer.WithBaseURL(envOr("LERIAN_TRACER_URL", "http://localhost:3003/v1")),
			tracer.WithAPIKey(os.Getenv("LERIAN_TRACER_API_KEY")),
		),

		// Matcher: reconciliation
		lerian.WithMatcher(
			matcher.WithBaseURL(envOr("LERIAN_MATCHER_URL", "http://localhost:3002/v1")),
			matcher.WithAPIKey(os.Getenv("LERIAN_MATCHER_API_KEY")),
		),

		// Shared: debug logging across all products.
		// In production, replace WithDebug with WithObservability for
		// structured OTel telemetry:
		//   lerian.WithObservability(true, true, true),
		//   lerian.WithCollectorEndpoint("http://otel-collector:4318"),
		lerian.WithDebug(os.Getenv("LERIAN_DEBUG") == "true"),
	)
	if err != nil {
		log.Fatalf("Failed to create multi-product client: %v", err)
	}

	defer client.Shutdown(context.Background())

	ctx := context.Background()

	// Verify all three products are available.
	fmt.Println("Multi-product client initialized:")
	fmt.Printf("  Midaz:   %v\n", client.Midaz != nil)
	fmt.Printf("  Tracer:  %v\n", client.Tracer != nil)
	fmt.Printf("  Matcher: %v\n", client.Matcher != nil)

	// =======================================================================
	// Phase 1: Set up the Midaz ledger and create a transaction.
	// =======================================================================
	fmt.Println("\n=== Phase 1: Midaz - Create Transaction ===")

	// Create an organization to scope all ledger operations.
	org, err := client.Midaz.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
		LegalName:     "Multi-Product Demo Corp",
		LegalDocument: "98765432000199",
	})
	if err != nil {
		log.Fatalf("Failed to create organization: %v", err)
	}

	fmt.Printf("Organization: %s (ID: %s)\n", org.LegalName, org.ID)

	// Create a ledger.
	ledger, err := client.Midaz.Ledgers.Create(ctx, org.ID, &midaz.CreateLedgerInput{
		Name: "Cross-Product Ledger",
	})
	if err != nil {
		log.Fatalf("Failed to create ledger: %v", err)
	}

	fmt.Printf("Ledger: %s (ID: %s)\n", ledger.Name, ledger.ID)

	// Create a BRL asset.
	_, err = client.Midaz.Assets.Create(ctx, org.ID, ledger.ID, &midaz.CreateAssetInput{
		Name: "Brazilian Real",
		Code: "BRL",
		Type: "currency",
	})
	if err != nil {
		log.Fatalf("Failed to create asset: %v", err)
	}

	// Create sender and receiver accounts.
	sender, err := client.Midaz.Accounts.Create(ctx, org.ID, ledger.ID, &midaz.CreateAccountInput{
		Name:      "Treasury",
		AssetCode: "BRL",
		Type:      "deposit",
	})
	if err != nil {
		log.Fatalf("Failed to create sender: %v", err)
	}

	receiver, err := client.Midaz.Accounts.Create(ctx, org.ID, ledger.ID, &midaz.CreateAccountInput{
		Name:      "Vendor Payout",
		AssetCode: "BRL",
		Type:      "deposit",
	})
	if err != nil {
		log.Fatalf("Failed to create receiver: %v", err)
	}

	// Create a transaction (R$ 25,000.00).
	txn, err := client.Midaz.Transactions.Create(ctx, org.ID, ledger.ID, &midaz.CreateTransactionInput{
		AssetCode: "BRL",
		Amount:    2500000, // R$ 25,000.00
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

	fmt.Printf("Transaction: %s (amount: %d, scale: %d, status: %s)\n",
		txn.ID, txn.Amount, txn.AmountScale, txn.Status.Code)

	// =======================================================================
	// Phase 2: Validate the transaction against Tracer compliance rules.
	// =======================================================================
	fmt.Println("\n=== Phase 2: Tracer - Validate Transaction ===")

	// Submit the transaction data for compliance validation.
	// The transaction map mirrors the financial details so rules can
	// inspect amount thresholds, asset codes, account types, etc.
	validation, err := client.Tracer.Validations.Create(ctx, &tracer.CreateValidationInput{
		Transaction: map[string]any{
			"transactionId":  txn.ID,
			"amount":         txn.Amount,
			"assetCode":      txn.AssetCode,
			"type":           "wire_transfer",
			"senderID":       sender.ID,
			"receiverID":     receiver.ID,
			"organizationID": org.ID,
		},
	})
	if err != nil {
		log.Fatalf("Failed to validate transaction: %v", err)
	}

	fmt.Printf("Validation result: %s (status: %s)\n", validation.Result, validation.Status)

	if len(validation.RulesApplied) > 0 {
		fmt.Printf("  Rules applied: %v\n", validation.RulesApplied)
	}

	if len(validation.Violations) > 0 {
		fmt.Printf("  Violations: %v\n", validation.Violations)
	}

	// Based on the validation result, decide whether to commit or cancel.
	if validation.Result == "rejected" {
		fmt.Println("Transaction REJECTED by compliance -- cancelling...")

		cancelled, cErr := client.Midaz.Transactions.Cancel(ctx, org.ID, ledger.ID, txn.ID)
		if cErr != nil {
			log.Fatalf("Failed to cancel transaction: %v", cErr)
		}

		fmt.Printf("Transaction cancelled: %s (status: %s)\n", cancelled.ID, cancelled.Status.Code)
	} else {
		fmt.Println("Transaction APPROVED by compliance -- committing...")

		committed, cErr := client.Midaz.Transactions.Commit(ctx, org.ID, ledger.ID, txn.ID)
		if cErr != nil {
			log.Fatalf("Failed to commit transaction: %v", cErr)
		}

		fmt.Printf("Transaction committed: %s (status: %s)\n", committed.ID, committed.Status.Code)
	}

	// =======================================================================
	// Phase 3: Set up Matcher for reconciliation.
	// =======================================================================
	fmt.Println("\n=== Phase 3: Matcher - Reconciliation ===")

	// Create a reconciliation context.
	tolerance := 0.01
	matchCtx, err := client.Matcher.Contexts.Create(ctx, &matcher.CreateContextInput{
		Name: "Midaz-to-Bank Reconciliation",
		Config: &matcher.ContextConfig{
			MatchingStrategy: "amount_date",
			Tolerance:        &tolerance,
			AutoApprove:      true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create reconciliation context: %v", err)
	}

	fmt.Printf("Reconciliation context: %s (ID: %s)\n", matchCtx.Name, matchCtx.ID)

	// Create a matching rule to compare amounts exactly.
	_, err = client.Matcher.Rules.Create(ctx, &matcher.CreateRuleInput{
		ContextID:  matchCtx.ID,
		Name:       "Exact Amount Match",
		Priority:   1,
		Expression: "source.amount == target.amount AND source.reference == target.reference",
	})
	if err != nil {
		log.Fatalf("Failed to create matching rule: %v", err)
	}

	fmt.Println("Created matching rule: Exact Amount Match (priority: 1)")

	// Trigger the reconciliation run.
	result, err := client.Matcher.Matching.Run(ctx, matchCtx.ID)
	if err != nil {
		log.Fatalf("Failed to run reconciliation: %v", err)
	}

	fmt.Printf("\nReconciliation result:\n")
	fmt.Printf("  Matched:    %d\n", result.MatchedCount)
	fmt.Printf("  Unmatched:  %d\n", result.UnmatchedCount)
	fmt.Printf("  Exceptions: %d\n", result.ExceptionCount)
	fmt.Printf("  Duration:   %d ms\n", result.Duration)

	// Get the reconciliation summary.
	summary, err := client.Matcher.Reports.GetSummary(ctx, matchCtx.ID)
	if err != nil {
		log.Fatalf("Failed to get summary: %v", err)
	}

	fmt.Printf("\nReconciliation summary:\n")
	fmt.Printf("  Total records:  %d\n", summary.TotalRecords)
	fmt.Printf("  Match rate:     %.2f%%\n", summary.MatchRate*100)

	// =======================================================================
	// Error handling demonstration
	// =======================================================================
	fmt.Println("\n=== Error Handling ===")

	// All products share the same sentinel errors, enabling product-agnostic
	// error handling. This is one of the key benefits of the unified SDK
	// architecture -- a single errors.Is() check works across Midaz, Tracer,
	// and Matcher.
	testCases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Midaz: get non-existent organization",
			fn: func() error {
				_, err := client.Midaz.Organizations.Get(ctx, "does-not-exist")
				return err
			},
		},
		{
			name: "Tracer: get non-existent rule",
			fn: func() error {
				_, err := client.Tracer.Rules.Get(ctx, "does-not-exist")
				return err
			},
		},
		{
			name: "Matcher: get non-existent context",
			fn: func() error {
				_, err := client.Matcher.Contexts.Get(ctx, "does-not-exist")
				return err
			},
		},
	}

	for _, tc := range testCases {
		err := tc.fn()
		if err != nil {
			if errors.Is(err, lerian.ErrNotFound) {
				fmt.Printf("  [OK] %s -> ErrNotFound\n", tc.name)
			} else {
				fmt.Printf("  [??] %s -> %v\n", tc.name, err)
			}
		}
	}

	fmt.Println("\nMulti-product workflow complete!")
}

// envOr returns the value of the environment variable named by key, or
// the fallback value if the variable is empty or unset.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
