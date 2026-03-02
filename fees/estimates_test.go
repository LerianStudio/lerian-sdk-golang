package fees

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Shared fixture
// ---------------------------------------------------------------------------

var testEstimate = Estimate{
	ID:        "est-001",
	PackageID: "pkg-001",
	Amount:    10000,
	Scale:     2,
	Currency:  "USD",
	FeeResults: []FeeResult{
		{RuleType: "flat", Amount: 100, Scale: 2, Currency: "USD", Applied: true},
	},
	TotalFee:      100,
	TotalFeeScale: 2,
	CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
}

// ---------------------------------------------------------------------------
// EstimatesService.Calculate — success
// ---------------------------------------------------------------------------

func TestEstimatesCalculate(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/estimates/calculate", path)
			assert.NotNil(t, body)

			return jsonInto(testEstimate, result)
		},
	}

	svc := newEstimatesService(backend)
	input := &CalculateEstimateInput{
		PackageID: "pkg-001",
		Amount:    10000,
		Scale:     2,
		Currency:  "USD",
	}

	est, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, est)
	assert.Equal(t, "est-001", est.ID)
	assert.Equal(t, "pkg-001", est.PackageID)
	assert.Equal(t, int64(10000), est.Amount)
	assert.Equal(t, "USD", est.Currency)
	assert.Len(t, est.FeeResults, 1)
	assert.Equal(t, int64(100), est.TotalFee)
}

// ---------------------------------------------------------------------------
// EstimatesService.Calculate — nil input guard clause
// ---------------------------------------------------------------------------

func TestEstimatesCalculateNilInput(t *testing.T) {
	svc := newEstimatesService(&mockBackend{})

	est, err := svc.Calculate(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, est)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	// Verify the error has the right operation and resource.
	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Estimates.Calculate", sdkErr.Operation)
	assert.Equal(t, "Estimate", sdkErr.Resource)
	assert.Contains(t, sdkErr.Message, "input is required")
}

// ---------------------------------------------------------------------------
// EstimatesService.Calculate — backend error propagation
// ---------------------------------------------------------------------------

func TestEstimatesCalculateBackendError(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("service unavailable")
		},
	}

	svc := newEstimatesService(backend)
	input := &CalculateEstimateInput{
		PackageID: "pkg-001",
		Amount:    10000,
		Scale:     2,
		Currency:  "USD",
	}

	est, err := svc.Calculate(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, est)
	assert.Contains(t, err.Error(), "service unavailable")
}

// ---------------------------------------------------------------------------
// EstimatesService.Calculate — verifies input body is forwarded
// ---------------------------------------------------------------------------

func TestEstimatesCalculateInputForwarded(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, body, result any) error {
			// Verify the input was passed through by checking body is not nil
			// and is the correct type.
			input, ok := body.(*CalculateEstimateInput)
			require.True(t, ok, "body should be *CalculateEstimateInput")
			assert.Equal(t, "pkg-002", input.PackageID)
			assert.Equal(t, int64(5000), input.Amount)
			assert.Equal(t, "EUR", input.Currency)

			return jsonInto(testEstimate, result)
		},
	}

	svc := newEstimatesService(backend)
	input := &CalculateEstimateInput{
		PackageID: "pkg-002",
		Amount:    5000,
		Scale:     2,
		Currency:  "EUR",
	}

	_, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ EstimatesService = (*estimatesService)(nil)
