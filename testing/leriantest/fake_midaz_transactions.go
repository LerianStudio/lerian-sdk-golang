package leriantest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// ---------------------------------------------------------------------------
// Balances
// ---------------------------------------------------------------------------

type fakeBalances struct {
	store    *fakeStore[midaz.Balance]
	accounts *fakeStore[midaz.Account]
	cfg      *fakeConfig
}

func newFakeBalances(cfg *fakeConfig, accounts *fakeStore[midaz.Account]) *fakeBalances {
	return &fakeBalances{store: newFakeStore[midaz.Balance](), accounts: accounts, cfg: cfg}
}

func (f *fakeBalances) accountForScope(accountID, orgID, ledgerID string) (midaz.Account, bool) {
	account, ok := f.accounts.Get(accountID)
	if !ok {
		return midaz.Account{}, false
	}

	if account.OrganizationID != orgID || account.LedgerID != ledgerID {
		return midaz.Account{}, false
	}

	return account, true
}

func (f *fakeBalances) hydrateBalance(balance midaz.Balance) midaz.Balance {
	account, ok := f.accountForScope(balance.AccountID, balance.OrganizationID, balance.LedgerID)
	if !ok {
		return balance
	}

	balance.AccountAlias = account.Alias
	if balance.AssetCode == "" {
		balance.AssetCode = account.AssetCode
	}

	return balance
}

