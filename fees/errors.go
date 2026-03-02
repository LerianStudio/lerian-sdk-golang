package fees

import (
	"encoding/json"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// feesErrorResponse is the wire format returned by the Fees API for error
// responses. Codes follow the FEE-XXXX convention.
type feesErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ParseError converts a raw Fees HTTP error response into a structured
// SDK [sdkerrors.Error]. It is injected into the [core.BackendConfig] as
// the ErrorParser, giving the backend product-aware error classification.
//
// If the response body cannot be parsed as JSON, the raw body is used as
// the error message with the appropriate category derived from the status code.
func ParseError(statusCode int, body []byte) *sdkerrors.Error {
	var resp feesErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return &sdkerrors.Error{
			Product:    "fees",
			Category:   sdkerrors.CategoryFromStatus(statusCode),
			Message:    sdkerrors.TruncateBody(body),
			StatusCode: statusCode,
		}
	}

	return &sdkerrors.Error{
		Product:    "fees",
		Category:   sdkerrors.CategoryFromStatus(statusCode),
		Code:       sdkerrors.ErrorCode(resp.Code),
		Message:    resp.Message,
		StatusCode: statusCode,
	}
}
