// Package observability provides OpenTelemetry-based tracing, metrics, and
// structured logging for the SDK.
//
// When an observability [Provider] is attached to the root [lerian.Client],
// the [core.BackendImpl] HTTP transport layer uses it to create a child span
// for every outbound HTTP request, record request duration histograms and
// request count metrics, and propagate W3C trace-context headers via the
// OTel propagator injected into each request's context.
//
// The [Provider] interface is intentionally narrow so consumers can supply
// their own implementations without pulling in the full OTel SDK:
//
//	type Provider interface {
//	    Tracer() trace.Tracer
//	    Meter() metric.Meter
//	    Logger() *slog.Logger
//	    Shutdown(ctx context.Context) error
//	    IsEnabled() bool
//	}
//
// Use [NewProvider] to create a production-grade provider backed by OTLP/HTTP
// exporters, or [NewNoopProvider] for a zero-cost disabled variant that is
// useful in tests or when no collector endpoint is configured.
//
// # Usage
//
// Attach an observability provider when creating the client:
//
//	provider, _ := observability.NewProvider(observability.ProviderConfig{
//	    ServiceName:       "my-service",
//	    CollectorEndpoint: "http://localhost:4318",
//	    EnableTraces:      true,
//	    EnableMetrics:     true,
//	    EnableLogs:        true,
//	})
//	client, _ := lerian.New(
//	    lerian.WithObservability(provider),
//	    // ... product options ...
//	)
//	defer client.Shutdown(ctx) // also shuts down the provider
package observability
