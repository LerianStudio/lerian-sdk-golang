package fees

type nilStringer struct{}

func (*nilStringer) String() string { return "10.00" }

func testTransactionDSL() TransactionDSL {
	return TransactionDSL{
		Description: "TED transfer",
		Route:       "ted_out",
		Pending:     true,
		Metadata: map[string]any{
			"transferType": "TED_OUT",
		},
		Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "15000",
			Source: TransactionDSLSource{
				From: []TransactionDSLLeg{
					{
						AccountAlias: "sender",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
				},
			},
			Distribute: TransactionDSLDistribute{
				To: []TransactionDSLLeg{
					{
						AccountAlias: "recipient",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
				},
			},
		},
	}
}

var testFeeCalculateResponse = FeeCalculate{
	SegmentID: strPtr("seg-retail"),
	LedgerID:  "ledger-001",
	Transaction: TransactionDSL{
		Description: "TED transfer",
		Route:       "ted_out",
		Pending:     true,
		Metadata: map[string]any{
			"transferType":     "TED_OUT",
			"packageAppliedID": "fee-package-uuid",
		},
		Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "15500",
			Source: TransactionDSLSource{
				From: []TransactionDSLLeg{
					{
						AccountAlias: "sender",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
					{
						AccountAlias: "sender",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "500"},
						Metadata: map[string]any{
							"feeLabel": "ted_transfer_fee",
						},
					},
				},
			},
			Distribute: TransactionDSLDistribute{
				To: []TransactionDSLLeg{
					{
						AccountAlias: "recipient",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
					{
						AccountAlias: "platform-fee-account",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "500"},
						Metadata: map[string]any{
							"feeLabel": "ted_transfer_fee",
						},
					},
				},
			},
		},
	},
}

var testEstimateResponse = FeeEstimateResponse{
	Message: "fees calculated successfully",
	FeesApplied: &FeeCalculate{
		SegmentID: strPtr("seg-retail"),
		LedgerID:  "ledger-001",
		Transaction: TransactionDSL{
			Description: "TED transfer",
			Route:       "ted_out",
			Pending:     true,
			Send: TransactionDSLSend{
				Asset: "BRL",
				Value: "15500",
				Source: TransactionDSLSource{
					From: []TransactionDSLLeg{
						{
							AccountAlias: "sender",
							Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
						},
						{
							AccountAlias: "sender",
							Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "500"},
							Metadata:     map[string]any{"feeLabel": "ted_transfer_fee"},
						},
					},
				},
				Distribute: TransactionDSLDistribute{
					To: []TransactionDSLLeg{
						{
							AccountAlias: "recipient",
							Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
						},
						{
							AccountAlias: "platform-fee-account",
							Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "500"},
							Metadata:     map[string]any{"feeLabel": "ted_transfer_fee"},
						},
					},
				},
			},
		},
	},
}
