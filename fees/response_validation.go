package fees

import (
	"strings"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

const (
	estimateNoMatchMessage       = "no matching fee rules found"
	estimateNoMatchDetailedLabel = "no fee or gratuity rules were found for the given parameters"
)

func validateCalculatedTransaction(operation string, _ *FeeCalculate, resp *FeeCalculate) error {
	if resp == nil {
		return sdkerrors.NewInternal("fees", operation, "response body is empty", nil)
	}

	if strings.TrimSpace(resp.LedgerID) == "" {
		return sdkerrors.NewInternal("fees", operation, "response contained no ledger ID", nil)
	}

	if err := validateTransactionForCalculation(operation, feeResource, resp.Transaction); err != nil {
		return sdkerrors.NewInternal("fees", operation, "response contained invalid transaction data", err)
	}

	return nil
}

func validateEstimateResponse(operation string, resp *FeeEstimateResponse) error {
	if resp == nil {
		return sdkerrors.NewInternal("fees", operation, "response body is empty", nil)
	}

	if resp.FeesApplied == nil {
		if isNoMatchEstimateMessage(resp.Message) {
			return nil
		}

		return sdkerrors.NewInternal("fees", operation, "response contained no fees payload", nil)
	}

	return validateCalculatedTransaction(operation, resp.FeesApplied, resp.FeesApplied)
}

func isNoMatchEstimateMessage(message string) bool {
	normalized := strings.TrimSpace(strings.ToLower(message))
	return normalized == estimateNoMatchMessage || normalized == estimateNoMatchDetailedLabel
}
