package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (b *BackendImpl) doRequest(ctx context.Context, req Request) (*Response, error) {
	operation := req.Method + " " + req.Path

	ctx, span := b.provider.Tracer().Start(ctx, operation)
	defer span.End()

	var bodyBytes []byte
	if len(req.BodyBytes) > 0 {
		bodyBytes = req.BodyBytes
	} else if req.Body != nil {
		var err error

		bodyBytes, err = b.jsonPool.Marshal(req.Body)
		if err != nil {
			return nil, sdkerrors.NewInternal("sdk", operation, "failed to marshal request body", err)
		}
	}

	var lastErr error

	for attempt := 0; attempt <= b.retryConfig.MaxRetries; attempt++ {
		httpReq, err := b.buildRequest(ctx, req, bodyBytes)
		if err != nil {
			return nil, err
		}

		b.logRequest(req.Method, req.Path, httpReq)

		resp, err := b.httpClient.Do(httpReq)
		if err != nil {
			lastErr = b.classifyNetworkError(ctx, operation, err)
			if ctx.Err() != nil {
				return nil, lastErr
			}

			if attempt < b.retryConfig.MaxRetries {
				if sleepErr := b.backoffSleep(ctx, attempt, 0); sleepErr != nil {
					return nil, b.classifyNetworkError(ctx, operation, sleepErr)
				}

				continue
			}

			return nil, lastErr
		}

		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
		resp.Body.Close()

		if readErr != nil {
			span.SetStatus(codes.Error, "failed to read response body")
			return nil, sdkerrors.NewInternal("sdk", operation, "failed to read response body", readErr)
		}

		b.logResponse(resp.StatusCode, req.Path)
		span.SetAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", b.baseURL+req.Path),
			attribute.Int("http.status_code", resp.StatusCode),
		)

		if resp.StatusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		}

		if resp.StatusCode >= 400 {
			requestID := resp.Header.Get("X-Request-ID")

			var retryAfter time.Duration
			if resp.StatusCode == http.StatusTooManyRequests {
				retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
			}

			if retry.IsRetryable(resp.StatusCode) && attempt < b.retryConfig.MaxRetries {
				lastErr = b.buildHTTPError(resp.StatusCode, respBody, requestID, operation)
				if sleepErr := b.backoffSleep(ctx, attempt, retryAfter); sleepErr != nil {
					return nil, b.classifyNetworkError(ctx, operation, sleepErr)
				}

				continue
			}

			return nil, b.buildHTTPError(resp.StatusCode, respBody, requestID, operation)
		}

		return &Response{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: respBody}, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, sdkerrors.NewInternal("sdk", operation, "all retry attempts exhausted", nil)
}
