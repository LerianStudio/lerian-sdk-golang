// ledgers.go implements the LedgersService for managing isolated
// double-entry ledgers within an organization. All accounts, transactions,
// and balances belong to exactly one ledger.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// LedgersService provides CRUD operations for ledgers scoped to an organization.
type LedgersService interface {
	// Create creates a new ledger within the specified organization.
	Create(ctx context.Context, orgID string, input *CreateLedgerInput) (*Ledger, error)

	// Get retrieves a ledger by its unique identifier within an organization.
	Get(ctx context.Context, orgID, ledgerID string) (*Ledger, error)

	// List returns a paginated iterator over ledgers in an organization.
	List(ctx context.Context, orgID string, opts *models.ListOptions) *pagination.Iterator[Ledger]

	// Update partially updates an existing ledger within an organization.
	Update(ctx context.Context, orgID, ledgerID string, input *UpdateLedgerInput) (*Ledger, error)

	// Delete removes a ledger by its unique identifier within an organization.
	Delete(ctx context.Context, orgID, ledgerID string) error
}

// ledgersService is the concrete implementation of [LedgersService].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type ledgersService struct {
	core.BaseService
}

// newLedgersService creates a new [LedgersService] backed by the
// given onboarding [core.Backend].
func newLedgersService(backend core.Backend) LedgersService {
	return &ledgersService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ LedgersService = (*ledgersService)(nil)

// basePath builds the ledgers collection path for the given organization.
func ledgersBasePath(orgID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers"
}

// itemPath builds the path for a specific ledger within an organization.
func ledgersItemPath(orgID, ledgerID string) string {
	return ledgersBasePath(orgID) + "/" + url.PathEscape(ledgerID)
}

const ledgerResource = "Ledger"

// Create creates a new ledger within the specified organization.
func (s *ledgersService) Create(ctx context.Context, orgID string, input *CreateLedgerInput) (*Ledger, error) {
	const operation = "Ledgers.Create"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, ledgerResource, "organization id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, ledgerResource, "input is required")
	}

	return core.Create[Ledger, CreateLedgerInput](ctx, &s.BaseService, ledgersBasePath(orgID), input)
}

// Get retrieves a ledger by its unique identifier within an organization.
func (s *ledgersService) Get(ctx context.Context, orgID, ledgerID string) (*Ledger, error) {
	const operation = "Ledgers.Get"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, ledgerResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, ledgerResource, "ledger id is required")
	}

	return core.Get[Ledger](ctx, &s.BaseService, ledgersItemPath(orgID, ledgerID))
}

// List returns a paginated iterator over ledgers in an organization.
func (s *ledgersService) List(ctx context.Context, orgID string, opts *models.ListOptions) *pagination.Iterator[Ledger] {
	if orgID == "" {
		return pagination.NewIterator[Ledger](func(_ context.Context, _ string) ([]Ledger, string, error) {
			return nil, "", sdkerrors.NewValidation("Ledgers.List", ledgerResource, "organization ID is required")
		})
	}

	return core.List[Ledger](ctx, &s.BaseService, ledgersBasePath(orgID), opts)
}

// Update partially updates an existing ledger within an organization.
func (s *ledgersService) Update(ctx context.Context, orgID, ledgerID string, input *UpdateLedgerInput) (*Ledger, error) {
	const operation = "Ledgers.Update"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, ledgerResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, ledgerResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, ledgerResource, "input is required")
	}

	return core.Update[Ledger, UpdateLedgerInput](ctx, &s.BaseService, ledgersItemPath(orgID, ledgerID), input)
}

// Delete removes a ledger by its unique identifier within an organization.
func (s *ledgersService) Delete(ctx context.Context, orgID, ledgerID string) error {
	const operation = "Ledgers.Delete"

	if orgID == "" {
		return sdkerrors.NewValidation(operation, ledgerResource, "organization id is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, ledgerResource, "ledger id is required")
	}

	return core.Delete(ctx, &s.BaseService, ledgersItemPath(orgID, ledgerID))
}
