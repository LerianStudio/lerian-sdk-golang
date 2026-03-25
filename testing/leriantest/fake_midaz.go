package leriantest

import (
	"strconv"
	"strings"

	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

func validateRequiredField(operation, resource, value, field string) error {
	if strings.TrimSpace(value) == "" {
		return sdkerrors.NewValidation(operation, resource, field+" is required")
	}

	return nil
}

func validateTransactionVariantLegsFake(operation, field string, legs []midaz.TransactionOperationLeg) error {
	if len(legs) == 0 {
		return sdkerrors.NewValidation(operation, "Transaction", field+" legs are required")
	}

	for _, leg := range legs {
		if strings.TrimSpace(leg.AccountAlias) == "" && strings.TrimSpace(leg.BalanceKey) == "" {
			return sdkerrors.NewValidation(operation, "Transaction", field+" leg identifier is required")
		}
	}

	return nil
}

func balanceInScope(balance midaz.Balance, orgID, ledgerID string) bool {
	return balance.OrganizationID == orgID && balance.LedgerID == ledgerID
}

func transactionInScope(tx midaz.Transaction, orgID, ledgerID string) bool {
	return tx.OrganizationID == orgID && tx.LedgerID == ledgerID
}

func parseScaledValue(operation, resource, raw string) (int64, int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, 0, sdkerrors.NewValidation(operation, resource, "send value is required")
	}

	negative := false
	if strings.HasPrefix(trimmed, "-") {
		negative = true
		trimmed = strings.TrimPrefix(trimmed, "-")
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 {
		return 0, 0, sdkerrors.NewValidation(operation, resource, "send value must be a valid decimal string")
	}

	whole := parts[0]
	if whole == "" {
		whole = "0"
	}

	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}

	combined := whole + frac
	if combined == "" {
		combined = "0"
	}

	amount, err := strconv.ParseInt(combined, 10, 64)
	if err != nil {
		return 0, 0, sdkerrors.NewValidation(operation, resource, "send value must be a valid decimal string")
	}

	if negative {
		amount = -amount
	}

	return amount, len(frac), nil
}

// newFakeMidazClient constructs a [midaz.Client] with all service fields
// backed by in-memory fakes.
func newFakeMidazClient(cfg *fakeConfig) *midaz.Client {
	organizations := newFakeOrganizations(cfg)
	ledgers := newFakeLedgers(cfg)
	accounts := newFakeAccounts(cfg)
	assets := newFakeAssets(cfg)
	assetRates := newFakeAssetRates(cfg)
	balances := newFakeBalances(cfg, accounts.store)
	portfolios := newFakePortfolios(cfg)
	segments := newFakeSegments(cfg)
	transactions := newFakeTransactions(cfg)
	operations := newFakeOperations(cfg)
	transactionRoutes := newFakeTransactionRoutes(cfg)
	operationRoutes := newFakeOperationRoutes(cfg)
	holders := newFakeHolders(cfg)

	return &midaz.Client{
		Onboarding: &midaz.OnboardingClient{
			Organizations: organizations,
			Ledgers:       ledgers,
			Accounts:      accounts,
			AccountTypes:  newFakeAccountTypes(cfg),
			Assets:        assets,
			Portfolios:    portfolios,
			Segments:      segments,
		},
		Transactions: &midaz.TransactionsClient{
			AssetRates:        assetRates,
			Balances:          balances,
			Transactions:      transactions,
			TransactionRoutes: transactionRoutes,
			Operations:        operations,
			OperationRoutes:   operationRoutes,
		},
		CRM: &midaz.CRMClient{
			Holders: holders,
			Aliases: newFakeAliases(cfg, holders.store, holders.orgByID),
		},
	}
}
