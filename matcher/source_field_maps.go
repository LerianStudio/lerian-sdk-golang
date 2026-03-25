// source_field_maps.go implements the sourceFieldMapsServiceAPI for managing
// source field mappings. Source field maps define how a field from a data
// source maps to the canonical reconciliation schema, with optional
// transforms for normalization.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// sourceFieldMapsServiceAPI provides CRUD operations for source field mappings.
type sourceFieldMapsServiceAPI interface {
	// Create creates a new source field mapping from the given input.
	Create(ctx context.Context, input *CreateSourceFieldMapInput) (*SourceFieldMap, error)

	// Get retrieves a source field mapping by its unique identifier.
	Get(ctx context.Context, id string) (*SourceFieldMap, error)

	// List returns a paginated iterator over source field mappings.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[SourceFieldMap]

	// Update partially updates an existing source field mapping.
	Update(ctx context.Context, id string, input *UpdateSourceFieldMapInput) (*SourceFieldMap, error)

	// Delete removes a source field mapping by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// sourceFieldMapsService is the concrete implementation of [sourceFieldMapsServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type sourceFieldMapsService struct {
	core.BaseService
}

// newSourceFieldMapsService creates a new [sourceFieldMapsServiceAPI] backed by
// the given Matcher [core.Backend].
func newSourceFieldMapsService(backend core.Backend) sourceFieldMapsServiceAPI {
	return &sourceFieldMapsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ sourceFieldMapsServiceAPI = (*sourceFieldMapsService)(nil)

// Create creates a new source field mapping from the given input.
func (s *sourceFieldMapsService) Create(ctx context.Context, input *CreateSourceFieldMapInput) (*SourceFieldMap, error) {
	const operation = "SourceFieldMaps.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "SourceFieldMap", "input is required")
	}

	return core.Create[SourceFieldMap, CreateSourceFieldMapInput](ctx, &s.BaseService, "/source-field-maps", input)
}

// Get retrieves a source field mapping by its unique identifier.
func (s *sourceFieldMapsService) Get(ctx context.Context, id string) (*SourceFieldMap, error) {
	const operation = "SourceFieldMaps.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "SourceFieldMap", "id is required")
	}

	return core.Get[SourceFieldMap](ctx, &s.BaseService, "/source-field-maps/"+url.PathEscape(id))
}

// List returns a paginated iterator over source field mappings.
func (s *sourceFieldMapsService) List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[SourceFieldMap] {
	return core.List[SourceFieldMap](ctx, &s.BaseService, "/source-field-maps", opts)
}

// Update partially updates an existing source field mapping.
func (s *sourceFieldMapsService) Update(ctx context.Context, id string, input *UpdateSourceFieldMapInput) (*SourceFieldMap, error) {
	const operation = "SourceFieldMaps.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "SourceFieldMap", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "SourceFieldMap", "input is required")
	}

	return core.Update[SourceFieldMap, UpdateSourceFieldMapInput](ctx, &s.BaseService, "/source-field-maps/"+url.PathEscape(id), input)
}

// Delete removes a source field mapping by its unique identifier.
func (s *sourceFieldMapsService) Delete(ctx context.Context, id string) error {
	const operation = "SourceFieldMaps.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "SourceFieldMap", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/source-field-maps/"+url.PathEscape(id))
}
