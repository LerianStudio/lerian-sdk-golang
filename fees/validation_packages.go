package fees

import (
	"strings"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/shopspring/decimal"
)

// Validate checks whether the package creation payload is complete and consistent.
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

// Validate checks whether the package update payload contains valid field values.
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

	if opts.PageNumber < 0 {
		return sdkerrors.NewValidation(operation, packageResource, "pageNumber must be greater than or equal to zero")
	}

	if opts.PageSize < 0 {
		return sdkerrors.NewValidation(operation, packageResource, "pageSize must be greater than or equal to zero")
	}

	if opts.SortOrder != "" && opts.SortOrder != "asc" && opts.SortOrder != "desc" {
		return sdkerrors.NewValidation(operation, packageResource, "sortOrder must be either asc or desc")
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
