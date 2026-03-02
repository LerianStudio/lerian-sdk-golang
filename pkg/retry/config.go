package retry

import "time"

// Config holds retry parameters.
type Config struct {
	// MaxRetries is the maximum number of retry attempts after the initial call.
	// A value of 3 means 4 total attempts (1 initial + 3 retries).
	MaxRetries int

	// BaseDelay is the initial delay before the first retry.
	// Subsequent retries use exponential backoff: BaseDelay * 2^attempt.
	BaseDelay time.Duration

	// MaxDelay caps the computed backoff so that no single sleep exceeds
	// this duration (before jitter is applied).
	MaxDelay time.Duration

	// JitterRatio controls the random jitter added to each delay.
	// The actual delay is multiplied by (1.0 + rand * JitterRatio),
	// so a ratio of 0.25 means the delay varies by up to +25%.
	JitterRatio float64
}

// DefaultConfig returns the default retry configuration.
//
// Defaults:
//   - MaxRetries:  3          (4 total attempts)
//   - BaseDelay:   500ms
//   - MaxDelay:    30s
//   - JitterRatio: 0.25       (+0-25% random jitter)
func DefaultConfig() Config {
	return Config{
		MaxRetries:  3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    30 * time.Second,
		JitterRatio: 0.25,
	}
}
