// Example: Tracer Validation Workflow
//
// Demonstrates the Tracer compliance and audit-trail product lifecycle
// using the Lerian SDK v3. The example walks through the full rule
// lifecycle (Draft -> Active -> Inactive), submitting a transaction for
// validation, inspecting the validation result, and querying the
// audit trail with integrity verification.
//
// Configure via environment variables:
//
//	LERIAN_TRACER_URL           (default: http://localhost:3003/v1)
//	LERIAN_TRACER_CLIENT_ID     (required with secret + token URL for OAuth2)
//	LERIAN_TRACER_CLIENT_SECRET (required with client ID + token URL for OAuth2)
//	LERIAN_TRACER_TOKEN_URL     (required with client ID + secret for OAuth2)
//	LERIAN_DEBUG                (optional, set "true" for verbose logging)
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	lerian "github.com/LerianStudio/lerian-sdk-golang"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
)

func main() {
	// -----------------------------------------------------------------------
	// Step 1: Create the SDK client configured for the Tracer product.
	//
	// Tracer uses OAuth2 client credentials. The base URL covers all Tracer
	// endpoints (rules, validations, limits, audit events).
	// -----------------------------------------------------------------------
	// NOTE: Use HTTPS URLs in production. HTTP is only for local development.
	client, err := lerian.New(lerian.Config{
		Debug: os.Getenv("LERIAN_DEBUG") == "true",
		Tracer: &tracer.Config{
			BaseURL: envOr("LERIAN_TRACER_URL", "http://localhost:3003/v1"),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer client.Shutdown(context.Background())

	ctx := context.Background()

	// -----------------------------------------------------------------------
	// Step 2: Create a compliance rule in DRAFT status.
	//
	// Rules contain conditions that are evaluated against transactions.
	// New rules start in DRAFT status and must be explicitly activated
	// before they take effect. Each condition specifies a field, an
	// operator (eq, gt, lt, contains, etc.), and a comparison value.
	// -----------------------------------------------------------------------
	ruleDesc := "Blocks transfers above R$ 50,000 for review"

	rule, err := client.Tracer.Rules.Create(ctx, &tracer.CreateRuleInput{
		Name:        "High-Value Transfer Check",
		Description: &ruleDesc,
		Priority:    1,
		Conditions: []tracer.RuleCondition{
			{
				Field:    "amount",
				Operator: "gt",
				Value:    5000000, // R$ 50,000.00 in cents
			},
			{
				Field:    "assetCode",
				Operator: "eq",
				Value:    "BRL",
			},
		},
		Actions: []string{"flag", "require_approval"},
	})
	if err != nil {
		log.Fatalf("Failed to create rule: %v", err)
	}

	fmt.Printf("Created rule: %s (status: %s)\n", rule.Name, rule.Status)

	// -----------------------------------------------------------------------
	// Step 3: Activate the rule (DRAFT -> ACTIVE).
	//
	// Only ACTIVE rules are evaluated during validation requests. The
	// state machine supports Draft -> Active -> Inactive transitions,
	// and Inactive -> Active re-activation.
	// -----------------------------------------------------------------------
	rule, err = client.Tracer.Rules.Activate(ctx, rule.ID)
	if err != nil {
		log.Fatalf("Failed to activate rule: %v", err)
	}

	fmt.Printf("Activated rule: %s (status: %s)\n", rule.Name, rule.Status)

	// -----------------------------------------------------------------------
	// Step 4: Submit a transaction for validation.
	//
	// The validation engine evaluates the submitted transaction map against
	// all active rules and limits. The result indicates whether the
	// transaction is approved, rejected, or pending review. RuleIDs is
	// optional; omitting it evaluates against ALL active rules.
	// -----------------------------------------------------------------------
	validation, err := client.Tracer.Validations.Create(ctx, &tracer.CreateValidationInput{
		Transaction: map[string]any{
			"amount":     7500000, // R$ 75,000.00 -- exceeds the R$ 50,000 threshold
			"assetCode":  "BRL",
			"type":       "transfer",
			"senderID":   "acc-sender-001",
			"receiverID": "acc-receiver-002",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create validation: %v", err)
	}

	fmt.Printf("\nValidation result: %s (status: %s)\n", validation.Result, validation.Status)
	fmt.Printf("  Rules applied: %v\n", validation.RulesApplied)

	if len(validation.Violations) > 0 {
		fmt.Printf("  Violations:    %v\n", validation.Violations)
	}

	// -----------------------------------------------------------------------
	// Step 5: Retrieve the validation by ID to inspect full details.
	//
	// The Get method returns the complete validation record including the
	// original transaction data, applied rules, and any violation details.
	// -----------------------------------------------------------------------
	fetched, err := client.Tracer.Validations.Get(ctx, validation.ID)
	if err != nil {
		log.Fatalf("Failed to get validation: %v", err)
	}

	fmt.Printf("\nFetched validation %s: result=%s\n", fetched.ID, fetched.Result)

	// -----------------------------------------------------------------------
	// Step 6: Deactivate the rule (ACTIVE -> INACTIVE).
	//
	// Deactivated rules are preserved for audit purposes but no longer
	// evaluated during validation. This is useful for temporarily
	// disabling rules without deleting them.
	// -----------------------------------------------------------------------
	rule, err = client.Tracer.Rules.Deactivate(ctx, rule.ID)
	if err != nil {
		log.Fatalf("Failed to deactivate rule: %v", err)
	}

	fmt.Printf("\nDeactivated rule: %s (status: %s)\n", rule.Name, rule.Status)

	// -----------------------------------------------------------------------
	// Step 7: List audit events.
	//
	// Every mutation in the Tracer service is recorded as an audit event.
	// The paginated iterator fetches events lazily, keeping memory usage
	// low even for large audit trails.
	// -----------------------------------------------------------------------
	fmt.Println("\n--- Recent Audit Events ---")

	auditIter := client.Tracer.AuditEvents.List(ctx, nil)
	count := 0

	for auditIter.Next(ctx) {
		event := auditIter.Item()
		fmt.Printf("  [%s] %s %s/%s by %s\n",
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Action,
			event.Resource,
			event.ResourceID,
			event.Actor,
		)

		// Verify the integrity of the first audit event as a demonstration.
		if count == 0 {
			verification, vErr := client.Tracer.AuditEvents.Verify(ctx, event.ID)
			if vErr != nil {
				fmt.Printf("  [!] Failed to verify event: %v\n", vErr)
			} else {
				fmt.Printf("  [v] Integrity check: valid=%t, hash=%s\n",
					verification.Valid, verification.Hash)
			}
		}

		count++
		// Limit output for readability.
		if count >= 5 {
			fmt.Println("  ... (truncated)")
			break
		}
	}

	if err := auditIter.Err(); err != nil {
		log.Fatalf("Failed to list audit events: %v", err)
	}

	// -----------------------------------------------------------------------
	// Step 8: Error handling -- demonstrate ErrNotFound.
	// -----------------------------------------------------------------------
	_, err = client.Tracer.Rules.Get(ctx, "non-existent-rule-id")
	if err != nil {
		if errors.Is(err, lerian.ErrNotFound) {
			fmt.Println("\nCorrectly caught ErrNotFound for non-existent rule")
		} else {
			fmt.Printf("\nUnexpected error: %v\n", err)
		}
	}

	fmt.Println("\nTracer validation workflow complete!")
}

// envOr returns the value of the environment variable named by key, or
// the fallback value if the variable is empty or unset.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
