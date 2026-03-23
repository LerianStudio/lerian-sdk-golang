// Package lerian provides the official Go SDK for the Lerian platform.
//
// The SDK supports five products: Midaz (financial ledger), Matcher
// (reconciliation), Tracer (compliance and audit trails), Reporter
// (analytics), and Fees (billing and fee calculation). Each product is
// configured independently and accessed through a single umbrella client.
//
// # Quick Start
//
// Create a client and configure one or more products:
//
//	client, err := lerian.New(
//	    lerian.WithMidaz(
//	        midaz.WithOnboardingURL("http://localhost:3000/v1"),
//	        midaz.WithTransactionURL("http://localhost:3001/v1"),
//	    ),
//	)
//	// Optional OAuth2 credentials can be loaded from:
//	// LERIAN_MIDAZ_CLIENT_ID, LERIAN_MIDAZ_CLIENT_SECRET, LERIAN_MIDAZ_TOKEN_URL
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Shutdown(context.Background())
//
//	// Use product APIs
//	org, err := client.Midaz.Organizations.Create(ctx, &midaz.CreateOrganizationInput{...})
//
// # Multi-Product Configuration
//
// Multiple products can be configured in a single client:
//
//	client, err := lerian.New(
//	    lerian.WithMidaz(midaz.WithOnboardingURL("..."), midaz.WithTransactionURL("..."), midaz.WithClientCredentials("...", "...", "...")),
//	    lerian.WithMatcher(matcher.WithBaseURL("..."), matcher.WithClientCredentials("...", "...", "...")),
//	)
//
// # Error Handling
//
// All SDK errors implement the standard error interface and support
// [errors.Is] for category-level matching:
//
//	if errors.Is(err, lerian.ErrNotFound) {
//	    // handle not found
//	}
//
// For richer inspection, use [errors.As] to extract the full [*Error]:
//
//	var sdkErr *lerian.Error
//	if errors.As(err, &sdkErr) {
//	    log.Printf("operation=%s resource=%s", sdkErr.Operation, sdkErr.Resource)
//	}
//
// # Environment Variables
//
// Configuration can be provided via LERIAN_* environment variables as
// fallback when explicit options are not set. The precedence order is:
//
//	explicit option > environment variable > empty (fails validation if required)
//
// See the .env.example file for the full list of supported variables.
//
// # Pointer Helpers
//
// The package provides convenience functions ([String], [Int], [Bool],
// [Float64], [Time]) that return pointers to their arguments, useful when
// constructing input types with optional fields.
package lerian
