// sources.go implements the SourcesService for managing data sources
// connected to a reconciliation context. Sources define where record data
// comes from (e.g., bank feed, ERP system) and how it is accessed.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// SourcesService provides CRUD operations for data sources.
type SourcesService interface {
	// Create creates a new data source from the given input.
	Create(ctx context.Context, input *CreateSourceInput) (*Source, error)

	// Get retrieves a data source by its unique identifier.
	Get(ctx context.Context, id string) (*Source, error)

	// List returns a paginated iterator over data sources.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Source]

	// Update partially updates an existing data source.
	Update(ctx context.Context, id string, input *UpdateSourceInput) (*Source, error)

	// Delete removes a data source by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// sourcesService is the concrete implementation of [SourcesService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type sourcesService struct {
	core.BaseService
}

// newSourcesService creates a new [SourcesService] backed by the given
// Matcher [core.Backend].
func newSourcesService(backend core.Backend) SourcesService {
	return &sourcesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ SourcesService = (*sourcesService)(nil)

// Create creates a new data source from the given input.
func (s *sourcesService) Create(ctx context.Context, input *CreateSourceInput) (*Source, error) {
	const operation = "Sources.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Source", "input is required")
	}

	return core.Create[Source, CreateSourceInput](ctx, &s.BaseService, "/sources", input)
}

// Get retrieves a data source by its unique identifier.
func (s *sourcesService) Get(ctx context.Context, id string) (*Source, error) {
	const operation = "Sources.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Source", "id is required")
	}

	return core.Get[Source](ctx, &s.BaseService, "/sources/"+url.PathEscape(id))
}

// List returns a paginated iterator over data sources.
func (s *sourcesService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Source] {
	return core.List[Source](ctx, &s.BaseService, "/sources", opts)
}

// Update partially updates an existing data source.
func (s *sourcesService) Update(ctx context.Context, id string, input *UpdateSourceInput) (*Source, error) {
	const operation = "Sources.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Source", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Source", "input is required")
	}

	return core.Update[Source, UpdateSourceInput](ctx, &s.BaseService, "/sources/"+url.PathEscape(id), input)
}

// Delete removes a data source by its unique identifier.
func (s *sourcesService) Delete(ctx context.Context, id string) error {
	const operation = "Sources.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Source", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/sources/"+url.PathEscape(id))
}
