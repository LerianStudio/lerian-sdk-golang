// Package retry provides retry logic with exponential backoff and jitter
// for transient HTTP failures.
//
// The default policy retries on 429 (Too Many Requests) and 5xx status
// codes with decorrelated-jitter exponential backoff. Callers may supply a
// custom [Config] to override the maximum number of attempts and the
// backoff timing parameters.
//
// # Configuration
//
// The [Config] struct controls retry behaviour:
//
//   - MaxRetries -- maximum retry attempts after the initial call (default 3)
//   - InitialBackoff -- base delay before the first retry (default 500ms)
//   - MaxBackoff -- ceiling for backoff duration (default 30s)
//
// # Usage
//
// Retry configuration is typically set via functional options on the
// umbrella client:
//
//	client, _ := lerian.New(
//	    lerian.WithRetry(retry.Config{
//	        MaxRetries:     5,
//	        InitialBackoff: time.Second,
//	        MaxBackoff:     time.Minute,
//	    }),
//	    lerian.WithMidaz(...),
//	)
//
// The retry logic is applied transparently by [core.Backend] and does not
// require any action from individual service callers.
package retry
