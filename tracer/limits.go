package tracer

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// LimitsService manages rate and amount limits with a state-machine lifecycle
// (Draft -> Active -> Inactive). Limits enforce caps on operations such as
// maximum transaction amounts per period, daily transfer counts, and similar
// constraints.
type LimitsService interface {
	// Create creates a new limit in DRAFT status.
	Create(ctx context.Context, input *CreateLimitInput) (*Limit, error)

	// Get retrieves a limit by its unique identifier.
	Get(ctx context.Context, id string) (*Limit, error)

	// List returns a paginated iterator over all limits.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Limit]

	// Update partially updates an existing limit.
	Update(ctx context.Context, id string, input *UpdateLimitInput) (*Limit, error)

	// Delete removes a limit by its unique identifier.
	Delete(ctx context.Context, id string) error

	// Activate transitions a limit from DRAFT or INACTIVE to ACTIVE status.
	Activate(ctx context.Context, id string) (*Limit, error)

	// Deactivate transitions a limit from ACTIVE to INACTIVE status.
	Deactivate(ctx context.Context, id string) (*Limit, error)

	// Draft transitions a limit back to DRAFT status for editing.
	Draft(ctx context.Context, id string) (*Limit, error)
}

// limitsService is the concrete implementation of [LimitsService].
type limitsService struct {
	core.BaseService
}

// newLimitsService creates a new [LimitsService] backed by the given [core.Backend].
func newLimitsService(backend core.Backend) LimitsService {
	return &limitsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface check.
var _ LimitsService = (*limitsService)(nil)

func (s *limitsService) Create(ctx context.Context, input *CreateLimitInput) (*Limit, error) {
	const operation = "Limits.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Limit", "input is required")
	}

	return core.Create[Limit, CreateLimitInput](ctx, &s.BaseService, "/limits", input)
}

func (s *limitsService) Get(ctx context.Context, id string) (*Limit, error) {
	const operation = "Limits.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Limit", "id is required")
	}

	return core.Get[Limit](ctx, &s.BaseService, "/limits/"+url.PathEscape(id))
}

func (s *limitsService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Limit] {
	return core.List[Limit](ctx, &s.BaseService, "/limits", opts)
}

func (s *limitsService) Update(ctx context.Context, id string, input *UpdateLimitInput) (*Limit, error) {
	const operation = "Limits.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Limit", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Limit", "input is required")
	}

	return core.Update[Limit, UpdateLimitInput](ctx, &s.BaseService, "/limits/"+url.PathEscape(id), input)
}

func (s *limitsService) Delete(ctx context.Context, id string) error {
	const operation = "Limits.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Limit", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/limits/"+url.PathEscape(id))
}

func (s *limitsService) Activate(ctx context.Context, id string) (*Limit, error) {
	const operation = "Limits.Activate"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Limit", "id is required")
	}

	return core.Action[Limit](ctx, &s.BaseService, "/limits/"+url.PathEscape(id)+"/activate", nil)
}

func (s *limitsService) Deactivate(ctx context.Context, id string) (*Limit, error) {
	const operation = "Limits.Deactivate"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Limit", "id is required")
	}

	return core.Action[Limit](ctx, &s.BaseService, "/limits/"+url.PathEscape(id)+"/deactivate", nil)
}

func (s *limitsService) Draft(ctx context.Context, id string) (*Limit, error) {
	const operation = "Limits.Draft"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Limit", "id is required")
	}

	return core.Action[Limit](ctx, &s.BaseService, "/limits/"+url.PathEscape(id)+"/draft", nil)
}
