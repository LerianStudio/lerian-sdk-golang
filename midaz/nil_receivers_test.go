package midaz

import (
	"context"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtendedServicesRejectNilReceivers(t *testing.T) {
	t.Parallel()

	t.Run("balances create for account", func(t *testing.T) {
		t.Parallel()

		var svc *balancesService

		balance, err := svc.CreateForAccount(context.Background(), "org-1", "ledger-1", "account-1", &CreateBalanceInput{Key: "main"})
		require.Error(t, err)
		assert.Nil(t, balance)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("balances lookup helper", func(t *testing.T) {
		t.Parallel()

		var svc *balancesService

		balances, err := svc.ListByAlias(context.Background(), "org-1", "ledger-1", "alias-1")
		require.Error(t, err)
		assert.Nil(t, balances)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("transactions annotation", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.CreateAnnotation(context.Background(), "org-1", "ledger-1", &CreateTransactionInput{})
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("transactions create", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.Create(context.Background(), "org-1", "ledger-1", &CreateTransactionInput{})
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("transactions dsl", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.CreateDSL(context.Background(), "org-1", "ledger-1", []byte("dsl"))
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("transactions inflow", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.CreateInflow(context.Background(), "org-1", "ledger-1", &CreateTransactionInflowInput{})
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("transactions outflow", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.CreateOutflow(context.Background(), "org-1", "ledger-1", &CreateTransactionOutflowInput{})
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("operations update", func(t *testing.T) {
		t.Parallel()

		var svc *operationsService

		op, err := svc.Update(context.Background(), "org-1", "ledger-1", "tx-1", "op-1", &UpdateOperationInput{})
		require.Error(t, err)
		assert.Nil(t, op)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("accounts external code", func(t *testing.T) {
		t.Parallel()

		var svc *accountsService

		account, err := svc.GetByExternalCode(context.Background(), "org-1", "ledger-1", "ext-1")
		require.Error(t, err)
		assert.Nil(t, account)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("asset rates external id", func(t *testing.T) {
		t.Parallel()

		var svc *assetRatesService

		rate, err := svc.GetByExternalID(context.Background(), "org-1", "ledger-1", "ext-1")
		require.Error(t, err)
		assert.Nil(t, rate)
		assert.ErrorIs(t, err, core.ErrNilService)
	})
}

func TestMetricsServicesRejectNilReceivers(t *testing.T) {
	t.Parallel()

	t.Run("organizations count", func(t *testing.T) {
		t.Parallel()

		var svc *organizationsService

		count, err := svc.Count(context.Background())
		require.Error(t, err)
		assert.Zero(t, count)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("ledgers count", func(t *testing.T) {
		t.Parallel()

		var svc *ledgersService

		count, err := svc.Count(context.Background(), "org-1")
		require.Error(t, err)
		assert.Zero(t, count)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("accounts count", func(t *testing.T) {
		t.Parallel()

		var svc *accountsService

		count, err := svc.Count(context.Background(), "org-1", "ledger-1")
		require.Error(t, err)
		assert.Zero(t, count)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("assets count", func(t *testing.T) {
		t.Parallel()

		var svc *assetsService

		count, err := svc.Count(context.Background(), "org-1", "ledger-1")
		require.Error(t, err)
		assert.Zero(t, count)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("portfolios count", func(t *testing.T) {
		t.Parallel()

		var svc *portfoliosService

		count, err := svc.Count(context.Background(), "org-1", "ledger-1")
		require.Error(t, err)
		assert.Zero(t, count)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("segments count", func(t *testing.T) {
		t.Parallel()

		var svc *segmentsService

		count, err := svc.Count(context.Background(), "org-1", "ledger-1")
		require.Error(t, err)
		assert.Zero(t, count)
		assert.ErrorIs(t, err, core.ErrNilService)
	})
}

func TestListServicesRejectNilReceivers(t *testing.T) {
	t.Parallel()

	t.Run("organizations list", func(t *testing.T) {
		t.Parallel()

		var svc *organizationsService

		items, err := svc.List(context.Background(), nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("ledgers list", func(t *testing.T) {
		t.Parallel()

		var svc *ledgersService

		items, err := svc.List(context.Background(), "org-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("account types list", func(t *testing.T) {
		t.Parallel()

		var svc *accountTypesService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("accounts list", func(t *testing.T) {
		t.Parallel()

		var svc *accountsService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("assets list", func(t *testing.T) {
		t.Parallel()

		var svc *assetsService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("asset rates list", func(t *testing.T) {
		t.Parallel()

		var svc *assetRatesService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("balances list", func(t *testing.T) {
		t.Parallel()

		var svc *balancesService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("operations list", func(t *testing.T) {
		t.Parallel()

		var svc *operationsService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("operations list by transaction", func(t *testing.T) {
		t.Parallel()

		var svc *operationsService

		items, err := svc.ListByTransaction(context.Background(), "org-1", "ledger-1", "tx-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("operations list by account", func(t *testing.T) {
		t.Parallel()

		var svc *operationsService

		items, err := svc.ListByAccount(context.Background(), "org-1", "ledger-1", "acc-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("portfolios list", func(t *testing.T) {
		t.Parallel()

		var svc *portfoliosService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("segments list", func(t *testing.T) {
		t.Parallel()

		var svc *segmentsService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("transactions list", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("transaction routes list", func(t *testing.T) {
		t.Parallel()

		var svc *transactionRoutesService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("operation routes list", func(t *testing.T) {
		t.Parallel()

		var svc *operationRoutesService

		items, err := svc.List(context.Background(), "org-1", "ledger-1", nil).Collect(context.Background())
		require.Error(t, err)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, core.ErrNilService)
	})
}

func TestTransactionsLifecycleRejectNilReceivers(t *testing.T) {
	t.Parallel()

	t.Run("get", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.Get(context.Background(), "org-1", "ledger-1", "txn-1")
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.Update(context.Background(), "org-1", "ledger-1", "txn-1", &UpdateTransactionInput{})
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("commit", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.Commit(context.Background(), "org-1", "ledger-1", "txn-1")
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("cancel", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.Cancel(context.Background(), "org-1", "ledger-1", "txn-1")
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})

	t.Run("revert", func(t *testing.T) {
		t.Parallel()

		var svc *transactionsService

		tx, err := svc.Revert(context.Background(), "org-1", "ledger-1", "txn-1")
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, core.ErrNilService)
	})
}
