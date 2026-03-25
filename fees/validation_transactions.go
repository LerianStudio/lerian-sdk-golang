package fees

import (
	"strconv"
	"strings"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// Validate checks whether the fee calculation request is complete and valid.
func (input *FeeCalculate) Validate() error {
	const operation = "Fees.Calculate"

	if input == nil {
		return sdkerrors.NewValidation(operation, feeResource, "input is required")
	}

	if strings.TrimSpace(input.LedgerID) == "" {
		return sdkerrors.NewValidation(operation, feeResource, "ledger ID is required")
	}

	return validateTransactionForCalculation(operation, feeResource, input.Transaction)
}

// Validate checks whether the fee estimate request is complete and valid.
func (input *FeeEstimateInput) Validate() error {
	const operation = "Estimates.Calculate"

	if input == nil {
		return sdkerrors.NewValidation(operation, estimateResource, "input is required")
	}

	if strings.TrimSpace(input.PackageID) == "" {
		return sdkerrors.NewValidation(operation, estimateResource, "package ID is required")
	}

	if strings.TrimSpace(input.LedgerID) == "" {
		return sdkerrors.NewValidation(operation, estimateResource, "ledger ID is required")
	}

	return validateTransactionForCalculation(operation, estimateResource, input.Transaction)
}

func validateTransactionForCalculation(operation, resource string, tx TransactionDSL) error {
	if strings.TrimSpace(tx.Send.Asset) == "" {
		return sdkerrors.NewValidation(operation, resource, "transaction send asset is required")
	}

	if err := validateTransactionValue(operation, resource, "transaction send value", tx.Send.Value); err != nil {
		return err
	}

	if len(tx.Send.Source.From) == 0 {
		return sdkerrors.NewValidation(operation, resource, "transaction source legs are required")
	}

	if len(tx.Send.Distribute.To) == 0 {
		return sdkerrors.NewValidation(operation, resource, "transaction distribute legs are required")
	}

	if err := validateTransactionLegs(operation, resource, "source", tx.Send.Source.From); err != nil {
		return err
	}

	return validateTransactionLegs(operation, resource, "distribute", tx.Send.Distribute.To)
}

func validateTransactionLegs(operation, resource, side string, legs []TransactionDSLLeg) error {
	for i, leg := range legs {
		index := strconv.Itoa(i)

		if strings.TrimSpace(leg.AccountAlias) == "" && strings.TrimSpace(leg.BalanceKey) == "" {
			return sdkerrors.NewValidation(operation, resource, "transaction "+side+" leg identifier is required at index "+index)
		}

		if leg.Amount == nil && leg.Share == nil {
			return sdkerrors.NewValidation(operation, resource, "transaction "+side+" leg amount or share is required at index "+index)
		}

		if leg.Amount != nil {
			if strings.TrimSpace(leg.Amount.Asset) == "" {
				return sdkerrors.NewValidation(operation, resource, "transaction "+side+" leg amount asset is required at index "+index)
			}

			if err := validateTransactionIndexedValue(operation, resource, "transaction "+side+" leg amount value", index, leg.Amount.Value); err != nil {
				return err
			}
		}

		if leg.Share != nil && leg.Share.Percentage <= 0 && leg.Share.PercentageOfPercentage <= 0 {
			return sdkerrors.NewValidation(operation, resource, "transaction "+side+" leg share must be greater than zero at index "+index)
		}
	}

	return nil
}
