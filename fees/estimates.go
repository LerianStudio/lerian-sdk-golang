package fees

import (
	"context"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// EstimatesService provides RPC-style fee estimation (preview) operations.
// Estimates are computed without being associated with a real transaction
// and are useful for previewing fee charges before committing.
type EstimatesService interface {
	// Calculate previews the fees that would be charged for a given
	// package, amount, and currency without creating a transaction.
	Calculate(ctx context.Context, input *CalculateEstimateInput) (*Estimate, error)
}

// estimatesService is the concrete implementation of [EstimatesService].
type estimatesService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ EstimatesService = (*estimatesService)(nil)

// newEstimatesService constructs an [EstimatesService] backed by the given
// [core.Backend].
func newEstimatesService(backend core.Backend) EstimatesService {
	return &estimatesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Calculate performs an RPC-style fee estimation by POSTing to
// /estimates/calculate.
func (s *estimatesService) Calculate(ctx context.Context, input *CalculateEstimateInput) (*Estimate, error) {
	const operation = "Estimates.Calculate"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Estimate", "input is required")
	}

	return core.Create[Estimate, CalculateEstimateInput](ctx, &s.BaseService, "/estimates/calculate", input)
}
