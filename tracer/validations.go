package tracer

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// validationsServiceAPI evaluates transactions against active rules and limits.
// Create is an RPC-style operation that submits a transaction for validation
// and returns the resulting decision (approved, rejected, or pending).
type validationsServiceAPI interface {
	// Create submits a transaction for validation against active rules and limits.
	Create(ctx context.Context, input *CreateValidationInput) (*Validation, error)

	// Get retrieves a validation result by its unique identifier.
	Get(ctx context.Context, id string) (*Validation, error)

	// List returns a paginated iterator over all validation results.
	List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Validation]
}

// validationsService is the concrete implementation of [validationsServiceAPI].
type validationsService struct {
	core.BaseService
}

// newValidationsService creates a new [validationsServiceAPI] backed by the given [core.Backend].
func newValidationsService(backend core.Backend) validationsServiceAPI {
	return &validationsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface check.
var _ validationsServiceAPI = (*validationsService)(nil)

func (s *validationsService) Create(ctx context.Context, input *CreateValidationInput) (*Validation, error) {
	const operation = "Validations.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Validation", "input is required")
	}

	return core.Create[Validation, CreateValidationInput](ctx, &s.BaseService, "/validations", input)
}

func (s *validationsService) Get(ctx context.Context, id string) (*Validation, error) {
	const operation = "Validations.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Validation", "id is required")
	}

	return core.Get[Validation](ctx, &s.BaseService, "/validations/"+url.PathEscape(id))
}

func (s *validationsService) List(ctx context.Context, opts *models.CursorListOptions) *pagination.Iterator[Validation] {
	return core.List[Validation](ctx, &s.BaseService, "/validations", opts)
}
