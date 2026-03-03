package tracer

import (
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
			name:       "400 validation error with TRC code",
			statusCode: 400,
			body:       []byte(`{"code":"TRC-0001","message":"invalid rule condition"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-0001"),
			wantCat:    sdkerrors.CategoryValidation,
			wantMsg:    "invalid rule condition",
		},
		{
			name:       "401 authentication error",
			statusCode: 401,
			body:       []byte(`{"code":"TRC-0010","message":"invalid API key"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-0010"),
			wantCat:    sdkerrors.CategoryAuthentication,
			wantMsg:    "invalid API key",
		},
		{
			name:       "403 authorization error",
			statusCode: 403,
			body:       []byte(`{"code":"TRC-0020","message":"insufficient permissions"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-0020"),
			wantCat:    sdkerrors.CategoryAuthorization,
			wantMsg:    "insufficient permissions",
		},
		{
			name:       "404 not found",
			statusCode: 404,
			body:       []byte(`{"code":"TRC-0030","message":"rule not found"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-0030"),
			wantCat:    sdkerrors.CategoryNotFound,
			wantMsg:    "rule not found",
		},
		{
			name:       "409 conflict",
			statusCode: 409,
			body:       []byte(`{"code":"TRC-0040","message":"duplicate rule name"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-0040"),
			wantCat:    sdkerrors.CategoryConflict,
			wantMsg:    "duplicate rule name",
		},
		{
			name:       "429 rate limit",
			statusCode: 429,
			body:       []byte(`{"code":"TRC-0050","message":"rate limit exceeded"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-0050"),
			wantCat:    sdkerrors.CategoryRateLimit,
			wantMsg:    "rate limit exceeded",
		},
		{
			name:       "500 internal error",
			statusCode: 500,
			body:       []byte(`{"code":"TRC-9000","message":"internal server error"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-9000"),
			wantCat:    sdkerrors.CategoryInternal,
			wantMsg:    "internal server error",
		},
		{
			name:       "502 bad gateway maps to internal",
			statusCode: 502,
			body:       []byte(`{"code":"TRC-9002","message":"upstream timeout"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-9002"),
			wantCat:    sdkerrors.CategoryInternal,
			wantMsg:    "upstream timeout",
		},
		{
			name:       "503 service unavailable maps to internal",
			statusCode: 503,
			body:       []byte(`{"code":"TRC-9003","message":"service unavailable"}`),
			wantCode:   sdkerrors.ErrorCode("TRC-9003"),
			wantCat:    sdkerrors.CategoryInternal,
			wantMsg:    "service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ParseError(tt.statusCode, tt.body)

			require.NotNil(t, err)
			assert.Equal(t, "tracer", err.Product)
			assert.Equal(t, tt.wantCat, err.Category)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMsg, err.Message)
			assert.Equal(t, tt.statusCode, err.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseError — malformed JSON body
// ---------------------------------------------------------------------------

func TestParseErrorMalformedJSON(t *testing.T) {
	t.Parallel()

	err := ParseError(400, []byte(`not json at all`))

	require.NotNil(t, err)
	assert.Equal(t, "tracer", err.Product)
	assert.Equal(t, sdkerrors.CategoryValidation, err.Category)
	assert.Equal(t, sdkerrors.ErrorCode(""), err.Code)
	assert.Equal(t, "not json at all", err.Message)
	assert.Equal(t, 400, err.StatusCode)
}

func TestParseErrorEmptyBody(t *testing.T) {
	t.Parallel()

	err := ParseError(500, []byte(""))

	require.NotNil(t, err)
	assert.Equal(t, "tracer", err.Product)
	assert.Equal(t, sdkerrors.CategoryInternal, err.Category)
	assert.Equal(t, sdkerrors.ErrorCode(""), err.Code)
	assert.Equal(t, "", err.Message)
	assert.Equal(t, 500, err.StatusCode)
}

// ---------------------------------------------------------------------------
// ParseError — with details field
// ---------------------------------------------------------------------------

func TestParseErrorWithDetails(t *testing.T) {
	t.Parallel()

	body := []byte(`{"code":"TRC-0001","message":"multiple violations","details":{"fields":["amount","currency"]}}`)
	err := ParseError(400, body)

	require.NotNil(t, err)
	assert.Equal(t, sdkerrors.ErrorCode("TRC-0001"), err.Code)
	assert.Equal(t, "multiple violations", err.Message)
	// Details are parsed into the tracerErrorResponse but not propagated
	// to the SDK error (by design). The structured error is sufficient.
}

// ---------------------------------------------------------------------------
// ParseError — unknown status code
// ---------------------------------------------------------------------------

func TestParseErrorUnknownStatusCode(t *testing.T) {
	t.Parallel()

	err := ParseError(418, []byte(`{"code":"TRC-TEAPOT","message":"I'm a teapot"}`))

	require.NotNil(t, err)
	assert.Equal(t, "tracer", err.Product)
	// Unknown status codes default to internal.
	assert.Equal(t, sdkerrors.CategoryInternal, err.Category)
	assert.Equal(t, sdkerrors.ErrorCode("TRC-TEAPOT"), err.Code)
	assert.Equal(t, "I'm a teapot", err.Message)
	assert.Equal(t, 418, err.StatusCode)
}

// ---------------------------------------------------------------------------
// categoryFromStatus — exhaustive
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
		{418, sdkerrors.CategoryInternal},
		{422, sdkerrors.CategoryInternal},
	}

	for _, tt := range tests {
		got := categoryFromStatus(tt.status)
		assert.Equal(t, tt.want, got, "status %d", tt.status)
	}
}

// ---------------------------------------------------------------------------
// ParseError — errors.Is compatibility
// ---------------------------------------------------------------------------

func TestParseErrorIsCompatible(t *testing.T) {
	t.Parallel()

	err := ParseError(404, []byte(`{"code":"TRC-0030","message":"not found"}`))

	// Should match the sentinel for not-found.
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)
	// Should NOT match unrelated sentinels.
	assert.NotErrorIs(t, err, sdkerrors.ErrValidation)
	assert.NotErrorIs(t, err, sdkerrors.ErrAuthentication)
}
