package fees

import (
	"reflect"
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

func validateTransactionValue(operation, resource, field string, value any) error {
	if isMissingValue(value) {
		return sdkerrors.NewValidation(operation, resource, field+" is required")
	}

	return nil
}

func validateTransactionIndexedValue(operation, resource, field, index string, value any) error {
	if isMissingValue(value) {
		return sdkerrors.NewValidation(operation, resource, field+" is required at index "+index)
	}

	return validateTransactionValue(operation, resource, field, value)
}

func parseOptionalPackageAmount(operation string, minimumAmount *string) (decimal.Decimal, error) {
	if minimumAmount == nil || *minimumAmount == "" {
		return decimal.Zero, nil
	}

	return parsePackageAmount(operation, "minimum amount", *minimumAmount)
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