func (f *fakeBalances) CreateForAccount(_ context.Context, orgID, ledgerID, accountID string, input *midaz.CreateBalanceInput) (*midaz.Balance, error) {
	if err := f.cfg.injectedError("midaz.Balances.Create"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Balances.CreateForAccount", "Balance", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Balances.CreateForAccount", "Balance", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Balances.CreateForAccount", "Balance", "input is required")
	}

	if input.Key == "" {
		return nil, sdkerrors.NewValidation("Balances.CreateForAccount", "Balance", "key is required")
	}

	if accountID == "" {
		return nil, sdkerrors.NewValidation("Balances.CreateForAccount", "Balance", "account id is required")
	}

	account, ok := f.accountForScope(accountID, orgID, ledgerID)
	if !ok {
		return nil, sdkerrors.NewNotFound("Balances.CreateForAccount", "Account", accountID)
	}

	now := time.Now()
	bal := midaz.Balance{
		ID:             generateID("bal"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		AccountID:      accountID,
		AssetCode:      account.AssetCode,
		AccountAlias:   account.Alias,
		Status:         models.Status{Code: "active"},
		AllowSending:   input.AllowSending == nil || *input.AllowSending,
		AllowReceiving: input.AllowReceiving == nil || *input.AllowReceiving,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(bal.ID, bal)

	return &bal, nil
}

func (f *fakeBalances) Get(_ context.Context, orgID, ledgerID string, id string) (*midaz.Balance, error) {
	if err := f.cfg.injectedError("midaz.Balances.Get"); err != nil {
		return nil, err
	}

	bal, ok := f.store.Get(id)
	if !ok || !balanceInScope(bal, orgID, ledgerID) {
		return nil, sdkerrors.NewNotFound("Balances.Get", "Balance", id)
	}

	return &bal, nil
}

func (f *fakeBalances) List(_ context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[midaz.Balance] {
	filtered := make([]midaz.Balance, 0)
	for _, bal := range f.store.List() {
		if balanceInScope(bal, orgID, ledgerID) {
			filtered = append(filtered, bal)
		}
	}

	return pagination.NewIteratorFromSlice(filtered)
}

func (f *fakeBalances) Update(_ context.Context, orgID, ledgerID string, id string, input *midaz.UpdateBalanceInput) (*midaz.Balance, error) {
	if err := f.cfg.injectedError("midaz.Balances.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Balances.Update", "Balance", "input is required")
	}

	return fakeScopedMutateStored(f.cfg, "", "Balances.Update", "Balance", id, f.store, func(b midaz.Balance) bool {
		return balanceInScope(b, orgID, ledgerID)
	}, func(bal *midaz.Balance) error {
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
		return nil
	})
}

func (f *fakeBalances) Delete(_ context.Context, orgID, ledgerID string, id string) error {
	if err := f.cfg.injectedError("midaz.Balances.Delete"); err != nil {
		return err
	}

	bal, ok := f.store.Get(id)
	if !ok || !balanceInScope(bal, orgID, ledgerID) {
		return sdkerrors.NewNotFound("Balances.Delete", "Balance", id)
	}

	f.store.Delete(id)

	return nil
}

func (f *fakeBalances) ListByAlias(_ context.Context, orgID, ledgerID, alias string) ([]midaz.Balance, error) {
	if err := validateRequiredField("Balances.ListByAlias", "Balance", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Balances.ListByAlias", "Balance", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Balances.ListByAlias", "Balance", alias, "alias"); err != nil {
		return nil, err
	}

	var balances []midaz.Balance
	for _, bal := range f.store.List() {
		if bal.OrganizationID != orgID || bal.LedgerID != ledgerID {
			continue
		}

		account, ok := f.accountForScope(bal.AccountID, orgID, ledgerID)
		if !ok || account.Alias == nil || *account.Alias != alias {
			continue
		}

		balances = append(balances, f.hydrateBalance(bal))
	}

	return balances, nil
}

func (f *fakeBalances) ListByExternalCode(_ context.Context, orgID, ledgerID string, code string) ([]midaz.Balance, error) {
	if err := validateRequiredField("Balances.ListByExternalCode", "Balance", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Balances.ListByExternalCode", "Balance", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Balances.ListByExternalCode", "Balance", code, "external code"); err != nil {
		return nil, err
	}

	accountIDs := make(map[string]struct{})
	for _, account := range f.accounts.List() {
		if account.OrganizationID == orgID && account.LedgerID == ledgerID && account.ExternalCode != nil && *account.ExternalCode == code {
			accountIDs[account.ID] = struct{}{}
		}
	}

	var balances []midaz.Balance
	for _, bal := range f.store.List() {
		if bal.OrganizationID == orgID && bal.LedgerID == ledgerID {
			if _, ok := accountIDs[bal.AccountID]; ok {
				balances = append(balances, f.hydrateBalance(bal))
			}
		}
	}

	return balances, nil
}

func (f *fakeBalances) ListByAccountID(_ context.Context, orgID, ledgerID string, accountID string) ([]midaz.Balance, error) {
	if err := validateRequiredField("Balances.ListByAccountID", "Balance", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Balances.ListByAccountID", "Balance", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Balances.ListByAccountID", "Balance", accountID, "account id"); err != nil {
		return nil, err
	}

	var balances []midaz.Balance
	for _, bal := range f.store.List() {
		if bal.OrganizationID == orgID && bal.LedgerID == ledgerID && bal.AccountID == accountID {
			balances = append(balances, f.hydrateBalance(bal))
		}
	}

	return balances, nil
}

// ---------------------------------------------------------------------------
// Transactions
// ---------------------------------------------------------------------------

type fakeTransactions struct {
	store *fakeStore[midaz.Transaction]
	cfg   *fakeConfig
}

func newFakeTransactions(cfg *fakeConfig) *fakeTransactions {
	return &fakeTransactions{store: newFakeStore[midaz.Transaction](), cfg: cfg}
}

func buildFakeTransactionAmounts(operation string, input *midaz.CreateTransactionInput) (string, int64, int, error) {
	amount, scale, err := parseScaledValue(operation, "Transaction", input.Send.Value)
	if err != nil {
		return "", 0, 0, err
	}

	return input.Send.Asset, amount, scale, nil
}

func buildFakeVariantTransactionAmounts(operation, resource, asset, value string) (string, int64, int, error) {
	amount, scale, err := parseScaledValue(operation, resource, value)
	if err != nil {
		return "", 0, 0, err
	}

	return asset, amount, scale, nil
}

func (f *fakeTransactions) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateTransactionInput) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Create"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Create", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Create", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Transactions.Create", "Transaction", "input is required")
	}

	if input.Send == nil {
		return nil, sdkerrors.NewValidation("Transactions.Create", "Transaction", "send is required")
	}

	assetCode, amount, scale, err := buildFakeTransactionAmounts("Transactions.Create", input)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	tx := midaz.Transaction{
		ID:             generateID("tx"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Description:    input.Description,
		AssetCode:      assetCode,
		Amount:         amount,
		AmountScale:    scale,
		Status:         models.Status{Code: "pending"},
		Metadata:       input.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if input.ChartOfAccountsGroupName != nil {
		tx.ChartOfAccountsGroupName = input.ChartOfAccountsGroupName
	}

	if input.ParentTransactionID != nil {
		tx.ParentTransactionID = input.ParentTransactionID
	}

	f.store.Set(tx.ID, tx)

	return &tx, nil
}

func (f *fakeTransactions) CreateAnnotation(ctx context.Context, orgID, ledgerID string, input *midaz.CreateTransactionInput) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.CreateAnnotation"); err != nil {
		return nil, err
	}

	return f.Create(ctx, orgID, ledgerID, input)
}

