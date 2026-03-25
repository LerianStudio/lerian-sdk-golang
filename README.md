# Lerian Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/LerianStudio/lerian-sdk-golang.svg)](https://pkg.go.dev/github.com/LerianStudio/lerian-sdk-golang)
[![CI](https://github.com/LerianStudio/lerian-sdk-golang/actions/workflows/go-combined-analysis.yml/badge.svg)](https://github.com/LerianStudio/lerian-sdk-golang/actions/workflows/go-combined-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/LerianStudio/lerian-sdk-golang)](https://goreportcard.com/report/github.com/LerianStudio/lerian-sdk-golang)
[![License: Elastic-2.0](https://img.shields.io/badge/License-Elastic--2.0-blue.svg)](LICENSE.md)

The official Go SDK for the **Lerian** financial infrastructure platform. This SDK provides a unified, type-safe client for all Lerian products -- from core ledger operations to transaction matching, tracing, reporting, and fee management. Built for production use in financial systems with enterprise-grade observability, retry logic, and error handling.

## Installation

```bash
go get github.com/LerianStudio/lerian-sdk-golang
```

Requires **Go 1.24** or later.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    lerian "github.com/LerianStudio/lerian-sdk-golang"
    "github.com/LerianStudio/lerian-sdk-golang/midaz"
    "github.com/LerianStudio/lerian-sdk-golang/models"
)

func main() {
    ctx := context.Background()

    // Create a client with the Midaz product enabled.
    client, err := lerian.New(lerian.Config{
        Midaz: &midaz.Config{
            OnboardingURL:  "http://localhost:3000/v1",
            TransactionURL: "http://localhost:3001/v1",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Shutdown(ctx)

    // Create an organization
    org, err := client.Midaz.Onboarding.Organizations.Create(ctx, &midaz.CreateOrganizationInput{
        LegalName:     "Acme Corp",
        LegalDocument: "12345678000100",
        Status:        &models.Status{Code: "ACTIVE"},
        Address: &models.Address{
            Country: "BR",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created organization: %s (%s)\n", org.LegalName, org.ID)
}
```

## Products

The SDK provides access to the full suite of Lerian products through a single client:

| Product | Services | Description |
|---------|----------|-------------|
| **Midaz** | 13 | Core ledger -- organizations, ledgers, accounts, portfolios, assets, segments, transactions, operations, balances, and more |
| **Matcher** | 14 | Transaction matching engine -- rules, reconciliation pipelines, match results, and manual overrides |
| **Tracer** | 4 | Transaction tracing and audit trail -- trace queries, event streams, and validation |
| **Reporter** | 3 | Financial reporting and analytics -- report generation, templates, and scheduling |
| **Fees** | 3 | Fee management -- fee rules, estimation, and billing |

Enable only the products you need:

```go
client, err := lerian.New(lerian.Config{
    Midaz: &midaz.Config{
        OnboardingURL:  "http://localhost:3000/v1",
        TransactionURL: "http://localhost:3001/v1",
    },
    Matcher: &matcher.Config{BaseURL: "http://localhost:3002/v1"},
    Tracer:  &tracer.Config{BaseURL: "http://localhost:3003/v1"},
    Reporter: &reporter.Config{
        BaseURL:        "http://localhost:3004/v1",
        OrganizationID: "org-uuid",
    },
    Fees: &fees.Config{
        BaseURL:        "http://localhost:3005/v1",
        OrganizationID: "org-uuid",
    },
})
```

Products with nil config remain nil on the client. This keeps resource usage minimal -- only backends you configure are initialized.
When authentication is required, populate each product config explicitly or load it from the matching `LERIAN_*_CLIENT_ID`, `LERIAN_*_CLIENT_SECRET`, and `LERIAN_*_TOKEN_URL` environment variables.

## Configuration

### Environment Variables

The SDK supports `LERIAN_*` environment variables through `lerian.LoadConfigFromEnv()`. Build the config from env, then pass that config into `New()`.

```bash
# Global
LERIAN_DEBUG=false

# Midaz (Ledger)
LERIAN_MIDAZ_ONBOARDING_URL=http://localhost:3000/v1
LERIAN_MIDAZ_TRANSACTION_URL=http://localhost:3001/v1
LERIAN_MIDAZ_CLIENT_ID=my-client-id
LERIAN_MIDAZ_CLIENT_SECRET=my-client-secret
LERIAN_MIDAZ_TOKEN_URL=https://auth.example.com/token

# Matcher (Rule Engine)
LERIAN_MATCHER_URL=http://localhost:3002/v1
LERIAN_MATCHER_CLIENT_ID=my-client-id
LERIAN_MATCHER_CLIENT_SECRET=my-client-secret
LERIAN_MATCHER_TOKEN_URL=https://auth.example.com/token

# Tracer (Audit Trail)
LERIAN_TRACER_URL=http://localhost:3003/v1
LERIAN_TRACER_CLIENT_ID=my-client-id
LERIAN_TRACER_CLIENT_SECRET=my-client-secret
LERIAN_TRACER_TOKEN_URL=https://auth.example.com/token

# Reporter (Analytics)
LERIAN_REPORTER_URL=http://localhost:3004/v1
LERIAN_REPORTER_ORG_ID=org-uuid
LERIAN_REPORTER_CLIENT_ID=my-client-id
LERIAN_REPORTER_CLIENT_SECRET=my-client-secret
LERIAN_REPORTER_TOKEN_URL=https://auth.example.com/token

# Fees (Billing)
LERIAN_FEES_URL=http://localhost:3005/v1
LERIAN_FEES_ORG_ID=org-uuid
LERIAN_FEES_CLIENT_ID=my-client-id
LERIAN_FEES_CLIENT_SECRET=my-client-secret
LERIAN_FEES_TOKEN_URL=https://auth.example.com/token
```

### Explicit Root Config

Every aspect of the root client is configured through `lerian.Config`:

```go
retryCfg := retry.DefaultConfig()
retryCfg.MaxRetries = 3
retryCfg.BaseDelay = 500 * time.Millisecond

client, err := lerian.New(lerian.Config{
    HTTPClient:  customHTTPClient,
    RetryConfig: &retryCfg,
    Observability: lerian.ObservabilityConfig{
        Traces:            true,
        Metrics:           true,
        CollectorEndpoint: "http://localhost:4318",
    },
    Midaz: &midaz.Config{
        OnboardingURL:  "http://localhost:3000/v1",
        TransactionURL: "http://localhost:3001/v1",
    },
})
```

Or load from env first:

```go
cfg := lerian.LoadConfigFromEnv()
client, err := lerian.New(cfg)
```

## Error Handling

The SDK uses a hierarchical error system. Use `errors.Is()` to match errors by category, regardless of which product returned them:

```go
import "errors"

_, err := client.Midaz.Onboarding.Accounts.Get(ctx, orgID, ledgerID, accountID)
if err != nil {
    switch {
    case errors.Is(err, lerian.ErrNotFound):
        // Handle not-found for any product
        fmt.Println("Resource not found")

    case errors.Is(err, lerian.ErrAuthentication):
        // Handle auth errors
        fmt.Println("Check your credentials")

    case errors.Is(err, lerian.ErrRateLimit):
        // Back off and retry
        fmt.Println("Rate limited, retrying...")

    case errors.Is(err, lerian.ErrValidation):
        // Inspect validation details
        var sdkErr *lerian.Error
        if errors.As(err, &sdkErr) {
            fmt.Printf("Validation: %s (code: %s)\n", sdkErr.Message, sdkErr.Code)
        }

    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

Available sentinel errors: `ErrValidation`, `ErrNotFound`, `ErrAuthentication`, `ErrAuthorization`, `ErrConflict`, `ErrRateLimit`, `ErrNetwork`, `ErrTimeout`, `ErrCancellation`, `ErrInternal`.

Each product also defines product-specific error codes (e.g., `midaz.ErrAccountNotFound`) that chain to the category sentinels, so both specific and broad matching work with `errors.Is()`.

## Iterator Pattern

List operations return an `Iterator[T]` that supports lazy, memory-efficient pagination. The iterator implements Go 1.23+ range-over-function via `All()`:

```go
// Iterate lazily over all accounts (fetches pages on demand)
for account, err := range client.Midaz.Onboarding.Accounts.List(ctx, orgID, ledgerID).All() {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(account.Name)
}

// Collect all results into a slice
accounts, err := client.Midaz.Onboarding.Accounts.List(ctx, orgID, ledgerID).Collect()

// Collect up to N results
first10, err := client.Midaz.Onboarding.Accounts.List(ctx, orgID, ledgerID).CollectN(10)

// Process concurrently with bounded parallelism
it := client.Midaz.Onboarding.Accounts.List(ctx, orgID, ledgerID)
errs := pagination.ForEachConcurrent(ctx, it, 8,
    func(ctx context.Context, account models.Account) error {
        return processAccount(ctx, account)
    },
)
```

## Testing

The SDK ships with a `leriantest` package that provides a complete fake client for all 37 services. No network calls, no mocking frameworks required:

```go
import (
    "testing"

    "github.com/LerianStudio/lerian-sdk-golang/midaz"
    "github.com/LerianStudio/lerian-sdk-golang/testing/leriantest"
    "github.com/stretchr/testify/assert"
)

func TestMyService(t *testing.T) {
    // Create a fake client pre-loaded with test data (no network, no mocks)
    fake := leriantest.NewFakeClient(
        leriantest.WithSeedOrganizations(midaz.Organization{
            ID:        "org-uuid",
            LegalName: "Test Corp",
        }),
    )

    // fake is a *lerian.Client -- pass it directly to code under test
    result, err := myFunction(ctx, fake)
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## Examples

The `examples/` directory contains runnable programs demonstrating each product:

| Example | Description |
|---------|-------------|
| [`midaz-workflow`](examples/midaz-workflow/) | End-to-end ledger workflow: org, ledger, accounts, transactions |
| [`matcher-reconciliation`](examples/matcher-reconciliation/) | Set up matching rules and run a reconciliation pipeline |
| [`tracer-validation`](examples/tracer-validation/) | Trace a transaction through the system |
| [`reporter-usage`](examples/reporter-usage/) | Generate and retrieve financial reports |
| [`fees-estimation`](examples/fees-estimation/) | Estimate and apply fees to transactions |
| [`multi-product`](examples/multi-product/) | Combine multiple products in a single workflow |

Run an example:

```bash
cd examples/midaz-workflow
go run .
```

## Contributing

We welcome contributions. Please follow these guidelines:

1. Fork the repository and create a feature branch.
2. Follow [Conventional Commits](https://www.conventionalcommits.org/): `feat(midaz): add balance caching`.
3. Ensure all checks pass: `make fmt lint test verify-sdk`.
4. Write tests for new code (target 80%+ coverage).
5. Open a pull request with a clear description of the change.

See [docs/PROJECT_RULES.md](docs/PROJECT_RULES.md) for architectural decisions, enforced patterns, and coding conventions.

## License

This project is licensed under the [Elastic License 2.0 (ELv2)](https://www.elastic.co/licensing/elastic-license). See [LICENSE.md](LICENSE.md) for the full text.

Copyright 2025-2026 [Lerian Studio](https://lerian.studio).
