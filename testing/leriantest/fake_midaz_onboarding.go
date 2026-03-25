package leriantest

import (
	"context"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// ---------------------------------------------------------------------------
// Organizations
// ---------------------------------------------------------------------------

type fakeOrganizations struct {
	store *fakeStore[midaz.Organization]
	cfg   *fakeConfig
}

func newFakeOrganizations(cfg *fakeConfig) *fakeOrganizations {
	return &fakeOrganizations{store: newFakeStore[midaz.Organization](), cfg: cfg}
}

func (f *fakeOrganizations) Create(_ context.Context, input *midaz.CreateOrganizationInput) (*midaz.Organization, error) {
	if err := f.cfg.injectedError("midaz.Organizations.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Organizations.Create", "Organization", "input is required")
	}

	now := time.Now()

	org := midaz.Organization{
		ID:            generateID("org"),
		LegalName:     input.LegalName,
		LegalDocument: input.LegalDocument,
		Status:        models.Status{Code: "active"},
		Address:       input.Address,
		Metadata:      input.Metadata,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if input.ParentOrganizationID != nil {
		org.ParentOrganizationID = input.ParentOrganizationID
	}

	f.store.Set(org.ID, org)

	return &org, nil
}

func (f *fakeOrganizations) Get(_ context.Context, id string) (*midaz.Organization, error) {
	if err := f.cfg.injectedError("midaz.Organizations.Get"); err != nil {
		return nil, err
	}

	org, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Organizations.Get", "Organization", id)
	}

	return &org, nil
}

func (f *fakeOrganizations) List(_ context.Context, opts *models.CursorListOptions) *pagination.Iterator[midaz.Organization] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeOrganizations) Update(_ context.Context, id string, input *midaz.UpdateOrganizationInput) (*midaz.Organization, error) {
	if err := f.cfg.injectedError("midaz.Organizations.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Organizations.Update", "Organization", "input is required")
	}

	return fakeMutateStored(f.cfg, "", "Organizations.Update", "Organization", id, f.store, func(org *midaz.Organization) {
		if input.LegalName != nil {
			org.LegalName = *input.LegalName
		}
		if input.LegalDocument != nil {
			org.LegalDocument = *input.LegalDocument
		}
		if input.Address != nil {
			org.Address = input.Address
		}
		if input.Status != nil {
			org.Status = *input.Status
		}
		if input.Metadata != nil {
			org.Metadata = input.Metadata
		}
		org.UpdatedAt = time.Now()
	})
}

func (f *fakeOrganizations) Delete(_ context.Context, id string) error {
	if err := f.cfg.injectedError("midaz.Organizations.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Organizations.Delete", "Organization", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeOrganizations) Count(_ context.Context) (int, error) {
	if err := f.cfg.injectedError("midaz.Organizations.Count"); err != nil {
		return 0, err
	}

	return f.store.Len(), nil
}

// ---------------------------------------------------------------------------
// Ledgers
// ---------------------------------------------------------------------------

type fakeLedgers struct {
	store *fakeStore[midaz.Ledger]
	cfg   *fakeConfig
}

func newFakeLedgers(cfg *fakeConfig) *fakeLedgers {
	return &fakeLedgers{store: newFakeStore[midaz.Ledger](), cfg: cfg}
}

func (f *fakeLedgers) Create(_ context.Context, orgID string, input *midaz.CreateLedgerInput) (*midaz.Ledger, error) {
	if err := f.cfg.injectedError("midaz.Ledgers.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Ledgers.Create", "Ledger", "input is required")
	}

	now := time.Now()

	ledger := midaz.Ledger{
		ID:             generateID("ledger"),
		OrganizationID: orgID,
		Name:           input.Name,
		Status:         models.Status{Code: "active"},
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(ledger.ID, ledger)

	return &ledger, nil
}

func (f *fakeLedgers) Get(_ context.Context, _ string, ledgerID string) (*midaz.Ledger, error) {
	if err := f.cfg.injectedError("midaz.Ledgers.Get"); err != nil {
		return nil, err
	}

	ledger, ok := f.store.Get(ledgerID)
	if !ok {
		return nil, sdkerrors.NewNotFound("Ledgers.Get", "Ledger", ledgerID)
	}

	return &ledger, nil
}

func (f *fakeLedgers) List(_ context.Context, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.Ledger] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeLedgers) Update(_ context.Context, _ string, ledgerID string, input *midaz.UpdateLedgerInput) (*midaz.Ledger, error) {
	if err := f.cfg.injectedError("midaz.Ledgers.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Ledgers.Update", "Ledger", "input is required")
	}

	return fakeMutateStored(f.cfg, "", "Ledgers.Update", "Ledger", ledgerID, f.store, func(ledger *midaz.Ledger) {
		if input.Name != nil {
			ledger.Name = *input.Name
		}
		if input.Status != nil {
			ledger.Status = *input.Status
		}
		if input.Metadata != nil {
			ledger.Metadata = input.Metadata
		}
		ledger.UpdatedAt = time.Now()
	})
}

