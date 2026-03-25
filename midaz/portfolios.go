// portfolios.go implements the portfoliosServiceAPI for managing portfolio
// resources within a Midaz ledger. Portfolios group related accounts under a
// single logical unit, enabling organisational reporting and access control.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// portfoliosServiceAPI provides CRUD operations for portfolios within a ledger.
type portfoliosServiceAPI interface {
	// Create creates a new portfolio within the specified ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreatePortfolioInput) (*Portfolio, error)

	// Get retrieves a single portfolio by its ID.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Portfolio, error)

	// List returns a paginated iterator over portfolios in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Portfolio]

	// Update modifies an existing portfolio. Only non-nil fields in the
	// input are sent in the PATCH request.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdatePortfolioInput) (*Portfolio, error)

	// Delete removes a portfolio by its ID.
	Delete(ctx context.Context, orgID, ledgerID, id string) error

	// Count returns the total number of portfolios in a ledger.
	Count(ctx context.Context, orgID, ledgerID string) (int, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

// portfoliosService is the concrete implementation of [portfoliosServiceAPI].
// It embeds [core.BaseService] for shared HTTP infrastructure and delegates
// all transport work to the generic core helpers.
type portfoliosService struct {
	core.BaseService
}

// newPortfoliosService creates a [portfoliosServiceAPI] backed by the given
// [core.Backend] (expected to point at the onboarding API).
func newPortfoliosService(backend core.Backend) portfoliosServiceAPI {
	return &portfoliosService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ portfoliosServiceAPI = (*portfoliosService)(nil)

// ---------------------------------------------------------------------------
// Path helpers
// ---------------------------------------------------------------------------

// portfoliosBasePath returns the collection URL for portfolios within a ledger.
func portfoliosBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/portfolios"
}

// portfoliosItemPath returns the resource URL for a specific portfolio.
func portfoliosItemPath(orgID, ledgerID, id string) string {
	return portfoliosBasePath(orgID, ledgerID) + "/" + url.PathEscape(id)
}

// ---------------------------------------------------------------------------
// CRUD methods
// ---------------------------------------------------------------------------

const portfolioResource = "Portfolio"

// Create creates a new portfolio.
func (s *portfoliosService) Create(ctx context.Context, orgID, ledgerID string, input *CreatePortfolioInput) (*Portfolio, error) {
	const operation = "Portfolios.Create"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "organization ID is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "ledger ID is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "input is required")
	}

	return core.Create[Portfolio, CreatePortfolioInput](ctx, &s.BaseService, portfoliosBasePath(orgID, ledgerID), input)
}

// Get retrieves a single portfolio by ID.
func (s *portfoliosService) Get(ctx context.Context, orgID, ledgerID, id string) (*Portfolio, error) {
	const operation = "Portfolios.Get"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "organization ID is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "ledger ID is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "portfolio ID is required")
	}

	return core.Get[Portfolio](ctx, &s.BaseService, portfoliosItemPath(orgID, ledgerID, id))
}

// List returns a paginated iterator over portfolios.
func (s *portfoliosService) List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Portfolio] {
	if orgID == "" || ledgerID == "" {
		return pagination.NewIterator[Portfolio](func(_ context.Context, _ string) ([]Portfolio, string, error) {
			return nil, "", sdkerrors.NewValidation("Portfolios.List", portfolioResource, "organization ID and ledger ID are required")
		})
	}

	return core.List[Portfolio](ctx, &s.BaseService, portfoliosBasePath(orgID, ledgerID), opts)
}

// Update modifies an existing portfolio.
func (s *portfoliosService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdatePortfolioInput) (*Portfolio, error) {
	const operation = "Portfolios.Update"

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "organization ID is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "ledger ID is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "portfolio ID is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, portfolioResource, "input is required")
	}

	return core.Update[Portfolio, UpdatePortfolioInput](ctx, &s.BaseService, portfoliosItemPath(orgID, ledgerID, id), input)
}

// Delete removes a portfolio by ID.
func (s *portfoliosService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "Portfolios.Delete"

	if orgID == "" {
		return sdkerrors.NewValidation(operation, portfolioResource, "organization ID is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, portfolioResource, "ledger ID is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, portfolioResource, "portfolio ID is required")
	}

	return core.Delete(ctx, &s.BaseService, portfoliosItemPath(orgID, ledgerID, id))
}

// Count returns the total number of portfolios in a ledger.
func (s *portfoliosService) Count(ctx context.Context, orgID, ledgerID string) (int, error) {
	const operation = "Portfolios.Count"

	if err := ensureService(s); err != nil {
		return 0, err
	}

	if orgID == "" {
		return 0, sdkerrors.NewValidation(operation, portfolioResource, "organization ID is required")
	}

	if ledgerID == "" {
		return 0, sdkerrors.NewValidation(operation, portfolioResource, "ledger ID is required")
	}

	return core.Count(ctx, &s.BaseService, portfoliosBasePath(orgID, ledgerID)+"/metrics/count")
}
