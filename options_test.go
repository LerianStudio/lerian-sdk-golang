package lerian

import (
	"net/http"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigUsesSharedHTTPClient(t *testing.T) {
	t.Parallel()

	custom := &http.Client{Timeout: 99 * time.Second}

	client, err := New(Config{HTTPClient: custom})
	require.NoError(t, err)
	assert.Same(t, custom, client.httpClient)
}

func TestConfigUsesSharedHTTPClientInProduct(t *testing.T) {
	t.Parallel()

	custom := &http.Client{Timeout: 42 * time.Second}

	client, err := New(Config{
		HTTPClient: custom,
		Matcher:    mustMatcherConfig(t, matcher.WithBaseURL("http://localhost:3002/v1")),
	})
	require.NoError(t, err)
	assert.Same(t, custom, client.httpClient)
	assert.NotNil(t, client.Matcher)
}

func TestConfigUsesRetryConfig(t *testing.T) {
	t.Parallel()

	custom := &retry.Config{MaxRetries: 5, BaseDelay: time.Second}

	client, err := New(Config{RetryConfig: custom})
	require.NoError(t, err)
	assert.Equal(t, *custom, client.retryConfig)
}

func TestConfigRetryZeroDisables(t *testing.T) {
	t.Parallel()

	custom := &retry.Config{MaxRetries: 0, BaseDelay: 0}

	client, err := New(Config{RetryConfig: custom})
	require.NoError(t, err)
	assert.Equal(t, 0, client.retryConfig.MaxRetries)
	assert.Equal(t, time.Duration(0), client.retryConfig.BaseDelay)
}

func TestConfigUsesDebugFlag(t *testing.T) {
	t.Parallel()

	client, err := New(Config{Debug: true})
	require.NoError(t, err)
	assert.True(t, client.debug)
}

func TestConfigUsesObservabilityDefaults(t *testing.T) {
	t.Parallel()

	client, err := New(Config{})
	require.NoError(t, err)
	assert.NotNil(t, client.observability)
	assert.False(t, client.observability.IsEnabled())
}

func TestConfigUsesDefaultValues(t *testing.T) {
	t.Parallel()

	client, err := New(Config{})
	require.NoError(t, err)
	assert.Equal(t, retry.DefaultConfig(), client.retryConfig)
	assert.Equal(t, defaultHTTPTimeout, client.httpClient.Timeout)
	assert.NotNil(t, client.jsonPool)
}
