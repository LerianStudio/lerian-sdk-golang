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
//	        fees.WithAuthToken("my-token"),
//	        fees.WithOrganizationID("org-uuid"),
//	    ),
//	)
//
//	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
//	    Name: "Standard Billing",
//	})
//
// # Available Services
//
//   - Packages -- fee package management (groups of fee rules)
//   - Fees -- individual fee rule configuration
//   - Estimates -- fee estimation for pending transactions
package fees
