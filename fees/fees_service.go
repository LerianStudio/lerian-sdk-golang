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

	// TransformTransaction sends a transaction DSL to the fees service for fee
	// injection. The service returns the mutated DSL with fee legs added to the
	// source and distribute arrays.
	TransformTransaction(ctx context.Context, input *TransformTransactionInput) (*TransformTransactionOutput, error)
}

// transformResponse handles the two response envelope formats returned by the
// fees transformation endpoint.
type transformResponse struct {
	Transaction *TransactionDSL `json:"transaction,omitempty"`
	Data        *struct {
		Transaction *TransactionDSL `json:"transaction,omitempty"`
	} `json:"data,omitempty"`
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

// TransformTransaction performs a DSL-based fee transformation by POSTing to
// /fees and normalizing the two supported response envelopes.
func (s *feesCalcService) TransformTransaction(ctx context.Context, input *TransformTransactionInput) (*TransformTransactionOutput, error) {
	const operation = "Fees.TransformTransaction"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Fee", "input is required")
	}

	if input.LedgerID == "" {
		return nil, sdkerrors.NewValidation(operation, "Fee", "ledger ID is required")
	}

	resp, err := core.Create[transformResponse, TransformTransactionInput](ctx, &s.BaseService, "/fees", input)
	if err != nil {
		return nil, err
	}

	if resp.Transaction != nil {
		return &TransformTransactionOutput{Transaction: *resp.Transaction}, nil
	}

	if resp.Data != nil && resp.Data.Transaction != nil {
		return &TransformTransactionOutput{Transaction: *resp.Data.Transaction}, nil
	}

	return nil, sdkerrors.NewInternal("fees", operation, "response contained no transaction data", nil)
}
