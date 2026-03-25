package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func (b *BackendImpl) buildRequest(ctx context.Context, req Request, bodyBytes []byte) (*http.Request, error) {
	operation := req.Method + " " + req.Path
	url := b.baseURL + req.Path

	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, sdkerrors.NewInternal("sdk", operation, "failed to create HTTP request", err)
	}

	if req.Accept != "" {
		httpReq.Header.Set("Accept", req.Accept)
	} else {
		httpReq.Header.Set("Accept", "application/json")
	}

	if bodyBytes != nil {
		contentType := req.ContentType
		if contentType == "" {
			contentType = "application/json"
		}

		httpReq.Header.Set("Content-Type", contentType)
	}

	for k, v := range b.defaultHeaders {
		httpReq.Header.Set(k, v)
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	if tenantID, ok := tenantIDFromContext(ctx); ok {
		httpReq.Header.Set("X-Tenant-ID", tenantID)
	}

	if key, ok := idempotencyKeyFromContext(ctx); ok {
		httpReq.Header.Set("X-Idempotency-Key", key)
	}

	if err := b.auth.Enrich(ctx, httpReq); err != nil {
		return nil, sdkerrors.NewAuthentication("sdk", operation,
			fmt.Sprintf("authentication enrichment failed: %v", err))
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	return httpReq, nil
}
