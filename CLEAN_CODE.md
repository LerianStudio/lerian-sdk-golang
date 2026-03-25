# Clean Code Refactor Plan

## Goal

Remove accommodation-driven code, collapse unnecessary indirection, and make the SDK reflect the real product/API shapes instead of preserving legacy surfaces.

This plan assumes breaking changes are allowed when they materially improve structure.

## Architecture Snapshot

- Product-first Go SDK with umbrella facade in `client.go`
- Shared transport/service substrate in `pkg/core/`
- Largest complexity hotspot in `midaz/`
- Most inherited abstraction ceremony in `matcher/` and `tracer/`
- Public API shape heavily influenced by backward-compatibility decisions

---

## Tier 1 - Highest Leverage

These changes remove the deepest structural causes of complexity.

### Phase 1 - Rebuild Root Construction

#### Problem

- `Client` is both runtime object and mutable builder state
- Empty `With<Product>()` calls act as product activation sentinels
- Env merging and explicit config merging happen inside construction
- Repeated product bootstrapping is duplicated across all products

#### Targets

- `client.go`
- `options.go`
- `env.go`

#### Actions

1. Replace root functional options with an explicit root config model.
2. Move env loading/resolution out of `New()` into a separate config-loading step.
3. Remove `requested` flags and builder-only state from `Client`.
4. Remove empty activation calls like `WithMidaz()` as enablement signals.
5. Collapse duplicated product initialization into a data-driven bootstrap path.

#### Expected Outcome

- Root becomes a real composition layer instead of a compatibility state machine.
- Product enablement becomes explicit.
- Constructor behavior becomes predictable.

---

### Phase 2 - Replace the Current Core Transport Shape

Status: completed

#### Problem

- Historically, `pkg/core.Backend` was mode-split across multiple call helpers instead of one request-oriented transport path
- Historically, raw multipart handling relied on a special compatibility wrapper instead of explicit request bytes
- Shared helpers are too specific to one JSON/list contract and force downstream mini-cores

#### Targets

- `pkg/core/backend.go`
- `pkg/core/backend_impl.go`
- `pkg/core/service.go`
- `pkg/core/core.go`

#### Actions

1. Replace the current backend interface with a single request-oriented transport API.
2. Remove the raw-body compatibility pattern and use explicit request bytes instead.
3. Rebuild shared service helpers so they are request-shape driven, not envelope-driven.
4. Push multipart/raw/download concerns down into transport helpers, not product service methods.
5. Internalize or delete secondary retry abstractions that are not the real control point.

#### Expected Outcome

- Fewer escape hatches.
- Less duplicated logic in product packages.
- Cleaner path for raw, multipart, JSON, and header-aware operations.

---

## Tier 2 - Major Product Cleanup

These changes remove the largest accommodation clusters after the core is fixed.

### Phase 3 - Reset Midaz Public Surface

#### Problem

- `midaz/` carries the highest amount of compatibility debt
- Split interfaces (`*Metrics`, `*Extended`, `*Variants`) preserve old shapes on the same concrete services
- Flat client fields hide the real multi-backend topology

#### Targets

- `midaz/client.go`
- `midaz/organizations.go`
- `midaz/accounts.go`
- `midaz/balances.go`
- `midaz/operations.go`
- `midaz/transactions.go`

#### Actions

1. Collapse all split service interfaces into one service per resource.
2. Remove duplicate compatibility accessors from `midaz.Client`.
3. Replace flat surface with backend-scoped subclients that reflect reality:
   - `Midaz.Onboarding`
   - `Midaz.Transactions`
   - `Midaz.CRM`
4. Remove compatibility constructors that only preserve pre-CRM shape.

#### Expected Outcome

- Midaz becomes structurally honest.
- Public API surface shrinks.
- Fake/test wiring becomes simpler.

---

### Phase 4 - Delete Midaz Legacy Compatibility Paths

#### Problem

- Legacy payload conversions and fallback routes are embedded in default behavior
- New callers pay ongoing complexity cost for old server/client contracts

#### Targets

- `midaz/models_input.go`
- `midaz/transactions.go`
- `midaz/accounts.go`
- `midaz/asset_rates.go`

#### Actions

1. Remove flattened transaction payload support and require canonical request shape.
2. Remove `/transactions/json` -> `/transactions` fallback.
3. Remove account external-code legacy path fallback.
4. Remove asset-rate PUT -> POST fallback.
5. If legacy support must exist at all, make it explicit opt-in, not default behavior.

#### Expected Outcome

- Midaz request flow becomes direct.
- Hidden fallback behavior disappears.
- Runtime behavior matches declared API.

---

### Phase 5 - Align Midaz Balance APIs With Reality

#### Problem

- Balance creation infers account scope from compatibility-only input fields
- Singleton lookup helpers sit on top of plural endpoints and invent extra semantics

#### Targets

- `midaz/balances.go`
- `midaz/models_input.go`

#### Actions