func (f *fakeLedgers) Delete(_ context.Context, _ string, ledgerID string) error {
	if err := f.cfg.injectedError("midaz.Ledgers.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(ledgerID); !ok {
		return sdkerrors.NewNotFound("Ledgers.Delete", "Ledger", ledgerID)
	}

	f.store.Delete(ledgerID)

	return nil
}

func (f *fakeLedgers) Count(_ context.Context, orgID string) (int, error) {
	if err := f.cfg.injectedError("midaz.Ledgers.Count"); err != nil {
		return 0, err
	}

	if orgID == "" {
		return 0, sdkerrors.NewValidation("Ledgers.Count", "Ledger", "organization id is required")
	}

	count := 0
	for _, ledger := range f.store.List() {
		if ledger.OrganizationID == orgID {
			count++
		}
	}

	return count, nil
}

// ---------------------------------------------------------------------------
// Accounts
// ---------------------------------------------------------------------------

type fakeAccounts struct {
	store *fakeStore[midaz.Account]
	cfg   *fakeConfig
}

func newFakeAccounts(cfg *fakeConfig) *fakeAccounts {
	return &fakeAccounts{store: newFakeStore[midaz.Account](), cfg: cfg}
}

func (f *fakeAccounts) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateAccountInput) (*midaz.Account, error) {
	if err := f.cfg.injectedError("midaz.Accounts.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Accounts.Create", "Account", "input is required")
	}

	now := time.Now()

	acct := midaz.Account{
		ID:             generateID("acct"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Name:           input.Name,
		Type:           input.Type,
		AssetCode:      input.AssetCode,
		Alias:          input.Alias,
		ExternalCode:   input.ExternalCode,
		PortfolioID:    input.PortfolioID,
		SegmentID:      input.SegmentID,
		EntityID:       input.EntityID,
		Status:         models.Status{Code: "active"},
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(acct.ID, acct)

	return &acct, nil
}

func (f *fakeAccounts) Get(_ context.Context, _, _ string, id string) (*midaz.Account, error) {
	if err := f.cfg.injectedError("midaz.Accounts.Get"); err != nil {
		return nil, err
	}

	acct, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Accounts.Get", "Account", id)
	}

	return &acct, nil
}

func (f *fakeAccounts) List(_ context.Context, _, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.Account] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeAccounts) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateAccountInput) (*midaz.Account, error) {
	if err := f.cfg.injectedError("midaz.Accounts.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Accounts.Update", "Account", "input is required")
	}

	return fakeMutateStored(f.cfg, "", "Accounts.Update", "Account", id, f.store, func(acct *midaz.Account) {
		if input.Name != nil {
			acct.Name = *input.Name
		}
		if input.Alias != nil {
			acct.Alias = input.Alias
		}
		if input.ExternalCode != nil {
			acct.ExternalCode = input.ExternalCode
		}
		if input.PortfolioID != nil {
			acct.PortfolioID = input.PortfolioID
		}
		if input.SegmentID != nil {
			acct.SegmentID = input.SegmentID
		}
		if input.ParentAccountID != nil {
			acct.ParentAccountID = input.ParentAccountID
		}
		if input.EntityID != nil {
			acct.EntityID = input.EntityID
		}
		if input.Status != nil {
			acct.Status = *input.Status
		}
		if input.Metadata != nil {
			acct.Metadata = input.Metadata
		}
		acct.UpdatedAt = time.Now()
	})
}

