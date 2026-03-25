package core

import (
	"context"
	"errors"
	"net/http"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

func (b *BackendImpl) classifyNetworkError(ctx context.Context, operation string, err error) *sdkerrors.Error {
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return sdkerrors.NewCancellation("sdk", operation, ctx.Err())
		}

		return sdkerrors.NewTimeout("sdk", operation, ctx.Err())
	}

	return sdkerrors.NewNetwork("sdk", operation, err)
}

func (b *BackendImpl) buildHTTPError(statusCode int, body []byte, requestID, operation string) *sdkerrors.Error {
	if b.errorParser != nil {
		sdkErr := b.errorParser(statusCode, body)
		if sdkErr != nil {
			if sdkErr.RequestID == "" {
				sdkErr.RequestID = requestID
			}

			if sdkErr.Operation == "" {
				sdkErr.Operation = operation
			}

			return sdkErr
		}
	}

	return b.genericHTTPError(statusCode, body, requestID, operation)
}

func (b *BackendImpl) genericHTTPError(statusCode int, body []byte, requestID, operation string) *sdkerrors.Error {
	message := http.StatusText(statusCode)
	if len(body) > 0 {
		message = string(body)
		if len(message) > sdkerrors.MaxErrorBodyBytes {
			message = message[:sdkerrors.MaxErrorBodyBytes] + "... [truncated]"
		}
	}

	sdkErr := &sdkerrors.Error{
		Product:    "sdk",
		Operation:  operation,
		StatusCode: statusCode,
		RequestID:  requestID,
		Message:    message,
	}

	switch {
	case statusCode == http.StatusBadRequest:
		sdkErr.Category = sdkerrors.CategoryValidation
	case statusCode == http.StatusUnauthorized:
		sdkErr.Category = sdkerrors.CategoryAuthentication
	case statusCode == http.StatusForbidden:
		sdkErr.Category = sdkerrors.CategoryAuthorization
	case statusCode == http.StatusNotFound:
		sdkErr.Category = sdkerrors.CategoryNotFound
	case statusCode == http.StatusConflict:
		sdkErr.Category = sdkerrors.CategoryConflict
	case statusCode == http.StatusTooManyRequests:
		sdkErr.Category = sdkerrors.CategoryRateLimit
	case statusCode >= 500:
		sdkErr.Category = sdkerrors.CategoryInternal
	default:
		sdkErr.Category = sdkerrors.CategoryInternal
	}

	return sdkErr
}
