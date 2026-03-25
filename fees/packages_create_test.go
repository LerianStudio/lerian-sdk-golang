package fees

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesCreate(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/packages", path)
			assert.NotNil(t, body)

			return jsonInto(testPackage, result)
		},
	}

	svc := newPackagesService(backend)
	input := &CreatePackageInput{
		FeeGroupLabel: "standard_fees",
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
				IsDeductibleFrom: boolPtr(false),
				CreditAccount:    "platform-fee-account",
			},
		},
		Enable: boolPtr(true),
	}

	pkg, err := svc.Create(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "pkg-001", pkg.ID)
	assert.Equal(t, "standard_fees", pkg.FeeGroupLabel)
	assert.Equal(t, "ledger-001", pkg.LedgerID)
	assert.Contains(t, pkg.Fees, "ted_fee")
	assert.True(t, *pkg.Enable)
}

func TestPackagesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Create(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesCreateBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("network failure")
		},
	}

	svc := newPackagesService(backend)
	input := &CreatePackageInput{
		FeeGroupLabel: "test",
		LedgerID:      "ledger-001",
		MinimumAmount: "0",
		MaximumAmount: "100",
		Fees:          map[string]Fee{},
		Enable:        boolPtr(true),
	}

	pkg, err := svc.Create(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "network failure")
}

func TestPackagesCreateRequiresCoreFields(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	tests := []struct {
		name    string
		input   *CreatePackageInput
		message string
	}{
		{name: "missing fee group label", input: &CreatePackageInput{LedgerID: "ledger-001", MinimumAmount: "0", MaximumAmount: "100", Enable: boolPtr(true)}, message: "fee group label is required"},
		{name: "missing ledger id", input: &CreatePackageInput{FeeGroupLabel: "fees", MinimumAmount: "0", MaximumAmount: "100", Enable: boolPtr(true)}, message: "ledger ID is required"},
		{name: "missing minimum amount", input: &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MaximumAmount: "100", Enable: boolPtr(true)}, message: "minimum amount is required"},
		{name: "missing maximum amount", input: &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0", Enable: boolPtr(true)}, message: "maximum amount is required"},
		{name: "missing enable flag", input: &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0", MaximumAmount: "100"}, message: "enable flag is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkg, err := svc.Create(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, pkg)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestPackagesCreateNormalizesNilFeesMap(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/packages", path)

			input, ok := body.(*CreatePackageInput)
			require.True(t, ok)
			require.NotNil(t, input.Fees)
			assert.Empty(t, input.Fees)

			return jsonInto(Package{
				ID:            "pkg-001",
				FeeGroupLabel: input.FeeGroupLabel,
				LedgerID:      input.LedgerID,
				MinimumAmount: input.MinimumAmount,
				MaximumAmount: input.MaximumAmount,
				Fees:          input.Fees,
				Enable:        input.Enable,
			}, result)
		},
	}

	svc := newPackagesService(backend)
	pkg, err := svc.Create(context.Background(), &CreatePackageInput{
		FeeGroupLabel: "fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "0.00",
		MaximumAmount: "100.00",
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)
	require.NotNil(t, pkg)
	require.NotNil(t, pkg.Fees)
	assert.Empty(t, pkg.Fees)
}

func TestPackagesCreateValidatesNestedFeeContract(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	tests := []struct {
		name    string
		input   *CreatePackageInput
		message string
	}{
		{
			name:    "invalid minimum amount ordering",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "10.00", MaximumAmount: "5.00", Enable: boolPtr(true)},
			message: "minimum amount must be less than or equal to maximum amount",
		},
		{
			name:    "negative minimum amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "-1.00", MaximumAmount: "5.00", Enable: boolPtr(true)},
			message: "minimum amount must be greater than or equal to zero",
		},
		{
			name:    "negative maximum amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "-5.00", Enable: boolPtr(true)},
			message: "maximum amount must be greater than or equal to zero",
		},
		{
			name:    "invalid reference amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "send.value", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "reference amount must be originalAmount or afterFeesAmount",
		},
		{
			name:    "negative calculation value",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "-1.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "calculation value must be greater than zero",
		},
		{
			name:    "zero calculation value",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "0"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "calculation value must be greater than zero",
		},
		{
			name:    "deductible after fees amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "10.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "afterFeesAmount", IsDeductibleFrom: boolPtr(true), CreditAccount: "acct"}}},
			message: "deductible fees must use originalAmount",
		},
		{
			name:    "invalid calculation model",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "10.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "percentual", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "percentual requires a percentage calculation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkg, err := svc.Create(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, pkg)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestPackagesCreateValidationAdditionalMatrix(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	tests := []struct {
		name    string
		input   *CreatePackageInput
		message string
	}{
		{
			name:    "whitespace fee group label",
			input:   &CreatePackageInput{FeeGroupLabel: "   ", LedgerID: "ledger-001", MinimumAmount: "0.00", MaximumAmount: "1.00", Enable: boolPtr(true)},
			message: "fee group label is required",
		},
		{
			name:    "comma decimal formatting",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "10,00", MaximumAmount: "100.00", Enable: boolPtr(true)},
			message: "minimum amount must use dot decimal formatting",
		},
		{
			name:    "missing calculation model",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "calculation model is required",
		},
		{
			name:    "unsupported application rule",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "mystery", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "unsupported application rule",
		},
		{
			name:    "empty calculations",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee"}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "at least one calculation is required",
		},
		{
			name:    "max between needs two calculations",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "maxBetweenTypes", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "maxBetweenTypes requires at least two calculations",
		},
		{
			name:    "priority one must use original amount",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "1.00"}}}, ReferenceAmount: "afterFeesAmount", Priority: 1, IsDeductibleFrom: boolPtr(false), CreditAccount: "acct"}}},
			message: "priority 1 fees must use originalAmount",
		},
		{
			name:    "deductible percentage above 100",
			input:   &CreatePackageInput{FeeGroupLabel: "fees", LedgerID: "ledger-001", MinimumAmount: "1.00", MaximumAmount: "100.00", Enable: boolPtr(true), Fees: map[string]Fee{"fee": {FeeLabel: "fee", CalculationModel: &CalculationModel{ApplicationRule: "percentual", Calculations: []Calculation{{Type: "percentage", Value: "101.00"}}}, ReferenceAmount: "originalAmount", IsDeductibleFrom: boolPtr(true), CreditAccount: "acct"}}},
			message: "deductible percentage cannot be greater than 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkg, err := svc.Create(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, pkg)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}
