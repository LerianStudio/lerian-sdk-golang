package core

import "net/http"

func (b *BackendImpl) logRequest(method, path string, req *http.Request) {
	if !b.debug {
		return
	}

	attrs := []any{"method", method, "path", path}
	if authVal := req.Header.Get("Authorization"); authVal != "" {
		attrs = append(attrs, "authorization", maskHeader(authVal))
	}

	b.logger.Debug("sdk request", attrs...)
}

func (b *BackendImpl) logResponse(statusCode int, path string) {
	if !b.debug {
		return
	}

	b.logger.Debug("sdk response", "status", statusCode, "path", path)
}

func maskHeader(value string) string {
	const visibleChars = 4
	if len(value) <= visibleChars {
		return value + "***"
	}

	return value[:visibleChars] + "***"
}

// MaskAuthorizationHeader is an exported wrapper around the internal
// maskHeader function, useful for custom debug middleware.
func MaskAuthorizationHeader(value string) string {
	return maskHeader(value)
}
