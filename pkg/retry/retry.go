package retry

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// IsRetryable reports whether the given HTTP status code is considered
// transient and therefore eligible for retry.
//
// Retryable codes:
//   - 429 (Too Many Requests)
//   - Any 5xx (server error)
func IsRetryable(statusCode int) bool {
	return statusCode == 429 || statusCode >= 500
}

// Do executes fn with retry logic governed by cfg.
//
// It calls fn up to cfg.MaxRetries+1 times. If fn returns nil on any
// attempt, Do returns nil immediately. Between attempts, it sleeps
// using exponential backoff with jitter, respecting context cancellation.
//
// If the context is cancelled during a backoff sleep, Do returns
// ctx.Err(). If all attempts are exhausted, Do returns the last
// error from fn.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Don't sleep after the final attempt.
		if attempt == cfg.MaxRetries {
			break
		}

		delay := CalculateDelay(cfg, attempt)

		select {
		case <-time.After(delay):
			// Backoff elapsed, proceed to next attempt.
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// CalculateDelay computes the backoff duration for the given attempt.
//
// The base delay grows exponentially: cfg.BaseDelay * 2^attempt.
// It is then capped at cfg.MaxDelay and randomised by jitter:
//
//	delay = min(base * 2^attempt, maxDelay) * (1.0 + rand * jitterRatio)
//
// This function is also used by [core.BackendImpl] to compute its HTTP
// retry backoff, ensuring a single source of truth for the delay algorithm.
func CalculateDelay(cfg Config, attempt int) time.Duration {
	// Exponential backoff: BaseDelay * 2^attempt.
	shift := math.Pow(2, float64(attempt))
	base := time.Duration(float64(cfg.BaseDelay) * shift)

	// Cap at MaxDelay.
	if base > cfg.MaxDelay {
		base = cfg.MaxDelay
	}

	// Apply jitter: multiply by (1.0 + rand * JitterRatio).
	jitter := 1.0 + rand.Float64()*cfg.JitterRatio //nolint:gosec // jitter does not need crypto rand
	delay := time.Duration(float64(base) * jitter)

	return delay
}