func (f *fakeTransactions) CreateDSL(_ context.Context, orgID, ledgerID string, dslContent []byte) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.CreateDSL"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.CreateDSL", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.CreateDSL", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if len(dslContent) == 0 {
		return nil, sdkerrors.NewValidation("Transactions.CreateDSL", "Transaction", "DSL content is required")
	}

	if len(dslContent) > 10<<20 {
		return nil, sdkerrors.NewValidation("Transactions.CreateDSL", "Transaction", "DSL content exceeds maximum allowed size")
	}

	now := time.Now()

	tx := midaz.Transaction{
		ID:             generateID("tx"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Status:         models.Status{Code: "pending"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	f.store.Set(tx.ID, tx)

	return &tx, nil
}

func (f *fakeTransactions) CreateInflow(_ context.Context, orgID, ledgerID string, input *midaz.CreateTransactionInflowInput) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.CreateInflow"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.CreateInflow", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.CreateInflow", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Transactions.CreateInflow", "Transaction", "input is required")
	}

	if strings.TrimSpace(input.Send.Asset) == "" {
		return nil, sdkerrors.NewValidation("Transactions.CreateInflow", "Transaction", "send asset is required")
	}

	if strings.TrimSpace(input.Send.Value) == "" {
		return nil, sdkerrors.NewValidation("Transactions.CreateInflow", "Transaction", "send value is required")
	}

	if err := validateTransactionVariantLegsFake("Transactions.CreateInflow", "distribute", input.Send.Distribute.To); err != nil {
		return nil, err
	}

	assetCode, amount, scale, err := buildFakeVariantTransactionAmounts("Transactions.CreateInflow", "Transaction", input.Send.Asset, input.Send.Value)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	tx := midaz.Transaction{
		ID:             generateID("tx"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Description:    input.Description,
		AssetCode:      assetCode,
		Amount:         amount,
		AmountScale:    scale,
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

func (f *fakeTransactions) CreateOutflow(_ context.Context, orgID, ledgerID string, input *midaz.CreateTransactionOutflowInput) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.CreateOutflow"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.CreateOutflow", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.CreateOutflow", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Transactions.CreateOutflow", "Transaction", "input is required")
	}

	if strings.TrimSpace(input.Send.Asset) == "" {
		return nil, sdkerrors.NewValidation("Transactions.CreateOutflow", "Transaction", "send asset is required")
	}

	if strings.TrimSpace(input.Send.Value) == "" {
		return nil, sdkerrors.NewValidation("Transactions.CreateOutflow", "Transaction", "send value is required")
	}

	if err := validateTransactionVariantLegsFake("Transactions.CreateOutflow", "source", input.Send.Source.From); err != nil {
		return nil, err
	}

	assetCode, amount, scale, err := buildFakeVariantTransactionAmounts("Transactions.CreateOutflow", "Transaction", input.Send.Asset, input.Send.Value)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	tx := midaz.Transaction{
		ID:             generateID("tx"),
		OrganizationID: orgID,
		LedgerID:       ledgerID,
		Description:    input.Description,
		AssetCode:      assetCode,
		Amount:         amount,
		AmountScale:    scale,
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

func (f *fakeTransactions) Get(_ context.Context, orgID, ledgerID string, id string) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Get"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Get", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Get", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Get", "Transaction", id, "transaction id"); err != nil {
		return nil, err
	}

	tx, ok := f.store.Get(id)
	if !ok || !transactionInScope(tx, orgID, ledgerID) {
		return nil, sdkerrors.NewNotFound("Transactions.Get", "Transaction", id)
	}

	return &tx, nil
}

