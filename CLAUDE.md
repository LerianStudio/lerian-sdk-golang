# Repository Guidelines

## Project Structure & Module Organization

- `midaz/` -- Midaz product client (13 services: organizations, ledgers, accounts, portfolios, assets, segments, transactions, operations, balances, etc.)
- `matcher/` -- Matcher product client (14 services: rules, reconciliation, match results, overrides, etc.)
- `tracer/` -- Tracer product client (4 services: traces, events, queries, validation)
- `reporter/` -- Reporter product client (3 services: reports, templates, schedules)
- `fees/` -- Fees product client (3 services: fee rules, estimation, billing)
- `models/` -- Shared model types used across products (request/response payloads)
- `pkg/` -- SDK infrastructure utilities:
  - `pkg/core/` -- Generic CRUD base service, backend interface, iterator, HTTP transport
  - `pkg/errors/` -- Hierarchical error types and sentinel errors
  - `pkg/auth/` -- Authentication providers
  - `pkg/retry/` -- Retry logic with exponential backoff
  - `pkg/performance/` -- JSON pool, buffer management
  - `pkg/observability/` -- OpenTelemetry integration (traces, metrics)
- `testing/leriantest/` -- Fake client for all 37 services (no mocks, no network)
- `examples/` -- Runnable example programs (one per product + multi-product)
- Root files: `client.go` (top-level Lerian client), `options.go`, `env.go`, `errors.go`, `Makefile`, `go.mod`, `.env.example`

## Build, Test, and Development Commands

- `make set-env` -- Create `.env` from `.env.example`.
- `make build` -- Build all packages.
- `make test` / `make test-fast` -- Run all/short tests with race detection.
- `make coverage` -- Produce HTML coverage report in `artifacts/`.
- `make lint` -- Run golangci-lint with the full 35+ linter suite.
- `make fmt` -- Format all Go source files.
- `make tidy` -- Tidy module dependencies.
- `make gosec` -- Run gosec security scanner.
- `make verify-sdk` -- Quick build + vet check.
- `go test -v ./path/to/package -run TestName` -- Run a specific test.
- `make godoc` -- Start a godoc server (http://localhost:6060).
- Run an example: `cd examples/midaz-workflow && go run .`

## Coding Style & Naming Conventions

- Go 1.24.x. Run `make fmt` before committing.
- Follow standard Go style and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Package names: lowercase, single-word.
- Exported names: CamelCase with first letter capitalized.
- Unexported names: camelCase with first letter lowercase.
- Import order: standard library, external packages, internal packages.
- Keep functions small, context-aware (`context.Context` first param), and return rich errors.
- Use functional options pattern for configuration (see `options.go`).
- Use the deferred options pattern for product configuration (options applied after all products init).
- Use interfaces for external dependencies (see `pkg/core/` Backend interface).
- Document all exported functions, types, and variables.
- Lint with `golangci-lint` (`make lint`). No panics in library code.
- Module path: `github.com/LerianStudio/lerian-sdk-golang`.

## Key Patterns

- **Deferred options**: `New()` accepts Options that run after all product backends initialize (order-independent config).
- **Generic CRUD**: `BaseService` in `pkg/core/service.go` provides `Get[T]`, `Create[T,I]`, `Update[T,I]`, `Delete`, `List[T]`, `Action[T]`.
- **Iterator[T]**: Lazy pagination with `All()` (iter.Seq2), `Collect`, `CollectN`, `ForEachConcurrent`.
- **Hierarchical errors**: `errors.Is()` matching on sentinel chains (Category + optional product Code).
- **Env precedence**: explicit Option > `LERIAN_*` env var > default.
- **RawBody sentinel**: Non-JSON multipart payloads bypass the JSON serializer.

## Testing Guidelines

- Use Go's `testing` with `testify` for assertions.
- Name tests `*_test.go`; functions `TestXxx` and table-driven where appropriate.
- Use `testing/leriantest` for fake clients instead of gomock for service-level tests.
- Write unit tests for all new code (minimum 80% coverage).
- Run `make test` locally; target >80% coverage for new critical logic; generate report with `make coverage`.
- Test helpers: `NewIteratorFromSlice`, `NewErrorIterator`, `fakeStore[T]` in `testing/leriantest/`.

## Commit & Pull Request Guidelines

- Conventional Commits: `<type>(<scope>): <description>`
  - Types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`
  - Examples: `feat(midaz): add balance caching`, `fix(matcher): handle empty rule set`
- PRs must include: purpose, scope, key changes, how-to-test, and linked issues.
- Run `make fmt lint test verify-sdk` before opening a PR.

## Security & Configuration Tips

- Never commit secrets. Configure via `.env` (copy from `.env.example`).
- Environment variables use the `LERIAN_` prefix (e.g., `LERIAN_MIDAZ_AUTH_TOKEN`).
- Ensure idempotent requests by setting `X-Idempotency` header via context when needed.

## Agent-Specific Instructions

- Keep changes minimal and scoped.
- Touch only relevant packages; follow existing folder conventions.
- After example changes, verify by running the example (`cd examples/<name> && go run .`).
