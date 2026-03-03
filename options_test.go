package lerian

import (
	"net/http"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Shared infrastructure options
// ---------------------------------------------------------------------------

func TestWithDebug(t *testing.T) {
	t.Parallel()

	client, err := New(WithDebug(true))
	require.NoError(t, err)
	assert.True(t, client.debug, "debug should be true")
}

func TestWithDebugFalse(t *testing.T) {
	t.Parallel()

	client, err := New(WithDebug(false))
	require.NoError(t, err)
	assert.False(t, client.debug, "debug should be false")
}

func TestWithRetry(t *testing.T) {
	t.Parallel()

	client, err := New(WithRetry(5, 1*time.Second))
	require.NoError(t, err)
	assert.Equal(t, 5, client.retryConfig.MaxRetries)
	assert.Equal(t, 1*time.Second, client.retryConfig.BaseDelay)
}

func TestWithRetryZeroDisables(t *testing.T) {
	t.Parallel()

	client, err := New(WithRetry(0, 0))
	require.NoError(t, err)
	assert.Equal(t, 0, client.retryConfig.MaxRetries)
	assert.Equal(t, time.Duration(0), client.retryConfig.BaseDelay)
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	custom := &http.Client{Timeout: 99 * time.Second}

	client, err := New(WithHTTPClient(custom))
	require.NoError(t, err)
	assert.Same(t, custom, client.httpClient, "custom HTTP client should be stored")
}

func TestWithHTTPClientUsedInProduct(t *testing.T) {
	t.Parallel()

	// Verify that the custom HTTP client propagates to product backends.
	custom := &http.Client{Timeout: 42 * time.Second}

	client, err := New(
		WithHTTPClient(custom),
		WithMatcher(
			matcher.WithBaseURL("http://localhost:3002/v1"),
		),
	)
	require.NoError(t, err)
	assert.Same(t, custom, client.httpClient)
	assert.NotNil(t, client.Matcher)
}

func TestWithObservabilityAllDisabled(t *testing.T) {
	t.Parallel()

	client, err := New(WithObservability(false, false, false))
	require.NoError(t, err)
	assert.NotNil(t, client.observability)
	assert.False(t, client.observability.IsEnabled())
}

func TestWithCollectorEndpoint(t *testing.T) {
	t.Parallel()

	client, err := New(WithCollectorEndpoint("http://collector:4318"))
	require.NoError(t, err)
	assert.Equal(t, "http://collector:4318", client.otelEndpoint)
}

// ---------------------------------------------------------------------------
// Default values
// ---------------------------------------------------------------------------

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()

	client, err := New()
	require.NoError(t, err)
	// DefaultConfig() values from pkg/retry.
	assert.Equal(t, 3, client.retryConfig.MaxRetries)
	assert.Equal(t, 500*time.Millisecond, client.retryConfig.BaseDelay)
}

func TestDefaultHTTPClient(t *testing.T) {
	t.Parallel()

	client, err := New()
	require.NoError(t, err)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, defaultHTTPTimeout, client.httpClient.Timeout)
}

func TestDefaultJSONPool(t *testing.T) {
	t.Parallel()

	client, err := New()
	require.NoError(t, err)
	assert.NotNil(t, client.jsonPool)
}

// ---------------------------------------------------------------------------
// Option ordering
// ---------------------------------------------------------------------------

func TestOptionsAppliedInOrder(t *testing.T) {
	t.Parallel()

	// Later options should override earlier ones.
	client, err := New(
		WithRetry(1, 100*time.Millisecond),
		WithRetry(10, 2*time.Second),
	)
	require.NoError(t, err)
	assert.Equal(t, 10, client.retryConfig.MaxRetries)
	assert.Equal(t, 2*time.Second, client.retryConfig.BaseDelay)
}

func TestOptionErrorPropagation(t *testing.T) {
	t.Parallel()

	// A failing option should prevent client creation.
	badOption := Option(func(_ *Client) error {
		return assert.AnError
	})

	_, err := New(badOption)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "applying option")
}