func (f *fakeTransactions) List(_ context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[midaz.Transaction] {
	filtered := make([]midaz.Transaction, 0)
	for _, tx := range f.store.List() {
		if transactionInScope(tx, orgID, ledgerID) {
			filtered = append(filtered, tx)
		}
	}

	return pagination.NewIteratorFromSlice(filtered)
}

func (f *fakeTransactions) Update(_ context.Context, orgID, ledgerID string, id string, input *midaz.UpdateTransactionInput) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Update"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Update", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Update", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Update", "Transaction", id, "transaction id"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Transactions.Update", "Transaction", "input is required")
	}

	return fakeScopedMutateStored(f.cfg, "", "Transactions.Update", "Transaction", id, f.store, func(tx midaz.Transaction) bool {
		return transactionInScope(tx, orgID, ledgerID)
	}, func(tx *midaz.Transaction) error {
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
		return nil
	})
}

func (f *fakeTransactions) Commit(_ context.Context, orgID, ledgerID string, id string) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Commit"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Commit", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Commit", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Commit", "Transaction", id, "transaction id"); err != nil {
		return nil, err
	}

	return fakeScopedMutateStored(f.cfg, "", "Transactions.Commit", "Transaction", id, f.store, func(tx midaz.Transaction) bool {
		return transactionInScope(tx, orgID, ledgerID)
	}, func(tx *midaz.Transaction) error {
		if tx.Status.Code != "pending" {
			return sdkerrors.NewConflict("midaz", "Transactions.Commit", "Transaction",
				fmt.Sprintf("cannot commit transaction %s: current status is %q, expected \"pending\"", id, tx.Status.Code))
		}
		tx.Status = models.Status{Code: "committed"}
		tx.UpdatedAt = time.Now()
		return nil
	})
}

func (f *fakeTransactions) Cancel(_ context.Context, orgID, ledgerID string, id string) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Cancel"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Cancel", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Cancel", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Cancel", "Transaction", id, "transaction id"); err != nil {
		return nil, err
	}

	return fakeScopedMutateStored(f.cfg, "", "Transactions.Cancel", "Transaction", id, f.store, func(tx midaz.Transaction) bool {
		return transactionInScope(tx, orgID, ledgerID)
	}, func(tx *midaz.Transaction) error {
		if tx.Status.Code != "pending" {
			return sdkerrors.NewConflict("midaz", "Transactions.Cancel", "Transaction",
				fmt.Sprintf("cannot cancel transaction %s: current status is %q, expected \"pending\"", id, tx.Status.Code))
		}
		tx.Status = models.Status{Code: "cancelled"}
		tx.UpdatedAt = time.Now()
		return nil
	})
}

