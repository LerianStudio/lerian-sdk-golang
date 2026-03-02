// matching.go implements the MatchingService for RPC-style reconciliation
// operations including automated matching runs, manual matching, and
// monetary adjustments.
package matcher

import (
	"context"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// MatchingService provides manual matching and adjustment operations
// for reconciliation contexts. All methods are RPC-style POST endpoints.
type MatchingService interface {
	// Run triggers an automated reconciliation run for the given context.
	Run(ctx context.Context, contextID string) (*MatchResult, error)

	// Manual performs a manual match of specific source records to target
	// records within a reconciliation context.
	Manual(ctx context.Context, input *ManualMatchInput) (*MatchResult, error)

	// Adjust creates a monetary adjustment to correct discrepancies
	// within a reconciliation context.
	Adjust(ctx context.Context, input *AdjustmentInput) (*Adjustment, error)
}

// matchRunInput is the request body for the matching Run endpoint.
// It carries only the contextId field required to scope the run.
type matchRunInput struct {
	ContextID string `json:"contextId"`
}

// matchingService is the concrete implementation of [MatchingService].
type matchingService struct {
	core.BaseService
}

// newMatchingService creates a new [MatchingService] backed by the given
// [core.Backend].
func newMatchingService(backend core.Backend) MatchingService {
	return &matchingService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ MatchingService = (*matchingService)(nil)

// Run triggers an automated reconciliation run for the given context.
func (s *matchingService) Run(ctx context.Context, contextID string) (*MatchResult, error) {
	const operation = "Matching.Run"

	if contextID == "" {
		return nil, sdkerrors.NewValidation(operation, "MatchResult", "contextID is required")
	}

	input := &matchRunInput{ContextID: contextID}

	return core.Create[MatchResult, matchRunInput](ctx, &s.BaseService, "/matching/run", input)
}

// Manual performs a manual match of specific source records to target
// records within a reconciliation context.
func (s *matchingService) Manual(ctx context.Context, input *ManualMatchInput) (*MatchResult, error) {
	const operation = "Matching.Manual"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "MatchResult", "input is required")
	}

	return core.Create[MatchResult, ManualMatchInput](ctx, &s.BaseService, "/matching/manual", input)
}

// Adjust creates a monetary adjustment to correct discrepancies
// within a reconciliation context.
func (s *matchingService) Adjust(ctx context.Context, input *AdjustmentInput) (*Adjustment, error) {
	const operation = "Matching.Adjust"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Adjustment", "input is required")
	}

	return core.Create[Adjustment, AdjustmentInput](ctx, &s.BaseService, "/matching/adjust", input)
}
