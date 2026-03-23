// Example: Fees Estimation Workflow
//
// Demonstrates the Fees calculation product using the Lerian SDK v3.
// The example creates a fee package with multiple rule types, previews
// charges via the estimation endpoint, and then calculates a fee linked
// to a transaction.
//
// Configure via environment variables:
//
//	LERIAN_FEES_URL           (default: http://localhost:3005/v1)
//	LERIAN_FEES_ORG_ID        (required -- the organization scope)
//	LERIAN_FEES_CLIENT_ID     (required with secret + token URL for OAuth2)
//	LERIAN_FEES_CLIENT_SECRET (required with client ID + token URL for OAuth2)
//	LERIAN_FEES_TOKEN_URL     (required with client ID + secret for OAuth2)
//	LERIAN_DEBUG              (optional, set "true" for verbose logging)
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	lerian "github.com/LerianStudio/lerian-sdk-golang"
	"github.com/LerianStudio/lerian-sdk-golang/fees"
)

func main() {
	// -----------------------------------------------------------------------
	// Step 1: Create the SDK client configured for the Fees product.
	//
	// Like Reporter, Fees requires an OrganizationID sent as the
	// X-Organization-Id header. Authentication uses OAuth2 client credentials.
	// -----------------------------------------------------------------------
	orgID := envOr("LERIAN_FEES_ORG_ID", "org-placeholder-id")

	// NOTE: Use HTTPS URLs in production. HTTP is only for local development.
	// OAuth2 credentials are read from LERIAN_FEES_CLIENT_ID,
	// LERIAN_FEES_CLIENT_SECRET, and LERIAN_FEES_TOKEN_URL when set.
	client, err := lerian.New(
		lerian.WithFees(
			fees.WithBaseURL(envOr("LERIAN_FEES_URL", "http://localhost:3005/v1")),
			fees.WithOrganizationID(orgID),
		),
		lerian.WithDebug(os.Getenv("LERIAN_DEBUG") == "true"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer client.Shutdown(context.Background())

	ctx := context.Background()

	// -----------------------------------------------------------------------
	// Step 2: Create a fee package with multiple rule types.
	//
	// A fee package groups one or more FeeRule definitions. Rules can be:
	//   - "flat":       a fixed amount per transaction
	//   - "percentage": a percentage of the transaction amount
	//   - "tiered":     a combination with min/max caps
	//
	// Amounts are in the smallest currency unit (e.g. cents for BRL).
	// -----------------------------------------------------------------------
	flatAmount := int64(150) // R$ 1.50 flat fee
	pctValue := "1.5"        // 1.5% of the transaction amount
	pkgDesc := "Standard fees for BRL wire transfers"

	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		Name:        "Wire Transfer Fees - BRL",
		Description: &pkgDesc,
		Rules: []fees.FeeRule{
			{
				Type:     "flat",
				Amount:   &flatAmount,
				Currency: "BRL",
			},
			{
				Type:       "percentage",
				Percentage: &pctValue,
				Currency:   "BRL",
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create fee package: %v", err)
	}

	fmt.Printf("Created fee package: %s (ID: %s)\n", pkg.Name, pkg.ID)
	fmt.Printf("  Rules: %d\n", len(pkg.Rules))

	for i, r := range pkg.Rules {
		switch r.Type {
		case "flat":
			fmt.Printf("  [%d] flat: %d cents %s\n", i+1, derefInt64(r.Amount), r.Currency)
		case "percentage":
			fmt.Printf("  [%d] percentage: %s%% %s\n", i+1, derefStr(r.Percentage), r.Currency)
		}
	}

	// -----------------------------------------------------------------------
	// Step 3: Calculate an estimate (preview) without a real transaction.
	//
	// The Estimates.Calculate method is RPC-style: it accepts a package ID,
	// amount, scale, and currency, then returns the computed fee breakdown
	// without persisting anything or linking to a transaction.
	// -----------------------------------------------------------------------
	estimate, err := client.Fees.Estimates.Calculate(ctx, &fees.CalculateEstimateInput{
		PackageID: pkg.ID,
		Amount:    50000, // R$ 500.00
		Scale:     2,
		Currency:  "BRL",
	})
	if err != nil {
		log.Fatalf("Failed to calculate estimate: %v", err)
	}

	fmt.Printf("\n--- Fee Estimate for R$ 500.00 ---\n")
	fmt.Printf("  Total fee: %d (scale: %d, currency: %s)\n",
		estimate.TotalFee, estimate.TotalFeeScale, estimate.Currency)

	for _, fr := range estimate.FeeResults {
		fmt.Printf("  [%s] %d %s (applied: %t)\n",
			fr.RuleType, fr.Amount, fr.Currency, fr.Applied)
	}

	// -----------------------------------------------------------------------
	// Step 4: Calculate a fee linked to a transaction.
	//
	// Unlike estimates, calculated fees can be linked to an actual
	// transaction via TransactionID. The resulting Fee object tracks
	// its lifecycle through a status field and is persisted.
	// -----------------------------------------------------------------------
	txnID := "txn-example-12345"

	fee, err := client.Fees.Fees.Calculate(ctx, &fees.CalculateFeeInput{
		PackageID:     pkg.ID,
		TransactionID: &txnID,
		Amount:        100000, // R$ 1,000.00
		Scale:         2,
		Currency:      "BRL",
	})
	if err != nil {
		log.Fatalf("Failed to calculate fee: %v", err)
	}

	fmt.Printf("\n--- Fee Calculation for R$ 1,000.00 ---\n")
	fmt.Printf("  Fee ID:    %s\n", fee.ID)
	fmt.Printf("  Total fee: %d (scale: %d)\n", fee.TotalFee, fee.TotalFeeScale)
	fmt.Printf("  Status:    %s\n", fee.Status)

	if fee.TransactionID != nil {
		fmt.Printf("  Linked to: %s\n", *fee.TransactionID)
	}

	for _, fr := range fee.FeeResults {
		fmt.Printf("  [%s] %d %s (applied: %t", fr.RuleType, fr.Amount, fr.Currency, fr.Applied)
		if fr.Reason != "" {
			fmt.Printf(", reason: %s", fr.Reason)
		}

		fmt.Println(")")
	}

	// -----------------------------------------------------------------------
	// Step 5: List all fee packages.
	// -----------------------------------------------------------------------
	fmt.Println("\n--- All Fee Packages ---")

	pkgIter := client.Fees.Packages.List(ctx, nil)
	for pkgIter.Next(ctx) {
		p := pkgIter.Item()
		fmt.Printf("  - %s (status: %s, rules: %d)\n", p.Name, p.Status, len(p.Rules))
	}

	if err := pkgIter.Err(); err != nil {
		log.Fatalf("Failed to list packages: %v", err)
	}

	// -----------------------------------------------------------------------
	// Step 6: Error handling -- demonstrate ErrNotFound.
	// -----------------------------------------------------------------------
	_, err = client.Fees.Packages.Get(ctx, "non-existent-package-id")
	if err != nil {
		if errors.Is(err, lerian.ErrNotFound) {
			fmt.Println("\nCorrectly caught ErrNotFound for non-existent package")
		} else {
			fmt.Printf("\nUnexpected error: %v\n", err)
		}
	}

	fmt.Println("\nFees estimation workflow complete!")
}

// envOr returns the value of the environment variable named by key, or
// the fallback value if the variable is empty or unset.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// derefInt64 safely dereferences a *int64 pointer, returning 0 if nil.
func derefInt64(p *int64) int64 {
	if p != nil {
		return *p
	}

	return 0
}

// derefStr safely dereferences a *string pointer, returning "" if nil.
func derefStr(p *string) string {
	if p != nil {
		return *p
	}

	return ""
}
