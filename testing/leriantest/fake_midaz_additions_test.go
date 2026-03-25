package leriantest

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	midazmodels "github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeMidazBalancesCreatePreservesFlagsAndSupportsLookups(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	alias := "@primary"
	externalCode := "EXT-ACC-1"
	account, err := client.Onboarding.Accounts.Create(ctx, "org-1", "ledger-1", &midaz.CreateAccountInput{
		Name:         "Primary",
		Type:         "asset",
		AssetCode:    "BRL",
		Alias:        &alias,
		ExternalCode: &externalCode,
	})
	require.NoError(t, err)
	allowSending := false
	allowReceiving := false

	created, err := client.Transactions.Balances.CreateForAccount(ctx, "org-1", "ledger-1", account.ID, &midaz.CreateBalanceInput{
		Key:            "asset-freeze",
		AllowSending:   &allowSending,
		AllowReceiving: &allowReceiving,
	})
	require.NoError(t, err)
	assert.False(t, created.AllowSending)
	assert.False(t, created.AllowReceiving)
	assert.Equal(t, "BRL", created.AssetCode)

	balancesByAlias, err := client.Transactions.Balances.ListByAlias(ctx, "org-1", "ledger-1", alias)
	require.NoError(t, err)
	require.Len(t, balancesByAlias, 1)
	assert.Equal(t, created.ID, balancesByAlias[0].ID)

	balancesByAccount, err := client.Transactions.Balances.ListByAccountID(ctx, "org-1", "ledger-1", account.ID)
	require.NoError(t, err)
	assert.Len(t, balancesByAccount, 1)

	balancesByExternal, err := client.Transactions.Balances.ListByExternalCode(ctx, "org-1", "ledger-1", externalCode)
	require.NoError(t, err)
	assert.Len(t, balancesByExternal, 1)

	updatedAlias := "@primary-renamed"
	_, err = client.Onboarding.Accounts.Update(ctx, "org-1", "ledger-1", account.ID, &midaz.UpdateAccountInput{Alias: &updatedAlias})
	require.NoError(t, err)

	balancesByAlias, err = client.Transactions.Balances.ListByAlias(ctx, "org-1", "ledger-1", alias)
	require.NoError(t, err)
	assert.Empty(t, balancesByAlias)

	byUpdatedAlias, err := client.Transactions.Balances.ListByAlias(ctx, "org-1", "ledger-1", updatedAlias)
	require.NoError(t, err)
	require.Len(t, byUpdatedAlias, 1)
	assert.Equal(t, created.ID, byUpdatedAlias[0].ID)
	require.NotNil(t, byUpdatedAlias[0].AccountAlias)
	assert.Equal(t, updatedAlias, *byUpdatedAlias[0].AccountAlias)
}

func TestFakeMidazBalancesCreateIgnoresStatusAndMetadataLikeProduction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	account, err := client.Onboarding.Accounts.Create(ctx, "org-1", "ledger-1", &midaz.CreateAccountInput{
		Name:      "Primary",
		Type:      "asset",
		AssetCode: "BRL",
	})
	require.NoError(t, err)

	status := &midazmodels.Status{Code: "suspended"}
	metadata := midazmodels.Metadata{"reason": "manual-hold"}
	created, err := client.Transactions.Balances.CreateForAccount(ctx, "org-1", "ledger-1", account.ID, &midaz.CreateBalanceInput{
		Key:      "asset-freeze",
		Status:   status,
		Metadata: metadata,
	})
	require.NoError(t, err)
	assert.Equal(t, "active", created.Status.Code)
	assert.Nil(t, created.Metadata)
}

func TestFakeMidazBalancesNilInputValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	_, err := client.Transactions.Balances.CreateForAccount(ctx, "org-1", "ledger-1", "acc-1", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestFakeMidazTransactionsVariantValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	_, err := client.Transactions.Transactions.CreateDSL(ctx, "org-1", "ledger-1", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	_, err = client.Transactions.Transactions.CreateInflow(ctx, "org-1", "ledger-1", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	_, err = client.Transactions.Transactions.CreateOutflow(ctx, "org-1", "ledger-1", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	_, err = client.Transactions.Transactions.CreateInflow(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInflowInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send asset is required")

	_, err = client.Transactions.Transactions.CreateInflow(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInflowInput{
		Send: midaz.TransactionInflowSend{Asset: "BRL", Value: "10.00"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "distribute legs are required")

	_, err = client.Transactions.Transactions.CreateOutflow(ctx, "org-1", "ledger-1", &midaz.CreateTransactionOutflowInput{
		Send: midaz.TransactionOutflowSend{Asset: "BRL", Value: "10.00", Source: midaz.TransactionOutflowSource{From: []midaz.TransactionOperationLeg{{}}}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source leg identifier is required")

	_, err = client.Transactions.Transactions.CreateDSL(ctx, "org-1", "ledger-1", bytes.Repeat([]byte("x"), 10<<20+1))
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestFakeMidazBalancesRespectOrgAndLedgerScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	account, err := client.Onboarding.Accounts.Create(ctx, "org-1", "ledger-1", &midaz.CreateAccountInput{Name: "Primary", Type: "asset", AssetCode: "BRL"})
	require.NoError(t, err)
	balance, err := client.Transactions.Balances.CreateForAccount(ctx, "org-1", "ledger-1", account.ID, &midaz.CreateBalanceInput{Key: "reserve"})
	require.NoError(t, err)

	_, err = client.Transactions.Balances.Get(ctx, "org-2", "ledger-1", balance.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	_, err = client.Transactions.Balances.Update(ctx, "org-1", "ledger-2", balance.ID, &midaz.UpdateBalanceInput{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	err = client.Transactions.Balances.Delete(ctx, "org-2", "ledger-2", balance.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	items, err := client.Transactions.Balances.List(ctx, "org-2", "ledger-2", nil).Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazTransactionsRespectOrgAndLedgerScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	tx, err := client.Transactions.Transactions.Create(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInput{Send: &midaz.TransactionSend{Asset: "BRL", Value: "10.00"}})
	require.NoError(t, err)

	_, err = client.Transactions.Transactions.Get(ctx, "org-2", "ledger-1", tx.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	_, err = client.Transactions.Transactions.Update(ctx, "org-1", "ledger-2", tx.ID, &midaz.UpdateTransactionInput{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	_, err = client.Transactions.Transactions.Commit(ctx, "org-2", "ledger-1", tx.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	_, err = client.Transactions.Transactions.Cancel(ctx, "org-1", "ledger-2", tx.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	committed, err := client.Transactions.Transactions.Commit(ctx, "org-1", "ledger-1", tx.ID)
	require.NoError(t, err)
	require.NotNil(t, committed)

	_, err = client.Transactions.Transactions.Revert(ctx, "org-2", "ledger-1", tx.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	items, err := client.Transactions.Transactions.List(ctx, "org-2", "ledger-1", nil).Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazSecondaryLookupsRespectOrgAndLedgerScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	alias := "@shared"
	externalCode := "EXT-SHARED"
	foreignAlias := "@foreign"
	foreignExternal := "EXT-FOREIGN"

	account1, err := client.Onboarding.Accounts.Create(ctx, "org-1", "ledger-1", &midaz.CreateAccountInput{Name: "Primary", Type: "asset", AssetCode: "BRL", Alias: &alias, ExternalCode: &externalCode})
	require.NoError(t, err)
	account2, err := client.Onboarding.Accounts.Create(ctx, "org-2", "ledger-2", &midaz.CreateAccountInput{Name: "Foreign", Type: "asset", AssetCode: "USD", Alias: &foreignAlias, ExternalCode: &foreignExternal})
	require.NoError(t, err)

	lookedUp, err := client.Onboarding.Accounts.GetByAlias(ctx, "org-1", "ledger-1", alias)
	require.NoError(t, err)
	assert.Equal(t, account1.ID, lookedUp.ID)

	_, err = client.Onboarding.Accounts.GetByAlias(ctx, "org-1", "ledger-1", foreignAlias)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	lookedUp, err = client.Onboarding.Accounts.GetByExternalCode(ctx, "org-2", "ledger-2", foreignExternal)
	require.NoError(t, err)
	assert.Equal(t, account2.ID, lookedUp.ID)

	assetCode := "BRL"
	externalID := "rate-1"
	_, err = client.Transactions.AssetRates.Create(ctx, "org-1", "ledger-1", &midaz.CreateAssetRateInput{BaseAssetCode: assetCode, CounterAssetCode: "USD", Amount: 100, Scale: 2, ExternalID: &externalID})
	require.NoError(t, err)
	foreignExternalID := "rate-2"
	_, err = client.Transactions.AssetRates.Create(ctx, "org-2", "ledger-2", &midaz.CreateAssetRateInput{BaseAssetCode: "EUR", CounterAssetCode: "USD", Amount: 100, Scale: 2, ExternalID: &foreignExternalID})
	require.NoError(t, err)

	rate, err := client.Transactions.AssetRates.GetByExternalID(ctx, "org-1", "ledger-1", externalID)
	require.NoError(t, err)
	assert.Equal(t, externalID, *rate.ExternalID)

	_, err = client.Transactions.AssetRates.GetByExternalID(ctx, "org-1", "ledger-1", foreignExternalID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	rate, err = client.Transactions.AssetRates.GetFromAssetCode(ctx, "org-1", "ledger-1", assetCode)
	require.NoError(t, err)
	assert.Equal(t, assetCode, rate.BaseAssetCode)

	_, err = client.Transactions.AssetRates.GetFromAssetCode(ctx, "org-1", "ledger-1", "EUR")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))
}

func TestFakeMidazTransactionsCreateSupportsSendPayload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	tx, err := client.Transactions.Transactions.Create(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInput{
		Send: &midaz.TransactionSend{
			Asset: "BRL",
			Value: "12.34",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, tx)
	assert.Equal(t, "BRL", tx.AssetCode)
	assert.Equal(t, int64(1234), tx.Amount)
	assert.Equal(t, 2, tx.AmountScale)
}

func TestFakeMidazLegacyNilInputValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	tests := []struct {
		name string
		call func() error
	}{
		{name: "organizations create", call: func() error { _, err := client.Onboarding.Organizations.Create(ctx, nil); return err }},
		{name: "organizations update", call: func() error { _, err := client.Onboarding.Organizations.Update(ctx, "org-1", nil); return err }},
		{name: "ledgers create", call: func() error { _, err := client.Onboarding.Ledgers.Create(ctx, "org-1", nil); return err }},
		{name: "ledgers update", call: func() error { _, err := client.Onboarding.Ledgers.Update(ctx, "org-1", "ledger-1", nil); return err }},
		{name: "accounts create", call: func() error { _, err := client.Onboarding.Accounts.Create(ctx, "org-1", "ledger-1", nil); return err }},
		{name: "accounts update", call: func() error {
			_, err := client.Onboarding.Accounts.Update(ctx, "org-1", "ledger-1", "acct-1", nil)
			return err
		}},
		{name: "account types create", call: func() error {
			_, err := client.Onboarding.AccountTypes.Create(ctx, "org-1", "ledger-1", nil)
			return err
		}},
		{name: "account types update", call: func() error {
			_, err := client.Onboarding.AccountTypes.Update(ctx, "org-1", "ledger-1", "type-1", nil)
			return err
		}},
		{name: "assets create", call: func() error { _, err := client.Onboarding.Assets.Create(ctx, "org-1", "ledger-1", nil); return err }},
		{name: "assets update", call: func() error {
			_, err := client.Onboarding.Assets.Update(ctx, "org-1", "ledger-1", "asset-1", nil)
			return err
		}},
		{name: "portfolios create", call: func() error { _, err := client.Onboarding.Portfolios.Create(ctx, "org-1", "ledger-1", nil); return err }},
		{name: "portfolios update", call: func() error {
			_, err := client.Onboarding.Portfolios.Update(ctx, "org-1", "ledger-1", "portfolio-1", nil)
			return err
		}},
		{name: "segments create", call: func() error { _, err := client.Onboarding.Segments.Create(ctx, "org-1", "ledger-1", nil); return err }},
		{name: "segments update", call: func() error {
			_, err := client.Onboarding.Segments.Update(ctx, "org-1", "ledger-1", "segment-1", nil)
			return err
		}},
		{name: "transaction routes create", call: func() error {
			_, err := client.Transactions.TransactionRoutes.Create(ctx, "org-1", "ledger-1", nil)
			return err
		}},
		{name: "transaction routes update", call: func() error {
			_, err := client.Transactions.TransactionRoutes.Update(ctx, "org-1", "ledger-1", "route-1", nil)
			return err
		}},
		{name: "operation routes create", call: func() error {
			_, err := client.Transactions.OperationRoutes.Create(ctx, "org-1", "ledger-1", nil)
			return err
		}},
		{name: "operation routes update", call: func() error {
			_, err := client.Transactions.OperationRoutes.Update(ctx, "org-1", "ledger-1", "route-1", nil)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call()
			require.Error(t, err)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
			assert.Contains(t, err.Error(), "input is required")
		})
	}
}

func TestFakeMidazTransactionsCreateRejectsMissingSend(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	tx, err := client.Transactions.Transactions.Create(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInput{})
	require.Error(t, err)
	assert.Nil(t, tx)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestFakeMidazTransactionsCreatePreservesParentTransactionID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})
	parentID := "txn-parent"

	tx, err := client.Transactions.Transactions.Create(ctx, "org-1", "ledger-1", &midaz.CreateTransactionInput{
		ParentTransactionID: &parentID,
		Send: &midaz.TransactionSend{
			Asset: "BRL",
			Value: "12.34",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, tx)
	require.NotNil(t, tx.ParentTransactionID)
	assert.Equal(t, parentID, *tx.ParentTransactionID)
}

func TestFakeMidazBalancesCreateRejectsMissingOrWrongScopeAccounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	foreignAlias := "@foreign"
	foreign, err := client.Onboarding.Accounts.Create(ctx, "org-2", "ledger-2", &midaz.CreateAccountInput{
		Name:      "Foreign",
		Type:      "asset",
		AssetCode: "USD",
		Alias:     &foreignAlias,
	})
	require.NoError(t, err)

	tests := []struct {
		name      string
		accountID string
	}{
		{name: "missing account", accountID: "acct-missing"},
		{name: "wrong scope account", accountID: foreign.ID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := client.Transactions.Balances.CreateForAccount(ctx, "org-1", "ledger-1", tt.accountID, &midaz.CreateBalanceInput{
				Key: "reserve",
			})
			require.Error(t, err)
			assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))
			assert.Contains(t, err.Error(), "Account")
		})
	}
}

func TestFakeMidazBalanceLookupValidationParity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	tests := []struct {
		name string
		call func() error
	}{
		{name: "list by alias missing org", call: func() error {
			_, err := client.Transactions.Balances.ListByAlias(ctx, "", "ledger-1", "@alias")
			return err
		}},
		{name: "list by alias missing ledger", call: func() error {
			_, err := client.Transactions.Balances.ListByAlias(ctx, "org-1", "", "@alias")
			return err
		}},
		{name: "list by alias missing alias", call: func() error {
			_, err := client.Transactions.Balances.ListByAlias(ctx, "org-1", "ledger-1", "")
			return err
		}},
		{name: "list by external missing code", call: func() error {
			_, err := client.Transactions.Balances.ListByExternalCode(ctx, "org-1", "ledger-1", "")
			return err
		}},
		{name: "list by external missing org", call: func() error {
			_, err := client.Transactions.Balances.ListByExternalCode(ctx, "", "ledger-1", "EXT")
			return err
		}},
		{name: "list by account missing account", call: func() error {
			_, err := client.Transactions.Balances.ListByAccountID(ctx, "org-1", "ledger-1", "")
			return err
		}},
		{name: "list by account missing ledger", call: func() error {
			_, err := client.Transactions.Balances.ListByAccountID(ctx, "org-1", "", "acct-1")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call()
			require.Error(t, err)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
		})
	}
}

func TestFakeMidazTransactionStateValidationParity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	validInflow := &midaz.CreateTransactionInflowInput{
		Send: midaz.TransactionInflowSend{
			Asset:      "BRL",
			Value:      "10.00",
			Distribute: midaz.TransactionInflowDistribution{To: []midaz.TransactionOperationLeg{{AccountAlias: "@target"}}},
		},
	}
	validOutflow := &midaz.CreateTransactionOutflowInput{
		Send: midaz.TransactionOutflowSend{
			Asset:  "BRL",
			Value:  "10.00",
			Source: midaz.TransactionOutflowSource{From: []midaz.TransactionOperationLeg{{AccountAlias: "@source"}}},
		},
	}
	validCreate := &midaz.CreateTransactionInput{Send: &midaz.TransactionSend{Asset: "BRL", Value: "10.00"}}

	tests := []struct {
		name string
		call func() error
	}{
		{name: "create missing org", call: func() error {
			_, err := client.Transactions.Transactions.Create(ctx, "", "ledger-1", validCreate)
			return err
		}},
		{name: "create annotation missing ledger", call: func() error {
			_, err := client.Transactions.Transactions.CreateAnnotation(ctx, "org-1", "", validCreate)
			return err
		}},
		{name: "create dsl missing org", call: func() error {
			_, err := client.Transactions.Transactions.CreateDSL(ctx, "", "ledger-1", []byte("SEND 10 BRL"))
			return err
		}},
		{name: "create inflow missing ledger", call: func() error {
			_, err := client.Transactions.Transactions.CreateInflow(ctx, "org-1", "", validInflow)
			return err
		}},
		{name: "create outflow missing org", call: func() error {
			_, err := client.Transactions.Transactions.CreateOutflow(ctx, "", "ledger-1", validOutflow)
			return err
		}},
		{name: "get missing transaction id", call: func() error { _, err := client.Transactions.Transactions.Get(ctx, "org-1", "ledger-1", ""); return err }},
		{name: "update missing org", call: func() error {
			_, err := client.Transactions.Transactions.Update(ctx, "", "ledger-1", "tx-1", &midaz.UpdateTransactionInput{})
			return err
		}},
		{name: "commit missing ledger", call: func() error { _, err := client.Transactions.Transactions.Commit(ctx, "org-1", "", "tx-1"); return err }},
		{name: "cancel missing transaction id", call: func() error {
			_, err := client.Transactions.Transactions.Cancel(ctx, "org-1", "ledger-1", "")
			return err
		}},
		{name: "revert missing org", call: func() error {
			_, err := client.Transactions.Transactions.Revert(ctx, "", "ledger-1", "tx-1")
			return err
		}},
		{name: "operations update missing transaction id", call: func() error {
			_, err := client.Transactions.Operations.Update(ctx, "org-1", "ledger-1", "", "op-1", &midaz.UpdateOperationInput{})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call()
			require.Error(t, err)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
		})
	}
}

func TestFakeMidazOperationsUpdateChecksTransactionScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	description := "original"
	client := NewFakeClient(
		WithSeedOperations(midaz.Operation{
			ID:             "op-1",
			OrganizationID: "org-1",
			LedgerID:       "ledger-1",
			TransactionID:  "tx-actual",
			Description:    &description,
		}),
	)

	_, err := client.Midaz.Transactions.Operations.Update(ctx, "org-1", "ledger-1", "wrong-tx", "op-1", &midaz.UpdateOperationInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))
}

func TestFakeMidazOperationsUpdateHappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	originalDescription := "original"
	client := NewFakeClient(
		WithSeedOperations(midaz.Operation{
			ID:             "op-1",
			OrganizationID: "org-1",
			LedgerID:       "ledger-1",
			TransactionID:  "tx-1",
			Description:    &originalDescription,
		}),
	)

	updatedDescription := "updated"
	updated, err := client.Midaz.Transactions.Operations.Update(ctx, "org-1", "ledger-1", "tx-1", "op-1", &midaz.UpdateOperationInput{
		Description: &updatedDescription,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "op-1", updated.ID)
	require.NotNil(t, updated.Description)
	assert.Equal(t, updatedDescription, *updated.Description)
}

func TestFakeMidazOperationsGetChecksOrgAndLedgerScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := NewFakeClient(
		WithSeedOperations(midaz.Operation{
			ID:             "op-1",
			OrganizationID: "org-1",
			LedgerID:       "ledger-1",
			TransactionID:  "tx-1",
		}),
	)

	_, err := client.Midaz.Transactions.Operations.Get(ctx, "org-2", "ledger-1", "op-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))

	_, err = client.Midaz.Transactions.Operations.Get(ctx, "org-1", "ledger-2", "op-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))
}

func TestFakeMidazOperationsListRespectsOrgAndLedgerScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := NewFakeClient(
		WithSeedOperations(
			midaz.Operation{ID: "op-1", OrganizationID: "org-1", LedgerID: "ledger-1", TransactionID: "tx-1", AccountID: "acct-1"},
			midaz.Operation{ID: "op-2", OrganizationID: "org-2", LedgerID: "ledger-1", TransactionID: "tx-1", AccountID: "acct-1"},
		),
	)

	items, err := client.Midaz.Transactions.Operations.List(ctx, "org-1", "ledger-1", nil).Collect(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "op-1", items[0].ID)
}

func TestFakeMidazHoldersSoftDeleteAndIncludeDeleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	holder, err := client.CRM.Holders.Create(ctx, "org-1", &midaz.CreateHolderInput{Name: "Alice", Type: "individual", Document: "123"})
	require.NoError(t, err)

	err = client.CRM.Holders.Delete(ctx, "org-1", holder.ID, nil)
	require.NoError(t, err)

	_, err = client.CRM.Holders.Get(ctx, "org-1", holder.ID, nil)
	require.Error(t, err)

	got, err := client.CRM.Holders.Get(ctx, "org-1", holder.ID, &midaz.CRMGetOptions{IncludeDeleted: true})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.DeletedAt)
}

func TestFakeMidazAliasesDeleteRelatedParty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	holder, err := client.CRM.Holders.Create(ctx, "org-1", &midaz.CreateHolderInput{Name: "Alice", Type: "individual", Document: "123"})
	require.NoError(t, err)

	alias, err := client.CRM.Aliases.Create(ctx, "org-1", holder.ID, &midaz.CreateAliasInput{
		LedgerID:  "ledger-1",
		AccountID: "acc-1",
		RelatedParties: []midaz.RelatedParty{{
			ID:        "rp-1",
			Name:      "Bob",
			Document:  "999",
			Role:      "owner",
			StartDate: "2024-01-01",
		}},
	})
	require.NoError(t, err)

	err = client.CRM.Aliases.DeleteRelatedParty(ctx, "org-1", holder.ID, alias.ID, "rp-1")
	require.NoError(t, err)

	updated, err := client.CRM.Aliases.Get(ctx, "org-1", holder.ID, alias.ID, nil)
	require.NoError(t, err)
	assert.Empty(t, updated.RelatedParties)
}

func TestFakeMidazAliasesUpdateCanClearRelatedParties(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	holder, err := client.CRM.Holders.Create(ctx, "org-1", &midaz.CreateHolderInput{Name: "Alice", Type: "individual", Document: "123"})
	require.NoError(t, err)

	alias, err := client.CRM.Aliases.Create(ctx, "org-1", holder.ID, &midaz.CreateAliasInput{
		LedgerID:  "ledger-1",
		AccountID: "acc-1",
		RelatedParties: []midaz.RelatedParty{{
			ID:        "rp-1",
			Name:      "Bob",
			Document:  "999",
			Role:      "owner",
			StartDate: "2024-01-01",
		}},
	})
	require.NoError(t, err)

	cleared := []midaz.RelatedParty{}
	updated, err := client.CRM.Aliases.Update(ctx, "org-1", holder.ID, alias.ID, &midaz.UpdateAliasInput{RelatedParties: &cleared})
	require.NoError(t, err)
	assert.Empty(t, updated.RelatedParties)
}

func TestFakeMidazCRMIdentifierValidationAndParentChecks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	tests := []struct {
		name string
		call func() error
	}{
		{name: "holders get missing org", call: func() error { _, err := client.CRM.Holders.Get(ctx, "", "holder-1", nil); return err }},
		{name: "holders get missing id", call: func() error { _, err := client.CRM.Holders.Get(ctx, "org-1", "", nil); return err }},
		{name: "holders update missing id", call: func() error {
			_, err := client.CRM.Holders.Update(ctx, "org-1", "", &midaz.UpdateHolderInput{})
			return err
		}},
		{name: "aliases create missing holder id", call: func() error {
			_, err := client.CRM.Aliases.Create(ctx, "org-1", "", &midaz.CreateAliasInput{LedgerID: "ledger-1", AccountID: "acc-1"})
			return err
		}},
		{name: "aliases get missing alias id", call: func() error { _, err := client.CRM.Aliases.Get(ctx, "org-1", "holder-1", "", nil); return err }},
		{name: "aliases delete related party missing related party id", call: func() error { return client.CRM.Aliases.DeleteRelatedParty(ctx, "org-1", "holder-1", "alias-1", "") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call()
			require.Error(t, err)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
		})
	}

	_, err := client.CRM.Aliases.Create(ctx, "org-1", "holder-missing", &midaz.CreateAliasInput{LedgerID: "ledger-1", AccountID: "acc-1"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))
	assert.Contains(t, err.Error(), "Holder")

	foreignHolder, err := client.CRM.Holders.Create(ctx, "org-2", &midaz.CreateHolderInput{Name: "Foreign", Type: "individual", Document: "999"})
	require.NoError(t, err)

	_, err = client.CRM.Aliases.Create(ctx, "org-1", foreignHolder.ID, &midaz.CreateAliasInput{LedgerID: "ledger-1", AccountID: "acc-1"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound))
}

func TestFakeMidazAccountsCountScoped(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	_, err := client.Onboarding.Accounts.Create(ctx, "org-1", "ledger-1", &midaz.CreateAccountInput{Name: "A1", Type: "asset", AssetCode: "BRL"})
	require.NoError(t, err)
	_, err = client.Onboarding.Accounts.Create(ctx, "org-2", "ledger-2", &midaz.CreateAccountInput{Name: "A2", Type: "asset", AssetCode: "BRL"})
	require.NoError(t, err)

	count, err := client.Onboarding.Accounts.Count(ctx, "org-1", "ledger-1")
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestFakeMidazAssetRatesCreateUpsertsByPair(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newFakeMidazClient(&fakeConfig{errorInjections: make(map[string]error)})

	rateA, err := client.Transactions.AssetRates.Create(ctx, "org-1", "ledger-1", &midaz.CreateAssetRateInput{BaseAssetCode: "BRL", CounterAssetCode: "USD", Amount: 100, Scale: 2})
	require.NoError(t, err)
	rateB, err := client.Transactions.AssetRates.Create(ctx, "org-1", "ledger-1", &midaz.CreateAssetRateInput{BaseAssetCode: "BRL", CounterAssetCode: "USD", Amount: 200, Scale: 2})
	require.NoError(t, err)
	assert.Equal(t, rateA.ID, rateB.ID)
	assert.Equal(t, int64(200), rateB.Amount)
}
