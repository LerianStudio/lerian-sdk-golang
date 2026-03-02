package leriantest

import (
	"context"
	"fmt"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// newFakeMidazClient constructs a [midaz.Client] with all service fields
// backed by in-memory fakes.
func newFakeMidazClient(cfg *fakeConfig) *midaz.Client {
	return &midaz.Client{
		Organizations:     newFakeOrganizations(cfg),
		Ledgers:           newFakeLedgers(cfg),
		Accounts:          newFakeAccounts(cfg),
		AccountTypes:      newFakeAccountTypes(cfg),
		Assets:            newFakeAssets(cfg),
		AssetRates:        newFakeAssetRates(cfg),
		Portfolios:        newFakePortfolios(cfg),
		Segments:          newFakeSegments(cfg),
		Balances:          newFakeBalances(cfg),
		Transactions:      newFakeTransactions(cfg),
		TransactionRoutes: newFakeTransactionRoutes(cfg),
		Operations:        newFakeOperations(cfg),
		OperationRoutes:   newFakeOperationRoutes(cfg),
	}
}

// ---------------------------------------------------------------------------
// Organizations
// ---------------------------------------------------------------------------

type fakeOrganizations struct {
	store *fakeStore[midaz.Organization]
	cfg   *fakeConfig
}

var _ midaz.OrganizationsService = (*fakeOrganizations)(nil)

func newFakeOrganizations(cfg *fakeConfig) *fakeOrganizations {
	return &fakeOrganizations{store: newFakeStore[midaz.Organization](), cfg: cfg}
}

func (f *fakeOrganizations) Create(_ context.Context, input *midaz.CreateOrganizationInput) (*midaz.Organization, error) {
	if err := f.cfg.injectedError("midaz.Organizations.Create"); err != nil {
		return nil, err
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

func (f *fakeOrganizations) List(_ context.Context, opts *models.ListOptions) *pagination.Iterator[midaz.Organization] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeOrganizations) Update(_ context.Context, id string, input *midaz.UpdateOrganizationInput) (*midaz.Organization, error) {
	if err := f.cfg.injectedError("midaz.Organizations.Update"); err != nil {
		return nil, err
	}

	org, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Organizations.Update", "Organization", id)
	}

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
	f.store.Set(id, org)

	return &org, nil
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

// ---------------------------------------------------------------------------
// Ledgers
// ---------------------------------------------------------------------------

type fakeLedgers struct {
	store *fakeStore[midaz.Ledger]
	cfg   *fakeConfig
}

var _ midaz.LedgersService = (*fakeLedgers)(nil)

func newFakeLedgers(cfg *fakeConfig) *fakeLedgers {
	return &fakeLedgers{store: newFakeStore[midaz.Ledger](), cfg: cfg}
}

func (f *fakeLedgers) Create(_ context.Context, orgID string, input *midaz.CreateLedgerInput) (*midaz.Ledger, error) {
	if err := f.cfg.injectedError("midaz.Ledgers.Create"); err != nil {
		return nil, err
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

func (f *fakeLedgers) List(_ context.Context, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.Ledger] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeLedgers) Update(_ context.Context, _ string, ledgerID string, input *midaz.UpdateLedgerInput) (*midaz.Ledger, error) {
	if err := f.cfg.injectedError("midaz.Ledgers.Update"); err != nil {
		return nil, err
	}

	ledger, ok := f.store.Get(ledgerID)
	if !ok {
		return nil, sdkerrors.NewNotFound("Ledgers.Update", "Ledger", ledgerID)
	}

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
	f.store.Set(ledgerID, ledger)

	return &ledger, nil
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

// ---------------------------------------------------------------------------
// Accounts
// ---------------------------------------------------------------------------

type fakeAccounts struct {
	store *fakeStore[midaz.Account]
	cfg   *fakeConfig
}

var _ midaz.AccountsService = (*fakeAccounts)(nil)

func newFakeAccounts(cfg *fakeConfig) *fakeAccounts {
	return &fakeAccounts{store: newFakeStore[midaz.Account](), cfg: cfg}
}

func (f *fakeAccounts) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateAccountInput) (*midaz.Account, error) {
	if err := f.cfg.injectedError("midaz.Accounts.Create"); err != nil {
		return nil, err
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

func (f *fakeAccounts) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.Account] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeAccounts) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateAccountInput) (*midaz.Account, error) {
	if err := f.cfg.injectedError("midaz.Accounts.Update"); err != nil {
		return nil, err
	}

	acct, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Accounts.Update", "Account", id)
	}

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
	f.store.Set(id, acct)

	return &acct, nil
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

func (f *fakeAccounts) GetByAlias(_ context.Context, _, _, alias string) (*midaz.Account, error) {
	for _, acct := range f.store.List() {
		if acct.Alias != nil && *acct.Alias == alias {
			return &acct, nil
		}
	}

	return nil, sdkerrors.NewNotFound("Accounts.GetByAlias", "Account", alias)
}

func (f *fakeAccounts) GetByExternalCode(_ context.Context, _, _, code string) (*midaz.Account, error) {
	for _, acct := range f.store.List() {
		if acct.ExternalCode != nil && *acct.ExternalCode == code {
			return &acct, nil
		}
	}

	return nil, sdkerrors.NewNotFound("Accounts.GetByExternalCode", "Account", code)
}

// ---------------------------------------------------------------------------
// AccountTypes
// ---------------------------------------------------------------------------

type fakeAccountTypes struct {
	store *fakeStore[midaz.AccountType]
	cfg   *fakeConfig
}

var _ midaz.AccountTypesService = (*fakeAccountTypes)(nil)

func newFakeAccountTypes(cfg *fakeConfig) *fakeAccountTypes {
	return &fakeAccountTypes{store: newFakeStore[midaz.AccountType](), cfg: cfg}
}

func (f *fakeAccountTypes) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateAccountTypeInput) (*midaz.AccountType, error) {
	if err := f.cfg.injectedError("midaz.AccountTypes.Create"); err != nil {
		return nil, err
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

func (f *fakeAccountTypes) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.AccountType] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeAccountTypes) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateAccountTypeInput) (*midaz.AccountType, error) {
	if err := f.cfg.injectedError("midaz.AccountTypes.Update"); err != nil {
		return nil, err
	}

	at, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("AccountTypes.Update", "AccountType", id)
	}

	if input.Name != nil {
		at.Name = *input.Name
	}

	if input.Description != nil {
		at.Description = input.Description
	}

	at.UpdatedAt = time.Now()
	f.store.Set(id, at)

	return &at, nil
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

