package fees

import "time"

var testPackage = Package{
	ID:            "pkg-001",
	FeeGroupLabel: "standard_fees",
	Description:   strPtr("Default fee package for TED transfers"),
	SegmentID:     strPtr("seg-retail"),
	LedgerID:      "ledger-001",
	MinimumAmount: "0.00",
	MaximumAmount: "1000000.00",
	Fees: map[string]Fee{
		"ted_fee": {
			FeeLabel: "TED Transfer Fee",
			CalculationModel: &CalculationModel{
				ApplicationRule: "flatFee",
				Calculations: []Calculation{
					{Type: "flat", Value: "5.00"},
				},
			},
			ReferenceAmount:  "originalAmount",
			Priority:         1,
			IsDeductibleFrom: boolPtr(false),
			CreditAccount:    "platform-fee-account",
		},
	},
	Enable:    boolPtr(true),
	CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
}