func (f *fakeAccounts) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.Accounts.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Accounts.Delete", "Account", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeAccounts) GetByAlias(_ context.Context, orgID, ledgerID, alias string) (*midaz.Account, error) {
	for _, acct := range f.store.List() {
		if acct.OrganizationID == orgID && acct.LedgerID == ledgerID && acct.Alias != nil && *acct.Alias == alias {
			return &acct, nil
		}
	}

	return nil, sdkerrors.NewNotFound("Accounts.GetByAlias", "Account", alias)
}

func (f *fakeAccounts) GetByExternalCode(_ context.Context, orgID, ledgerID, code string) (*midaz.Account, error) {
	for _, acct := range f.store.List() {
		if acct.OrganizationID == orgID && acct.LedgerID == ledgerID && acct.ExternalCode != nil && *acct.ExternalCode == code {
			return &acct, nil
		}
	}

	return nil, sdkerrors.NewNotFound("Accounts.GetByExternalCode", "Account", code)
}

func (f *fakeAccounts) Count(_ context.Context, orgID, ledgerID string) (int, error) {
	if err := f.cfg.injectedError("midaz.Accounts.Count"); err != nil {
		return 0, err
	}

	if orgID == "" {
		return 0, sdkerrors.NewValidation("Accounts.Count", "Account", "organization id is required")
	}

	if ledgerID == "" {
		return 0, sdkerrors.NewValidation("Accounts.Count", "Account", "ledger id is required")
	}

	count := 0
	for _, account := range f.store.List() {
		if account.OrganizationID == orgID && account.LedgerID == ledgerID {
			count++
		}
	}

	return count, nil
}

// ---------------------------------------------------------------------------
// AccountTypes
// ---------------------------------------------------------------------------

type fakeAccountTypes struct {
	store *fakeStore[midaz.AccountType]
	cfg   *fakeConfig
}

func newFakeAccountTypes(cfg *fakeConfig) *fakeAccountTypes {
	return &fakeAccountTypes{store: newFakeStore[midaz.AccountType](), cfg: cfg}
}

func (f *fakeAccountTypes) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateAccountTypeInput) (*midaz.AccountType, error) {
	if err := f.cfg.injectedError("midaz.AccountTypes.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("AccountTypes.Create", "AccountType", "input is required")
	}

	now := time.Now()

	at := midaz.AccountType{
		ID:             generateID("accttype"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Name:           input.Name,
		Description:    input.Description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(at.ID, at)

	return &at, nil
}

func (f *fakeAccountTypes) Get(_ context.Context, _, _ string, id string) (*midaz.AccountType, error) {
	if err := f.cfg.injectedError("midaz.AccountTypes.Get"); err != nil {
		return nil, err
	}

	at, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("AccountTypes.Get", "AccountType", id)
	}

	return &at, nil
}

func (f *fakeAccountTypes) List(_ context.Context, _, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.AccountType] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeAccountTypes) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateAccountTypeInput) (*midaz.AccountType, error) {
	if err := f.cfg.injectedError("midaz.AccountTypes.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("AccountTypes.Update", "AccountType", "input is required")
	}

	return fakeMutateStored(f.cfg, "", "AccountTypes.Update", "AccountType", id, f.store, func(at *midaz.AccountType) {
		if input.Name != nil {
			at.Name = *input.Name
		}
		if input.Description != nil {
			at.Description = input.Description
		}
		at.UpdatedAt = time.Now()
	})
}

