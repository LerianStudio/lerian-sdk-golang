package lerian

import (
	"fmt"
	"net/http"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/reporter"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
)

// Option configures the Lerian SDK [Client]. Options are applied in order
// during [New]; later options override earlier ones.
type Option func(*Client) error

// ---------------------------------------------------------------------------
// Product configuration options
// ---------------------------------------------------------------------------

// WithMidaz configures the Midaz product client. The supplied options are
// deferred until [New] validates and constructs the Midaz backends.
//
// Example:
//
//	lerian.WithMidaz(
//	    midaz.WithOnboardingURL("http://localhost:3000/v1"),
//	    midaz.WithTransactionURL("http://localhost:3001/v1"),
//	    midaz.WithAuthToken("my-token"),
//	)
func WithMidaz(opts ...midaz.Option) Option {
	return func(c *Client) error {
		c.midazRequested = true
		c.midazOpts = append(c.midazOpts, opts...)

		return nil
	}
}

// WithMatcher configures the Matcher product client. The supplied options are
// deferred until [New] validates and constructs the Matcher backend.
func WithMatcher(opts ...matcher.Option) Option {
	return func(c *Client) error {
		c.matcherRequested = true
		c.matcherOpts = append(c.matcherOpts, opts...)

		return nil
	}
}

// WithTracer configures the Tracer product client. The supplied options are
// deferred until [New] validates and constructs the Tracer backend.
func WithTracer(opts ...tracer.Option) Option {
	return func(c *Client) error {
		c.tracerRequested = true
		c.tracerOpts = append(c.tracerOpts, opts...)

		return nil
	}
}

// WithReporter configures the Reporter product client. The supplied options
// are deferred until [New] validates and constructs the Reporter backend.
func WithReporter(opts ...reporter.Option) Option {
	return func(c *Client) error {
		c.reporterRequested = true
		c.reporterOpts = append(c.reporterOpts, opts...)

		return nil
	}
}

// WithFees configures the Fees product client. The supplied options are
// deferred until [New] validates and constructs the Fees backend.
func WithFees(opts ...fees.Option) Option {
	return func(c *Client) error {
		c.feesRequested = true
		c.feesOpts = append(c.feesOpts, opts...)

		return nil
	}
}

// ---------------------------------------------------------------------------
// Shared infrastructure options
// ---------------------------------------------------------------------------

// WithDebug enables or disables verbose request/response logging across all
// product clients. Authorization header values are masked in debug output.
//
// Calling WithDebug explicitly (even with false) takes precedence over the
// LERIAN_DEBUG environment variable, ensuring callers can suppress debug
// output regardless of the runtime environment.
func WithDebug(debug bool) Option {
	return func(c *Client) error {
		c.debug = debug
		c.debugExplicit = true

		return nil
	}
}

// WithRetry configures the retry policy for all product clients. The
// retryConfig uses exponential backoff with jitter; maxRetries=0 disables
// retries entirely.
func WithRetry(maxRetries int, baseDelay time.Duration) Option {
	return func(c *Client) error {
		c.retryConfig.MaxRetries = maxRetries
		c.retryConfig.BaseDelay = baseDelay

		return nil
	}
}

// WithHTTPClient replaces the default HTTP client used by all product
// backends. This is useful for custom TLS configuration, proxies, or
// testing with a fake transport.
//
// Passing nil returns an error; use the zero-value defaults instead.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return fmt.Errorf("lerian: HTTP client must not be nil")
		}

		c.httpClient = client

		return nil
	}
}

// WithObservability enables OpenTelemetry observability pillars. Setting all
// three flags to false is equivalent to noop observability (the default).
//
// When any flag is true, [New] creates an OTel provider that exports
// telemetry to the configured collector endpoint (see [WithCollectorEndpoint]).
func WithObservability(traces, metrics, logs bool) Option {
	return func(c *Client) error {
		c.otelTraces = traces
		c.otelMetrics = metrics
		c.otelLogs = logs

		return nil
	}
}

// WithCollectorEndpoint sets the OTLP collector endpoint URL used when
// observability is enabled. The default is "http://localhost:4318".
func WithCollectorEndpoint(endpoint string) Option {
	return func(c *Client) error {
		c.otelEndpoint = endpoint
		return nil
	}
}
