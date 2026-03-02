package reporter

import (
	"encoding/json"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// reporterErrorResponse mirrors the JSON error body returned by the Reporter API.
type reporterErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ParseError converts a raw HTTP error response from the Reporter API into
// a structured [sdkerrors.Error]. It attempts to unmarshal the body as a
// JSON object with "code" and "message" fields. If unmarshalling fails
// (e.g. the body is plain text or HTML), the raw body is used as the
// message and no code is set.
//
// ParseError is wired into the [core.BackendImpl] error parser so that
// every HTTP 4xx/5xx response is automatically translated into the SDK
// error taxonomy.
func ParseError(statusCode int, body []byte) *sdkerrors.Error {
	var resp reporterErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return &sdkerrors.Error{
			Product:    "reporter",
			Category:   sdkerrors.CategoryFromStatus(statusCode),
			Message:    sdkerrors.TruncateBody(body),
			StatusCode: statusCode,
		}
	}

	return &sdkerrors.Error{
		Product:    "reporter",
		Category:   sdkerrors.CategoryFromStatus(statusCode),
		Code:       sdkerrors.ErrorCode(resp.Code),
		Message:    resp.Message,
		StatusCode: statusCode,
	}
}
