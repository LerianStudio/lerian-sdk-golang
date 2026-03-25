package leriantest

import (
	"reflect"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
)

func clonePackage(input fees.Package) fees.Package {
	output := input
	output.Description = cloneStringPointer(input.Description)
	output.SegmentID = cloneStringPointer(input.SegmentID)
	output.TransactionRoute = cloneStringPointer(input.TransactionRoute)
	output.WaivedAccounts = cloneStringSlicePointer(input.WaivedAccounts)
	output.Fees = cloneFeeMap(input.Fees)
	output.Enable = cloneBoolPointer(input.Enable)

	if input.DeletedAt != nil {
		deletedAt := *input.DeletedAt
		output.DeletedAt = &deletedAt
	}

	return output
}

func clonePackagePointer(input *fees.Package) *fees.Package {
	if input == nil {
		return nil
	}

	output := clonePackage(*input)
	return &output
}

func clonePackageSlice(input []fees.Package) []fees.Package {
	if input == nil {
		return nil
	}

	output := make([]fees.Package, 0, len(input))
	for _, item := range input {
		output = append(output, clonePackage(item))
	}

	return output
}

// ---------------------------------------------------------------------------
// Clone helpers — deep-copy TransactionDSL types so the fake never
// shares mutable state with caller code.
// ---------------------------------------------------------------------------

func cloneTransactionDSL(input fees.TransactionDSL) fees.TransactionDSL {
	output := input
	output.Metadata = cloneMap(input.Metadata)
	output.TransactionDate = cloneAny(input.TransactionDate)
	output.Send = cloneTransactionDSLSend(input.Send)

	return output
}

func cloneTransactionDSLSend(input fees.TransactionDSLSend) fees.TransactionDSLSend {
	return fees.TransactionDSLSend{
		Asset:      input.Asset,
		Value:      cloneAny(input.Value),
		Source:     cloneTransactionDSLSource(input.Source),
		Distribute: cloneTransactionDSLDistribute(input.Distribute),
	}
}

func cloneTransactionDSLSource(input fees.TransactionDSLSource) fees.TransactionDSLSource {
	output := fees.TransactionDSLSource{
		Remaining: input.Remaining,
		From:      make([]fees.TransactionDSLLeg, 0, len(input.From)),
	}

	for _, leg := range input.From {
		output.From = append(output.From, cloneTransactionDSLLeg(leg))
	}

	return output
}

func cloneTransactionDSLDistribute(input fees.TransactionDSLDistribute) fees.TransactionDSLDistribute {
	output := fees.TransactionDSLDistribute{
		Remaining: input.Remaining,
		To:        make([]fees.TransactionDSLLeg, 0, len(input.To)),
	}

	for _, leg := range input.To {
		output.To = append(output.To, cloneTransactionDSLLeg(leg))
	}

	return output
}

func cloneTransactionDSLLeg(input fees.TransactionDSLLeg) fees.TransactionDSLLeg {
	output := input
	output.Amount = cloneTransactionDSLAmount(input.Amount)
	output.Share = cloneTransactionDSLShare(input.Share)
	output.Rate = cloneTransactionDSLRate(input.Rate)
	output.Metadata = cloneMap(input.Metadata)

	return output
}

func cloneTransactionDSLAmount(input *fees.TransactionDSLAmount) *fees.TransactionDSLAmount {
	if input == nil {
		return nil
	}

	output := *input
	output.Value = cloneAny(input.Value)

	return &output
}

func cloneTransactionDSLShare(input *fees.TransactionDSLShare) *fees.TransactionDSLShare {
	if input == nil {
		return nil
	}

	output := *input

	return &output
}

func cloneTransactionDSLRate(input *fees.TransactionDSLRate) *fees.TransactionDSLRate {
	if input == nil {
		return nil
	}

	output := *input
	output.Value = cloneAny(input.Value)

	return &output
}

// ---------------------------------------------------------------------------
// Primitive clone helpers
// ---------------------------------------------------------------------------

func cloneStringPointer(input *string) *string {
	if input == nil {
		return nil
	}

	value := *input

	return &value
}

func cloneBoolPointer(input *bool) *bool {
	if input == nil {
		return nil
	}

	value := *input

	return &value
}

func cloneStringSlicePointer(input *[]string) *[]string {
	if input == nil {
		return nil
	}

	cloned := make([]string, len(*input))
	copy(cloned, *input)

	return &cloned
}

func cloneFeeMap(input map[string]fees.Fee) map[string]fees.Fee {
	if input == nil {
		return map[string]fees.Fee{}
	}

	output := make(map[string]fees.Fee, len(input))
	for k, v := range input {
		output[k] = cloneFee(v)
	}

	return output
}

func cloneFee(input fees.Fee) fees.Fee {
	output := input
	output.RouteFrom = cloneStringPointer(input.RouteFrom)
	output.RouteTo = cloneStringPointer(input.RouteTo)
	output.IsDeductibleFrom = cloneBoolPointer(input.IsDeductibleFrom)

	if input.CalculationModel != nil {
		cm := *input.CalculationModel
		cm.Calculations = make([]fees.Calculation, len(input.CalculationModel.Calculations))
		copy(cm.Calculations, input.CalculationModel.Calculations)
		output.CalculationModel = &cm
	}

	return output
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = cloneAny(value)
	}

	return output
}

func cloneAny(input any) any {
	if input == nil {
		return nil
	}

	return cloneReflectValue(reflect.ValueOf(input)).Interface()
}

func cloneReflectValue(input reflect.Value) reflect.Value {
	if !input.IsValid() {
		return input
	}

	switch input.Kind() {
	case reflect.Interface:
		if input.IsNil() {
			return reflect.Zero(input.Type())
		}

		cloned := cloneReflectValue(input.Elem())
		output := reflect.New(input.Type()).Elem()
		output.Set(cloned)

		return output
	case reflect.Pointer:
		if input.IsNil() {
			return reflect.Zero(input.Type())
		}

		output := reflect.New(input.Type().Elem())
		output.Elem().Set(cloneReflectValue(input.Elem()))

		return output
	case reflect.Map:
		if input.IsNil() {
			return reflect.Zero(input.Type())
		}

		output := reflect.MakeMapWithSize(input.Type(), input.Len())
		iter := input.MapRange()
		for iter.Next() {
			output.SetMapIndex(iter.Key(), cloneReflectValue(iter.Value()))
		}

		return output
	case reflect.Slice:
		if input.IsNil() {
			return reflect.Zero(input.Type())
		}

		output := reflect.MakeSlice(input.Type(), input.Len(), input.Len())
		for i := 0; i < input.Len(); i++ {
			output.Index(i).Set(cloneReflectValue(input.Index(i)))
		}

		return output
	case reflect.Array:
		output := reflect.New(input.Type()).Elem()
		for i := 0; i < input.Len(); i++ {
			output.Index(i).Set(cloneReflectValue(input.Index(i)))
		}

		return output
	default:
		return input
	}
}
