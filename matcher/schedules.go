// schedules.go implements the SchedulesService for managing recurring
// reconciliation schedules. Schedules use cron expressions to define when
// automated reconciliation runs are triggered.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// SchedulesService provides CRUD operations for reconciliation schedules.
type SchedulesService interface {
	// Create creates a new reconciliation schedule from the given input.
	Create(ctx context.Context, input *CreateScheduleInput) (*Schedule, error)

	// Get retrieves a reconciliation schedule by its unique identifier.
	Get(ctx context.Context, id string) (*Schedule, error)

	// List returns a paginated iterator over reconciliation schedules.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Schedule]

	// Update partially updates an existing reconciliation schedule.
	Update(ctx context.Context, id string, input *UpdateScheduleInput) (*Schedule, error)

	// Delete removes a reconciliation schedule by its unique identifier.
	Delete(ctx context.Context, id string) error
}

// schedulesService is the concrete implementation of [SchedulesService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type schedulesService struct {
	core.BaseService
}

// newSchedulesService creates a new [SchedulesService] backed by the given
// Matcher [core.Backend].
func newSchedulesService(backend core.Backend) SchedulesService {
	return &schedulesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ SchedulesService = (*schedulesService)(nil)

// Create creates a new reconciliation schedule from the given input.
func (s *schedulesService) Create(ctx context.Context, input *CreateScheduleInput) (*Schedule, error) {
	const operation = "Schedules.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Schedule", "input is required")
	}

	return core.Create[Schedule, CreateScheduleInput](ctx, &s.BaseService, "/schedules", input)
}

// Get retrieves a reconciliation schedule by its unique identifier.
func (s *schedulesService) Get(ctx context.Context, id string) (*Schedule, error) {
	const operation = "Schedules.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Schedule", "id is required")
	}

	return core.Get[Schedule](ctx, &s.BaseService, "/schedules/"+url.PathEscape(id))
}

// List returns a paginated iterator over reconciliation schedules.
func (s *schedulesService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Schedule] {
	return core.List[Schedule](ctx, &s.BaseService, "/schedules", opts)
}

// Update partially updates an existing reconciliation schedule.
func (s *schedulesService) Update(ctx context.Context, id string, input *UpdateScheduleInput) (*Schedule, error) {
	const operation = "Schedules.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Schedule", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Schedule", "input is required")
	}

	return core.Update[Schedule, UpdateScheduleInput](ctx, &s.BaseService, "/schedules/"+url.PathEscape(id), input)
}

// Delete removes a reconciliation schedule by its unique identifier.
func (s *schedulesService) Delete(ctx context.Context, id string) error {
	const operation = "Schedules.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Schedule", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/schedules/"+url.PathEscape(id))
}
