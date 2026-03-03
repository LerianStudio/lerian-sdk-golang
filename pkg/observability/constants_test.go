package observability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Attribute key namespace verification
// ---------------------------------------------------------------------------

func TestAttributeKeyConstants(t *testing.T) {
	t.Parallel()

	keys := []struct {
		name  string
		value string
	}{
		{"KeySDKVersion", KeySDKVersion},
		{"KeySDKLanguage", KeySDKLanguage},
		{"KeyProduct", KeyProduct},
		{"KeyOperationName", KeyOperationName},
		{"KeyOperationType", KeyOperationType},
		{"KeyResourceType", KeyResourceType},
		{"KeyResourceID", KeyResourceID},
	}

	for _, k := range keys {
		t.Run(k.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, strings.HasPrefix(k.value, "lerian."),
				"attribute key %s (%q) must start with 'lerian.'", k.name, k.value)
		})
	}
}

// ---------------------------------------------------------------------------
// Metric name namespace verification
// ---------------------------------------------------------------------------

func TestMetricNameConstants(t *testing.T) {
	t.Parallel()

	metrics := []struct {
		name  string
		value string
	}{
		{"MetricRequestTotal", MetricRequestTotal},
		{"MetricRequestDuration", MetricRequestDuration},
		{"MetricRequestErrorTotal", MetricRequestErrorTotal},
	}

	for _, m := range metrics {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, strings.HasPrefix(m.value, "lerian.sdk."),
				"metric %s (%q) must start with 'lerian.sdk.'", m.name, m.value)
		})
	}
}

// ---------------------------------------------------------------------------
// No "midaz" references in source files
// ---------------------------------------------------------------------------

func TestNoMidazReferences(t *testing.T) {
	t.Parallel()

	// Walk every .go source file in the package directory and ensure none
	// of them contain the string "midaz" (case-insensitive). Test files
	// are excluded because the assertion message itself mentions the word.
	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	for _, e := range entries {
		name := e.Name()

		// Only check Go source files, skip test files.
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data, rErr := os.ReadFile(filepath.Clean(name))
			require.NoError(t, rErr)

			content := strings.ToLower(string(data))
			assert.NotContains(t, content, "midaz",
				"source file %s must not contain any 'midaz' references", name)
		})
	}
}
