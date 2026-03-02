package fees

import (
	"context"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// FeesService provides RPC-style fee calculation operations. Unlike
// [EstimatesService], a calculated fee may be linked to an actual
// transaction and tracks its lifecycle via a status field.
type FeesService interface {
	// Calculate computes fees for a given package, amount, and currency,
	// optionally linking the result to a transaction.
	Calculate(ctx context.Context, input *CalculateFeeInput) (*Fee, error)
}

// feesCalcService is the concrete implementation of [FeesService].
// The type is named feesCalcService (not feesService) to avoid shadowing
// the package-level FeesService interface.
type feesCalcService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ FeesService = (*feesCalcService)(nil)

// newFeesCalcService constructs a [FeesService] backed by the given
// [core.Backend].
func newFeesCalcService(backend core.Backend) FeesService {
	return &feesCalcService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Calculate performs an RPC-style fee calculation by POSTing to
// /fees/calculate.
func (s *feesCalcService) Calculate(ctx context.Context, input *CalculateFeeInput) (*Fee, error) {
	const operation = "Fees.Calculate"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Fee", "input is required")
	}

	return core.Create[Fee, CalculateFeeInput](ctx, &s.BaseService, "/fees/calculate", input)
}
