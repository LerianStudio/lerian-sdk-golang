package midaz

import (
	stderrors "errors"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ParseError — table-driven
// ---------------------------------------------------------------------------

func TestParseError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantCode   sdkerrors.ErrorCode
		wantCat    sdkerrors.ErrorCategory
		wantMsg    string
	}{
		{
			name:       "400 validation with JSON body",
			statusCode: 400,
			body:       []byte(`{"code":"0003","title":"Validation Error","message":"name is required"}`),
			wantCode:   CodeValidationError,
			wantCat:    sdkerrors.CategoryValidation,
			wantMsg:    "name is required",
		},
		{
			name:       "401 authentication",
			statusCode: 401,
			body:       []byte(`{"code":"0013","title":"Unauthorized","message":"invalid bearer token"}`),
			wantCode:   CodeAuthenticationError,
			wantCat:    sdkerrors.CategoryAuthentication,
			wantMsg:    "invalid bearer token",
		},
		{
			name:       "403 authorization",
			statusCode: 403,
			body:       []byte(`{"code":"0010","title":"Forbidden","message":"insufficient permissions"}`),
			wantCode:   sdkerrors.ErrorCode("0010"),
			wantCat:    sdkerrors.CategoryAuthorization,
			wantMsg:    "insufficient permissions",
		},
		{
			name:       "404 not found",
			statusCode: 404,
			body:       []byte(`{"code":"0040","title":"Not Found","message":"account not found"}`),
			wantCode:   CodeNotFound,
			wantCat:    sdkerrors.CategoryNotFound,
			wantMsg:    "account not found",
		},
		{
			name:       "409 conflict (already exists)",
			statusCode: 409,
			body:       []byte(`{"code":"0005","title":"Conflict","message":"organization already exists"}`),
			wantCode:   CodeAlreadyExists,
			wantCat:    sdkerrors.CategoryConflict,
			wantMsg:    "organization already exists",
		},
		{
			name:       "429 rate limit",
			statusCode: 429,
			body:       []byte(`{"code":"0050","title":"Too Many Requests","message":"rate limit exceeded"}`),
			wantCode:   sdkerrors.ErrorCode("0050"),
			wantCat:    sdkerrors.CategoryRateLimit,
			wantMsg:    "rate limit exceeded",
		},
		{
			name:       "500 internal server error",
			statusCode: 500,
			body:       []byte(`{"code":"9999","title":"Internal Error","message":"unexpected failure"}`),
			wantCode:   sdkerrors.ErrorCode("9999"),
			wantCat:    sdkerrors.CategoryInternal,
			wantMsg:    "unexpected failure",
		},
		{
			name:       "502 bad gateway",
			statusCode: 502,
			body:       []byte(`{"code":"5002","title":"Bad Gateway","message":"upstream unavailable"}`),
			wantCode:   sdkerrors.ErrorCode("5002"),
			wantCat:    sdkerrors.CategoryInternal,
			wantMsg:    "upstream unavailable",
		},
		{
			name:       "418 unknown status defaults to internal",
			statusCode: 418,
			body:       []byte(`{"code":"0418","title":"Teapot","message":"I'm a teapot"}`),
			wantCode:   sdkerrors.ErrorCode("0418"),
			wantCat:    sdkerrors.CategoryInternal,
			wantMsg:    "I'm a teapot",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ParseError(tc.statusCode, tc.body)

			require.NotNil(t, err)
			assert.Equal(t, "midaz", err.Product)
			assert.Equal(t, tc.wantCat, err.Category)
			assert.Equal(t, tc.wantCode, err.Code)
			assert.Equal(t, tc.wantMsg, err.Message)
			assert.Equal(t, tc.statusCode, err.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseError — invalid JSON fallback
// ---------------------------------------------------------------------------

func TestParseErrorInvalidJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       []byte
	}{
		{
			name:       "plain text body",
			statusCode: 500,
			body:       []byte("Internal Server Error"),
		},
		{
			name:       "empty body",
			statusCode: 404,
			body:       []byte(""),
		},
		{
			name:       "malformed JSON",
			statusCode: 400,
			body:       []byte(`{"code": "0003", "message": `),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ParseError(tc.statusCode, tc.body)

			require.NotNil(t, err)
			assert.Equal(t, "midaz", err.Product)
			assert.Equal(t, tc.statusCode, err.StatusCode)
			// Message should be the raw body string.
			assert.Equal(t, string(tc.body), err.Message)
			// Code should be empty since JSON parsing failed.
			assert.Empty(t, err.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseError — sentinel matching via errors.Is
// ---------------------------------------------------------------------------

func TestParseErrorSentinelMatching(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		sentinel *sdkerrors.Error
	}{
		{"400 matches ErrValidation", 400, sdkerrors.ErrValidation},
		{"401 matches ErrAuthentication", 401, sdkerrors.ErrAuthentication},
		{"403 matches ErrAuthorization", 403, sdkerrors.ErrAuthorization},
		{"404 matches ErrNotFound", 404, sdkerrors.ErrNotFound},
		{"409 matches ErrConflict", 409, sdkerrors.ErrConflict},
		{"429 matches ErrRateLimit", 429, sdkerrors.ErrRateLimit},
		{"500 matches ErrInternal", 500, sdkerrors.ErrInternal},
	}

	body := []byte(`{"code":"0001","title":"Error","message":"test"}`)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ParseError(tc.status, body)
			assert.True(t, stderrors.Is(err, tc.sentinel),
				"ParseError(%d, ...) should match %v", tc.status, tc.sentinel)
		})
	}
}

// ---------------------------------------------------------------------------
// Midaz error codes — constant values
// ---------------------------------------------------------------------------

func TestMidazErrorCodeConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, sdkerrors.ErrorCode("0040"), CodeNotFound)
	assert.Equal(t, sdkerrors.ErrorCode("0005"), CodeAlreadyExists)
	assert.Equal(t, sdkerrors.ErrorCode("0029"), CodeInsufficientBalance)
	assert.Equal(t, sdkerrors.ErrorCode("0036"), CodeAssetMismatch)
	assert.Equal(t, sdkerrors.ErrorCode("0003"), CodeValidationError)
	assert.Equal(t, sdkerrors.ErrorCode("0013"), CodeAuthenticationError)
}

// ---------------------------------------------------------------------------
// CategoryFromStatus — exhaustive status coverage (delegated to sdkerrors)
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
		{418, sdkerrors.CategoryInternal}, // fallback
		{200, sdkerrors.CategoryInternal}, // fallback for unexpected success code
	}

	for _, tc := range tests {
		got := sdkerrors.CategoryFromStatus(tc.status)
		assert.Equal(t, tc.want, got, "status %d", tc.status)
	}
}