var _ midaz.AssetsService = (*fakeAssets)(nil)

func newFakeAssets(cfg *fakeConfig) *fakeAssets {
	return &fakeAssets{store: newFakeStore[midaz.Asset](), cfg: cfg}
}

func (f *fakeAssets) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateAssetInput) (*midaz.Asset, error) {
	if err := f.cfg.injectedError("midaz.Assets.Create"); err != nil {
		return nil, err
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

func (f *fakeAssets) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.Asset] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeAssets) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateAssetInput) (*midaz.Asset, error) {
	if err := f.cfg.injectedError("midaz.Assets.Update"); err != nil {
		return nil, err
	}

	asset, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Assets.Update", "Asset", id)
	}

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
	f.store.Set(id, asset)

	return &asset, nil
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

// ---------------------------------------------------------------------------
// AssetRates
// ---------------------------------------------------------------------------

type fakeAssetRates struct {
	store *fakeStore[midaz.AssetRate]
	cfg   *fakeConfig
}

var _ midaz.AssetRatesService = (*fakeAssetRates)(nil)

func newFakeAssetRates(cfg *fakeConfig) *fakeAssetRates {
	return &fakeAssetRates{store: newFakeStore[midaz.AssetRate](), cfg: cfg}
}

func (f *fakeAssetRates) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateAssetRateInput) (*midaz.AssetRate, error) {
	if err := f.cfg.injectedError("midaz.AssetRates.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

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

func (f *fakeAssetRates) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.AssetRate] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeAssetRates) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateAssetRateInput) (*midaz.AssetRate, error) {
	if err := f.cfg.injectedError("midaz.AssetRates.Update"); err != nil {
		return nil, err
	}

	rate, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("AssetRates.Update", "AssetRate", id)
	}

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
	f.store.Set(id, rate)

	return &rate, nil
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

