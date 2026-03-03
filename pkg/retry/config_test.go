package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, 3, cfg.MaxRetries, "MaxRetries should default to 3")
	assert.Equal(t, 500*time.Millisecond, cfg.BaseDelay, "BaseDelay should default to 500ms")
	assert.Equal(t, 30*time.Second, cfg.MaxDelay, "MaxDelay should default to 30s")
	assert.InDelta(t, 0.25, cfg.JitterRatio, 1e-9, "JitterRatio should default to 0.25")
}

func TestDefaultConfigMaxRetriesIsBelowBackendCap(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	// The backend caps MaxRetries at 10 (core.MaxRetriesLimit).
	// Ensure the default is well below that limit to avoid surprises.
	assert.LessOrEqual(t, cfg.MaxRetries, 10,
		"default MaxRetries should be at or below the backend cap of 10")
}
