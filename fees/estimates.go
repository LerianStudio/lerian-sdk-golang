package fees

import (
	"context"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

// estimatesServiceAPI provides fee estimation (preview) operations.
// Estimates are computed without being associated with a real transaction
// and are useful for previewing fee charges before committing.
type estimatesServiceAPI interface {
	// Calculate previews the fees that would be charged for a given
	// package and transaction without creating a transaction.
	Calculate(ctx context.Context, input *FeeEstimateInput) (*FeeEstimateResponse, error)
}

// estimatesService is the concrete implementation of [estimatesServiceAPI].
type estimatesService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ estimatesServiceAPI = (*estimatesService)(nil)

// newEstimatesService constructs an [estimatesServiceAPI] backed by the given
// [core.Backend].
func newEstimatesService(backend core.Backend) estimatesServiceAPI {
	return &estimatesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Calculate performs fee estimation by POSTing to /estimates.
func (s *estimatesService) Calculate(ctx context.Context, input *FeeEstimateInput) (*FeeEstimateResponse, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	return core.Create[FeeEstimateResponse, FeeEstimateInput](ctx, &s.BaseService, "/estimates", input)
}
