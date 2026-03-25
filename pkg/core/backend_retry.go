package core

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"
)

func (b *BackendImpl) backoffSleep(ctx context.Context, attempt int, retryAfter time.Duration) error {
	delay := retry.CalculateDelay(b.retryConfig, attempt)
	if retryAfter > delay {
		delay = retryAfter
	}

	if delay > b.retryConfig.MaxDelay {
		delay = b.retryConfig.MaxDelay
	}

	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	if t, err := http.ParseTime(value); err == nil {
		delay := time.Until(t)
		if delay > 0 {
			return delay
		}
	}

	return 0
}
