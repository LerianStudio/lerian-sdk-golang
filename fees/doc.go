// Package fees provides the client for the Fees calculation and billing
// service.
//
// Fees manages fee packages, fee rules, fee estimation, and fee settlement.
// Fee packages group related rules, and the estimation service calculates
// projected fees for a given transaction before it is committed.
//
// # Usage
//
// Access Fees services through the umbrella client:
//
//	client, _ := lerian.New(
//	    lerian.WithFees(
//	        fees.WithBaseURL("http://localhost:3005/v1"),
//	        fees.WithOrganizationID("org-uuid"),
//	    ),
//	)
//	// Optional OAuth2 credentials can be loaded from the matching LERIAN_FEES_* env vars.
//
//	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
//	    Name: "Standard Billing",
//	})
//
// # Available Services
//
//   - Packages -- fee package management (groups of fee rules)
//   - Fees -- fee calculation and transaction transformation
//   - Estimates -- fee estimation for pending transactions
package fees
