// asset_rates.go implements the assetRatesServiceAPI for managing exchange rates
// between pairs of assets at specific points in time.
//
// In addition to standard CRUD, the service provides lookup-by-external-id
// and lookup-by-from-asset-code endpoints for resolving rates via business
// identifiers rather than the Midaz-internal UUID.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// assetRatesServiceAPI provides CRUD and lookup operations for asset rates.
type assetRatesServiceAPI interface {
	// Create creates a new asset rate within the specified ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateAssetRateInput) (*AssetRate, error)

	// Get retrieves an asset rate by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*AssetRate, error)

	// List returns a paginated iterator over asset rates in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[AssetRate]

	// Update partially updates an existing asset rate.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateAssetRateInput) (*AssetRate, error)

	// Delete removes an asset rate by its unique identifier.
	Delete(ctx context.Context, orgID, ledgerID, id string) error

	// GetByExternalID retrieves an asset rate by its external identifier.
	GetByExternalID(ctx context.Context, orgID, ledgerID, externalID string) (*AssetRate, error)

	// GetFromAssetCode retrieves an asset rate by the source (from) asset code.
	GetFromAssetCode(ctx context.Context, orgID, ledgerID, assetCode string) (*AssetRate, error)
}

// assetRatesService is the concrete implementation of [assetRatesServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type assetRatesService struct {
	core.BaseService
}

// newAssetRatesService creates a new [assetRatesServiceAPI] backed by the given
// transaction [core.Backend].
func newAssetRatesService(backend core.Backend) assetRatesServiceAPI {
	return &assetRatesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ assetRatesServiceAPI = (*assetRatesService)(nil)

// assetRatesBasePath builds the base path for asset rate operations scoped to
// an organization and ledger.
func assetRatesBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/asset-rates"
}

const assetRateResource = "AssetRate"

// Create creates a new asset rate within the specified ledger.
func (s *assetRatesService) Create(ctx context.Context, orgID, ledgerID string, input *CreateAssetRateInput) (*AssetRate, error) {
	const operation = "AssetRates.Create"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "input is required")
	}

	return core.Create[AssetRate, CreateAssetRateInput](ctx, &s.BaseService, assetRatesBasePath(orgID, ledgerID), input)
}

// Get retrieves an asset rate by its unique identifier.
func (s *assetRatesService) Get(ctx context.Context, orgID, ledgerID, id string) (*AssetRate, error) {
	const operation = "AssetRates.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "id is required")
	}

	return core.Get[AssetRate](ctx, &s.BaseService, assetRatesBasePath(orgID, ledgerID)+"/"+url.PathEscape(id))
}

// List returns a paginated iterator over asset rates in a ledger.
func (s *assetRatesService) List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[AssetRate] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[AssetRate](err)
	}

	if orgID == "" || ledgerID == "" {
		return newErrorIterator[AssetRate](sdkerrors.NewValidation("AssetRates.List", assetRateResource, "organization ID and ledger ID are required"))
	}

	return core.List[AssetRate](ctx, &s.BaseService, assetRatesBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing asset rate.
func (s *assetRatesService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateAssetRateInput) (*AssetRate, error) {
	const operation = "AssetRates.Update"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "input is required")
	}

	return core.Update[AssetRate, UpdateAssetRateInput](ctx, &s.BaseService, assetRatesBasePath(orgID, ledgerID)+"/"+url.PathEscape(id), input)
}

// Delete removes an asset rate by its unique identifier.
func (s *assetRatesService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "AssetRates.Delete"

	if err := ensureService(s); err != nil {
		return err
	}

	if orgID == "" {
		return sdkerrors.NewValidation(operation, assetRateResource, "organization id is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, assetRateResource, "ledger id is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, assetRateResource, "id is required")
	}

	return core.Delete(ctx, &s.BaseService, assetRatesBasePath(orgID, ledgerID)+"/"+url.PathEscape(id))
}

// GetByExternalID retrieves an asset rate by its external identifier.
func (s *assetRatesService) GetByExternalID(ctx context.Context, orgID, ledgerID, externalID string) (*AssetRate, error) {
	const operation = "AssetRates.GetByExternalID"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "ledger id is required")
	}

	if externalID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "external id is required")
	}

	return core.Get[AssetRate](ctx, &s.BaseService, assetRatesBasePath(orgID, ledgerID)+"/external-id/"+url.PathEscape(externalID))
}

// GetFromAssetCode retrieves an asset rate by the source (from) asset code.
func (s *assetRatesService) GetFromAssetCode(ctx context.Context, orgID, ledgerID, assetCode string) (*AssetRate, error) {
	const operation = "AssetRates.GetFromAssetCode"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "ledger id is required")
	}

	if assetCode == "" {
		return nil, sdkerrors.NewValidation(operation, assetRateResource, "asset code is required")
	}

	return core.Get[AssetRate](ctx, &s.BaseService, assetRatesBasePath(orgID, ledgerID)+"/from/"+url.PathEscape(assetCode))
}
