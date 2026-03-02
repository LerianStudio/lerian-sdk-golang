// disputes.go implements the DisputesService for managing disputes in the
// Matcher service. Disputes represent formal challenges raised against
// reconciliation results or exceptions, and track the resolution workflow
// from creation through final resolution or escalation.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// DisputesService provides CRUD operations plus resolution and escalation
// workflows for reconciliation disputes.
type DisputesService interface {
	// Create creates a new dispute from the given input.
	Create(ctx context.Context, input *CreateDisputeInput) (*Dispute, error)

	// Get retrieves a dispute by its unique identifier.
	Get(ctx context.Context, id string) (*Dispute, error)

	// List returns a paginated iterator over disputes.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Dispute]

	// Update partially updates an existing dispute.
	Update(ctx context.Context, id string, input *UpdateDisputeInput) (*Dispute, error)

	// Resolve resolves a dispute with the provided resolution text.
	Resolve(ctx context.Context, id string, input *ResolveDisputeInput) (*Dispute, error)

	// Escalate escalates a dispute to a higher-level review process.
	Escalate(ctx context.Context, id string) (*Dispute, error)
}

// disputesService is the concrete implementation of [DisputesService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type disputesService struct {
	core.BaseService
}

// newDisputesService creates a new [DisputesService] backed by the given
// Matcher [core.Backend].
func newDisputesService(backend core.Backend) DisputesService {
	return &disputesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ DisputesService = (*disputesService)(nil)

// Create creates a new dispute from the given input.
func (s *disputesService) Create(ctx context.Context, input *CreateDisputeInput) (*Dispute, error) {
	const operation = "Disputes.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Dispute", "input is required")
	}

	return core.Create[Dispute, CreateDisputeInput](ctx, &s.BaseService, "/disputes", input)
}

// Get retrieves a dispute by its unique identifier.
func (s *disputesService) Get(ctx context.Context, id string) (*Dispute, error) {
	const operation = "Disputes.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Dispute", "id is required")
	}

	return core.Get[Dispute](ctx, &s.BaseService, "/disputes/"+url.PathEscape(id))
}

// List returns a paginated iterator over disputes.
func (s *disputesService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Dispute] {
	return core.List[Dispute](ctx, &s.BaseService, "/disputes", opts)
}

// Update partially updates an existing dispute.
func (s *disputesService) Update(ctx context.Context, id string, input *UpdateDisputeInput) (*Dispute, error) {
	const operation = "Disputes.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Dispute", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Dispute", "input is required")
	}

	return core.Update[Dispute, UpdateDisputeInput](ctx, &s.BaseService, "/disputes/"+url.PathEscape(id), input)
}

// Resolve resolves a dispute with the provided resolution text.
func (s *disputesService) Resolve(ctx context.Context, id string, input *ResolveDisputeInput) (*Dispute, error) {
	const operation = "Disputes.Resolve"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Dispute", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Dispute", "input is required")
	}

	return core.Action[Dispute](ctx, &s.BaseService, "/disputes/"+url.PathEscape(id)+"/resolve", input)
}

// Escalate escalates a dispute to a higher-level review process.
func (s *disputesService) Escalate(ctx context.Context, id string) (*Dispute, error) {
	const operation = "Disputes.Escalate"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Dispute", "id is required")
	}

	return core.Action[Dispute](ctx, &s.BaseService, "/disputes/"+url.PathEscape(id)+"/escalate", nil)
}
