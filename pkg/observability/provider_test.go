package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Noop provider tests
// ---------------------------------------------------------------------------

func TestNoopProvider(t *testing.T) {
	t.Parallel()

	p := NewNoopProvider()

	assert.False(t, p.IsEnabled(), "noop provider must report IsEnabled()==false")
	assert.NotNil(t, p.Tracer(), "noop provider must return a non-nil Tracer")
	assert.NotNil(t, p.Meter(), "noop provider must return a non-nil Meter")
	assert.NotNil(t, p.Logger(), "noop provider must return a non-nil Logger")
	assert.NoError(t, p.Shutdown(context.Background()), "noop Shutdown must return nil")
}

func TestNoopProviderShutdownIdempotent(t *testing.T) {
	t.Parallel()

	p := NewNoopProvider()

	for i := 0; i < 3; i++ {
		assert.NoError(t, p.Shutdown(context.Background()),
			"noop Shutdown call %d must not error", i+1)
	}
}

// ---------------------------------------------------------------------------
// NewProvider — all disabled → noop
// ---------------------------------------------------------------------------

func TestNewProviderAllDisabled(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(ProviderConfig{
		ServiceName:    "test-svc",
		ServiceVersion: "0.0.1",
		EnableTraces:   false,
		EnableMetrics:  false,
		EnableLogs:     false,
	})

	require.NoError(t, err)
	assert.False(t, p.IsEnabled(),
		"provider with all flags disabled must behave as noop")
}

// ---------------------------------------------------------------------------
// NewProvider — traces enabled
// ---------------------------------------------------------------------------

func TestNewProviderWithTraces(t *testing.T) {
	t.Parallel()

	// We point the exporter at an unreachable endpoint. The OTel SDK
	// creates the provider eagerly but exports lazily, so construction
	// must succeed even without a running collector.
	p, err := NewProvider(ProviderConfig{
		ServiceName:       "trace-test",
		ServiceVersion:    "1.0.0",
		CollectorEndpoint: "http://127.0.0.1:14318",
		EnableTraces:      true,
		EnableMetrics:     false,
		EnableLogs:        false,
	})

	require.NoError(t, err, "provider construction must not fail without a collector")
	assert.True(t, p.IsEnabled())
	assert.NotNil(t, p.Tracer())

	// Meter and Logger should still be usable (noop / discard).
	assert.NotNil(t, p.Meter())
	assert.NotNil(t, p.Logger())

	// Shutdown should not panic even though the collector is unreachable.
	// The batch exporter will fail silently (or return a context-cancelled
	// error) — either way we accept it.
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	_ = p.Shutdown(ctx)
}

// ---------------------------------------------------------------------------
// NewProvider — metrics enabled
// ---------------------------------------------------------------------------

func TestNewProviderWithMetrics(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(ProviderConfig{
		ServiceName:       "metric-test",
		ServiceVersion:    "1.0.0",
		CollectorEndpoint: "http://127.0.0.1:14318",
		EnableTraces:      false,
		EnableMetrics:     true,
		EnableLogs:        false,
	})

	require.NoError(t, err)
	assert.True(t, p.IsEnabled())
	assert.NotNil(t, p.Meter())

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	_ = p.Shutdown(ctx)
}

// ---------------------------------------------------------------------------
// NewProvider — logs enabled
// ---------------------------------------------------------------------------

func TestNewProviderWithLogs(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(ProviderConfig{
		ServiceName:    "log-test",
		ServiceVersion: "1.0.0",
		EnableTraces:   false,
		EnableMetrics:  false,
		EnableLogs:     true,
	})

	require.NoError(t, err)
	assert.True(t, p.IsEnabled())
	assert.NotNil(t, p.Logger())

	// The logger should actually write (not discard). We just verify it
	// does not panic when invoked.
	p.Logger().Info("hello from test")

	require.NoError(t, p.Shutdown(context.Background()))
}

// ---------------------------------------------------------------------------
// NewProvider — all enabled
// ---------------------------------------------------------------------------

func TestNewProviderAllEnabled(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(ProviderConfig{
		ServiceName:       "full-test",
		ServiceVersion:    "2.0.0",
		CollectorEndpoint: "http://127.0.0.1:14318",
		EnableTraces:      true,
		EnableMetrics:     true,
		EnableLogs:        true,
	})

	require.NoError(t, err)
	assert.True(t, p.IsEnabled())
	assert.NotNil(t, p.Tracer())
	assert.NotNil(t, p.Meter())
	assert.NotNil(t, p.Logger())

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	_ = p.Shutdown(ctx)
}

// ---------------------------------------------------------------------------
// Shutdown idempotency on the real provider
// ---------------------------------------------------------------------------

func TestShutdownIdempotency(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(ProviderConfig{
		ServiceName:       "idempotent-test",
		ServiceVersion:    "1.0.0",
		CollectorEndpoint: "http://127.0.0.1:14318",
		EnableTraces:      true,
		EnableMetrics:     true,
		EnableLogs:        true,
	})

	require.NoError(t, err)

	ctx := context.Background()

	// Calling Shutdown multiple times must not panic and must return the
	// same error (or nil) each time.
	err1 := p.Shutdown(ctx)
	err2 := p.Shutdown(ctx)
	err3 := p.Shutdown(ctx)

	assert.Equal(t, err1, err2, "repeated Shutdown must return the same error")
	assert.Equal(t, err2, err3, "repeated Shutdown must return the same error")
}

// ---------------------------------------------------------------------------
// Provider interface compliance (compile-time checks)
// ---------------------------------------------------------------------------

var (
	_ Provider = (*noopProvider)(nil)
	_ Provider = (*otelProvider)(nil)
)
