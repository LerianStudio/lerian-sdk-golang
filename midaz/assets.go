// assets.go implements the assetsServiceAPI for managing tradable instruments
// and currencies within a ledger (e.g. "BRL", "USD", "BTC").
//
// Assets define the denomination of account balances and the unit of measure
// for transaction amounts. Each asset is scoped to a single ledger.
package midaz

import (
	"context"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// assetsServiceAPI provides CRUD operations for assets.
type assetsServiceAPI interface {
	// Create creates a new asset within the specified ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateAssetInput) (*Asset, error)

	// Get retrieves an asset by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Asset, error)

	// List returns a paginated iterator over assets in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Asset]

	// Update partially updates an existing asset.
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateAssetInput) (*Asset, error)

	// Delete removes an asset by its unique identifier.
	Delete(ctx context.Context, orgID, ledgerID, id string) error

	// Count returns the total number of assets in a ledger.
	Count(ctx context.Context, orgID, ledgerID string) (int, error)
}

// assetsService is the concrete implementation of [assetsServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type assetsService struct {
	core.BaseService
}

// newAssetsService creates a new [assetsServiceAPI] backed by the given
// onboarding [core.Backend].
func newAssetsService(backend core.Backend) assetsServiceAPI {
	return &assetsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ assetsServiceAPI = (*assetsService)(nil)

// assetsBasePath builds the base path for asset operations scoped to
// an organization and ledger.
func assetsBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/assets"
}

const assetResource = "Asset"

// Create creates a new asset within the specified ledger.
func (s *assetsService) Create(ctx context.Context, orgID, ledgerID string, input *CreateAssetInput) (*Asset, error) {
	const operation = "Assets.Create"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, assetResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, assetResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, assetResource, "input is required")
	}

	return core.Create[Asset, CreateAssetInput](ctx, &s.BaseService, assetsBasePath(orgID, ledgerID), input)
}

// Get retrieves an asset by its unique identifier.
func (s *assetsService) Get(ctx context.Context, orgID, ledgerID, id string) (*Asset, error) {
	const operation = "Assets.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, assetResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, assetResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, assetResource, "id is required")
	}

	return core.Get[Asset](ctx, &s.BaseService, assetsBasePath(orgID, ledgerID)+"/"+url.PathEscape(id))
}

// List returns a paginated iterator over assets in a ledger.
func (s *assetsService) List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Asset] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[Asset](err)
	}

	if orgID == "" || ledgerID == "" {
		return newErrorIterator[Asset](sdkerrors.NewValidation("Assets.List", assetResource, "organization ID and ledger ID are required"))
	}

	return core.List[Asset](ctx, &s.BaseService, assetsBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing asset.
func (s *assetsService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateAssetInput) (*Asset, error) {
	const operation = "Assets.Update"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, assetResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, assetResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, assetResource, "id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, assetResource, "input is required")
	}

	return core.Update[Asset, UpdateAssetInput](ctx, &s.BaseService, assetsBasePath(orgID, ledgerID)+"/"+url.PathEscape(id), input)
}

// Delete removes an asset by its unique identifier.
func (s *assetsService) Delete(ctx context.Context, orgID, ledgerID, id string) error {
	const operation = "Assets.Delete"

	if err := ensureService(s); err != nil {
		return err
	}

	if orgID == "" {
		return sdkerrors.NewValidation(operation, assetResource, "organization id is required")
	}

	if ledgerID == "" {
		return sdkerrors.NewValidation(operation, assetResource, "ledger id is required")
	}

	if id == "" {
		return sdkerrors.NewValidation(operation, assetResource, "id is required")
	}

	return core.Delete(ctx, &s.BaseService, assetsBasePath(orgID, ledgerID)+"/"+url.PathEscape(id))
}

// Count returns the total number of assets in a ledger.
func (s *assetsService) Count(ctx context.Context, orgID, ledgerID string) (int, error) {
	const operation = "Assets.Count"

	if err := ensureService(s); err != nil {
		return 0, err
	}

	if orgID == "" {
		return 0, sdkerrors.NewValidation(operation, assetResource, "organization id is required")
	}

	if ledgerID == "" {
		return 0, sdkerrors.NewValidation(operation, assetResource, "ledger id is required")
	}

	return core.Count(ctx, &s.BaseService, assetsBasePath(orgID, ledgerID)+"/metrics/count")
}
