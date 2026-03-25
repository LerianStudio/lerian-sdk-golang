package core

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func secureSDKHTTPClient(base *http.Client) *http.Client {
	if base == nil {
		return &http.Client{Timeout: defaultHTTPTimeout, CheckRedirect: sdkRedirectPolicy(nil)}
	}

	cloned := *base
	cloned.CheckRedirect = sdkRedirectPolicy(base.CheckRedirect)

	return &cloned
}

func sdkRedirectPolicy(next func(*http.Request, []*http.Request) error) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if next != nil {
			if err := next(req, via); err != nil {
				return err
			}
		}

		return stripSensitiveOnRedirect(req, via)
	}
}

func stripSensitiveOnRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("stopped after 10 redirects")
	}

	if len(via) > 0 {
		previous := via[len(via)-1].URL
		hostChanged := !sameAuthority(req.URL, previous)

		downgradedScheme := previous.Scheme == "https" && req.URL.Scheme == "http"
		if hostChanged || downgradedScheme {
			for name := range req.Header {
				if shouldStripOnRedirect(name) {
					req.Header.Del(name)
				}
			}
		}
	}

	return nil
}

func sameAuthority(a, b *url.URL) bool {
	if a == nil || b == nil {
		return false
	}

	return strings.EqualFold(a.Hostname(), b.Hostname()) && effectivePort(a) == effectivePort(b)
}

func effectivePort(targetURL *url.URL) string {
	if targetURL == nil {
		return ""
	}

	if port := targetURL.Port(); port != "" {
		return port
	}

	switch strings.ToLower(targetURL.Scheme) {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}

func shouldStripOnRedirect(name string) bool {
	canonical := http.CanonicalHeaderKey(name)
	lower := strings.ToLower(canonical)

	switch canonical {
	case "Authorization", "Proxy-Authorization", "Cookie", "Cookie2",
		"X-Tenant-Id", "X-Organization-Id", "X-Idempotency-Key",
		"X-Api-Key", "Api-Key", "X-Auth-Token", "X-Amz-Security-Token":
		return true
	}

	return strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "api-key") || strings.Contains(lower, "apikey")
}
