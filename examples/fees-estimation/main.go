// Example: Fees Estimation Workflow
//
// Demonstrates the Fees calculation product using the Lerian SDK v3.
// The example creates a fee package with calculation rules, previews
// charges via the estimation endpoint, and then calculates fees by
// submitting a transaction DSL for fee injection.
//
// Configure via environment variables:
//
//	LERIAN_FEES_URL           (default: http://localhost:3005/v1)
//	LERIAN_FEES_ORG_ID        (required -- the organization scope)
//	LERIAN_FEES_CLIENT_ID     (required with secret + token URL for OAuth2)
//	LERIAN_FEES_CLIENT_SECRET (required with client ID + token URL for OAuth2)
//	LERIAN_FEES_TOKEN_URL     (required with client ID + secret for OAuth2)
//	LERIAN_DEBUG              (optional, set "true" for verbose logging; avoid enabling it with sensitive transaction/account data outside local development)
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
	// Fees requires an OrganizationID sent as the X-Organization-Id header.
	// Authentication uses OAuth2 client credentials.
	// -----------------------------------------------------------------------
	orgID := envOr("LERIAN_FEES_ORG_ID", "org-placeholder-id")

	// NOTE: Use HTTPS URLs in production. HTTP is only for local development.
	client, err := lerian.New(lerian.Config{
		Debug: os.Getenv("LERIAN_DEBUG") == "true",
		Fees: &fees.Config{
			BaseURL:        envOr("LERIAN_FEES_URL", "http://localhost:3005/v1"),
			OrganizationID: orgID,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer client.Shutdown(context.Background())

	ctx := context.Background()

	// -----------------------------------------------------------------------
	// Step 2: Create a fee package with calculation rules.
	//
	// A fee package groups one or more Fee definitions, each with a
	// CalculationModel that determines how the fee is computed:
	//   - "flatFee":          a fixed amount per transaction
	//   - "percentual":       a percentage of the transaction amount
	//   - "maxBetweenTypes":  multiple calculations, takes the maximum
	//
	// Packages are scoped to a ledger and optionally a segment, with
	// minimum/maximum transaction amount thresholds.
	// -----------------------------------------------------------------------
	pkgDesc := "Standard fees for BRL wire transfers"
	enablePkg := true

	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Wire Transfer Fees - BRL",
		Description:   &pkgDesc,
		LedgerID:      "00000000-0000-0000-0000-000000000001",
		MinimumAmount: "100.00",
		MaximumAmount: "100000.00",
		Enable:        &enablePkg,
		Fees: map[string]fees.Fee{
			"administrativeFee": {
				FeeLabel: "Taxa Administrativa",
				CalculationModel: &fees.CalculationModel{
					ApplicationRule: "flatFee",
					Calculations: []fees.Calculation{
						{Type: "flat", Value: "1.50"},
					},
				},
				ReferenceAmount:  "originalAmount",
				Priority:         1,
				IsDeductibleFrom: boolPtr(true),
				CreditAccount:    "@revenue_fees",
			},
			"serviceFee": {
				FeeLabel: "Taxa de Serviço",
				CalculationModel: &fees.CalculationModel{
					ApplicationRule: "percentual",
					Calculations: []fees.Calculation{
						{Type: "percentage", Value: "1.5"},
					},
				},
				ReferenceAmount:  "originalAmount",
				Priority:         2,
				IsDeductibleFrom: boolPtr(false),
				CreditAccount:    "@revenue_service",
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create fee package: %v", err)
	}

	fmt.Printf("Created fee package: %s (ID: %s)\n", pkg.FeeGroupLabel, pkg.ID)
	fmt.Printf("  Amount range: %s - %s\n", pkg.MinimumAmount, pkg.MaximumAmount)
	fmt.Printf("  Fees: %d\n", len(pkg.Fees))

	for key, feeDefinition := range pkg.Fees {
		rule := "<missing calculation model>"
		if feeDefinition.CalculationModel != nil {
			rule = feeDefinition.CalculationModel.ApplicationRule
		}

		fmt.Printf("  [%s] %s (rule: %s, priority: %d)\n",
			key, feeDefinition.FeeLabel, rule, feeDefinition.Priority)
	}

	// -----------------------------------------------------------------------
	// Step 3: Calculate an estimate (preview) for a transaction.
	//
	// The Estimates.Calculate method takes a package ID, ledger ID, and a
	// transaction DSL, then returns whether any fees would be applied
	// without persisting anything.
	// -----------------------------------------------------------------------
	estimateResp, err := client.Fees.Estimates.Calculate(ctx, &fees.FeeEstimateInput{
		PackageID: pkg.ID,
		LedgerID:  "00000000-0000-0000-0000-000000000001",
		Transaction: fees.TransactionDSL{
			Description: "Wire transfer estimate",
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: "500.00", // R$ 500.00
				Source: fees.TransactionDSLSource{
					From: []fees.TransactionDSLLeg{
						{AccountAlias: "@sender", Share: &fees.TransactionDSLShare{Percentage: 100}},
					},
				},
				Distribute: fees.TransactionDSLDistribute{
					To: []fees.TransactionDSLLeg{
						{AccountAlias: "@receiver", Share: &fees.TransactionDSLShare{Percentage: 100}},
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to calculate estimate: %v", err)
	}

	fmt.Printf("\n--- Fee Estimate ---\n")
	fmt.Printf("  Message: %s\n", estimateResp.Message)

	if estimateResp.FeesApplied != nil {
		fmt.Println("  Fees were applied to the transaction DSL")
	} else {
		fmt.Println("  No fees matched the given parameters")
	}

	// -----------------------------------------------------------------------
	// Step 4: Calculate fees for a real transaction via DSL injection.
	//
	// The Fees.Calculate method sends a transaction DSL to the fees service.
	// The service evaluates matching fee packages and returns the mutated
	// transaction with fee legs injected into source and distribute arrays.
	// -----------------------------------------------------------------------
	segmentID := "00000000-0000-0000-0000-000000000002"

	feeResult, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		SegmentID: &segmentID,
		LedgerID:  "00000000-0000-0000-0000-000000000001",
		Transaction: fees.TransactionDSL{
			Description: "Wire transfer with fees",
			Route:       "wire_transfer",
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: "1000.00", // R$ 1,000.00
				Source: fees.TransactionDSLSource{
					From: []fees.TransactionDSLLeg{
						{AccountAlias: "@sender", Share: &fees.TransactionDSLShare{Percentage: 100}},
					},
				},
				Distribute: fees.TransactionDSLDistribute{
					To: []fees.TransactionDSLLeg{
						{AccountAlias: "@receiver", Share: &fees.TransactionDSLShare{Percentage: 100}},
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to calculate fee: %v", err)
	}

	fmt.Printf("\n--- Fee Calculation for R$ 1,000.00 ---\n")
	fmt.Printf("  Ledger:   %s\n", feeResult.LedgerID)
	fmt.Printf("  Source legs:     %d\n", len(feeResult.Transaction.Send.Source.From))
	fmt.Printf("  Distribute legs: %d\n", len(feeResult.Transaction.Send.Distribute.To))

	// -----------------------------------------------------------------------
	// Step 5: List all fee packages with optional filters.
	// -----------------------------------------------------------------------
	fmt.Println("\n--- All Fee Packages ---")

	listResp, err := client.Fees.Packages.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list packages: %v", err)
	}

	fmt.Printf("  Total: %d (page: %d/%d, size: %d)\n", listResp.TotalItems, listResp.PageNumber, listResp.TotalPages, listResp.PageSize)

	for _, p := range listResp.Items {
		enabled := "disabled"
		if p.Enable != nil && *p.Enable {
			enabled = "enabled"
		}

		fmt.Printf("  - %s (%s, fees: %d)\n", p.FeeGroupLabel, enabled, len(p.Fees))
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

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool {
	return &b
}
