package observability

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// ---------------------------------------------------------------------------
// Provider interface
// ---------------------------------------------------------------------------

// Provider is the central observability abstraction for the SDK.
//
// It exposes the three pillars — tracing, metrics, and structured logging —
// behind a minimal interface. Consumers may use [NewProvider] for a
// full-featured OTel-backed implementation, or [NewNoopProvider] for a
// zero-cost disabled variant (useful in tests or when observability is not
// configured).
type Provider interface {
	// Tracer returns the [trace.Tracer] used for creating spans.
	Tracer() trace.Tracer

	// Meter returns the [metric.Meter] used for recording metrics.
	Meter() metric.Meter

	// Logger returns the structured logger for SDK-internal diagnostics.
	Logger() *slog.Logger

	// Shutdown gracefully drains buffered telemetry and releases resources.
	// It is safe to call Shutdown multiple times; only the first call
	// performs work.
	Shutdown(ctx context.Context) error

	// IsEnabled reports whether this provider is actually producing
	// telemetry. A noop provider always returns false.
	IsEnabled() bool
}

// ---------------------------------------------------------------------------
// ProviderConfig
// ---------------------------------------------------------------------------

// ProviderConfig holds the settings required to construct an [otelProvider].
//
// When all three Enable* flags are false, [NewProvider] short-circuits and
// returns a [noopProvider] — no OTel SDK objects are allocated.
type ProviderConfig struct {
	// ServiceName is the logical name attached to the OTel resource
	// (e.g. "lerian-sdk-go").
	ServiceName string

	// ServiceVersion is the version string attached to the OTel resource.
	ServiceVersion string

	// CollectorEndpoint is the base URL of the OTLP-compatible collector
	// (e.g. "http://localhost:4318"). Both traces and metrics are sent
	// using OTLP/HTTP (protobuf).
	CollectorEndpoint string

	// EnableTraces activates distributed tracing via an OTLP HTTP
	// span exporter + batching TracerProvider.
	EnableTraces bool

	// EnableMetrics activates metrics collection via an OTLP HTTP
	// metric exporter + periodic MeterProvider reader.
	EnableMetrics bool

	// EnableLogs activates structured logging to stdout using a
	// [slog.JSONHandler].
	EnableLogs bool
}

// ---------------------------------------------------------------------------
// Noop provider
// ---------------------------------------------------------------------------

// noopProvider is a zero-cost Provider that produces no telemetry.
//
// Method calls (Tracer, Meter, Logger) return pre-allocated noop/discard
// instances so there is no per-call allocation overhead.
type noopProvider struct {
	tracer trace.Tracer
	meter  metric.Meter
	logger *slog.Logger
}

