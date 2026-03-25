package tracer

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// rulesServiceAPI manages compliance rules with a state-machine lifecycle
// (Draft -> Active -> Inactive). Each rule contains conditions that are
// evaluated against transactions and other resources.
type rulesServiceAPI interface {
	// Create creates a new compliance rule in DRAFT status.
	Create(ctx context.Context, input *CreateRuleInput) (*Rule, error)

	// Get retrieves a rule by its unique identifier.
	Get(ctx context.Context, id string) (*Rule, error)

	// List returns a paginated iterator over all rules.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Rule]

	// Update partially updates an existing rule.
	Update(ctx context.Context, id string, input *UpdateRuleInput) (*Rule, error)

	// Delete removes a rule by its unique identifier.
	Delete(ctx context.Context, id string) error

	// Activate transitions a rule from DRAFT or INACTIVE to ACTIVE status.
	Activate(ctx context.Context, id string) (*Rule, error)

	// Deactivate transitions a rule from ACTIVE to INACTIVE status.
	Deactivate(ctx context.Context, id string) (*Rule, error)

	// Draft transitions a rule back to DRAFT status for editing.
	Draft(ctx context.Context, id string) (*Rule, error)
}

// rulesService is the concrete implementation of [rulesServiceAPI].
type rulesService struct {
	core.BaseService
}

// newRulesService creates a new [rulesServiceAPI] backed by the given [core.Backend].
func newRulesService(backend core.Backend) rulesServiceAPI {
	return &rulesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface check.
var _ rulesServiceAPI = (*rulesService)(nil)

func (s *rulesService) Create(ctx context.Context, input *CreateRuleInput) (*Rule, error) {
	const operation = "Rules.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Rule", "input is required")
	}

	return core.Create[Rule, CreateRuleInput](ctx, &s.BaseService, "/rules", input)
}

func (s *rulesService) Get(ctx context.Context, id string) (*Rule, error) {
	const operation = "Rules.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Rule", "id is required")
	}

	return core.Get[Rule](ctx, &s.BaseService, "/rules/"+url.PathEscape(id))
}

func (s *rulesService) List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Rule] {
	return core.List[Rule](ctx, &s.BaseService, "/rules", opts)
}

func (s *rulesService) Update(ctx context.Context, id string, input *UpdateRuleInput) (*Rule, error) {
	const operation = "Rules.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Rule", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Rule", "input is required")
	}

	return core.Update[Rule, UpdateRuleInput](ctx, &s.BaseService, "/rules/"+url.PathEscape(id), input)
}

func (s *rulesService) Delete(ctx context.Context, id string) error {
	const operation = "Rules.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Rule", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/rules/"+url.PathEscape(id))
}

func (s *rulesService) Activate(ctx context.Context, id string) (*Rule, error) {
	const operation = "Rules.Activate"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Rule", "id is required")
	}

	return core.Action[Rule](ctx, &s.BaseService, "/rules/"+url.PathEscape(id)+"/activate", nil)
}

func (s *rulesService) Deactivate(ctx context.Context, id string) (*Rule, error) {
	const operation = "Rules.Deactivate"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Rule", "id is required")
	}

	return core.Action[Rule](ctx, &s.BaseService, "/rules/"+url.PathEscape(id)+"/deactivate", nil)
}

func (s *rulesService) Draft(ctx context.Context, id string) (*Rule, error) {
	const operation = "Rules.Draft"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Rule", "id is required")
	}

	return core.Action[Rule](ctx, &s.BaseService, "/rules/"+url.PathEscape(id)+"/draft", nil)
}
