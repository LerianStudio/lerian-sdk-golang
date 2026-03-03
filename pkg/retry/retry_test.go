package retry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fastConfig returns a Config with very short delays for fast tests.
func fastConfig() Config {
	return Config{
		MaxRetries:  3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		JitterRatio: 0.25,
	}
}

func TestDoSuccessFirstAttempt(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	err := Do(context.Background(), fastConfig(), func() error {
		calls.Add(1)
		return nil
	})

	assert.NoError(t, err, "should succeed without error")
	assert.Equal(t, int32(1), calls.Load(), "fn should be called exactly once")
}

func TestDoRetryThenSuccess(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	errTransient := errors.New("transient failure")

	err := Do(context.Background(), fastConfig(), func() error {
		n := calls.Add(1)
		if n < 3 {
			return errTransient
		}

		return nil
	})

	assert.NoError(t, err, "should succeed after retries")
	assert.Equal(t, int32(3), calls.Load(), "fn should be called 3 times (2 failures + 1 success)")
}

func TestDoMaxRetriesExhausted(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	cfg := fastConfig()
	cfg.MaxRetries = 2

	errPersistent := errors.New("persistent failure")

	err := Do(context.Background(), cfg, func() error {
		calls.Add(1)
		return errPersistent
	})

	assert.ErrorIs(t, err, errPersistent, "should return the last error")
	assert.Equal(t, int32(3), calls.Load(),
		"fn should be called MaxRetries+1 times (1 initial + 2 retries)")
}

func TestDoContextCancellation(t *testing.T) {
	t.Parallel()

	cfg := Config{
		MaxRetries:  10,
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		JitterRatio: 0.0,
	}

	ctx, cancel := context.WithCancel(context.Background())

	var calls atomic.Int32

	// Cancel the context after a brief delay so the retry loop is
	// interrupted during the backoff sleep.
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, cfg, func() error {
		calls.Add(1)
		return errors.New("always fails")
	})

	assert.ErrorIs(t, err, context.Canceled, "should return context.Canceled")
	// With 50ms backoff and 20ms cancel, we expect 1 attempt (the initial
	// call) before the context is cancelled during the sleep.
	assert.Less(t, calls.Load(), int32(cfg.MaxRetries+1),
		"should not exhaust all retries when context is cancelled")
}

func TestIsRetryable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code int
		want bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, IsRetryable(tt.code),
			"IsRetryable(%d) should be %v", tt.code, tt.want)
	}
}

func TestJitterVariation(t *testing.T) {
	t.Parallel()

	cfg := Config{
		MaxRetries:  3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		JitterRatio: 0.5, // 50% jitter for clear variation
	}

	const runs = 20

	delays := make([]time.Duration, runs)
	for i := range runs {
		delays[i] = CalculateDelay(cfg, 1) // attempt 1 so base = 20ms
		_ = i
	}

	// At least two distinct values should appear. With 50% jitter over
	// 20 samples, the probability of all being identical is vanishingly
	// small (< 2^-60).
	allSame := true

	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}

	assert.False(t, allSame, "jitter should produce varying delays across %d calls", runs)
}

func TestCalculateDelayExponentialGrowth(t *testing.T) {
	t.Parallel()

	cfg := Config{
		MaxRetries:  5,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		JitterRatio: 0.0, // no jitter for deterministic checks
	}

	// With zero jitter, delay = BaseDelay * 2^attempt exactly.
	assert.Equal(t, 10*time.Millisecond, CalculateDelay(cfg, 0))  // 10ms * 1
	assert.Equal(t, 20*time.Millisecond, CalculateDelay(cfg, 1))  // 10ms * 2
	assert.Equal(t, 40*time.Millisecond, CalculateDelay(cfg, 2))  // 10ms * 4
	assert.Equal(t, 80*time.Millisecond, CalculateDelay(cfg, 3))  // 10ms * 8
	assert.Equal(t, 160*time.Millisecond, CalculateDelay(cfg, 4)) // 10ms * 16
}

func TestCalculateDelayCapAtMaxDelay(t *testing.T) {
	t.Parallel()

	cfg := Config{
		MaxRetries:  10,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    500 * time.Millisecond,
		JitterRatio: 0.0,
	}

	// attempt 3 => 100ms * 8 = 800ms, capped at 500ms
	d := CalculateDelay(cfg, 3)
	assert.Equal(t, 500*time.Millisecond, d,
		"delay should be capped at MaxDelay")

	// attempt 10 => way above max, still capped
	d = CalculateDelay(cfg, 10)
	assert.Equal(t, 500*time.Millisecond, d,
		"large attempt should still be capped at MaxDelay")
}
