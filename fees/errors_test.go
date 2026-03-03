package fees

import (
	stderrors "errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ParseError — table-driven tests
// ---------------------------------------------------------------------------

func TestParseError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		statusCode     int
		body           []byte
		wantProduct    string
		wantCategory   sdkerrors.ErrorCategory
		wantCode       sdkerrors.ErrorCode
		wantMessage    string
		wantStatusCode int
	}{
		{
			name:           "400 bad request with JSON body",
			statusCode:     400,
			body:           []byte(`{"code":"FEE-001","message":"invalid package ID"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryValidation,
			wantCode:       sdkerrors.ErrorCode("FEE-001"),
			wantMessage:    "invalid package ID",
			wantStatusCode: 400,
		},
		{
			name:           "401 unauthorized",
			statusCode:     401,
			body:           []byte(`{"code":"AUTH-001","message":"missing bearer token"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryAuthentication,
			wantCode:       sdkerrors.ErrorCode("AUTH-001"),
			wantMessage:    "missing bearer token",
			wantStatusCode: 401,
		},
		{
			name:           "403 forbidden",
			statusCode:     403,
			body:           []byte(`{"code":"PERM-001","message":"insufficient privileges"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryAuthorization,
			wantCode:       sdkerrors.ErrorCode("PERM-001"),
			wantMessage:    "insufficient privileges",
			wantStatusCode: 403,
		},
		{
			name:           "404 not found",
			statusCode:     404,
			body:           []byte(`{"code":"FEE-002","message":"package not found"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryNotFound,
			wantCode:       sdkerrors.ErrorCode("FEE-002"),
			wantMessage:    "package not found",
			wantStatusCode: 404,
		},
		{
			name:           "409 conflict",
			statusCode:     409,
			body:           []byte(`{"code":"FEE-003","message":"duplicate package name"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryConflict,
			wantCode:       sdkerrors.ErrorCode("FEE-003"),
			wantMessage:    "duplicate package name",
			wantStatusCode: 409,
		},
		{
			name:           "429 rate limited",
			statusCode:     429,
			body:           []byte(`{"code":"RATE-001","message":"too many requests"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryRateLimit,
			wantCode:       sdkerrors.ErrorCode("RATE-001"),
			wantMessage:    "too many requests",
			wantStatusCode: 429,
		},
		{
			name:           "500 internal server error",
			statusCode:     500,
			body:           []byte(`{"code":"INT-001","message":"unexpected error"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       sdkerrors.ErrorCode("INT-001"),
			wantMessage:    "unexpected error",
			wantStatusCode: 500,
		},
		{
			name:           "502 bad gateway maps to internal",
			statusCode:     502,
			body:           []byte(`{"code":"GW-001","message":"upstream unreachable"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       sdkerrors.ErrorCode("GW-001"),
			wantMessage:    "upstream unreachable",
			wantStatusCode: 502,
		},
		{
			name:           "503 service unavailable maps to internal",
			statusCode:     503,
			body:           []byte(`{"code":"SVC-001","message":"service unavailable"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       sdkerrors.ErrorCode("SVC-001"),
			wantMessage:    "service unavailable",
			wantStatusCode: 503,
		},
		{
			name:           "unknown status code defaults to internal",
			statusCode:     418,
			body:           []byte(`{"code":"TEA-001","message":"I'm a teapot"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       sdkerrors.ErrorCode("TEA-001"),
			wantMessage:    "I'm a teapot",
			wantStatusCode: 418,
		},
		{
			name:           "non-JSON body falls back to raw string",
			statusCode:     500,
			body:           []byte("Internal Server Error"),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryInternal,
			wantCode:       "",
			wantMessage:    "Internal Server Error",
			wantStatusCode: 500,
		},
		{
			name:           "malformed JSON falls back to raw string",
			statusCode:     400,
			body:           []byte(`{"code": "FEE-001", "message": `),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryValidation,
			wantCode:       "",
			wantMessage:    `{"code": "FEE-001", "message": `,
			wantStatusCode: 400,
		},
		{
			name:           "empty body falls back to empty string",
			statusCode:     404,
			body:           []byte(""),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryNotFound,
			wantCode:       "",
			wantMessage:    "",
			wantStatusCode: 404,
		},
		{
			name:           "JSON with empty code",
			statusCode:     400,
			body:           []byte(`{"code":"","message":"something went wrong"}`),
			wantProduct:    "fees",
			wantCategory:   sdkerrors.CategoryValidation,
			wantCode:       "",
			wantMessage:    "something went wrong",
			wantStatusCode: 400,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ParseError(tc.statusCode, tc.body)
			require.NotNil(t, got)

			assert.Equal(t, tc.wantProduct, got.Product)
			assert.Equal(t, tc.wantCategory, got.Category)
			assert.Equal(t, tc.wantCode, got.Code)
			assert.Equal(t, tc.wantMessage, got.Message)
			assert.Equal(t, tc.wantStatusCode, got.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseError — sentinel matching with errors.Is
// ---------------------------------------------------------------------------

func TestParseErrorMatchesSentinels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		body     []byte
		sentinel *sdkerrors.Error
	}{
		{
			name:     "400 matches ErrValidation",
			status:   400,
			body:     []byte(`{"code":"FEE-001","message":"bad input"}`),
			sentinel: sdkerrors.ErrValidation,
		},
		{
			name:     "401 matches ErrAuthentication",
			status:   401,
			body:     []byte(`{"code":"AUTH-001","message":"expired"}`),
			sentinel: sdkerrors.ErrAuthentication,
		},
		{
			name:     "403 matches ErrAuthorization",
			status:   403,
			body:     []byte(`{"code":"PERM-001","message":"denied"}`),
			sentinel: sdkerrors.ErrAuthorization,
		},
		{
			name:     "404 matches ErrNotFound",
			status:   404,
			body:     []byte(`{"code":"FEE-002","message":"missing"}`),
			sentinel: sdkerrors.ErrNotFound,
		},
		{
			name:     "409 matches ErrConflict",
			status:   409,
			body:     []byte(`{"code":"FEE-003","message":"dup"}`),
			sentinel: sdkerrors.ErrConflict,
		},
		{
			name:     "429 matches ErrRateLimit",
			status:   429,
			body:     []byte(`{"code":"RATE-001","message":"throttled"}`),
			sentinel: sdkerrors.ErrRateLimit,
		},
		{
			name:     "500 matches ErrInternal",
			status:   500,
			body:     []byte(`{"code":"INT-001","message":"oops"}`),
			sentinel: sdkerrors.ErrInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ParseError(tc.status, tc.body)
			assert.True(t, stderrors.Is(err, tc.sentinel),
				"ParseError(%d) should match %s", tc.status, tc.sentinel.Category)
		})
	}
}

// ---------------------------------------------------------------------------
// categoryFromStatus — exhaustive coverage
// ---------------------------------------------------------------------------

func TestCategoryFromStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status int
		want   sdkerrors.ErrorCategory
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
		{418, sdkerrors.CategoryInternal}, // unknown -> internal
		{200, sdkerrors.CategoryInternal}, // unexpected success code -> internal
		{301, sdkerrors.CategoryInternal}, // redirect -> internal
	}

	for _, tc := range tests {
		got := sdkerrors.CategoryFromStatus(tc.status)
		assert.Equal(t, tc.want, got, "categoryFromStatus(%d)", tc.status)
	}
}