// NewNoopProvider returns a [Provider] that is completely inert: no spans
// are created, no metrics are recorded, and log output is discarded.
//
// This is the default when no observability configuration is supplied and
// is also useful for unit testing.
func NewNoopProvider() Provider {
	return &noopProvider{
		tracer: tracenoop.NewTracerProvider().Tracer("noop"),
		meter:  metricnoop.NewMeterProvider().Meter("noop"),
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func (n *noopProvider) Tracer() trace.Tracer           { return n.tracer }
func (n *noopProvider) Meter() metric.Meter            { return n.meter }
func (n *noopProvider) Logger() *slog.Logger           { return n.logger }
func (n *noopProvider) Shutdown(context.Context) error { return nil }
func (n *noopProvider) IsEnabled() bool                { return false }

// ---------------------------------------------------------------------------
// OTel provider
// ---------------------------------------------------------------------------

// otelProvider is the production Provider backed by the OpenTelemetry SDK.
//
// It creates OTLP/HTTP exporters for traces and/or metrics and a
// structured JSON logger. The Shutdown method ensures all buffered
// telemetry is flushed exactly once (guarded by sync.Once).
type otelProvider struct {
	tracer       trace.Tracer
	meter        metric.Meter
	logger       *slog.Logger
	shutdownFns  []func(context.Context) error
	shutdownOnce sync.Once
	shutdownErr  error
}

// instrumentationName is the OTel instrumentation scope name attached to
// all tracers and meters produced by this package.
const instrumentationName = "github.com/LerianStudio/lerian-sdk-golang"

// NewProvider constructs a production [Provider] based on the supplied
// [ProviderConfig].
//
// If all three Enable* flags are false the function returns a
// [noopProvider] (and a nil error) without allocating any OTel SDK
// objects.
//
// For each enabled pillar the function creates the corresponding OTLP/HTTP
// exporter and SDK provider, wiring them to a shared OTel Resource that
// carries service.name and service.version attributes. Callers must invoke
// [Provider.Shutdown] to flush buffered telemetry on program exit.
func NewProvider(cfg ProviderConfig) (Provider, error) {
	// Fast path: everything disabled → return noop.
	if !cfg.EnableTraces && !cfg.EnableMetrics && !cfg.EnableLogs {
		return NewNoopProvider(), nil
	}

	ctx := context.Background()

	// Build the shared OTel resource.
	res, err := buildResource(ctx, cfg.ServiceName, cfg.ServiceVersion)
	if err != nil {
		return nil, fmt.Errorf("observability: build resource: %w", err)
	}

	p := &otelProvider{}

	// --- Traces -----------------------------------------------------------
	if cfg.EnableTraces {
		tp, shutdownFn, tErr := setupTracing(ctx, cfg, res)
		if tErr != nil {
			// Clean up anything already created.
			p.shutdownErr = p.shutdownAll(ctx)

			return nil, fmt.Errorf("observability: setup tracing: %w", tErr)
		}

		p.tracer = tp.Tracer(instrumentationName)
		p.shutdownFns = append(p.shutdownFns, shutdownFn)
	} else {
		p.tracer = tracenoop.NewTracerProvider().Tracer("noop")
	}

	// --- Metrics ----------------------------------------------------------
	if cfg.EnableMetrics {
		mp, shutdownFn, mErr := setupMetrics(ctx, cfg, res)
		if mErr != nil {
			p.shutdownErr = p.shutdownAll(ctx)

			return nil, fmt.Errorf("observability: setup metrics: %w", mErr)
		}

		p.meter = mp.Meter(instrumentationName)
		p.shutdownFns = append(p.shutdownFns, shutdownFn)
	} else {
		p.meter = metricnoop.NewMeterProvider().Meter("noop")
	}

	// --- Logs -------------------------------------------------------------
	if cfg.EnableLogs {
		p.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	} else {
		p.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return p, nil
}

func (o *otelProvider) Tracer() trace.Tracer { return o.tracer }
func (o *otelProvider) Meter() metric.Meter  { return o.meter }
func (o *otelProvider) Logger() *slog.Logger { return o.logger }
func (o *otelProvider) IsEnabled() bool      { return true }

// Shutdown drains all buffered telemetry. It is safe to call multiple
// times — only the first invocation performs work.
func (o *otelProvider) Shutdown(ctx context.Context) error {
	o.shutdownOnce.Do(func() {
		o.shutdownErr = o.shutdownAll(ctx)
	})

	return o.shutdownErr
}

// shutdownAll iterates over the registered shutdown functions and collects
// any errors. It is called both from Shutdown (guarded by Once) and as a
// cleanup path if provider construction fails partway through.
func (o *otelProvider) shutdownAll(ctx context.Context) error {
	var errs []error

	for _, fn := range o.shutdownFns {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	// Combine multiple errors into a single one for the caller.
	combined := errs[0]
	for _, e := range errs[1:] {
		combined = fmt.Errorf("%w; %w", combined, e)
	}

	return combined
}

// ---------------------------------------------------------------------------
// Internal setup helpers
// ---------------------------------------------------------------------------

// buildResource creates the OTel Resource shared by the TracerProvider and
// MeterProvider. It carries the semantic-convention attributes
// service.name and service.version.
func buildResource(ctx context.Context, serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
}

// setupTracing creates an OTLP/HTTP trace exporter, wraps it in a
// BatchSpanProcessor, and returns a TracerProvider plus a shutdown function.
func setupTracing(
	ctx context.Context,
	cfg ProviderConfig,
	res *resource.Resource,
) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	var opts []otlptracehttp.Option
	if cfg.CollectorEndpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpointURL(cfg.CollectorEndpoint))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return tp, tp.Shutdown, nil
}

// setupMetrics creates an OTLP/HTTP metric exporter, wraps it in a
// PeriodicReader, and returns a MeterProvider plus a shutdown function.
func setupMetrics(
	ctx context.Context,
	cfg ProviderConfig,
	res *resource.Resource,
) (*sdkmetric.MeterProvider, func(context.Context) error, error) {
	var opts []otlpmetrichttp.Option
	if cfg.CollectorEndpoint != "" {
		opts = append(opts, otlpmetrichttp.WithEndpointURL(cfg.CollectorEndpoint))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create OTLP metric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(exporter)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)

	return mp, mp.Shutdown, nil
}
