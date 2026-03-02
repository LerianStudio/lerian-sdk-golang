package matcher

import (
	"encoding/json"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// matcherErrorResponse is the wire format returned by the Matcher API for
// error responses. It extends the standard code/message pair with an
// optional entityType field that identifies the resource kind involved
// in the error (e.g., "Context", "Rule", "Exception").
type matcherErrorResponse struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	EntityType string `json:"entityType,omitempty"`
}

// ParseError converts a raw Matcher HTTP error response into a structured
// SDK [sdkerrors.Error]. It is injected into the [core.BackendConfig] as
// the ErrorParser, giving the backend product-aware error classification.
//
// The Matcher API returns a JSON body with "code", "message", and an
// optional "entityType" field. The entityType is mapped to the Error's
// Resource field so callers can inspect which entity kind triggered the
// failure.
//
// If the response body cannot be parsed as JSON, the raw body is used as
// the error message with the appropriate category derived from the status code.
func ParseError(statusCode int, body []byte) *sdkerrors.Error {
	var resp matcherErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return &sdkerrors.Error{
			Product:    "matcher",
			Category:   sdkerrors.CategoryFromStatus(statusCode),
			Message:    sdkerrors.TruncateBody(body),
			StatusCode: statusCode,
		}
	}

	return &sdkerrors.Error{
		Product:    "matcher",
		Category:   sdkerrors.CategoryFromStatus(statusCode),
		Code:       sdkerrors.ErrorCode(resp.Code),
		Message:    resp.Message,
		Resource:   resp.EntityType,
		StatusCode: statusCode,
	}
}
