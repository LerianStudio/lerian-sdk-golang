// segments.go implements the SegmentsService for managing segment resources
// within a Midaz ledger. Segments classify and group accounts for reporting
// and access-control purposes.
package midaz

import (
	"context"
	"fmt"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// SegmentsService provides CRUD operations for segments within a ledger.
type SegmentsService interface {
	// Create creates a new segment within the specified ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateSegmentInput) (*Segment, error)

	// Get retrieves a single segment by its ID.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Segment, error)

	// List returns a paginated iterator over segments in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[Segment]

	// Update modifies an existing segment. Only non-nil fields in the
	// input are sent in the PATCH request.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateSegmentInput) (*Segment, error)

	// Delete removes a segment by its ID.
	Delete(ctx context.Context, orgID, ledgerID, id string) error
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

// segmentsService is the concrete implementation of [SegmentsService].
// It embeds [core.BaseService] for shared HTTP infrastructure and delegates
// all transport work to the generic core helpers.
type segmentsService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ SegmentsService = (*segmentsService)(nil)

// newSegmentsService creates a [SegmentsService] backed by the given
// [core.Backend] (expected to point at the onboarding API).
func newSegmentsService(backend core.Backend) SegmentsService {
	return &segmentsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// ---------------------------------------------------------------------------
// Path helpers
// ---------------------------------------------------------------------------

// segmentsBasePath returns the collection URL for segments within a ledger.
func segmentsBasePath(orgID, ledgerID string) string {
	return fmt.Sprintf("/organizations/%s/ledgers/%s/segments", url.PathEscape(orgID), url.PathEscape(ledgerID))
}

// segmentsItemPath returns the resource URL for a specific segment.
func segmentsItemPath(orgID, ledgerID, id string) string {
	return fmt.Sprintf("/organizations/%s/ledgers/%s/segments/%s", url.PathEscape(orgID), url.PathEscape(ledgerID), url.PathEscape(id))
}

// ---------------------------------------------------------------------------
// CRUD methods
// ---------------------------------------------------------------------------

const segmentResource = "Segment"

// Create creates a new segment.
func (s *segmentsService) Create(ctx context.Context, orgID, ledgerID string, input *CreateSegmentInput) (*Segment, error) {
	const operation = "Segments.Create"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "organization ID is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "ledger ID is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "input is required")
	}

	return core.Create[Segment, CreateSegmentInput](ctx, &s.BaseService, segmentsBasePath(orgID, ledgerID), input)
}

// Get retrieves a single segment by ID.
func (s *segmentsService) Get(ctx context.Context, orgID, ledgerID, id string) (*Segment, error) {
	const operation = "Segments.Get"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "organization ID is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "ledger ID is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "segment ID is required")
	}

	return core.Get[Segment](ctx, &s.BaseService, segmentsItemPath(orgID, ledgerID, id))
}

// List returns a paginated iterator over segments.
func (s *segmentsService) List(ctx context.Context, orgID, ledgerID string, opts *models.ListOptions) *pagination.Iterator[Segment] {
	if orgID == "" || ledgerID == "" {
		return pagination.NewIterator[Segment](func(_ context.Context, _ string) ([]Segment, string, error) {
			return nil, "", sdkerrors.NewValidation("Segments.List", segmentResource, "organization ID and ledger ID are required")
		})
	}

	return core.List[Segment](ctx, &s.BaseService, segmentsBasePath(orgID, ledgerID), opts)
}

// Update modifies an existing segment.
func (s *segmentsService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateSegmentInput) (*Segment, error) {
	const operation = "Segments.Update"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "organization ID is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "ledger ID is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "segment ID is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, segmentResource, "input is required")
	}

	return core.Update[Segment, UpdateSegmentInput](ctx, &s.BaseService, segmentsItemPath(orgID, ledgerID, id), input)
}

// Delete removes a segment by ID.
func (s *segmentsService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "Segments.Delete"

	if orgID == "" {
		return sdkerrors.NewValidation(operation, segmentResource, "organization ID is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, segmentResource, "ledger ID is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, segmentResource, "segment ID is required")
	}

	return core.Delete(ctx, &s.BaseService, segmentsItemPath(orgID, ledgerID, id))
}