func (f *fakeAccountTypes) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.AccountTypes.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("AccountTypes.Delete", "AccountType", id)
	}

	f.store.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// Assets
// ---------------------------------------------------------------------------

type fakeAssets struct {
	store *fakeStore[midaz.Asset]
	cfg   *fakeConfig
}

func newFakeAssets(cfg *fakeConfig) *fakeAssets {
	return &fakeAssets{store: newFakeStore[midaz.Asset](), cfg: cfg}
}

func (f *fakeAssets) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateAssetInput) (*midaz.Asset, error) {
	if err := f.cfg.injectedError("midaz.Assets.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Assets.Create", "Asset", "input is required")
	}

	now := time.Now()

	asset := midaz.Asset{
		ID:             generateID("asset"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Name:           input.Name,
		Code:           input.Code,
		Type:           input.Type,
		Status:         models.Status{Code: "active"},
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(asset.ID, asset)

	return &asset, nil
}

func (f *fakeAssets) Get(_ context.Context, _, _ string, id string) (*midaz.Asset, error) {
	if err := f.cfg.injectedError("midaz.Assets.Get"); err != nil {
		return nil, err
	}

	asset, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Assets.Get", "Asset", id)
	}

	return &asset, nil
}

func (f *fakeAssets) List(_ context.Context, _, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.Asset] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeAssets) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateAssetInput) (*midaz.Asset, error) {
	if err := f.cfg.injectedError("midaz.Assets.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Assets.Update", "Asset", "input is required")
	}

	return fakeMutateStored(f.cfg, "", "Assets.Update", "Asset", id, f.store, func(asset *midaz.Asset) {
		if input.Name != nil {
			asset.Name = *input.Name
		}
		if input.Status != nil {
			asset.Status = *input.Status
		}
		if input.Metadata != nil {
			asset.Metadata = input.Metadata
		}
		asset.UpdatedAt = time.Now()
	})
}

func (f *fakeAssets) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.Assets.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Assets.Delete", "Asset", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeAssets) Count(_ context.Context, orgID, ledgerID string) (int, error) {
	if err := f.cfg.injectedError("midaz.Assets.Count"); err != nil {
		return 0, err
	}

	if orgID == "" {
		return 0, sdkerrors.NewValidation("Assets.Count", "Asset", "organization id is required")
	}

	if ledgerID == "" {
		return 0, sdkerrors.NewValidation("Assets.Count", "Asset", "ledger id is required")
	}

	count := 0
	for _, asset := range f.store.List() {
		if asset.OrganizationID == orgID && asset.LedgerID == ledgerID {
			count++
		}
	}

	return count, nil
}

// ---------------------------------------------------------------------------
// AssetRates
// ---------------------------------------------------------------------------

type fakeAssetRates struct {
	store *fakeStore[midaz.AssetRate]
	cfg   *fakeConfig
}

func newFakeAssetRates(cfg *fakeConfig) *fakeAssetRates {
	return &fakeAssetRates{store: newFakeStore[midaz.AssetRate](), cfg: cfg}
}

func (f *fakeAssetRates) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateAssetRateInput) (*midaz.AssetRate, error) {
	if err := f.cfg.injectedError("midaz.AssetRates.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("AssetRates.Create", "AssetRate", "input is required")
	}

	now := time.Now()
	for _, existing := range f.store.List() {
		if existing.OrganizationID == orgID && existing.LedgerID == ledgerID && existing.BaseAssetCode == input.BaseAssetCode && existing.CounterAssetCode == input.CounterAssetCode {
			existing.Amount = input.Amount
			existing.Scale = input.Scale
			existing.Source = input.Source
			existing.ExternalID = input.ExternalID
			existing.UpdatedAt = now
			f.store.Set(existing.ID, existing)

			return &existing, nil
		}
	}

	rate := midaz.AssetRate{
		ID:               generateID("rate"),
		OrganizationID:   orgID,
		LedgerID:         ledgerID,
		BaseAssetCode:    input.BaseAssetCode,
		CounterAssetCode: input.CounterAssetCode,
		Amount:           input.Amount,
		Scale:            input.Scale,
		Source:           input.Source,
		ExternalID:       input.ExternalID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	f.store.Set(rate.ID, rate)

	return &rate, nil
}

func (f *fakeAssetRates) Get(_ context.Context, _, _ string, id string) (*midaz.AssetRate, error) {
	if err := f.cfg.injectedError("midaz.AssetRates.Get"); err != nil {
		return nil, err
	}

	rate, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("AssetRates.Get", "AssetRate", id)
	}

	return &rate, nil
}

