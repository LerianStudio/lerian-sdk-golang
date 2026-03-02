// Package core provides the Backend transport abstraction and BaseService
// generic functions that underpin every product client in the SDK.
//
// # Backend
//
// [Backend] encapsulates the HTTP round-trip lifecycle: authentication
// header injection via [auth.Authenticator], retry with exponential backoff
// via [retry.Config], request/response serialisation through pooled JSON
// codecs from [performance.JSONPool], and OpenTelemetry instrumentation
// via [observability.Provider]. When a Provider is wired in, [BackendImpl]
// creates a child span per HTTP request, records duration histograms and
// request counters, and propagates W3C trace-context headers.
//
// Product clients never construct a Backend directly; the umbrella
// [lerian.Client] wires one per product during initialisation.
//
// # BaseService
//
// BaseService supplies generic CRUD helpers (Get, List, Create, Update,
// Delete) that product-specific service implementations compose rather
// than duplicate. Each helper builds the URL, delegates to the Backend,
// and returns the typed response:
//
//	type OrganizationsService struct {
//	    backend core.Backend
//	}
//
//	func (s *OrganizationsService) Get(ctx context.Context, id string) (*Organization, error) {
//	    return core.Get[Organization](ctx, s.backend, "/organizations/"+id)
//	}
package core
