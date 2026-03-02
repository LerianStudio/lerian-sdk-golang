// rules.go implements the RulesService for managing matching rules within
// a reconciliation context. Rules define expressions that are evaluated
// against record pairs to determine whether they match, ordered by priority.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// RulesService provides CRUD and reorder operations for matching rules.
type RulesService interface {
	// Create creates a new matching rule from the given input.
	Create(ctx context.Context, input *CreateRuleInput) (*Rule, error)

	// Get retrieves a matching rule by its unique identifier.
	Get(ctx context.Context, id string) (*Rule, error)

	// List returns a paginated iterator over matching rules.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Rule]

	// Update partially updates an existing matching rule.
	Update(ctx context.Context, id string, input *UpdateRuleInput) (*Rule, error)

	// Delete removes a matching rule by its unique identifier.
	Delete(ctx context.Context, id string) error

	// Reorder changes the evaluation priority of rules within a context
	// by providing an ordered list of rule IDs.
	Reorder(ctx context.Context, contextID string, input *ReorderRulesInput) error
}

// rulesService is the concrete implementation of [RulesService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type rulesService struct {
	core.BaseService
}

// newRulesService creates a new [RulesService] backed by the given
// Matcher [core.Backend].
func newRulesService(backend core.Backend) RulesService {
	return &rulesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ RulesService = (*rulesService)(nil)

// Create creates a new matching rule from the given input.
func (s *rulesService) Create(ctx context.Context, input *CreateRuleInput) (*Rule, error) {
	const operation = "Rules.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Rule", "input is required")
	}

	return core.Create[Rule, CreateRuleInput](ctx, &s.BaseService, "/rules", input)
}

// Get retrieves a matching rule by its unique identifier.
func (s *rulesService) Get(ctx context.Context, id string) (*Rule, error) {
	const operation = "Rules.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Rule", "id is required")
	}

	return core.Get[Rule](ctx, &s.BaseService, "/rules/"+url.PathEscape(id))
}

// List returns a paginated iterator over matching rules.
func (s *rulesService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Rule] {
	return core.List[Rule](ctx, &s.BaseService, "/rules", opts)
}

// Update partially updates an existing matching rule.
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

// Delete removes a matching rule by its unique identifier.
func (s *rulesService) Delete(ctx context.Context, id string) error {
	const operation = "Rules.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Rule", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/rules/"+url.PathEscape(id))
}

// Reorder changes the evaluation priority of rules within a context
// by providing an ordered list of rule IDs.
func (s *rulesService) Reorder(ctx context.Context, contextID string, input *ReorderRulesInput) error {
	const operation = "Rules.Reorder"

	if contextID == "" {
		return sdkerrors.NewValidation(operation, "Rule", "context id is required")
	}

	if input == nil {
		return sdkerrors.NewValidation(operation, "Rule", "input is required")
	}

	return s.Backend.Call(ctx, "POST", "/contexts/"+url.PathEscape(contextID)+"/rules/reorder", input, nil)
}