func (f *fakeAssetRates) GetByExternalID(_ context.Context, _, _, externalID string) (*midaz.AssetRate, error) {
	for _, rate := range f.store.List() {
		if rate.ExternalID != nil && *rate.ExternalID == externalID {
			return &rate, nil
		}
	}

	return nil, sdkerrors.NewNotFound("AssetRates.GetByExternalID", "AssetRate", externalID)
}

func (f *fakeAssetRates) GetFromAssetCode(_ context.Context, _, _, assetCode string) (*midaz.AssetRate, error) {
	for _, rate := range f.store.List() {
		if rate.BaseAssetCode == assetCode {
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

var _ midaz.PortfoliosService = (*fakePortfolios)(nil)

func newFakePortfolios(cfg *fakeConfig) *fakePortfolios {
	return &fakePortfolios{store: newFakeStore[midaz.Portfolio](), cfg: cfg}
}

func (f *fakePortfolios) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreatePortfolioInput) (*midaz.Portfolio, error) {
	if err := f.cfg.injectedError("midaz.Portfolios.Create"); err != nil {
		return nil, err
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

func (f *fakePortfolios) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.Portfolio] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakePortfolios) Update(_ context.Context, _, _ string, id string, input *midaz.UpdatePortfolioInput) (*midaz.Portfolio, error) {
	if err := f.cfg.injectedError("midaz.Portfolios.Update"); err != nil {
		return nil, err
	}

	p, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Portfolios.Update", "Portfolio", id)
	}

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
	f.store.Set(id, p)

	return &p, nil
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

// ---------------------------------------------------------------------------
// Segments
// ---------------------------------------------------------------------------

type fakeSegments struct {
	store *fakeStore[midaz.Segment]
	cfg   *fakeConfig
}

var _ midaz.SegmentsService = (*fakeSegments)(nil)

func newFakeSegments(cfg *fakeConfig) *fakeSegments {
	return &fakeSegments{store: newFakeStore[midaz.Segment](), cfg: cfg}
}

func (f *fakeSegments) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateSegmentInput) (*midaz.Segment, error) {
	if err := f.cfg.injectedError("midaz.Segments.Create"); err != nil {
		return nil, err
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

func (f *fakeSegments) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.Segment] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeSegments) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateSegmentInput) (*midaz.Segment, error) {
	if err := f.cfg.injectedError("midaz.Segments.Update"); err != nil {
		return nil, err
	}

	seg, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Segments.Update", "Segment", id)
	}

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
	f.store.Set(id, seg)

	return &seg, nil
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

// ---------------------------------------------------------------------------
// Balances
// ---------------------------------------------------------------------------

type fakeBalances struct {
	store *fakeStore[midaz.Balance]
	cfg   *fakeConfig
}

var _ midaz.BalancesService = (*fakeBalances)(nil)

func newFakeBalances(cfg *fakeConfig) *fakeBalances {
	return &fakeBalances{store: newFakeStore[midaz.Balance](), cfg: cfg}
}

