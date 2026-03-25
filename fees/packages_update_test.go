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

func TestPackagesUpdate(t *testing.T) {
	t.Parallel()

	updated := testPackage
	updated.FeeGroupLabel = "premium_fees"

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			switch method {
			case "GET":
				assert.Equal(t, "/packages/pkg-001", path)
				assert.Nil(t, body)

				return jsonInto(testPackage, result)
			case "PATCH":
				assert.Equal(t, "/packages/pkg-001", path)
				assert.NotNil(t, body)

				return jsonInto(updated, result)
			default:
				return fmt.Errorf("unexpected method %s", method)
			}
		},
	}

	svc := newPackagesService(backend)
	input := &UpdatePackageInput{FeeGroupLabel: "premium_fees"}

	pkg, err := svc.Update(context.Background(), "pkg-001", input)
	require.NoError(t, err)
	require.NotNil(t, pkg)
	assert.Equal(t, "premium_fees", pkg.FeeGroupLabel)
}

func TestPackagesUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})
	input := &UpdatePackageInput{FeeGroupLabel: "test"}

	pkg, err := svc.Update(context.Background(), "", input)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newPackagesService(&mockBackend{})

	pkg, err := svc.Update(context.Background(), "pkg-001", nil)
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestPackagesUpdateValidatesNestedFeeContract(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, _, result any) error {
			if method == "GET" {
				assert.Equal(t, "/packages/pkg-001", path)
				return jsonInto(testPackage, result)
			}

			return nil
		},
	}

	svc := newPackagesService(backend)
	minAmount := "10.00"
	negativeMin := "-1.00"

	pkg, err := svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{
		MinimumAmount: &minAmount,
		Fees: map[string]Fee{
			"fee": {
				CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "20.00"}}},
				IsDeductibleFrom: boolPtr(true),
			},
		},
	})
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "deductible flat fee cannot exceed minimum amount")

	pkg, err = svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{
		MinimumAmount: &negativeMin,
	})
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "minimum amount must be greater than or equal to zero")

	pkg, err = svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{
		Fees: map[string]Fee{
			"fee": {
				CalculationModel: &CalculationModel{ApplicationRule: "flatFee", Calculations: []Calculation{{Type: "flat", Value: "0"}}},
			},
		},
	})
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "calculation value must be greater than zero")
}

func TestPackagesUpdateBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, _, result any) error {
			switch method {
			case "GET":
				assert.Equal(t, "/packages/pkg-001", path)
				return jsonInto(testPackage, result)
			case "PATCH":
				return fmt.Errorf("server error")
			default:
				return fmt.Errorf("unexpected method %s", method)
			}
		},
	}

	svc := newPackagesService(backend)
	input := &UpdatePackageInput{FeeGroupLabel: "test"}

	pkg, err := svc.Update(context.Background(), "pkg-001", input)
	require.Error(t, err)
	assert.Nil(t, pkg)
}
