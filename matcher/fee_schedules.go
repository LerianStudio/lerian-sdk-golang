// fee_schedules.go implements the FeeSchedulesService for managing fee
// configurations within a reconciliation context. Fee schedules contain one
// or more fee rules that define how fees are calculated for matched records.
package matcher

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// FeeSchedulesService provides CRUD and simulation operations for fee
// schedules.
type FeeSchedulesService interface {
	// Create creates a new fee schedule from the given input.
	Create(ctx context.Context, input *CreateFeeScheduleInput) (*FeeSchedule, error)

	// Get retrieves a fee schedule by its unique identifier.
	Get(ctx context.Context, id string) (*FeeSchedule, error)

	// List returns a paginated iterator over fee schedules.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[FeeSchedule]

	// Update partially updates an existing fee schedule.
	Update(ctx context.Context, id string, input *UpdateFeeScheduleInput) (*FeeSchedule, error)

	// Delete removes a fee schedule by its unique identifier.
	Delete(ctx context.Context, id string) error

	// Simulate runs a fee calculation simulation against the specified fee
	// schedule using the provided amount and currency, returning a detailed
	// breakdown of computed fees.
	Simulate(ctx context.Context, id string, input *SimulateFeeScheduleInput) (*FeeSimulationResult, error)
}

// feeSchedulesService is the concrete implementation of [FeeSchedulesService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type feeSchedulesService struct {
	core.BaseService
}

// newFeeSchedulesService creates a new [FeeSchedulesService] backed by the
// given Matcher [core.Backend].
func newFeeSchedulesService(backend core.Backend) FeeSchedulesService {
	return &feeSchedulesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ FeeSchedulesService = (*feeSchedulesService)(nil)

// Create creates a new fee schedule from the given input.
func (s *feeSchedulesService) Create(ctx context.Context, input *CreateFeeScheduleInput) (*FeeSchedule, error) {
	const operation = "FeeSchedules.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "FeeSchedule", "input is required")
	}

	return core.Create[FeeSchedule, CreateFeeScheduleInput](ctx, &s.BaseService, "/fee-schedules", input)
}

// Get retrieves a fee schedule by its unique identifier.
func (s *feeSchedulesService) Get(ctx context.Context, id string) (*FeeSchedule, error) {
	const operation = "FeeSchedules.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "FeeSchedule", "id is required")
	}

	return core.Get[FeeSchedule](ctx, &s.BaseService, "/fee-schedules/"+url.PathEscape(id))
}

// List returns a paginated iterator over fee schedules.
func (s *feeSchedulesService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[FeeSchedule] {
	return core.List[FeeSchedule](ctx, &s.BaseService, "/fee-schedules", opts)
}

// Update partially updates an existing fee schedule.
func (s *feeSchedulesService) Update(ctx context.Context, id string, input *UpdateFeeScheduleInput) (*FeeSchedule, error) {
	const operation = "FeeSchedules.Update"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "FeeSchedule", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "FeeSchedule", "input is required")
	}

	return core.Update[FeeSchedule, UpdateFeeScheduleInput](ctx, &s.BaseService, "/fee-schedules/"+url.PathEscape(id), input)
}

// Delete removes a fee schedule by its unique identifier.
func (s *feeSchedulesService) Delete(ctx context.Context, id string) error {
	const operation = "FeeSchedules.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "FeeSchedule", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/fee-schedules/"+url.PathEscape(id))
}

// Simulate runs a fee calculation simulation against the specified fee
// schedule using the provided amount and currency.
func (s *feeSchedulesService) Simulate(ctx context.Context, id string, input *SimulateFeeScheduleInput) (*FeeSimulationResult, error) {
	const operation = "FeeSchedules.Simulate"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "FeeSchedule", "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "FeeSchedule", "input is required")
	}

	return core.Action[FeeSimulationResult](ctx, &s.BaseService, "/fee-schedules/"+url.PathEscape(id)+"/simulate", input)
}