1. Remove compatibility-only non-wire fields from balance create input.
2. Keep only account-scoped balance creation.
3. Remove singleton wrappers over collection endpoints.
4. Prefer collection-first or iterator-first lookups.

#### Expected Outcome

- Balance API reflects actual server behavior.
- Less wrapper logic in SDK and fake client.

---

## Tier 3 - Normalize Product Divergence Honestly

These changes reduce false uniformity across products.

### Phase 6 - Normalize Fees Listing Surface

#### Problem

- `fees` exposes backend-specific list/query quirks publicly
- It bypasses shared list contracts instead of normalizing them

#### Targets

- `fees/packages.go`
- `fees/models.go`

#### Actions

1. Keep page-based listing semantics.
2. Stop exposing backend query-name quirks publicly.
3. Return a normalized pagination model instead of transport-shaped response types.
4. Keep only package-specific filter concepts in the public options type.

#### Expected Outcome

- Fees becomes easier to use without pretending it shares every other product contract.
- Product-specific behavior stays honest, but transport quirks disappear.

---

### Phase 7 - Make Reporter Upload Semantics Explicit

#### Problem

- `Templates.Create` is really an upload endpoint hidden behind CRUD naming
- Multipart handling is open-coded at the service layer

#### Targets

- `reporter/templates.go`

#### Actions

1. Rename the operation to reflect upload semantics.
2. Move multipart assembly into shared helper/transport code.
3. Keep raw report download explicit as-is.

#### Expected Outcome

- Reporter surface becomes more truthful.
- Service code becomes thinner.

---

## Tier 4 - Remove Inherited Ceremony

These are worthwhile once the major structural work is done.

### Phase 8 - Reduce Interface Ceremony in Matcher and Tracer

#### Problem

- Many services are interface + private impl + constructor + client field with only one real implementation
- Most of the benefit is for fake wiring, not real runtime polymorphism

#### Targets

- `matcher/`
- `tracer/`
- `testing/leriantest/`

#### Actions

1. Collapse service-per-interface pattern where only one implementation exists.
2. Move the testing seam lower if needed instead of preserving interface-heavy public surfaces.
3. Remove trivial wrappers and aliases:
   - `matcher/client.go` `ErrorParser()`
   - `tracer/errors.go` `categoryFromStatus`

#### Expected Outcome

- Less template ceremony.
- Smaller public API surface.
- Clearer ownership between real abstractions and test-only seams.

---

### Phase 9 - Revisit Shared List Abstractions

#### Problem

- `models.ListOptions` tries to unify incompatible protocols
- Cursor and page semantics are mixed together
- Some products use iterators, some use page responses, some need custom shims

#### Targets

- `models/common.go`
- `pkg/core/service.go`
- Product-specific list callers
- `testing/leriantest/`
- `examples/`

#### Actions

1. Split cursor-style and page-style list concepts.
2. Stop using one global list abstraction to span incompatible product APIs.
3. Update examples so the SDK teaches one honest access pattern per endpoint type.

#### Expected Outcome

- Less fake uniformity.
- Cleaner product-specific APIs.
- Fewer side systems like CRM/Fees list shims.

---

## Tier 5 - Low-Risk Cleanup Wins

These can happen opportunistically at any point.

### Phase 10 - Delete Small Compatibility Leftovers

#### Targets

- `env.go`
- `client.go`
- `errors.go`
- small product wrappers/aliases

#### Actions

1. Remove legacy env migration guards once the new config path exists.
2. Remove one-line wrappers only kept for compatibility.
3. Reassess tiny facade sugar helpers and keep only those with clear ergonomic value.

#### Expected Outcome

- Less noise.
- Fewer stale compatibility branches.

---

## What Should Stay

These are not the enemy.

- `pkg/auth/auth.go` `Authenticator` abstraction
- `pkg/errors/errors.go` shared error model
- `pkg/pagination/iterator.go` as a reusable iterator primitive
- Real Midaz multi-backend routing
- CRM-specific helper logic where the remote API is genuinely different
- Raw download endpoints that are honestly raw

---

## Recommended Execution Order

1. Phase 1 - Rebuild root construction
2. Phase 2 - Replace core transport/service shape
3. Phase 3 - Reset Midaz public surface
4. Phase 4 - Delete Midaz legacy compatibility paths
5. Phase 5 - Align Midaz balance APIs with reality
6. Phase 6 - Normalize Fees listing surface
7. Phase 7 - Make Reporter upload semantics explicit
8. Phase 8 - Reduce interface ceremony in Matcher and Tracer
9. Phase 9 - Revisit shared list abstractions
10. Phase 10 - Delete small compatibility leftovers

---

## Refactor Principle

Do not preserve a lie because it is convenient.

If the backend is plural, expose plural.
If the endpoint is upload, call it upload.
If the product has different pagination semantics, model them honestly.
If compatibility is required, make it explicit and isolated - never implicit and ambient.
