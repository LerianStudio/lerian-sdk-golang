package fees

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesUpdateValidatesEffectiveAmountBounds(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, _, result any) error {
			switch method {
			case "GET":
				assert.Equal(t, "/packages/pkg-001", path)
				return jsonInto(testPackage, result)
			case "PATCH":
				return jsonInto(testPackage, result)
			default:
				return nil
			}
		},
	}

	svc := newPackagesService(backend)
	minimum := "1000001.00"

	pkg, err := svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{MinimumAmount: &minimum})
	require.Error(t, err)
	assert.Nil(t, pkg)
	assert.Contains(t, err.Error(), "minimum amount must be less than or equal to maximum amount")
}

func TestPackagesUpdateMergesPartialFeePatch(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			switch method {
			case "GET":
				assert.Equal(t, "/packages/pkg-001", path)
				return jsonInto(testPackage, result)
			case "PATCH":
				assert.Equal(t, "/packages/pkg-001", path)

				input, ok := body.(*UpdatePackageInput)
				require.True(t, ok)
				require.Contains(t, input.Fees, "ted_fee")
				assert.Equal(t, "TED Transfer Fee", input.Fees["ted_fee"].FeeLabel)
				assert.Equal(t, "platform-updated-account", input.Fees["ted_fee"].CreditAccount)

				updated := testPackage
				updated.Fees = input.Fees

				return jsonInto(updated, result)
			default:
				return nil
			}
		},
	}

	svc := newPackagesService(backend)
	pkg, err := svc.Update(context.Background(), "pkg-001", &UpdatePackageInput{
		Fees: map[string]Fee{
			"ted_fee": {CreditAccount: "platform-updated-account"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, pkg)
	require.Contains(t, pkg.Fees, "ted_fee")
	assert.Equal(t, "TED Transfer Fee", pkg.Fees["ted_fee"].FeeLabel)
	assert.Equal(t, "platform-updated-account", pkg.Fees["ted_fee"].CreditAccount)
}
