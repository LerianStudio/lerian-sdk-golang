// field_maps.go implements the FieldMapsService for managing field mappings
// within a reconciliation context. Field maps control how data is aligned
// across different sources for comparison.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// FieldMapsService provides CRUD operations for field mappings within a
// reconciliation context.
type FieldMapsService interface {
	// Create creates a new field mapping from the given input.
	Create(ctx context.Context, input *CreateFieldMapInput) (*FieldMap, error)

	// Get retrieves a field mapping by its unique identifier.
	Get(ctx context.Context, id string) (*FieldMap, error)

	// List returns a paginated iterator over field mappings.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[FieldMap]

	// Update partially updates an existing field mapping.
	Update(ctx context.Context, id string, input *UpdateFieldMapInput) (*FieldMap, error)

	// Delete removes a field mapping by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// fieldMapsService is the concrete implementation of [FieldMapsService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type fieldMapsService struct {
	core.BaseService
}

// newFieldMapsService creates a new [FieldMapsService] backed by the given
// Matcher [core.Backend].
func newFieldMapsService(backend core.Backend) FieldMapsService {
	return &fieldMapsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ FieldMapsService = (*fieldMapsService)(nil)

// Create creates a new field mapping from the given input.
func (s *fieldMapsService) Create(ctx context.Context, input *CreateFieldMapInput) (*FieldMap, error) {
	const operation = "FieldMaps.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "FieldMap", "input is required")
	}

	return core.Create[FieldMap, CreateFieldMapInput](ctx, &s.BaseService, "/field-maps", input)
}

// Get retrieves a field mapping by its unique identifier.
func (s *fieldMapsService) Get(ctx context.Context, id string) (*FieldMap, error) {
	const operation = "FieldMaps.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "FieldMap", "id is required")
	}

	return core.Get[FieldMap](ctx, &s.BaseService, "/field-maps/"+url.PathEscape(id))
}

// List returns a paginated iterator over field mappings.
func (s *fieldMapsService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[FieldMap] {
	return core.List[FieldMap](ctx, &s.BaseService, "/field-maps", opts)
}

// Update partially updates an existing field mapping.
func (s *fieldMapsService) Update(ctx context.Context, id string, input *UpdateFieldMapInput) (*FieldMap, error) {
	const operation = "FieldMaps.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "FieldMap", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "FieldMap", "input is required")
	}

	return core.Update[FieldMap, UpdateFieldMapInput](ctx, &s.BaseService, "/field-maps/"+url.PathEscape(id), input)
}

// Delete removes a field mapping by its unique identifier.
func (s *fieldMapsService) Delete(ctx context.Context, id string) error {
	const operation = "FieldMaps.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "FieldMap", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/field-maps/"+url.PathEscape(id))
}
