package matcher

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// fakeBackend satisfies core.Backend for unit tests.
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

// ---------------------------------------------------------------------------
// NewClient tests
// ---------------------------------------------------------------------------

func TestNewClient(t *testing.T) {
	t.Parallel()

	cfg := Config{
		BaseURL:      "http://localhost:3002/v1",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     "https://auth.example.com/token",
		Scopes:       []string{"matcher:read"},
		Timeout:      30 * time.Second,
	}

	client := NewClient(&fakeBackend{}, cfg)

	require.NotNil(t, client)
	assert.Equal(t, "http://localhost:3002/v1", client.config.BaseURL)
	assert.Equal(t, "client-id", client.config.ClientID)
	assert.Equal(t, "client-secret", client.config.ClientSecret)
	assert.Equal(t, "https://auth.example.com/token", client.config.TokenURL)
	assert.Equal(t, []string{"matcher:read"}, client.config.Scopes)
	assert.Equal(t, 30*time.Second, client.config.Timeout)

	// Config service accessors are wired by NewClient.
	assert.NotNil(t, client.Contexts)
	assert.NotNil(t, client.Rules)
	assert.NotNil(t, client.Schedules)
	assert.NotNil(t, client.Sources)
	assert.NotNil(t, client.SourceFieldMaps)
	assert.NotNil(t, client.FeeSchedules)
	assert.NotNil(t, client.FieldMaps)

	assert.NotNil(t, client.Reports)
	assert.NotNil(t, client.Governance)
	assert.NotNil(t, client.Imports)
	assert.NotNil(t, client.Matching)

	// T-21 service accessors are wired by NewClient.
	assert.NotNil(t, client.ExportJobs)
	assert.NotNil(t, client.Disputes)
	assert.NotNil(t, client.Exceptions)
}

// ---------------------------------------------------------------------------
// Option tests
// ---------------------------------------------------------------------------

func TestOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithBaseURL", func(t *testing.T) {
		t.Parallel()

		var cfg Config

		err := WithBaseURL("https://matcher.example.com/v1")(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "https://matcher.example.com/v1", cfg.BaseURL)
	})

	t.Run("WithClientCredentials", func(t *testing.T) {
		t.Parallel()

		var cfg Config

		err := WithClientCredentials("client-id", "client-secret", "https://auth.example.com/token")(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "client-id", cfg.ClientID)
		assert.Equal(t, "client-secret", cfg.ClientSecret)
		assert.Equal(t, "https://auth.example.com/token", cfg.TokenURL)
	})

	t.Run("WithScopes", func(t *testing.T) {
		t.Parallel()

		var cfg Config

		err := WithScopes("matcher:read", "matcher:write")(&cfg)
		require.NoError(t, err)
		assert.Equal(t, []string{"matcher:read", "matcher:write"}, cfg.Scopes)
	})

	t.Run("WithTimeout", func(t *testing.T) {
		t.Parallel()

		var cfg Config

		err := WithTimeout(45 * time.Second)(&cfg)
		require.NoError(t, err)
		assert.Equal(t, 45*time.Second, cfg.Timeout)
	})
}

// ---------------------------------------------------------------------------
// Config credential redaction — String()
// ---------------------------------------------------------------------------

func TestConfigStringRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		BaseURL:      "http://localhost:3002/v1",
		ClientID:     "client-id",
		ClientSecret: "super-secret-client-secret",
		TokenURL:     "https://auth.example.com/token",
		Scopes:       []string{"matcher:read"},
		Timeout:      30 * time.Second,
	}

	s := cfg.String()

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "MatcherConfig")
	assert.Contains(t, s, "http://localhost:3002/v1", "BaseURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
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
		BaseURL:      "http://localhost:3002/v1",
		ClientID:     "client-id",
		ClientSecret: "super-secret-client-secret",
		TokenURL:     "https://auth.example.com/token",
		Scopes:       []string{"matcher:read"},
		Timeout:      30 * time.Second,
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	s := string(data)

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "http://localhost:3002/v1", "BaseURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"MarshalJSON must not contain the actual client secret")
}

// ---------------------------------------------------------------------------
// ErrorParser returns a function
// ---------------------------------------------------------------------------

func TestErrorParserReturnsFunction(t *testing.T) {
	t.Parallel()

	parser := ErrorParser()
	require.NotNil(t, parser)

	// Quick smoke test to ensure it produces a valid error.
	err := parser(500, []byte(`{"code":"ERR","message":"boom"}`))
	require.NotNil(t, err)
	assert.Equal(t, "matcher", err.Product)
}
