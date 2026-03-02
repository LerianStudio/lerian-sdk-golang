package leriantest

import (
	lerian "github.com/LerianStudio/lerian-sdk-golang"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
)

// FakeOption configures the fake client.
type FakeOption func(*fakeConfig)

// fakeConfig holds configuration applied via FakeOption functions.
type fakeConfig struct {
	seedOrgs        []midaz.Organization
	seedLedgers     []midaz.Ledger
	seedAccounts    []midaz.Account
	errorInjections map[string]error // "midaz.Organizations.Create" -> error
}

// injectedError checks if an error was injected for the given operation key.
// Returns nil if no error is configured.
func (c *fakeConfig) injectedError(key string) error {
	if c.errorInjections == nil {
		return nil
	}

	return c.errorInjections[key]
}

// NewFakeClient creates a [lerian.Client] backed by in-memory fake services
// for all five products. No network access is required. All product fields
// (Midaz, Matcher, Tracer, Reporter, Fees) are populated and functional.
func NewFakeClient(opts ...FakeOption) *lerian.Client {
	cfg := &fakeConfig{
		errorInjections: make(map[string]error),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	client := &lerian.Client{}
	client.Midaz = newFakeMidazClient(cfg)
	client.Matcher = newFakeMatcherClient(cfg)
	client.Tracer = newFakeTracerClient(cfg)
	client.Reporter = newFakeReporterClient(cfg)
	client.Fees = newFakeFeesClient(cfg)

	// Apply seed data.
	applySeedData(client, cfg)

	return client
}

// applySeedData pre-populates the fake stores with any seed data provided
// via options. Seeded items are stored as-is, preserving whatever IDs and
// fields the caller supplied.
func applySeedData(client *lerian.Client, cfg *fakeConfig) {
	// Seed organizations.
	if orgs := client.Midaz.Organizations.(*fakeOrganizations); orgs != nil {
		for _, org := range cfg.seedOrgs {
			if org.ID == "" {
				org.ID = generateID("org")
			}

			orgs.store.Set(org.ID, org)
		}
	}

	// Seed ledgers.
	if ledgers := client.Midaz.Ledgers.(*fakeLedgers); ledgers != nil {
		for _, l := range cfg.seedLedgers {
			if l.ID == "" {
				l.ID = generateID("ledger")
			}

			ledgers.store.Set(l.ID, l)
		}
	}

	// Seed accounts.
	if accounts := client.Midaz.Accounts.(*fakeAccounts); accounts != nil {
		for _, a := range cfg.seedAccounts {
			if a.ID == "" {
				a.ID = generateID("acct")
			}

			accounts.store.Set(a.ID, a)
		}
	}
}

// ---------------------------------------------------------------------------
// FakeOption constructors
// ---------------------------------------------------------------------------

// WithSeedOrganizations pre-populates the Midaz Organizations store with
// the given organizations. If an organization's ID is empty, a unique ID
// is generated automatically.
func WithSeedOrganizations(orgs ...midaz.Organization) FakeOption {
	return func(cfg *fakeConfig) {
		cfg.seedOrgs = append(cfg.seedOrgs, orgs...)
	}
}

// WithSeedLedgers pre-populates the Midaz Ledgers store with the given
// ledgers. If a ledger's ID is empty, a unique ID is generated automatically.
func WithSeedLedgers(ledgers ...midaz.Ledger) FakeOption {
	return func(cfg *fakeConfig) {
		cfg.seedLedgers = append(cfg.seedLedgers, ledgers...)
	}
}

// WithSeedAccounts pre-populates the Midaz Accounts store with the given
// accounts. If an account's ID is empty, a unique ID is generated
// automatically.
func WithSeedAccounts(accts ...midaz.Account) FakeOption {
	return func(cfg *fakeConfig) {
		cfg.seedAccounts = append(cfg.seedAccounts, accts...)
	}
}

// WithErrorOn injects an error that will be returned when the specified
// operation key is called. The key format is "product.Service.Method",
// for example "midaz.Organizations.Create" or "fees.Packages.Create".
func WithErrorOn(key string, err error) FakeOption {
	return func(cfg *fakeConfig) {
		cfg.errorInjections[key] = err
	}
}
