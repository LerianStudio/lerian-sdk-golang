package fees

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// fakeBackend — minimal core.Backend stub for unit tests
// ---------------------------------------------------------------------------

type fakeBackend struct{}

func (f *fakeBackend) Call(_ context.Context, _, _ string, _, _ any) error {
	return nil
}

func (f *fakeBackend) CallWithHeaders(_ context.Context, _, _ string, _ map[string]string, _, _ any) error {
	return nil
}

func (f *fakeBackend) CallRaw(_ context.Context, _, _ string, _ any) ([]byte, error) {
	return nil, nil
}

// Compile-time check that fakeBackend satisfies core.Backend.
var _ core.Backend = (*fakeBackend)(nil)

// ---------------------------------------------------------------------------
// NewClient — basic construction tests
// ---------------------------------------------------------------------------

func TestNewClient(t *testing.T) {
	t.Parallel()

	backend := &fakeBackend{}
	cfg := Config{
		BaseURL:        "http://localhost:3005/v1",
		ClientID:       "client-id",
		ClientSecret:   "client-secret",
		TokenURL:       "https://auth.example.com/token",
		OrganizationID: "org-123",
	}

	client := NewClient(backend, cfg)

	require.NotNil(t, client)
	assert.Equal(t, cfg.BaseURL, client.config.BaseURL)
	assert.Equal(t, cfg.ClientID, client.config.ClientID)
	assert.Equal(t, cfg.ClientSecret, client.config.ClientSecret)
	assert.Equal(t, cfg.TokenURL, client.config.TokenURL)
	assert.Equal(t, cfg.OrganizationID, client.config.OrganizationID)
}

func TestNewClientServiceFieldsAreWired(t *testing.T) {
	t.Parallel()

	backend := &fakeBackend{}
	cfg := Config{
		BaseURL:        "http://localhost:3005/v1",
		OrganizationID: "org-123",
	}

	client := NewClient(backend, cfg)

	require.NotNil(t, client)

	// Service implementations are now wired during construction.
	assert.NotNil(t, client.Packages, "Packages should be wired")
	assert.NotNil(t, client.Estimates, "Estimates should be wired")
	assert.NotNil(t, client.Fees, "Fees should be wired")
}

func TestNewClientMinimalConfig(t *testing.T) {
	t.Parallel()

	backend := &fakeBackend{}
	cfg := Config{} // all zero values

	client := NewClient(backend, cfg)

	require.NotNil(t, client)
	assert.Empty(t, client.config.BaseURL)
	assert.Empty(t, client.config.ClientID)
	assert.Empty(t, client.config.ClientSecret)
	assert.Empty(t, client.config.TokenURL)
	assert.Empty(t, client.config.OrganizationID)
	assert.Zero(t, client.config.Timeout)
}

// ---------------------------------------------------------------------------
// Option functions — unit tests
// ---------------------------------------------------------------------------

func TestWithBaseURL(t *testing.T) {
	t.Parallel()

	var cfg Config

	err := WithBaseURL("http://localhost:3005/v1")(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:3005/v1", cfg.BaseURL)
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

func TestWithOrganizationID(t *testing.T) {
	t.Parallel()

	var cfg Config

	err := WithOrganizationID("org-uuid-456")(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "org-uuid-456", cfg.OrganizationID)
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	var cfg Config

	err := WithTimeout(5_000_000_000)(&cfg) // 5 seconds in nanoseconds
	require.NoError(t, err)
	assert.Equal(t, 5_000_000_000, int(cfg.Timeout))
}

// ---------------------------------------------------------------------------
// Config credential redaction — String()
// ---------------------------------------------------------------------------

func TestConfigStringRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		BaseURL:        "http://localhost:3005/v1",
		ClientID:       "client-id",
		ClientSecret:   "super-secret-client-secret",
		TokenURL:       "https://auth.example.com/token",
		OrganizationID: "org-456",
		Timeout:        30 * time.Second,
	}

	s := cfg.String()

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "FeesConfig")
	assert.Contains(t, s, "http://localhost:3005/v1", "BaseURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.Contains(t, s, "org-456", "OrganizationID should be visible")
	assert.Contains(t, s, "30s", "Timeout should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"String() must not contain the actual client secret")
}

// ---------------------------------------------------------------------------
// Config credential redaction — MarshalJSON()
// ---------------------------------------------------------------------------

func TestConfigMarshalJSONRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		BaseURL:        "http://localhost:3005/v1",
		ClientID:       "client-id",
		ClientSecret:   "super-secret-client-secret",
		TokenURL:       "https://auth.example.com/token",
		OrganizationID: "org-456",
		Timeout:        30 * time.Second,
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	s := string(data)

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "http://localhost:3005/v1", "BaseURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.Contains(t, s, "org-456", "OrganizationID should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"MarshalJSON must not contain the actual client secret")
}
