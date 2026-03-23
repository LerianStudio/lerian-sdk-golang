# Project Rules & Architectural Decisions

> Single source of truth for architectural decisions, enforced patterns, and project conventions
> for the Lerian Go SDK. Last updated: 2026-03-23.

## Table of Contents

- [ADR-001: Layered Hexagonal Plugin Architecture](#adr-001-layered-hexagonal-plugin-architecture)
- [ADR-002: Deferred Functional Options Pattern](#adr-002-deferred-functional-options-pattern)
- [ADR-003: Generic CRUD via BaseService](#adr-003-generic-crud-via-baseservice)
- [ADR-004: Hierarchical Error System](#adr-004-hierarchical-error-system)
- [ADR-005: Fake Clients Over Mocks](#adr-005-fake-clients-over-mocks)
- [ADR-006: Lazy Iterator Pagination](#adr-006-lazy-iterator-pagination)
- [ADR-007: Zero Cross-Product Dependencies](#adr-007-zero-cross-product-dependencies)
- [ADR-008: OAuth2 Client-Credentials Authentication](#adr-008-oauth2-client-credentials-authentication)
- [ADR-009: Environment Variable Precedence](#adr-009-environment-variable-precedence)
- [ADR-010: OpenTelemetry Observability](#adr-010-opentelemetry-observability)
- [Enforced Code Patterns](#enforced-code-patterns)
- [Linter Configuration](#linter-configuration)
- [Security Rules](#security-rules)
- [Testing Standards](#testing-standards)
- [Dependency Policy](#dependency-policy)

---

## ADR-001: Layered Hexagonal Plugin Architecture

**Status:** Accepted
**Date:** 2025-01
**Context:** The SDK needs to support multiple independent financial products through a single client while keeping infrastructure concerns separated from business logic.

**Decision:** Four-layer architecture:
1. **Root Orchestrator** (`client.go`) -- Single entry point, owns lifecycle
2. **Product Clients** (`midaz/`, `matcher/`, `tracer/`, `reporter/`, `fees/`) -- Domain-specific APIs
3. **Service Layer** (37+ services) -- CRUD operations per resource
4. **Infrastructure** (`pkg/`) -- Transport, auth, retry, observability, errors

**Consequences:**
- Products can evolve independently
- New products added without touching existing code
- Infrastructure shared via well-defined interfaces

---

## ADR-002: Deferred Functional Options Pattern

**Status:** Accepted
**Date:** 2025-01
**Context:** Product options (like OAuth2 credentials) depend on infrastructure being initialized first (backends, observability), but users configure everything in a single `New()` call.

**Decision:** Options are collected during `New()` but applied *after* all backends and infrastructure initialize. This makes configuration order-independent.

**Key files:** `client.go`, `options.go`

**Consequences:**
- Users can pass `WithMidaz()` and `WithObservability()` in any order
- Environment variables work as fallbacks even when product options are explicit
- Validation happens after full initialization

---

## ADR-003: Generic CRUD via BaseService

**Status:** Accepted
**Date:** 2025-01
**Context:** All 37+ services follow the same REST patterns (Create, Get, Update, Delete, List, Action). Duplicating HTTP call logic would be error-prone.

**Decision:** Package-level generic functions in `pkg/core/service.go`:
- `Get[T any](ctx, svc, path) (*T, error)`
- `Create[T, I any](ctx, svc, path, input) (*T, error)`
- `Update[T, I any](ctx, svc, path, input) (*T, error)`
- `Delete(ctx, svc, path) error`
- `List[T any](ctx, svc, path, opts) *Iterator[T]`
- `Action[T any](ctx, svc, path, input) (*T, error)`

All services embed `BaseService` and delegate to these functions.

**Consequences:**
- New services are ~50 lines of boilerplate
- Transport, retry, and observability applied uniformly
- Type safety via Go generics (no `interface{}` casting)

---

## ADR-004: Hierarchical Error System

**Status:** Accepted
**Date:** 2025-01
**Context:** Financial systems need precise error classification for programmatic handling across multiple products.

**Decision:** Structured `Error` type in `pkg/errors/` with:
- **Categories:** `validation`, `not_found`, `authentication`, `authorization`, `conflict`, `rate_limit`, `network`, `timeout`, `cancellation`, `internal`
- **Product codes:** Optional product-specific sub-codes (e.g., Midaz `0029` = InsufficientBalance)
- **`errors.Is()` matching:** Category-first, then optionally by code -- two SDK errors match when they share the same `Category`; if the target also specifies a `Code`, codes must match as well
- **Sentinel errors:** `ErrValidation`, `ErrNotFound`, `ErrAuthentication`, `ErrAuthorization`, `ErrConflict`, `ErrRateLimit`, `ErrNetwork`, `ErrTimeout`, `ErrCancellation`, `ErrInternal`
- **Factory functions:** `NewValidation()`, `NewNotFound()`, `NewAuthentication()`, `NewAuthorization()`, `NewConflict()`, `NewNetwork()`, `NewTimeout()`, `NewCancellation()`, `NewInternal()`
- **HTTP status mapping:** `CategoryFromStatus()` maps 400->Validation, 401->Authentication, 403->Authorization, 404->NotFound, 409->Conflict, 429->RateLimit, 500+->Internal
- **Body truncation:** `MaxErrorBodyBytes = 512` -- response bodies exceeding this limit are truncated with `"... [truncated]"` suffix

**Rule:** Product code MUST NOT use `errors.New()` -- use `sdkerrors` factories (enforced by forbidigo linter).

**Consequences:**
- Callers can match on broad categories or specific codes
- Standard `errors.Is()` / `errors.As()` chain works
- Error wrapping always uses `%w` (enforced by errorlint)

---

## ADR-005: Fake Clients Over Mocks

**Status:** Accepted
**Date:** 2025-02
**Context:** Mock-based testing (gomock, mockgen) creates brittle tests coupled to implementation details and doesn't test real service logic.

**Decision:** `testing/leriantest/` provides `NewFakeClient()` with in-memory stores for all 37 services. No network, no mocking framework.

**Key features:**
- `WithSeed*()` options to pre-populate data
- `WithErrorOn(key, err)` for error injection (`"product.Service.Method"` format)
- Generic `fakeStore[T]` with concurrent-safe CRUD + pagination
- Cursor-based pagination matching real API behavior

**Consequences:**
- Tests run in microseconds, no Docker/network needed
- Tests validate real business logic paths
- Refactoring internals doesn't break tests

---

## ADR-006: Lazy Iterator Pagination

**Status:** Accepted
**Date:** 2025-02
**Context:** List endpoints return paginated results. Fetching all pages upfront wastes memory and adds latency for consumers who only need the first few items.

**Decision:** `Iterator[T]` in `pkg/core/` fetches pages on-demand:
- `Next(ctx) bool` / `Item() T` -- standard iteration
- `All() iter.Seq2[T, error]` -- Go 1.23+ range-over-function
- `Collect()` / `CollectN(n)` -- convenience collectors
- `ForEachConcurrent(workers, fn)` -- parallel processing

**Consequences:**
- Only one page buffered in memory at a time
- Supports arbitrary pagination schemes via `PageFetcher[T]`
- Naturally integrates with Go 1.23+ range semantics

---

## ADR-007: Zero Cross-Product Dependencies

**Status:** Accepted
**Date:** 2025-01
**Context:** Products (Midaz, Matcher, Tracer, Reporter, Fees) are separate backend services that evolve independently.

**Decision:** No product package may import another product package. Products share only:
- `models/` -- Request/response types
- `pkg/` -- Infrastructure utilities

**Consequences:**
- Products can be versioned and released independently
- No circular dependency risk
- Clear ownership boundaries

---

## ADR-008: OAuth2 Client-Credentials Authentication

**Status:** Accepted
**Date:** 2025-03
**Context:** Lerian products use OAuth2 client-credentials flow for service-to-service auth.

**Decision:** `pkg/auth/` provides:
- `Authenticator` interface with single method: `Enrich(ctx, *http.Request) error`
- `OAuth2` implementation with automatic token caching and refresh
- `NoAuth` no-op for unauthenticated/development use
- 30-second expiry buffer (refresh before real expiration)
- Mutex-serialized refresh with concurrent reads
- Cross-host redirect protection (max 10 redirects)
- Credential redaction in `String()` and `MarshalJSON()`

**Consequences:**
- Thread-safe token management
- No accidental credential leaks in logs or serialization
- Pluggable -- custom auth providers implement `Authenticator`

---

## ADR-009: Environment Variable Precedence

**Status:** Accepted
**Date:** 2025-01
**Context:** SDK users need flexible configuration -- code-level options for libraries, env vars for containers/CI, and sensible defaults for development.

**Decision:** Three-tier precedence: **explicit Option > `LERIAN_*` env var > default value**.

**Env var pattern:** `LERIAN_{PRODUCT}_{SETTING}` (e.g., `LERIAN_MIDAZ_CLIENT_ID`)

**Global vars:**
- `LERIAN_DEBUG` -- Enable debug logging

**Per-product vars:**
- `LERIAN_{PRODUCT}_URL` -- Base URL
- `LERIAN_{PRODUCT}_CLIENT_ID` -- OAuth2 client ID
- `LERIAN_{PRODUCT}_CLIENT_SECRET` -- OAuth2 client secret
- `LERIAN_{PRODUCT}_TOKEN_URL` -- OAuth2 token endpoint

**Consequences:**
- Zero-config in containerized environments (just set env vars)
- Code-level options always win (no surprises)
- Defaults work for local development

---

## ADR-010: OpenTelemetry Observability

**Status:** Accepted
**Date:** 2025-02
**Context:** Financial transaction SDKs need distributed tracing and metrics for production debugging and SLA monitoring.

**Decision:** Optional OpenTelemetry integration via `pkg/observability/`:
- Traces, Metrics, Logs -- each independently toggleable
- OTLP collector endpoint (default: `http://localhost:4318`)
- `NoopProvider` when disabled (zero overhead)
- Backend instrumentation: span creation, error classification, operation attributes

**Configuration:** `WithObservability(traces, metrics, logs bool)` + `WithCollectorEndpoint(endpoint)`

**Consequences:**
- Zero overhead when disabled
- Full distributed tracing when enabled
- Standard OTel ecosystem compatibility (Jaeger, Grafana, Datadog, etc.)

---

## Enforced Code Patterns

These patterns are enforced by the 35+ linter suite and must be followed in all code:

| Pattern | Rule | Enforced By |
|---------|------|-------------|
| `context.Context` always first parameter | severity: error | revive: context-as-argument |
| `sdkerrors.NewValidation()` for input validation | no `errors.New()` in product code | forbidigo |
| `url.PathEscape()` on all URL path parameters | prevent injection | code review |
| `var _ Interface = (*impl)(nil)` compile-time checks | all service implementations | convention |
| Operation constants: `"Service.Method"` format | consistent tracing | convention |
| Error wrapping: always `%w`, never `%v` | proper error chains | errorlint |
| Import alias: `sdkerrors` for `pkg/errors` | consistent naming | revive: import-alias-naming (`^[a-z][a-z0-9]*$`) |
| Service receivers: `s` / Config receivers: `c` | consistent naming | revive: receiver-naming |
| `t.Parallel()` in all tests | except `t.Setenv`/global state | tparallel |
| `t.Helper()` in test helper functions | correct stack traces | thelper |
| No `panic()` / `os.Exit()` / `fmt.Print*` in library code | return errors, use structured logging | forbidigo |
| No `time.Sleep` in non-test code | use context timeout | forbidigo |
| No `http.Get/Post/Head` | use `http.NewRequestWithContext` | noctx + forbidigo |
| HTTP response bodies must be closed | prevent resource leaks | bodyclose |
| All error returns must be checked | exceptions: `io.Copy`, `Close`, stdout/stderr writes | errcheck |
| No `ioutil.*` | deprecated since Go 1.16, use `io` and `os` | forbidigo |
| No `log.Print/Fatal/Panic` | use structured logging via `pkg/observability` | forbidigo |
| No `unsafe.*` | forbidden in SDK code | forbidigo |
| No redundant import aliases | imports matching package name | revive: redundant-import-alias |
| No naked returns | all returns must be explicit | nakedret |
| Variable names >= 2 characters | common single-letter names allowlisted (`t`, `i`, `j`, `k`, `v`, `id`, `ok`, `wg`, `mu`, `fn`, `s`, `c`, `w`, `r`, `b`, `d`, `e`, `l`, `n`, `p`) | varnamelen |
| Use stdlib vars | `http.MethodGet` over `"GET"`, `http.StatusOK` over `200` | usestdlibvars |
| Preallocate slices | when capacity is known | prealloc |
| Named interface params | for readability in long signatures | inamedparam |
| Error type assertions checked | `check-type-assertions: true` | errcheck |
| Blank error returns checked | `check-blank: true` | errcheck |

---

## Linter Configuration

**File:** `.golangci.yml` (35+ linters, ZERO issues as of 2026-03-03)
**Golangci-lint version:** v2
**Run timeout:** 5 minutes
**Tests included:** Yes

### Complexity Limits

| Metric | Limit | Linter |
|--------|-------|--------|
| Cyclomatic complexity | 15 | gocyclo |
| Cognitive complexity | 20 | gocognit |
| Function length | 80 lines / 40 statements | funlen |
| Nesting depth | 4 | nestif |
| Cyclomatic (cyclop) | 15 | cyclop |
| Maintainability index | 20 (minimum) | maintidx |

### Import Organization
- **Formatter:** goimports
- **Order:** stdlib -> external -> internal
- **Alias rule:** lowercase only, pattern `^[a-z][a-z0-9]*$`
- **No redundant aliases** allowed

### Revive Rules (severity: error)

These revive rules are configured at `severity: error` and will fail the build:

- `import-shadowing` -- shadowing an imported package name
- `context-as-argument` -- `context.Context` must be first parameter
- `context-keys-type` -- context keys must use typed constants
- `error-return` -- error must be last return value
- `range-val-in-closure` -- loop variable captured in closure
- `range-val-address` -- address of loop variable taken
- `unreachable-code` -- code after return/panic
- `struct-tag` -- invalid struct tags
- `constant-logical-expr` -- logical expression always evaluates the same
- `redefines-builtin-id` -- redefining built-in identifier
- `unconditional-recursion` -- function unconditionally calls itself
- `identical-branches` -- if/else branches are identical
- `datarace` -- potential data race
- `waitgroup-by-value` -- passing WaitGroup by value
- `atomic` -- incorrect atomic operations
- `string-of-int` -- `string(int)` conversion
- `imports-blocklist` -- forbidden imports
- `redundant-import-alias` -- alias matches package name
- `import-alias-naming` -- alias does not match `^[a-z][a-z0-9]*$`

### Linter Exclusions

Certain linters are relaxed in specific paths to accommodate their unique requirements:

| Path | Relaxed Linters | Reason |
|------|----------------|--------|
| `*_test.go` | errcheck, gosec, funlen, gocognit, gocyclo, cyclop, maintidx, forbidigo, noctx, goconst, inamedparam, varnamelen | Test files need flexibility |
| `pkg/core/` | funlen, gocognit, gocyclo, cyclop, maintidx, nestif, revive (cognitive-complexity, empty-lines) | Core transport layer inherently complex |
| `pkg/observability/` | funlen, gocognit | Observability setup inherently complex |
| `examples/` | forbidigo, depguard, funlen, gocognit, gocyclo, cyclop, maintidx, nestif, errcheck, wsl_v5 | Examples are workflow demonstrations |
| `testing/` | forbidigo, gocognit, gocyclo, funlen, cyclop, wsl_v5 | Fake implementations satisfy interfaces |
| `pkg/core/` | forbidigo (`errors.New`) | Infrastructure cannot import sdkerrors (dependency direction) |
| `^errors\.go$` | forbidigo (`errors.New`) | Root sentinel errors |

---

## Security Rules

| Rule | Details |
|------|---------|
| No secrets in code | Configure via `.env` (copy from `.env.example`) |
| Env var prefix | `LERIAN_*` for all configuration |
| Credential redaction | `String()` and `MarshalJSON()` mask secrets |
| Auth header masking | Debug logs strip Authorization headers |
| Context-aware HTTP | All HTTP requests use `context.Context` (noctx linter) |
| GoSec scanning | Enabled in CI; excluded: G101 (env var names containing "token"/"key"), G117 (config struct fields with redaction), G704 (user-provided URLs required by SDK) |
| No unsafe package | Forbidden by forbidigo |
| Cross-host redirect protection | OAuth2 client refuses cross-host redirects |
| Response body limits | Error body truncated at 512 bytes (`MaxErrorBodyBytes`) |
| Error body check | `check-type-assertions: true`, `check-blank: true` in errcheck |
| Shadow detection | govet shadow: strict mode enabled (except `err` and `ok`) |

---

## Testing Standards

| Standard | Details |
|----------|---------|
| Framework | Go `testing` + `testify` v1.11.1 |
| Fake clients | `testing/leriantest/` -- no gomock |
| Coverage target | 80%+ for new critical logic |
| Test naming | `TestXxx` with table-driven subtests |
| Assertions | `require.NoError` for setup guards, `assert.*` for behavioral checks |
| Parallelism | `t.Parallel()` required (except `t.Setenv`) |
| Helpers | Must call `t.Helper()` |
| CI command | `make test` (race detection enabled) |
| Testifylint | Enabled with 9 checks deferred for incremental migration: `error-is-as` (315+ sites), `require-error` (28+), `float-compare` (8+), `go-require` (10+), `error-nil` (11+), `expected-actual` (9+), `empty` (6+), `len` (2+), `compares` |

---

## Dependency Policy

### Allowed Dependencies (depguard)
- Go standard library (`$gostd`)
- `github.com/LerianStudio/*` (internal)
- `github.com/stretchr/testify` (testing)
- `go.opentelemetry.io/*` (observability)

### Denied Dependencies
- `io/ioutil` -- deprecated since Go 1.16, use `io` and `os`
- `github.com/pkg/errors` -- use standard `errors` package with `fmt.Errorf` and `%w`

### Key Versions
- **Go:** 1.24.0
- **testify:** v1.11.1
- **OpenTelemetry:** v1.40.0 (otel, trace, metric, SDK, exporters)
- **google.golang.org/grpc:** v1.79.3 (indirect, via OTel exporters)
- **google.golang.org/protobuf:** v1.36.11 (indirect, via OTel exporters)

---

## Retry Configuration

| Parameter | Default |
|-----------|---------|
| Max retries | 3 (4 total attempts: 1 initial + 3 retries) |
| Base delay | 500ms |
| Max delay | 30s |
| Jitter ratio | 0.25 (+0-25% random jitter) |
| Max retries hard cap | 10 (BackendImpl) |
| Strategy | Exponential backoff: `BaseDelay * 2^attempt * (1 + rand * JitterRatio)` |
