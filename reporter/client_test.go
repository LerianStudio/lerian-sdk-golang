package reporter

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
// fakeBackend — minimal Backend for unit tests
// ---------------------------------------------------------------------------

// fakeBackend is a trivial core.Backend implementation that satisfies the
// interface without performing any HTTP calls. It is used only to verify
// that NewClient correctly assigns its fields.
type fakeBackend struct{}

func (f *fakeBackend) Do(_ context.Context, _ core.Request) (*core.Response, error) {
	return &core.Response{}, nil
}

// Compile-time check.
var _ core.Backend = (*fakeBackend)(nil)

// ---------------------------------------------------------------------------
// NewClient tests
// ---------------------------------------------------------------------------

func TestNewClientBasic(t *testing.T) {
	t.Parallel()

	cfg := Config{
		BaseURL:        "http://localhost:3004/v1",
		ClientID:       "client-id",
		ClientSecret:   "client-secret",
		TokenURL:       "https://auth.example.com/token",
		OrganizationID: "org-123",
	}

	client := NewClient(&fakeBackend{}, cfg)
	require.NotNil(t, client)

	// Service accessors are wired.
	assert.NotNil(t, client.DataSources)
	assert.NotNil(t, client.Reports)
	assert.NotNil(t, client.Templates)
}

func TestNewClientZeroConfig(t *testing.T) {
	t.Parallel()

	client := NewClient(&fakeBackend{}, Config{})
	require.NotNil(t, client)
	assert.NotNil(t, client.DataSources)
	assert.NotNil(t, client.Reports)
	assert.NotNil(t, client.Templates)
}

// ---------------------------------------------------------------------------
// Option tests — verify functional options work correctly
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

func TestWithOrganizationID(t *testing.T) {
	t.Parallel()

	var cfg Config

	err := WithOrganizationID("org-abc")(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "org-abc", cfg.OrganizationID)
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
		BaseURL:        "http://localhost:3004/v1",
		ClientID:       "client-id",
		ClientSecret:   "super-secret-client-secret",
		TokenURL:       "https://auth.example.com/token",
		OrganizationID: "org-123",
		Timeout:        30 * time.Second,
	}

	s := cfg.String()

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "ReporterConfig")
	assert.Contains(t, s, "http://localhost:3004/v1", "BaseURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.Contains(t, s, "org-123", "OrganizationID should be visible")
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
		BaseURL:        "http://localhost:3004/v1",
		ClientID:       "client-id",
		ClientSecret:   "super-secret-client-secret",
		TokenURL:       "https://auth.example.com/token",
		OrganizationID: "org-123",
		Timeout:        30 * time.Second,
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	s := string(data)

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "http://localhost:3004/v1", "BaseURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.Contains(t, s, "org-123", "OrganizationID should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"MarshalJSON must not contain the actual client secret")
}
