package fees

import (
	"reflect"
	"strconv"
	"strings"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	packageResource  = "Package"
	feeResource      = "Fee"
	estimateResource = "Estimate"

	referenceAmountOriginal = "originalAmount"
	referenceAmountAfter    = "afterFeesAmount"

	applicationRuleFlat       = "flatFee"
	applicationRulePercentual = "percentual"
	applicationRuleMaxBetween = "maxBetweenTypes"

	calculationTypeFlat       = "flat"
	calculationTypePercentage = "percentage"
)

// Validate checks whether the create package input is complete and internally consistent.
func (input *CreatePackageInput) Validate() error {
	const operation = "Packages.Create"

	if input == nil {
		return sdkerrors.NewValidation(operation, packageResource, "input is required")
	}

	if strings.TrimSpace(input.FeeGroupLabel) == "" {
		return sdkerrors.NewValidation(operation, packageResource, "fee group label is required")
	}

	if strings.TrimSpace(input.LedgerID) == "" {
		return sdkerrors.NewValidation(operation, packageResource, "ledger ID is required")
	}

	if strings.TrimSpace(input.MinimumAmount) == "" {
		return sdkerrors.NewValidation(operation, packageResource, "minimum amount is required")
	}

	if strings.TrimSpace(input.MaximumAmount) == "" {
		return sdkerrors.NewValidation(operation, packageResource, "maximum amount is required")
	}

	if input.Enable == nil {
		return sdkerrors.NewValidation(operation, packageResource, "enable flag is required")
	}

	if input.Fees == nil {
		input.Fees = map[string]Fee{}
	}

	minAmount, err := parsePackageAmount(operation, "minimum amount", input.MinimumAmount)
	if err != nil {
		return err
	}

	maxAmount, err := parsePackageAmount(operation, "maximum amount", input.MaximumAmount)
	if err != nil {
		return err
	}

	if minAmount.GreaterThan(maxAmount) {
		return sdkerrors.NewValidation(operation, packageResource, "minimum amount must be less than or equal to maximum amount")
	}

	for key, fee := range input.Fees {
		if err := validateCreateFeeDefinition(operation, key, fee, minAmount); err != nil {
			return err
		}
	}

	return nil
}

// Validate checks whether the update package input is internally consistent.
func (input *UpdatePackageInput) Validate() error {
	const operation = "Packages.Update"

	if input == nil {
		return sdkerrors.NewValidation(operation, packageResource, "input is required")
	}

	var (
		minAmount decimal.Decimal
		haveMin   bool
	)

	if input.MinimumAmount != nil {
		parsed, err := parsePackageAmount(operation, "minimum amount", *input.MinimumAmount)
		if err != nil {
			return err
		}

		minAmount = parsed
		haveMin = true
	}

	var (
		maxAmount decimal.Decimal
		haveMax   bool
	)

	if input.MaximumAmount != nil {
		parsed, err := parsePackageAmount(operation, "maximum amount", *input.MaximumAmount)
		if err != nil {
			return err
		}

		maxAmount = parsed
		haveMax = true
	}

	if haveMin && haveMax && minAmount.GreaterThan(maxAmount) {
		return sdkerrors.NewValidation(operation, packageResource, "minimum amount must be less than or equal to maximum amount")
	}

	for key, fee := range input.Fees {
		if err := validateUpdateFeeDefinition(operation, key, fee, input.MinimumAmount); err != nil {
			return err
		}
	}

	return nil
}

// Validate checks whether the package list filters are internally consistent.
func (opts *PackageListOptions) Validate() error {
	const operation = "Packages.List"

	if opts == nil {
		return nil
	}

	if (opts.CreatedFrom == nil) != (opts.CreatedTo == nil) {
		return sdkerrors.NewValidation(operation, packageResource, "createdFrom and createdTo must both be provided")
	}

	if opts.CreatedFrom != nil && opts.CreatedTo != nil && opts.CreatedFrom.After(*opts.CreatedTo) {
		return sdkerrors.NewValidation(operation, packageResource, "createdFrom must be before or equal to createdTo")
	}

	if opts.SortOrder != "" && opts.SortOrder != "asc" && opts.SortOrder != "desc" {
		return sdkerrors.NewValidation(operation, packageResource, "sortOrder must be either asc or desc")
	}

	return nil
}

// Validate checks whether the fee calculation request is complete and well formed.
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

// Validate checks whether the fee estimate request is complete and well formed.
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

	if isMissingValue(tx.Send.Value) {
		return sdkerrors.NewValidation(operation, resource, "transaction send value is required")
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

		if leg.Amount != nil && isMissingValue(leg.Amount.Value) {
			return sdkerrors.NewValidation(operation, resource, "transaction "+side+" leg amount value is required at index "+index)
		}

		if leg.Share != nil && leg.Share.Percentage <= 0 && leg.Share.PercentageOfPercentage <= 0 {
			return sdkerrors.NewValidation(operation, resource, "transaction "+side+" leg share must be greater than zero at index "+index)
		}
	}

	return nil
}

