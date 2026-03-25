package fees

import (
	"context"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

// feesServiceAPI provides fee calculation operations. The Calculate method
// sends a transaction DSL to the fees service, which evaluates matching
// fee packages and returns the transaction with fee legs injected into
// the source and distribute arrays.
type feesServiceAPI interface {
	// Calculate sends a transaction DSL to the fees service for fee
	// injection. The service evaluates matching fee packages for the
	// given organization, ledger, and (optional) segment, then returns
	// the mutated transaction with fee legs added.
	Calculate(ctx context.Context, input *FeeCalculate) (*FeeCalculate, error)
}

// feesCalcService is the concrete implementation of [feesServiceAPI].
type feesCalcService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ feesServiceAPI = (*feesCalcService)(nil)

// newFeesCalcService constructs a [feesServiceAPI] backed by the given
// [core.Backend].
func newFeesCalcService(backend core.Backend) feesServiceAPI {
	return &feesCalcService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Calculate performs fee calculation by POSTing to /fees.
// The service mutates the input transaction DSL and returns it with
// fee legs injected.
func (s *feesCalcService) Calculate(ctx context.Context, input *FeeCalculate) (*FeeCalculate, error) {
	if err := ensureService(s); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	resp, err := core.Create[FeeCalculate, FeeCalculate](ctx, &s.BaseService, "/fees", input)
	if err != nil {
		return nil, err
	}

	if err := validateCalculatedTransaction("Fees.Calculate", input, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