func (f *fakeTransactions) Revert(_ context.Context, orgID, ledgerID string, id string) (*midaz.Transaction, error) {
	if err := f.cfg.injectedError("midaz.Transactions.Revert"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Revert", "Transaction", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Revert", "Transaction", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Transactions.Revert", "Transaction", id, "transaction id"); err != nil {
		return nil, err
	}

	return fakeScopedMutateStored(f.cfg, "", "Transactions.Revert", "Transaction", id, f.store, func(tx midaz.Transaction) bool {
		return transactionInScope(tx, orgID, ledgerID)
	}, func(tx *midaz.Transaction) error {
		if tx.Status.Code != "committed" {
			return sdkerrors.NewConflict("midaz", "Transactions.Revert", "Transaction",
				fmt.Sprintf("cannot revert transaction %s: current status is %q, expected \"committed\"", id, tx.Status.Code))
		}
		tx.Status = models.Status{Code: "reverted"}
		tx.UpdatedAt = time.Now()
		return nil
	})
}

// ---------------------------------------------------------------------------
// TransactionRoutes
// ---------------------------------------------------------------------------

type fakeTransactionRoutes struct {
	store *fakeStore[midaz.TransactionRoute]
	cfg   *fakeConfig
}

func newFakeTransactionRoutes(cfg *fakeConfig) *fakeTransactionRoutes {
	return &fakeTransactionRoutes{store: newFakeStore[midaz.TransactionRoute](), cfg: cfg}
}

