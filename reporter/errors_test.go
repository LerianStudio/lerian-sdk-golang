package reporter

import (
	"errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ParseError — table-driven tests
// ---------------------------------------------------------------------------

func TestParseError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           []byte
		wantCategory   sdkerrors.ErrorCategory
		wantCode       sdkerrors.ErrorCode
		wantMessage    string
		wantProduct    string
		wantStatusCode int
	}{
		{
			name:       "400 validation with JSON body",
			statusCode: 400,
			body:       []byte(`{"code":"invalid_format","message":"report format must be pdf, csv, or xlsx"}`),
			wantCategory:   sdkerrors.CategoryValidation,
			wantCode:       sdkerrors.ErrorCode("invalid_format"),
			wantMessage:    "report format must be pdf, csv, or xlsx",
			wantProduct:    "reporter",
			wantStatusCode: 400,
		},
		{
			name:       "401 authentication with JSON body",
			statusCode: 401,
			body:       []byte(`{"code":"token_expired","message":"authentication token has expired"}`),
			wantCategory:   sdkerrors.CategoryAuthentication,
			wantCode:       sdkerrors.ErrorCode("token_expired"),
			wantMessage:    "authentication token has expired",
			wantProduct:    "reporter",
			wantStatusCode: 401,
		},
		{
			name:       "403 authorization with JSON body",
			statusCode: 403,
			body:       []byte(`{"code":"insufficient_permissions","message":"you do not have access to this report"}`),
			wantCategory:   sdkerrors.CategoryAuthorization,
			wantCode:       sdkerrors.ErrorCode("insufficient_permissions"),
			wantMessage:    "you do not have access to this report",
			wantProduct:    "reporter",
			wantStatusCode: 403,
		},
		{
			name:       "404 not found with JSON body",
			statusCode: 404,
			body:       []byte(`{"code":"report_not_found","message":"report rpt-999 does not exist"}`),
			wantCategory:   sdkerrors.CategoryNotFound,
			wantCode:       sdkerrors.ErrorCode("report_not_found"),
			wantMessage:    "report rpt-999 does not exist",
			wantProduct:    "reporter",
			wantStatusCode: 404,
		},
		{
			name:       "409 conflict with JSON body",
			statusCode: 409,
			body:       []byte(`{"code":"duplicate_name","message":"a report with this name already exists"}`),
			wantCategory:   sdkerrors.CategoryConflict,
			wantCode:       sdkerrors.ErrorCode("duplicate_name"),
			wantMessage:    "a report with this name already exists",
			wantProduct:    "reporter",
			wantStatusCode: 409,
		},
		{
			name:       "429 rate limit with JSON body",
			statusCode: 429,
			body:       []byte(`{"code":"rate_exceeded","message":"too many requests, try again later"}`),
			wantCategory:   sdkerrors.CategoryRateLimit,
			wantCode:       sdkerrors.ErrorCode("rate_exceeded"),
			wantMessage:    "too many requests, try again later",
			wantProduct:    "reporter",
			wantStatusCode: 429,
		},
		{
			name:       "500 internal server error with JSON body",
			statusCode: 500,
			body:       []byte(`{"code":"internal_error","message":"an unexpected error occurred"}`),
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       sdkerrors.ErrorCode("internal_error"),
			wantMessage:    "an unexpected error occurred",
			wantProduct:    "reporter",
			wantStatusCode: 500,
		},
		{
			name:       "502 bad gateway maps to internal",
			statusCode: 502,
			body:       []byte(`{"code":"upstream_error","message":"upstream service unavailable"}`),
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       sdkerrors.ErrorCode("upstream_error"),
			wantMessage:    "upstream service unavailable",
			wantProduct:    "reporter",
			wantStatusCode: 502,
		},
		{
			name:       "503 service unavailable maps to internal",
			statusCode: 503,
			body:       []byte(`{"code":"service_down","message":"service temporarily unavailable"}`),
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       sdkerrors.ErrorCode("service_down"),
			wantMessage:    "service temporarily unavailable",
			wantProduct:    "reporter",
			wantStatusCode: 503,
		},
		{
			name:       "invalid JSON body falls back to raw message",
			statusCode: 500,
			body:       []byte(`this is not json`),
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       "",
			wantMessage:    "this is not json",
			wantProduct:    "reporter",
			wantStatusCode: 500,
		},
		{
			name:       "empty body falls back to empty message",
			statusCode: 404,
			body:       []byte(``),
			wantCategory:   sdkerrors.CategoryNotFound,
			wantCode:       "",
			wantMessage:    "",
			wantProduct:    "reporter",
			wantStatusCode: 404,
		},
		{
			name:       "HTML body treated as invalid JSON",
			statusCode: 502,
			body:       []byte(`<html><body>Bad Gateway</body></html>`),
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       "",
			wantMessage:    "<html><body>Bad Gateway</body></html>",
			wantProduct:    "reporter",
			wantStatusCode: 502,
		},
		{
			name:       "unknown 4xx status defaults to internal category",
			statusCode: 418,
			body:       []byte(`{"code":"teapot","message":"I am a teapot"}`),
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       sdkerrors.ErrorCode("teapot"),
			wantMessage:    "I am a teapot",
			wantProduct:    "reporter",
			wantStatusCode: 418,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sdkErr := ParseError(tc.statusCode, tc.body)
			require.NotNil(t, sdkErr)

			assert.Equal(t, tc.wantProduct, sdkErr.Product)
			assert.Equal(t, tc.wantCategory, sdkErr.Category)
			assert.Equal(t, tc.wantCode, sdkErr.Code)
			assert.Equal(t, tc.wantMessage, sdkErr.Message)
			assert.Equal(t, tc.wantStatusCode, sdkErr.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseError — errors.Is compatibility
// ---------------------------------------------------------------------------

func TestParseErrorSentinelMatching(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		sentinel   *sdkerrors.Error
	}{
		{"400 matches ErrValidation", 400, sdkerrors.ErrValidation},
		{"401 matches ErrAuthentication", 401, sdkerrors.ErrAuthentication},
		{"403 matches ErrAuthorization", 403, sdkerrors.ErrAuthorization},
		{"404 matches ErrNotFound", 404, sdkerrors.ErrNotFound},
		{"409 matches ErrConflict", 409, sdkerrors.ErrConflict},
		{"429 matches ErrRateLimit", 429, sdkerrors.ErrRateLimit},
		{"500 matches ErrInternal", 500, sdkerrors.ErrInternal},
	}

	body := []byte(`{"code":"test","message":"test error"}`)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sdkErr := ParseError(tc.statusCode, body)
			assert.True(t, errors.Is(sdkErr, tc.sentinel),
				"ParseError(%d) should match %v via errors.Is", tc.statusCode, tc.sentinel.Category)
		})
	}
}

// ---------------------------------------------------------------------------
// categoryFromStatus — exhaustive coverage
// ---------------------------------------------------------------------------

func TestCategoryFromStatus(t *testing.T) {
	tests := []struct {
		status   int
		expected sdkerrors.ErrorCategory
	}{
		{400, sdkerrors.CategoryValidation},
		{401, sdkerrors.CategoryAuthentication},
		{403, sdkerrors.CategoryAuthorization},
		{404, sdkerrors.CategoryNotFound},
		{409, sdkerrors.CategoryConflict},
		{429, sdkerrors.CategoryRateLimit},
		{500, sdkerrors.CategoryInternal},
		{502, sdkerrors.CategoryInternal},
		{503, sdkerrors.CategoryInternal},
		{418, sdkerrors.CategoryInternal}, // unknown defaults to internal
		{422, sdkerrors.CategoryInternal}, // unprocessable entity also defaults
	}

	for _, tc := range tests {
		got := sdkerrors.CategoryFromStatus(tc.status)
		assert.Equal(t, tc.expected, got, "categoryFromStatus(%d)", tc.status)
	}
}