func (f *fakeBalances) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateBalanceInput) (*midaz.Balance, error) {
	if err := f.cfg.injectedError("midaz.Balances.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

	bal := midaz.Balance{
		ID:             generateID("bal"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		AccountID:      input.AccountID,
		AssetCode:      input.AssetCode,
		Status:         models.Status{Code: "active"},
		AllowSending:   true,
		AllowReceiving: true,
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(bal.ID, bal)

	return &bal, nil
}

func (f *fakeBalances) Get(_ context.Context, _, _ string, id string) (*midaz.Balance, error) {
	if err := f.cfg.injectedError("midaz.Balances.Get"); err != nil {
		return nil, err
	}

	bal, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Balances.Get", "Balance", id)
	}

	return &bal, nil
}

func (f *fakeBalances) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.Balance] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeBalances) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateBalanceInput) (*midaz.Balance, error) {
	if err := f.cfg.injectedError("midaz.Balances.Update"); err != nil {
		return nil, err
	}

	bal, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Balances.Update", "Balance", id)
	}

	if input.AllowSending != nil {
		bal.AllowSending = *input.AllowSending
	}

	if input.AllowReceiving != nil {
		bal.AllowReceiving = *input.AllowReceiving
	}

	if input.Status != nil {
		bal.Status = *input.Status
	}

	if input.Metadata != nil {
		bal.Metadata = input.Metadata
	}

	bal.UpdatedAt = time.Now()
	f.store.Set(id, bal)

	return &bal, nil
}

func (f *fakeBalances) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.Balances.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("Balances.Delete", "Balance", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeBalances) GetByAlias(_ context.Context, _, _, alias string) (*midaz.Balance, error) {
	for _, bal := range f.store.List() {
		if bal.AccountAlias != nil && *bal.AccountAlias == alias {
			return &bal, nil
		}
	}

	return nil, sdkerrors.NewNotFound("Balances.GetByAlias", "Balance", alias)
}

func (f *fakeBalances) GetByExternalCode(_ context.Context, _, _, code string) (*midaz.Balance, error) {
	// Balance model does not have ExternalCode; this is a passthrough lookup.
	// Return not-found for the fake.
	return nil, sdkerrors.NewNotFound("Balances.GetByExternalCode", "Balance", code)
}

func (f *fakeBalances) GetByAccountID(_ context.Context, _, _, accountID string) (*midaz.Balance, error) {
	for _, bal := range f.store.List() {
		if bal.AccountID == accountID {
			return &bal, nil
		}
	}

	return nil, sdkerrors.NewNotFound("Balances.GetByAccountID", "Balance", accountID)
}

// ---------------------------------------------------------------------------
// Transactions
// ---------------------------------------------------------------------------

type fakeTransactions struct {
	store *fakeStore[midaz.Transaction]
	cfg   *fakeConfig
}

var _ midaz.TransactionsService = (*fakeTransactions)(nil)

func newFakeTransactions(cfg *fakeConfig) *fakeTransactions {
	return &fakeTransactions{store: newFakeStore[midaz.Transaction](), cfg: cfg}
}

func (f *fakeTransactions) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateTransactionInput) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

	tx := midaz.Transaction{
		ID:             generateID("tx"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Description:    input.Description,
		AssetCode:      input.AssetCode,
		Amount:         input.Amount,
		AmountScale:    input.Scale,
		Status:         models.Status{Code: "pending"},
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if input.ChartOfAccountsGroupName != nil {
		tx.ChartOfAccountsGroupName = input.ChartOfAccountsGroupName
	}

	f.store.Set(tx.ID, tx)

	return &tx, nil
}

func (f *fakeTransactions) Get(_ context.Context, _, _ string, id string) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Get"); err != nil {
		return nil, err
	}

	tx, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Transactions.Get", "Transaction", id)
	}

	return &tx, nil
}

func (f *fakeTransactions) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.Transaction] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeTransactions) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateTransactionInput) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Update"); err != nil {
		return nil, err
	}

	tx, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Transactions.Update", "Transaction", id)
	}

	if input.Description != nil {
		tx.Description = input.Description
	}

	if input.Status != nil {
		tx.Status = *input.Status
	}

	if input.Metadata != nil {
		tx.Metadata = input.Metadata
	}

	tx.UpdatedAt = time.Now()
	f.store.Set(id, tx)

	return &tx, nil
}

