// Package fees provides the client for the Fees calculation plugin.
//
// Fees manages fee packages, fee calculation, and fee estimation.
// Fee packages group related fee rules (with calculation models such
// as flatFee, percentual, or maxBetweenTypes) scoped to a ledger and
// optionally a segment. The calculation service injects fee legs into
// a transaction DSL, while the estimation service previews fees without
// committing.
//
// # Usage
//
// Access Fees services through the umbrella client:
//
//	client, _ := lerian.New(lerian.Config{
//	    Fees: &fees.Config{
//	        BaseURL:        "http://localhost:3005/v1",
//	        OrganizationID: "org-uuid",
//	    },
//	})
//
//	enabled := true
//	deductible := true
//
//	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
//	    FeeGroupLabel: "Standard Fees",
//	    LedgerID:      "ledger-uuid",
//	    MinimumAmount: "100.00",
//	    MaximumAmount: "10000.00",
//	    Enable:        &enabled,
//	    Fees: map[string]fees.Fee{
//	        "adminFee": {
//	            FeeLabel: "Taxa Administrativa",
//	            CalculationModel: &fees.CalculationModel{
//	                ApplicationRule: "flatFee",
//	                Calculations:    []fees.Calculation{{Type: "flat", Value: "1.50"}},
//	            },
//	            ReferenceAmount:  "originalAmount",
//	            Priority:         1,
//	            IsDeductibleFrom: &deductible,
//	            CreditAccount:    "@revenue_fees",
//	        },
//	    },
//	})
//
// # Available Services
//
//   - Packages -- fee package management (groups of fee rules with calculation models)
//   - Fees -- fee calculation via transaction DSL injection
//   - Estimates -- fee estimation for previewing charges before committing
package fees