func (f *fakeAssetRates) List(_ context.Context, _, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.AssetRate] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeAssetRates) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateAssetRateInput) (*midaz.AssetRate, error) {
	if err := f.cfg.injectedError("midaz.AssetRates.Update"); err != nil {
		return nil, err
	}

	return fakeMutateStored(f.cfg, "", "AssetRates.Update", "AssetRate", id, f.store, func(rate *midaz.AssetRate) {
		if input.Amount != nil {
			rate.Amount = *input.Amount
		}
		if input.Scale != nil {
			rate.Scale = *input.Scale
		}
		if input.Source != nil {
			rate.Source = input.Source
		}
		if input.ExternalID != nil {
			rate.ExternalID = input.ExternalID
		}
		rate.UpdatedAt = time.Now()
	})
}

func (f *fakeAssetRates) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.AssetRates.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("AssetRates.Delete", "AssetRate", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeAssetRates) GetByExternalID(_ context.Context, orgID, ledgerID, externalID string) (*midaz.AssetRate, error) {
	for _, rate := range f.store.List() {
		if rate.OrganizationID == orgID && rate.LedgerID == ledgerID && rate.ExternalID != nil && *rate.ExternalID == externalID {
			return &rate, nil
		}
	}

	return nil, sdkerrors.NewNotFound("AssetRates.GetByExternalID", "AssetRate", externalID)
}

func (f *fakeAssetRates) GetFromAssetCode(_ context.Context, orgID, ledgerID, assetCode string) (*midaz.AssetRate, error) {
	for _, rate := range f.store.List() {
		if rate.OrganizationID == orgID && rate.LedgerID == ledgerID && rate.BaseAssetCode == assetCode {
			return &rate, nil
		}
	}

	return nil, sdkerrors.NewNotFound("AssetRates.GetFromAssetCode", "AssetRate", assetCode)
}

// ---------------------------------------------------------------------------
// Portfolios
// ---------------------------------------------------------------------------

type fakePortfolios struct {
	store *fakeStore[midaz.Portfolio]
	cfg   *fakeConfig
}

func newFakePortfolios(cfg *fakeConfig) *fakePortfolios {
	return &fakePortfolios{store: newFakeStore[midaz.Portfolio](), cfg: cfg}
}

func (f *fakePortfolios) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreatePortfolioInput) (*midaz.Portfolio, error) {
	if err := f.cfg.injectedError("midaz.Portfolios.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Portfolios.Create", "Portfolio", "input is required")
	}

	now := time.Now()

	p := midaz.Portfolio{
		ID:             generateID("portfolio"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Name:           input.Name,
		EntityID:       input.EntityID,
		Status:         models.Status{Code: "active"},
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(p.ID, p)

	return &p, nil
}

func (f *fakePortfolios) Get(_ context.Context, _, _ string, id string) (*midaz.Portfolio, error) {
	if err := f.cfg.injectedError("midaz.Portfolios.Get"); err != nil {
		return nil, err
	}

	p, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Portfolios.Get", "Portfolio", id)
	}

	return &p, nil
}

