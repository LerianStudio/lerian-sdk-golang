package tracer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// fakeBackend implements core.Backend for unit-testing the Client wiring.
// ---------------------------------------------------------------------------

type fakeBackend struct {
	lastMethod string
	lastPath   string
}

func (f *fakeBackend) Call(_ context.Context, method, path string, _, _ any) error {
	f.lastMethod = method
	f.lastPath = path

	return nil
}

func (f *fakeBackend) CallWithHeaders(_ context.Context, method, path string,
	_ map[string]string, _, _ any) error {
	f.lastMethod = method
	f.lastPath = path

	return nil
}

func (f *fakeBackend) CallRaw(_ context.Context, method, path string, _ any) ([]byte, error) {
	f.lastMethod = method
	f.lastPath = path

	return nil, nil
}

// ---------------------------------------------------------------------------
// NewClient tests
// ---------------------------------------------------------------------------

func TestNewClientReturnsNonNil(t *testing.T) {
	t.Parallel()

	backend := &fakeBackend{}
	cfg := Config{
		BaseURL:      "http://localhost:3003/v1",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     "https://auth.example.com/token",
		Scopes:       []string{"tracer:read"},
		Timeout:      5 * time.Second,
	}

	client := NewClient(backend, cfg)
	require.NotNil(t, client)
}

func TestNewClientStoresConfig(t *testing.T) {
	t.Parallel()

	backend := &fakeBackend{}
	cfg := Config{
		BaseURL:      "http://localhost:3003/v1",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     "https://auth.example.com/token",
		Scopes:       []string{"tracer:read"},
		Timeout:      15 * time.Second,
	}

	client := NewClient(backend, cfg)

	assert.Equal(t, "http://localhost:3003/v1", client.config.BaseURL)
	assert.Equal(t, "client-id", client.config.ClientID)
	assert.Equal(t, "client-secret", client.config.ClientSecret)
	assert.Equal(t, "https://auth.example.com/token", client.config.TokenURL)
	assert.Equal(t, []string{"tracer:read"}, client.config.Scopes)
	assert.Equal(t, 15*time.Second, client.config.Timeout)
}

func TestNewClientStoresBackend(t *testing.T) {
	t.Parallel()

	backend := &fakeBackend{}
	cfg := Config{BaseURL: "http://localhost:3003/v1"}

	client := NewClient(backend, cfg)

	assert.NotNil(t, client.backend)
}

// ---------------------------------------------------------------------------
// Service interface fields are wired
// ---------------------------------------------------------------------------

func TestNewClientServiceFieldsAreWired(t *testing.T) {
	t.Parallel()

	backend := &fakeBackend{}
	cfg := Config{BaseURL: "http://localhost:3003/v1"}

	client := NewClient(backend, cfg)

	// All service implementations are wired during construction.
	assert.NotNil(t, client.Rules, "Rules service should be wired")
	assert.NotNil(t, client.Limits, "Limits service should be wired")
	assert.NotNil(t, client.Validations, "Validations service should be wired")
	assert.NotNil(t, client.AuditEvents, "AuditEvents service should be wired")
}

// ---------------------------------------------------------------------------
// Option functions
// ---------------------------------------------------------------------------

func TestWithBaseURL(t *testing.T) {
	t.Parallel()

	var cfg Config

	err := WithBaseURL("http://example.com/v1")(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/v1", cfg.BaseURL)
}

func TestWithClientCredentials(t *testing.T) {
	t.Parallel()

	var cfg Config

	err := WithClientCredentials("client-id", "client-secret", "https://auth.example.com/token")(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "client-id", cfg.ClientID)
	assert.Equal(t, "client-secret", cfg.ClientSecret)
	assert.Equal(t, "https://auth.example.com/token", cfg.TokenURL)
}

func TestWithScopes(t *testing.T) {
	t.Parallel()

	var cfg Config

	err := WithScopes("tracer:read", "tracer:write")(&cfg)
	require.NoError(t, err)
	assert.Equal(t, []string{"tracer:read", "tracer:write"}, cfg.Scopes)
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	var cfg Config

	err := WithTimeout(30 * time.Second)(&cfg)
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}

func TestOptionsComposeCorrectly(t *testing.T) {
	t.Parallel()

	var cfg Config

	opts := []Option{
		WithBaseURL("http://tracer.local/v1"),
		WithClientCredentials("client-id", "client-secret", "https://auth.example.com/token"),
		WithScopes("tracer:read"),
		WithTimeout(20 * time.Second),
	}

	for _, opt := range opts {
		err := opt(&cfg)
		require.NoError(t, err)
	}

	assert.Equal(t, "http://tracer.local/v1", cfg.BaseURL)
	assert.Equal(t, "client-id", cfg.ClientID)
	assert.Equal(t, "client-secret", cfg.ClientSecret)
	assert.Equal(t, "https://auth.example.com/token", cfg.TokenURL)
	assert.Equal(t, []string{"tracer:read"}, cfg.Scopes)
	assert.Equal(t, 20*time.Second, cfg.Timeout)
}

func TestOptionsLastWins(t *testing.T) {
	t.Parallel()

	var cfg Config

	opts := []Option{
		WithBaseURL("http://first.com"),
		WithBaseURL("http://second.com"),
	}

	for _, opt := range opts {
		err := opt(&cfg)
		require.NoError(t, err)
	}

	assert.Equal(t, "http://second.com", cfg.BaseURL)
}

// ---------------------------------------------------------------------------
// Config credential redaction — String()
// ---------------------------------------------------------------------------

func TestConfigStringRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		BaseURL:      "http://localhost:3003/v1",
		ClientID:     "client-id",
		ClientSecret: "super-secret-client-secret",
		TokenURL:     "https://auth.example.com/token",
		Scopes:       []string{"tracer:read"},
		Timeout:      10 * time.Second,
	}

	s := cfg.String()

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "TracerConfig")
	assert.Contains(t, s, "http://localhost:3003/v1", "BaseURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.Contains(t, s, "10s", "Timeout should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"String() must not contain the actual client secret")
}

// ---------------------------------------------------------------------------
// Config credential redaction — MarshalJSON()
// ---------------------------------------------------------------------------

func TestConfigMarshalJSONRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		BaseURL:      "http://localhost:3003/v1",
		ClientID:     "client-id",
		ClientSecret: "super-secret-client-secret",
		TokenURL:     "https://auth.example.com/token",
		Scopes:       []string{"tracer:read"},
		Timeout:      10 * time.Second,
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	s := string(data)

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "http://localhost:3003/v1", "BaseURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"MarshalJSON must not contain the actual client secret")
}