func validateCreateFeeDefinition(operation, key string, fee Fee, minAmount decimal.Decimal) error {
	if strings.TrimSpace(fee.FeeLabel) == "" {
		return sdkerrors.NewValidation(operation, packageResource, "fee label is required for fee `"+key+"`")
	}

	if strings.TrimSpace(fee.ReferenceAmount) == "" {
		return sdkerrors.NewValidation(operation, packageResource, "reference amount is required for fee `"+key+"`")
	}

	if fee.ReferenceAmount != referenceAmountOriginal && fee.ReferenceAmount != referenceAmountAfter {
		return sdkerrors.NewValidation(operation, packageResource, "reference amount must be originalAmount or afterFeesAmount for fee `"+key+"`")
	}

	if fee.Priority < 0 {
		return sdkerrors.NewValidation(operation, packageResource, "priority must be greater than or equal to zero for fee `"+key+"`")
	}

	if fee.IsDeductibleFrom == nil {
		return sdkerrors.NewValidation(operation, packageResource, "is deductible flag is required for fee `"+key+"`")
	}

	if strings.TrimSpace(fee.CreditAccount) == "" {
		return sdkerrors.NewValidation(operation, packageResource, "credit account is required for fee `"+key+"`")
	}

	if fee.Priority == 1 && fee.ReferenceAmount != referenceAmountOriginal {
		return sdkerrors.NewValidation(operation, packageResource, "priority 1 fees must use originalAmount for fee `"+key+"`")
	}

	if fee.IsDeductibleFrom != nil && *fee.IsDeductibleFrom && fee.ReferenceAmount != referenceAmountOriginal {
		return sdkerrors.NewValidation(operation, packageResource, "deductible fees must use originalAmount for fee `"+key+"`")
	}

	return validateCalculationModel(operation, key, fee.CalculationModel, minAmount, fee.IsDeductibleFrom != nil && *fee.IsDeductibleFrom)
}

func validateUpdateFeeDefinition(operation, key string, fee Fee, minimumAmount *string) error {
	if err := validateUpdateFeeDefinitionBasics(operation, key, fee); err != nil {
		return err
	}

	if fee.CalculationModel == nil {
		return nil
	}

	minAmount, err := parseOptionalPackageAmount(operation, minimumAmount)
	if err != nil {
		return err
	}

	return validateCalculationModel(operation, key, fee.CalculationModel, minAmount, fee.IsDeductibleFrom != nil && *fee.IsDeductibleFrom)
}

func validateUpdateFeeDefinitionBasics(operation, key string, fee Fee) error {
	if fee.Priority < 0 {
		return sdkerrors.NewValidation(operation, packageResource, "priority must be greater than or equal to zero for fee `"+key+"`")
	}

	if fee.ReferenceAmount != "" && fee.ReferenceAmount != referenceAmountOriginal && fee.ReferenceAmount != referenceAmountAfter {
		return sdkerrors.NewValidation(operation, packageResource, "reference amount must be originalAmount or afterFeesAmount for fee `"+key+"`")
	}

	if fee.Priority == 1 && fee.ReferenceAmount != "" && fee.ReferenceAmount != referenceAmountOriginal {
		return sdkerrors.NewValidation(operation, packageResource, "priority 1 fees must use originalAmount for fee `"+key+"`")
	}

	if fee.IsDeductibleFrom != nil && *fee.IsDeductibleFrom && fee.ReferenceAmount == referenceAmountAfter {
		return sdkerrors.NewValidation(operation, packageResource, "deductible fees must use originalAmount for fee `"+key+"`")
	}

	return nil
}

func parseOptionalPackageAmount(operation string, minimumAmount *string) (decimal.Decimal, error) {
	if minimumAmount == nil || *minimumAmount == "" {
		return decimal.Zero, nil
	}

	return parsePackageAmount(operation, "minimum amount", *minimumAmount)
}

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

func parsePackageAmount(operation, field, value string) (decimal.Decimal, error) {
	value = strings.TrimSpace(value)

	if strings.Contains(value, ",") {
		return decimal.Zero, sdkerrors.NewValidation(operation, packageResource, field+" must use dot decimal formatting")
	}

	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, sdkerrors.NewValidation(operation, packageResource, field+" must be a valid decimal")
	}

	if parsed.IsNegative() {
		return decimal.Zero, sdkerrors.NewValidation(operation, packageResource, field+" must be greater than or equal to zero")
	}

	return parsed, nil
}

func isMissingValue(value any) bool {
	if value == nil {
		return true
	}

	if s, ok := value.(string); ok {
		return strings.TrimSpace(s) == ""
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