func (f *fakePortfolios) List(_ context.Context, _, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.Portfolio] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakePortfolios) Update(_ context.Context, _, _ string, id string, input *midaz.UpdatePortfolioInput) (*midaz.Portfolio, error) {
	if err := f.cfg.injectedError("midaz.Portfolios.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Portfolios.Update", "Portfolio", "input is required")
	}

	return fakeMutateStored(f.cfg, "", "Portfolios.Update", "Portfolio", id, f.store, func(p *midaz.Portfolio) {
		if input.Name != nil {
			p.Name = *input.Name
		}
		if input.EntityID != nil {
			p.EntityID = input.EntityID
		}
		if input.Status != nil {
			p.Status = *input.Status
		}
		if input.Metadata != nil {
			p.Metadata = input.Metadata
		}
		p.UpdatedAt = time.Now()
	})
}

func (f *fakePortfolios) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.Portfolios.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Portfolios.Delete", "Portfolio", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakePortfolios) Count(_ context.Context, orgID, ledgerID string) (int, error) {
	if err := f.cfg.injectedError("midaz.Portfolios.Count"); err != nil {
		return 0, err
	}

	if orgID == "" {
		return 0, sdkerrors.NewValidation("Portfolios.Count", "Portfolio", "organization ID is required")
	}

	if ledgerID == "" {
		return 0, sdkerrors.NewValidation("Portfolios.Count", "Portfolio", "ledger ID is required")
	}

	count := 0
	for _, portfolio := range f.store.List() {
		if portfolio.OrganizationID == orgID && portfolio.LedgerID == ledgerID {
			count++
		}
	}

	return count, nil
}

// ---------------------------------------------------------------------------
// Segments
// ---------------------------------------------------------------------------

type fakeSegments struct {
	store *fakeStore[midaz.Segment]
	cfg   *fakeConfig
}

func newFakeSegments(cfg *fakeConfig) *fakeSegments {
	return &fakeSegments{store: newFakeStore[midaz.Segment](), cfg: cfg}
}

func (f *fakeSegments) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateSegmentInput) (*midaz.Segment, error) {
	if err := f.cfg.injectedError("midaz.Segments.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Segments.Create", "Segment", "input is required")
	}

	now := time.Now()

	seg := midaz.Segment{
		ID:             generateID("segment"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Name:           input.Name,
		Status:         models.Status{Code: "active"},
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(seg.ID, seg)

	return &seg, nil
}

func (f *fakeSegments) Get(_ context.Context, _, _ string, id string) (*midaz.Segment, error) {
	if err := f.cfg.injectedError("midaz.Segments.Get"); err != nil {
		return nil, err
	}

	seg, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Segments.Get", "Segment", id)
	}

	return &seg, nil
}

func (f *fakeSegments) List(_ context.Context, _, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.Segment] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeSegments) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateSegmentInput) (*midaz.Segment, error) {
	if err := f.cfg.injectedError("midaz.Segments.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Segments.Update", "Segment", "input is required")
	}

	return fakeMutateStored(f.cfg, "", "Segments.Update", "Segment", id, f.store, func(seg *midaz.Segment) {
		if input.Name != nil {
			seg.Name = *input.Name
		}
		if input.Status != nil {
			seg.Status = *input.Status
		}
		if input.Metadata != nil {
			seg.Metadata = input.Metadata
		}
		seg.UpdatedAt = time.Now()
	})
}

func (f *fakeSegments) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.Segments.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Segments.Delete", "Segment", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeSegments) Count(_ context.Context, orgID, ledgerID string) (int, error) {
	if err := f.cfg.injectedError("midaz.Segments.Count"); err != nil {
		return 0, err
	}

	if orgID == "" {
		return 0, sdkerrors.NewValidation("Segments.Count", "Segment", "organization ID is required")
	}

	if ledgerID == "" {
		return 0, sdkerrors.NewValidation("Segments.Count", "Segment", "ledger ID is required")
	}

	count := 0
	for _, segment := range f.store.List() {
		if segment.OrganizationID == orgID && segment.LedgerID == ledgerID {
			count++
		}
	}

	return count, nil
}