func (f *fakeTransactions) Commit(_ context.Context, _, _ string, id string) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Commit"); err != nil {
		return nil, err
	}

	tx, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Transactions.Commit", "Transaction", id)
	}

	if tx.Status.Code != "pending" {
		return nil, sdkerrors.NewConflict("midaz", "Transactions.Commit", "Transaction",
			fmt.Sprintf("cannot commit transaction %s: current status is %q, expected \"pending\"", id, tx.Status.Code))
	}

	tx.Status = models.Status{Code: "committed"}
	tx.UpdatedAt = time.Now()
	f.store.Set(id, tx)

	return &tx, nil
}

func (f *fakeTransactions) Cancel(_ context.Context, _, _ string, id string) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Cancel"); err != nil {
		return nil, err
	}

	tx, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Transactions.Cancel", "Transaction", id)
	}

	if tx.Status.Code != "pending" {
		return nil, sdkerrors.NewConflict("midaz", "Transactions.Cancel", "Transaction",
			fmt.Sprintf("cannot cancel transaction %s: current status is %q, expected \"pending\"", id, tx.Status.Code))
	}

	tx.Status = models.Status{Code: "cancelled"}
	tx.UpdatedAt = time.Now()
	f.store.Set(id, tx)

	return &tx, nil
}

func (f *fakeTransactions) Revert(_ context.Context, _, _ string, id string) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Revert"); err != nil {
		return nil, err
	}

	tx, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Transactions.Revert", "Transaction", id)
	}

	if tx.Status.Code != "committed" {
		return nil, sdkerrors.NewConflict("midaz", "Transactions.Revert", "Transaction",
			fmt.Sprintf("cannot revert transaction %s: current status is %q, expected \"committed\"", id, tx.Status.Code))
	}

	tx.Status = models.Status{Code: "reverted"}
	tx.UpdatedAt = time.Now()
	f.store.Set(id, tx)

	return &tx, nil
}

// ---------------------------------------------------------------------------
// TransactionRoutes
// ---------------------------------------------------------------------------

type fakeTransactionRoutes struct {
	store *fakeStore[midaz.TransactionRoute]
	cfg   *fakeConfig
}

var _ midaz.TransactionRoutesService = (*fakeTransactionRoutes)(nil)

func newFakeTransactionRoutes(cfg *fakeConfig) *fakeTransactionRoutes {
	return &fakeTransactionRoutes{store: newFakeStore[midaz.TransactionRoute](), cfg: cfg}
}

func (f *fakeTransactionRoutes) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateTransactionRouteInput) (*midaz.TransactionRoute, error) {
	if err := f.cfg.injectedError("midaz.TransactionRoutes.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

	tr := midaz.TransactionRoute{
		ID:              generateID("txroute"),
		OrganizationID:  orgID,
		LedgerID:        ledgerID,
		TransactionType: input.TransactionType,
		Description:     input.Description,
		Code:            input.Code,
		Metadata:        input.Metadata,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	f.store.Set(tr.ID, tr)

	return &tr, nil
}

func (f *fakeTransactionRoutes) Get(_ context.Context, _, _ string, id string) (*midaz.TransactionRoute, error) {
	if err := f.cfg.injectedError("midaz.TransactionRoutes.Get"); err != nil {
		return nil, err
	}

	tr, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("TransactionRoutes.Get", "TransactionRoute", id)
	}

	return &tr, nil
}

func (f *fakeTransactionRoutes) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.TransactionRoute] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeTransactionRoutes) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateTransactionRouteInput) (*midaz.TransactionRoute, error) {
	if err := f.cfg.injectedError("midaz.TransactionRoutes.Update"); err != nil {
		return nil, err
	}

	tr, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("TransactionRoutes.Update", "TransactionRoute", id)
	}

	if input.Description != nil {
		tr.Description = input.Description
	}

	if input.Metadata != nil {
		tr.Metadata = input.Metadata
	}

	tr.UpdatedAt = time.Now()
	f.store.Set(id, tr)

	return &tr, nil
}

func (f *fakeTransactionRoutes) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.TransactionRoutes.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("TransactionRoutes.Delete", "TransactionRoute", id)
	}

	f.store.Delete(id)

	return nil
}

// ---------------------------------------------------------------------------
// Operations (read-only)
// ---------------------------------------------------------------------------