func (f *fakeTransactionRoutes) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateTransactionRouteInput) (*midaz.TransactionRoute, error) {
	if err := f.cfg.injectedError("midaz.TransactionRoutes.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("TransactionRoutes.Create", "TransactionRoute", "input is required")
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

func (f *fakeTransactionRoutes) List(_ context.Context, _, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.TransactionRoute] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeTransactionRoutes) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateTransactionRouteInput) (*midaz.TransactionRoute, error) {
	if err := f.cfg.injectedError("midaz.TransactionRoutes.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("TransactionRoutes.Update", "TransactionRoute", "input is required")
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
// Operations
// ---------------------------------------------------------------------------

type fakeOperations struct {
	store *fakeStore[midaz.Operation]
	cfg   *fakeConfig
}

func newFakeOperations(cfg *fakeConfig) *fakeOperations {
	return &fakeOperations{store: newFakeStore[midaz.Operation](), cfg: cfg}
}

func (f *fakeOperations) Get(_ context.Context, orgID, ledgerID string, id string) (*midaz.Operation, error) {
	if err := f.cfg.injectedError("midaz.Operations.Get"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Operations.Get", "Operation", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Operations.Get", "Operation", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Operations.Get", "Operation", id, "operation id"); err != nil {
		return nil, err
	}

	op, ok := f.store.Get(id)
	if !ok {
		return nil, sdkerrors.NewNotFound("Operations.Get", "Operation", id)
	}

	if op.OrganizationID != "" && op.OrganizationID != orgID {
		return nil, sdkerrors.NewNotFound("Operations.Get", "Operation", id)
	}

	if op.LedgerID != "" && op.LedgerID != ledgerID {
		return nil, sdkerrors.NewNotFound("Operations.Get", "Operation", id)
	}

	return &op, nil
}

func (f *fakeOperations) List(_ context.Context, orgID, ledgerID string, _ *models.CursorListOptions) *pagination.Iterator[midaz.Operation] {
	if orgID == "" || ledgerID == "" {
		return pagination.NewErrorIterator[midaz.Operation](sdkerrors.NewValidation("Operations.List", "Operation", "organization ID and ledger ID are required"))
	}

	var ops []midaz.Operation
	for _, op := range f.store.List() {
		if op.OrganizationID != "" && op.OrganizationID != orgID {
			continue
		}

		if op.LedgerID != "" && op.LedgerID != ledgerID {
			continue
		}

		ops = append(ops, op)
	}

	return pagination.NewIteratorFromSlice(ops)
}

func (f *fakeOperations) ListByTransaction(_ context.Context, orgID, ledgerID, transactionID string, _ *models.CursorListOptions) *pagination.Iterator[midaz.Operation] {
	if orgID == "" || ledgerID == "" || transactionID == "" {
		return pagination.NewErrorIterator[midaz.Operation](sdkerrors.NewValidation("Operations.ListByTransaction", "Operation", "organization ID, ledger ID, and transaction ID are required"))
	}

	var ops []midaz.Operation
	for _, op := range f.store.List() {
		if op.TransactionID == transactionID && (op.OrganizationID == "" || op.OrganizationID == orgID) && (op.LedgerID == "" || op.LedgerID == ledgerID) {
			ops = append(ops, op)
		}
	}

	return pagination.NewIteratorFromSlice(ops)
}

func (f *fakeOperations) ListByAccount(_ context.Context, orgID, ledgerID, accountID string, _ *models.CursorListOptions) *pagination.Iterator[midaz.Operation] {
	if orgID == "" || ledgerID == "" || accountID == "" {
		return pagination.NewErrorIterator[midaz.Operation](sdkerrors.NewValidation("Operations.ListByAccount", "Operation", "organization ID, ledger ID, and account ID are required"))
	}

	var ops []midaz.Operation
	for _, op := range f.store.List() {
		if op.AccountID == accountID && (op.OrganizationID == "" || op.OrganizationID == orgID) && (op.LedgerID == "" || op.LedgerID == ledgerID) {
			ops = append(ops, op)
		}
	}

	return pagination.NewIteratorFromSlice(ops)
}

func (f *fakeOperations) Update(_ context.Context, orgID, ledgerID string, transactionID, operationID string, input *midaz.UpdateOperationInput) (*midaz.Operation, error) {
	if err := f.cfg.injectedError("midaz.Operations.Update"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Operations.Update", "Operation", orgID, "organization id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Operations.Update", "Operation", ledgerID, "ledger id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Operations.Update", "Operation", transactionID, "transaction id"); err != nil {
		return nil, err
	}

	if err := validateRequiredField("Operations.Update", "Operation", operationID, "operation id"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("Operations.Update", "Operation", "input is required")
	}

	return fakeScopedMutateStored(f.cfg, "", "Operations.Update", "Operation", operationID, f.store, func(op midaz.Operation) bool {
		if op.OrganizationID != "" && op.OrganizationID != orgID {
			return false
		}
		if op.LedgerID != "" && op.LedgerID != ledgerID {
			return false
		}
		return op.TransactionID == transactionID
	}, func(op *midaz.Operation) error {
		if input.Description != nil {
			op.Description = input.Description
		}
		if input.Metadata != nil {
			op.Metadata = input.Metadata
		}
		op.UpdatedAt = time.Now()
		return nil
	})
}

// ---------------------------------------------------------------------------
// OperationRoutes
// ---------------------------------------------------------------------------

type fakeOperationRoutes struct {
	store *fakeStore[midaz.OperationRoute]
	cfg   *fakeConfig
}

func newFakeOperationRoutes(cfg *fakeConfig) *fakeOperationRoutes {
	return &fakeOperationRoutes{store: newFakeStore[midaz.OperationRoute](), cfg: cfg}
}

func (f *fakeOperationRoutes) Create(_ context.Context, orgID, ledgerID string, input *midaz.CreateOperationRouteInput) (*midaz.OperationRoute, error) {
	if err := f.cfg.injectedError("midaz.OperationRoutes.Create"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("OperationRoutes.Create", "OperationRoute", "input is required")
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

func (f *fakeOperationRoutes) List(_ context.Context, _, _ string, opts *models.CursorListOptions) *pagination.Iterator[midaz.OperationRoute] {
	return f.store.PaginatedIterator(opts)
}

func (f *fakeOperationRoutes) Update(_ context.Context, _, _ string, id string, input *midaz.UpdateOperationRouteInput) (*midaz.OperationRoute, error) {
	if err := f.cfg.injectedError("midaz.OperationRoutes.Update"); err != nil {
		return nil, err
	}

	if input == nil {
		return nil, sdkerrors.NewValidation("OperationRoutes.Update", "OperationRoute", "input is required")
	}

	return fakeMutateStored(f.cfg, "", "OperationRoutes.Update", "OperationRoute", id, f.store, func(or *midaz.OperationRoute) {
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
	})
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
