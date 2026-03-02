package tracer

import (
	"encoding/json"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// tracerErrorResponse is the wire format returned by the Tracer API for
// error responses. Codes follow the TRC-XXXX convention.
type tracerErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ParseError converts a raw Tracer HTTP error response into a structured
// SDK [sdkerrors.Error]. It is injected into the [core.BackendConfig] as
// the ErrorParser, giving the backend product-aware error classification.
//
// If the response body cannot be parsed as JSON, the raw body is used as
// the error message with the appropriate category derived from the status code.
func ParseError(statusCode int, body []byte) *sdkerrors.Error {
	var resp tracerErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return &sdkerrors.Error{
			Product:    "tracer",
			Category:   sdkerrors.CategoryFromStatus(statusCode),
			Message:    sdkerrors.TruncateBody(body),
			StatusCode: statusCode,
		}
	}

	return &sdkerrors.Error{
		Product:    "tracer",
		Category:   sdkerrors.CategoryFromStatus(statusCode),
		Code:       sdkerrors.ErrorCode(resp.Code),
		Message:    resp.Message,
		StatusCode: statusCode,
	}
}

// categoryFromStatus delegates to the centralized [sdkerrors.CategoryFromStatus].
// Kept as a thin alias so existing tests continue to compile.
var categoryFromStatus = sdkerrors.CategoryFromStatus
