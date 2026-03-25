package fees

import (
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/shopspring/decimal"
)

func validateCalculationModel(operation, feeKey string, model *CalculationModel, minAmount decimal.Decimal, deductible bool) error {
	if model == nil {
		return sdkerrors.NewValidation(operation, packageResource, "calculation model is required for fee `"+feeKey+"`")
	}

	if err := validateCalculationApplicationRule(operation, feeKey, model); err != nil {
		return err
	}

	if err := validateCalculationCardinality(operation, feeKey, model); err != nil {
		return err
	}

	for i, calc := range model.Calculations {
		value, err := validateCalculationValue(operation, feeKey, calc)
		if err != nil {
			return err
		}

		if err := validateFirstCalculationForRule(operation, feeKey, model.ApplicationRule, calc, i); err != nil {
			return err
		}

		if err := validateDeductibleCalculation(operation, feeKey, calc, value, minAmount, deductible); err != nil {
			return err
		}
	}

	return nil
}

func validateCalculationApplicationRule(operation, feeKey string, model *CalculationModel) error {
	switch model.ApplicationRule {
	case applicationRuleFlat, applicationRulePercentual, applicationRuleMaxBetween:
		return nil
	default:
		return sdkerrors.NewValidation(operation, packageResource, "unsupported application rule for fee `"+feeKey+"`")
	}
}

func validateCalculationCardinality(operation, feeKey string, model *CalculationModel) error {
	if len(model.Calculations) == 0 {
		return sdkerrors.NewValidation(operation, packageResource, "at least one calculation is required for fee `"+feeKey+"`")
	}

	switch model.ApplicationRule {
	case applicationRuleFlat, applicationRulePercentual:
		if len(model.Calculations) != 1 {
			return sdkerrors.NewValidation(operation, packageResource, "flatFee and percentual rules require exactly one calculation for fee `"+feeKey+"`")
		}
	case applicationRuleMaxBetween:
		if len(model.Calculations) < 2 {
			return sdkerrors.NewValidation(operation, packageResource, "maxBetweenTypes requires at least two calculations for fee `"+feeKey+"`")
		}
	}

	return nil
}

func validateCalculationValue(operation, feeKey string, calc Calculation) (decimal.Decimal, error) {
	if calc.Type != calculationTypeFlat && calc.Type != calculationTypePercentage {
		return decimal.Zero, sdkerrors.NewValidation(operation, packageResource, "unsupported calculation type for fee `"+feeKey+"`")
	}

	value, err := decimal.NewFromString(calc.Value)
	if err != nil {
		return decimal.Zero, sdkerrors.NewValidation(operation, packageResource, "invalid decimal value for fee `"+feeKey+"` calculation")
	}

	if !value.GreaterThan(decimal.Zero) {
		return decimal.Zero, sdkerrors.NewValidation(operation, packageResource, "calculation value must be greater than zero for fee `"+feeKey+"`")
	}

	return value, nil
}

func validateFirstCalculationForRule(operation, feeKey, rule string, calc Calculation, index int) error {
	if index != 0 {
		return nil
	}

	switch rule {
	case applicationRuleFlat:
		if calc.Type != calculationTypeFlat {
			return sdkerrors.NewValidation(operation, packageResource, "flatFee requires a flat calculation for fee `"+feeKey+"`")
		}
	case applicationRulePercentual:
		if calc.Type != calculationTypePercentage {
			return sdkerrors.NewValidation(operation, packageResource, "percentual requires a percentage calculation for fee `"+feeKey+"`")
		}
	}

	return nil
}

func validateDeductibleCalculation(operation, feeKey string, calc Calculation, value, minAmount decimal.Decimal, deductible bool) error {
	if !deductible {
		return nil
	}

	if calc.Type == calculationTypePercentage && value.GreaterThan(decimal.NewFromInt(100)) {
		return sdkerrors.NewValidation(operation, packageResource, "deductible percentage cannot be greater than 100 for fee `"+feeKey+"`")
	}

	if calc.Type == calculationTypeFlat && !minAmount.IsZero() && value.GreaterThan(minAmount) {
		return sdkerrors.NewValidation(operation, packageResource, "deductible flat fee cannot exceed minimum amount for fee `"+feeKey+"`")
	}

	return nil
}
