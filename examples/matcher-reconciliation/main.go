// Example: Matcher Reconciliation Workflow
//
// Demonstrates the Matcher reconciliation product lifecycle using the
// Lerian SDK v3. The example walks through creating a reconciliation
// context, defining matching rules with expressions, setting up a
// schedule for automated runs, triggering a manual reconciliation,
// reviewing exceptions, and retrieving the reconciliation summary.
//
// Configure via environment variables:
//
//	LERIAN_MATCHER_URL           (default: http://localhost:3002/v1)
//	LERIAN_MATCHER_CLIENT_ID     (required with secret + token URL for OAuth2)
//	LERIAN_MATCHER_CLIENT_SECRET (required with client ID + token URL for OAuth2)
//	LERIAN_MATCHER_TOKEN_URL     (required with client ID + secret for OAuth2)
//	LERIAN_DEBUG                 (optional, set "true" for verbose logging)
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	lerian "github.com/LerianStudio/lerian-sdk-golang"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
)

func main() {
	// -----------------------------------------------------------------------
	// Step 1: Create the SDK client configured for the Matcher product.
	//
	// Matcher uses OAuth2 client credentials. A single base URL covers all
	// Matcher endpoints unlike Midaz which has two microservices.
	// -----------------------------------------------------------------------
	// NOTE: Use HTTPS URLs in production. HTTP is only for local development.
	client, err := lerian.New(lerian.Config{
		Debug: os.Getenv("LERIAN_DEBUG") == "true",
		Matcher: &matcher.Config{
			BaseURL: envOr("LERIAN_MATCHER_URL", "http://localhost:3002/v1"),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer client.Shutdown(context.Background())

	ctx := context.Background()

	// -----------------------------------------------------------------------
	// Step 2: Create a reconciliation context.
	//
	// Contexts are the top-level scope for all reconciliation operations.
	// The ContextConfig controls the matching strategy, tolerance thresholds,
	// and whether matched results are automatically approved.
	// -----------------------------------------------------------------------
	tolerance := 0.01
	desc := "Reconciles bank feed against internal ledger entries"

	matchCtx, err := client.Matcher.Contexts.Create(ctx, &matcher.CreateContextInput{
		Name:        "Bank Reconciliation",
		Description: &desc,
		Config: &matcher.ContextConfig{
			MatchingStrategy: "amount_date",
			Tolerance:        &tolerance,
			AutoApprove:      false,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create context: %v", err)
	}

	fmt.Printf("Created reconciliation context: %s (ID: %s)\n", matchCtx.Name, matchCtx.ID)

	// -----------------------------------------------------------------------
	// Step 3: Create matching rules with expressions.
	//
	// Rules define how record pairs are compared. They are evaluated in
	// priority order (lower number = higher priority). Expressions use the
	// Matcher DSL to specify field comparisons and tolerance conditions.
	// -----------------------------------------------------------------------
	exactRule, err := client.Matcher.Rules.Create(ctx, &matcher.CreateRuleInput{
		ContextID:  matchCtx.ID,
		Name:       "Exact Amount Match",
		Priority:   1,
		Expression: "source.amount == target.amount AND source.date == target.date",
	})
	if err != nil {
		log.Fatalf("Failed to create exact-match rule: %v", err)
	}

	fmt.Printf("Created rule: %s (priority: %d)\n", exactRule.Name, exactRule.Priority)

	fuzzyDesc := "Allows small rounding differences between source and target amounts"

	_, err = client.Matcher.Rules.Create(ctx, &matcher.CreateRuleInput{
		ContextID:   matchCtx.ID,
		Name:        "Fuzzy Amount Match",
		Description: &fuzzyDesc,
		Priority:    2,
		Expression:  "abs(source.amount - target.amount) <= 100 AND source.date == target.date",
	})
	if err != nil {
		log.Fatalf("Failed to create fuzzy-match rule: %v", err)
	}

	fmt.Println("Created fuzzy amount match rule (priority: 2)")

	// -----------------------------------------------------------------------
	// Step 4: Create a schedule for automated reconciliation runs.
	//
	// Schedules use cron expressions to define recurring run times. The
	// Matcher service triggers a reconciliation run at each scheduled
	// interval, processing all unmatched records in the context.
	// -----------------------------------------------------------------------
	schedule, err := client.Matcher.Schedules.Create(ctx, &matcher.CreateScheduleInput{
		ContextID: matchCtx.ID,
		Name:      "Daily EOD Reconciliation",
		CronExpr:  "0 23 * * *", // Every day at 11 PM
	})
	if err != nil {
		log.Fatalf("Failed to create schedule: %v", err)
	}

	fmt.Printf("Created schedule: %s (cron: %s)\n", schedule.Name, schedule.CronExpr)

	// -----------------------------------------------------------------------
	// Step 5: Trigger a manual reconciliation run.
	//
	// The Matching.Run() method is an RPC-style action that executes the
	// reconciliation engine against all unmatched records in the context.
	// It returns a MatchResult with counts of matched, unmatched, and
	// exception records along with the run duration.
	// -----------------------------------------------------------------------
	result, err := client.Matcher.Matching.Run(ctx, matchCtx.ID)
	if err != nil {
		log.Fatalf("Failed to run matching: %v", err)
	}

	fmt.Printf("\nReconciliation result:\n")
	fmt.Printf("  Matched:   %d\n", result.MatchedCount)
	fmt.Printf("  Unmatched: %d\n", result.UnmatchedCount)
	fmt.Printf("  Exceptions: %d\n", result.ExceptionCount)
	fmt.Printf("  Duration:  %d ms\n", result.Duration)

	// -----------------------------------------------------------------------
	// Step 6: Review exceptions.
	//
	// Exceptions are anomalies detected during reconciliation that need
	// human review. We iterate over them using the paginated iterator,
	// then demonstrate approving a single exception.
	// -----------------------------------------------------------------------
	fmt.Println("\n--- Listing Exceptions ---")

	exIter := client.Matcher.Exceptions.List(ctx, nil)
	var firstExceptionID string

	for exIter.Next(ctx) {
		ex := exIter.Item()
		if firstExceptionID == "" {
			firstExceptionID = ex.ID
		}

		fmt.Printf("  - [%s] %s (priority: %s, status: %s)\n",
			ex.Type, deref(ex.Description), ex.Priority, ex.Status)
	}

	if err := exIter.Err(); err != nil {
		log.Fatalf("Failed to list exceptions: %v", err)
	}

	// Approve the first exception if one exists.
	if firstExceptionID != "" {
		approved, err := client.Matcher.Exceptions.Approve(ctx, firstExceptionID)
		if err != nil {
			log.Fatalf("Failed to approve exception: %v", err)
		}

		fmt.Printf("\nApproved exception %s (status: %s)\n", approved.ID, approved.Status)
	}

	// -----------------------------------------------------------------------
	// Step 7: Get the reconciliation summary report.
	//
	// The Reports service provides analytics scoped to a context. The
	// GetSummary method returns high-level counts and the overall match
	// rate for the reconciliation period.
	// -----------------------------------------------------------------------
	summary, err := client.Matcher.Reports.GetSummary(ctx, matchCtx.ID)
	if err != nil {
		log.Fatalf("Failed to get summary: %v", err)
	}

	fmt.Printf("\n--- Reconciliation Summary ---\n")
	fmt.Printf("  Period:     %s\n", summary.Period)
	fmt.Printf("  Total:      %d\n", summary.TotalRecords)
	fmt.Printf("  Matched:    %d\n", summary.MatchedRecords)
	fmt.Printf("  Match Rate: %.2f%%\n", summary.MatchRate*100)

	// -----------------------------------------------------------------------
	// Step 8: Error handling -- demonstrate ErrNotFound.
	// -----------------------------------------------------------------------
	_, err = client.Matcher.Contexts.Get(ctx, "non-existent-context-id")
	if err != nil {
		if errors.Is(err, lerian.ErrNotFound) {
			fmt.Println("\nCorrectly caught ErrNotFound for non-existent context")
		} else {
			fmt.Printf("\nUnexpected error: %v\n", err)
		}
	}

	fmt.Println("\nMatcher reconciliation workflow complete!")
}

// envOr returns the value of the environment variable named by key, or
// the fallback value if the variable is empty or unset.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// deref safely dereferences a *string pointer, returning "" if nil.
func deref(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}
