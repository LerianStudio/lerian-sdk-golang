// contexts.go implements the contextsServiceAPI for managing reconciliation
// contexts in the Matcher service. Contexts are the top-level scope for all
// reconciliation operations, rules, sources, and related configuration.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// contextsServiceAPI provides CRUD operations and cloning for reconciliation
// contexts.
type contextsServiceAPI interface {
	// Create creates a new reconciliation context from the given input.
	Create(ctx context.Context, input *CreateContextInput) (*Context, error)

	// Get retrieves a reconciliation context by its unique identifier.
	Get(ctx context.Context, id string) (*Context, error)

	// List returns a paginated iterator over reconciliation contexts.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Context]

	// Update partially updates an existing reconciliation context.
	Update(ctx context.Context, id string, input *UpdateContextInput) (*Context, error)

	// Delete removes a reconciliation context by its unique identifier.
	Delete(ctx context.Context, id string) error

	// Clone creates a deep copy of an existing reconciliation context,
	// including its rules, sources, and configuration.
	Clone(ctx context.Context, id string) (*Context, error)
}

// contextsService is the concrete implementation of [contextsServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type contextsService struct {
	core.BaseService
}

// newContextsService creates a new [contextsServiceAPI] backed by the given
// Matcher [core.Backend].
func newContextsService(backend core.Backend) contextsServiceAPI {
	return &contextsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ contextsServiceAPI = (*contextsService)(nil)

// Create creates a new reconciliation context from the given input.
func (s *contextsService) Create(ctx context.Context, input *CreateContextInput) (*Context, error) {
	const operation = "Contexts.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Context", "input is required")
	}

	return core.Create[Context, CreateContextInput](ctx, &s.BaseService, "/contexts", input)
}

// Get retrieves a reconciliation context by its unique identifier.
func (s *contextsService) Get(ctx context.Context, id string) (*Context, error) {
	const operation = "Contexts.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Context", "id is required")
	}

	return core.Get[Context](ctx, &s.BaseService, "/contexts/"+url.PathEscape(id))
}

// List returns a paginated iterator over reconciliation contexts.
func (s *contextsService) List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Context] {
	return core.List[Context](ctx, &s.BaseService, "/contexts", opts)
}

// Update partially updates an existing reconciliation context.
func (s *contextsService) Update(ctx context.Context, id string, input *UpdateContextInput) (*Context, error) {
	const operation = "Contexts.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Context", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Context", "input is required")
	}

	return core.Update[Context, UpdateContextInput](ctx, &s.BaseService, "/contexts/"+url.PathEscape(id), input)
}

// Delete removes a reconciliation context by its unique identifier.
func (s *contextsService) Delete(ctx context.Context, id string) error {
	const operation = "Contexts.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Context", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/contexts/"+url.PathEscape(id))
}

// Clone creates a deep copy of an existing reconciliation context,
// including its rules, sources, and configuration.
func (s *contextsService) Clone(ctx context.Context, id string) (*Context, error) {
	const operation = "Contexts.Clone"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Context", "id is required")
	}

	return core.Action[Context](ctx, &s.BaseService, "/contexts/"+url.PathEscape(id)+"/clone", nil)
}
