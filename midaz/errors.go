package midaz

import (
	"encoding/json"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// ---------------------------------------------------------------------------
// Midaz-specific error codes
// ---------------------------------------------------------------------------

// Midaz API error codes. These correspond to the "code" field in the Midaz
// JSON error response body, enabling programmatic handling of specific
// failure conditions.
const (
	CodeNotFound            sdkerrors.ErrorCode = "0040"
	CodeAlreadyExists       sdkerrors.ErrorCode = "0005"
	CodeInsufficientBalance sdkerrors.ErrorCode = "0029"
	CodeAssetMismatch       sdkerrors.ErrorCode = "0036"
	CodeValidationError     sdkerrors.ErrorCode = "0003"
	CodeAuthenticationError sdkerrors.ErrorCode = "0013"
)

// ---------------------------------------------------------------------------
// Error response wire type
// ---------------------------------------------------------------------------

// midazErrorResponse mirrors the Midaz API error response JSON shape.
type midazErrorResponse struct {
	Code    string `json:"code"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// ---------------------------------------------------------------------------
// ParseError
// ---------------------------------------------------------------------------

// ParseError converts a Midaz HTTP error response body into a structured
// [*sdkerrors.Error]. It is injected into the [core.BackendImpl] as the
// ErrorParser so that all Midaz API errors are automatically classified.
//
// If the body is not valid JSON, the raw body string is used as the error
// message with no product-specific code.
func ParseError(statusCode int, body []byte) *sdkerrors.Error {
	var resp midazErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return &sdkerrors.Error{
			Product:    "midaz",
			Category:   sdkerrors.CategoryFromStatus(statusCode),
			Message:    sdkerrors.TruncateBody(body),
			StatusCode: statusCode,
		}
	}

	return &sdkerrors.Error{
		Product:    "midaz",
		Category:   sdkerrors.CategoryFromStatus(statusCode),
		Code:       sdkerrors.ErrorCode(resp.Code),
		Message:    resp.Message,
		StatusCode: statusCode,
	}
}