type fakeOperations struct {
	store *fakeStore[midaz.Operation]
	cfg   *fakeConfig
}

var _ midaz.OperationsService = (*fakeOperations)(nil)

func newFakeOperations(cfg *fakeConfig) *fakeOperations {
	return &fakeOperations{store: newFakeStore[midaz.Operation](), cfg: cfg}
}

func (f *fakeOperations) Get(_ context.Context, _, _ string, id string) (*midaz.Operation, error) {
	if err := f.cfg.injectedError("midaz.Operations.Get"); err != nil {
		return nil, err
	}

	op, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Operations.Get", "Operation", id)
	}

	return &op, nil
}

func (f *fakeOperations) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.Operation] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeOperations) ListByTransaction(_ context.Context, _, _, transactionID string, _ *models.ListOptions) *pagination.Iterator[midaz.Operation] {
	var ops []midaz.Operation
	for _, op := range f.store.List() {
		if op.TransactionID == transactionID {
			ops = append(ops, op)
		}
	}

	return pagination.NewIteratorFromSlice(ops)
}

func (f *fakeOperations) ListByAccount(_ context.Context, _, _, accountID string, _ *models.ListOptions) *pagination.Iterator[midaz.Operation] {
	var ops []midaz.Operation
	for _, op := range f.store.List() {
		if op.AccountID == accountID {
			ops = append(ops, op)
		}
	}

	return pagination.NewIteratorFromSlice(ops)
}

// ---------------------------------------------------------------------------
// OperationRoutes
// ---------------------------------------------------------------------------

type fakeOperationRoutes struct {
	store *fakeStore[midaz.OperationRoute]
	cfg   *fakeConfig
}

var _ midaz.OperationRoutesService = (*fakeOperationRoutes)(nil)

func newFakeOperationRoutes(cfg *fakeConfig) *fakeOperationRoutes {
	return &fakeOperationRoutes{store: newFakeStore[midaz.OperationRoute](), cfg: cfg}
}

func (f *fakeOperationRoutes) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateOperationRouteInput) (*midaz.OperationRoute, error) {
	if err := f.cfg.injectedError("midaz.OperationRoutes.Create"); err != nil {
		return nil, err
	}

	now := time.Now()

	or := midaz.OperationRoute{
		ID:             generateID("oproute"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		AccountID:      input.AccountID,
		Type:           input.Type,
		Description:    input.Description,
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(or.ID, or)

	return &or, nil
}

func (f *fakeOperationRoutes) Get(_ context.Context, _, _ string, id string) (*midaz.OperationRoute, error) {
	if err := f.cfg.injectedError("midaz.OperationRoutes.Get"); err != nil {
		return nil, err
	}

	or, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("OperationRoutes.Get", "OperationRoute", id)
	}

	return &or, nil
}

func (f *fakeOperationRoutes) List(_ context.Context, _, _ string, opts *models.ListOptions) *pagination.Iterator[midaz.OperationRoute] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeOperationRoutes) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateOperationRouteInput) (*midaz.OperationRoute, error) {
	if err := f.cfg.injectedError("midaz.OperationRoutes.Update"); err != nil {
		return nil, err
	}

	or, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("OperationRoutes.Update", "OperationRoute", id)
	}

	if input.AccountID != nil {
		or.AccountID = *input.AccountID
	}

	if input.AccountAlias != nil {
		or.AccountAlias = input.AccountAlias
	}

	if input.Description != nil {
		or.Description = input.Description
	}

	if input.Metadata != nil {
		or.Metadata = input.Metadata
	}

	or.UpdatedAt = time.Now()
	f.store.Set(id, or)

	return &or, nil
}

func (f *fakeOperationRoutes) Delete(_ context.Context, _, _ string, id string) error {
	if err := f.cfg.injectedError("midaz.OperationRoutes.Delete"); err != nil {
		return err
	}

	if _, ok := f.store.Get(id); !ok {
		return sdkerrors.NewNotFound("OperationRoutes.Delete", "OperationRoute", id)
	}

	f.store.Delete(id)

	return nil
}
