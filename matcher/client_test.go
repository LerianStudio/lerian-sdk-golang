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
	cfg := Config{
		BaseURL: "http://localhost:3002/v1",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	client := NewClient(&fakeBackend{}, cfg)

	require.NotNil(t, client)
	assert.Equal(t, "http://localhost:3002/v1", client.config.BaseURL)
	assert.Equal(t, "test-key", client.config.APIKey)
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
	t.Run("WithBaseURL", func(t *testing.T) {
		var cfg Config
		err := WithBaseURL("https://matcher.example.com/v1")(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "https://matcher.example.com/v1", cfg.BaseURL)
	})

	t.Run("WithAPIKey", func(t *testing.T) {
		var cfg Config
		err := WithAPIKey("api-key-123")(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "api-key-123", cfg.APIKey)
	})

	t.Run("WithTimeout", func(t *testing.T) {
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
	cfg := Config{
		BaseURL: "http://localhost:3002/v1",
		APIKey:  "super-secret-api-key",
		Timeout: 30 * time.Second,
	}

	s := cfg.String()

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "MatcherConfig")
	assert.Contains(t, s, "http://localhost:3002/v1", "BaseURL should be visible")
	assert.Contains(t, s, "30s", "Timeout should be visible")
	assert.NotContains(t, s, "super-secret-api-key",
		"String() must not contain the actual API key")
}

// ---------------------------------------------------------------------------
// Config credential redaction — MarshalJSON()
// ---------------------------------------------------------------------------

func TestConfigMarshalJSONRedaction(t *testing.T) {
	cfg := Config{
		BaseURL: "http://localhost:3002/v1",
		APIKey:  "super-secret-api-key",
		Timeout: 30 * time.Second,
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	s := string(data)

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "http://localhost:3002/v1", "BaseURL should be visible")
	assert.NotContains(t, s, "super-secret-api-key",
		"MarshalJSON must not contain the actual API key")
}

// ---------------------------------------------------------------------------
// ErrorParser returns a function
// ---------------------------------------------------------------------------

func TestErrorParserReturnsFunction(t *testing.T) {
	parser := ErrorParser()
	require.NotNil(t, parser)

	// Quick smoke test to ensure it produces a valid error.
	err := parser(500, []byte(`{"code":"ERR","message":"boom"}`))
	require.NotNil(t, err)
	assert.Equal(t, "matcher", err.Product)
}
