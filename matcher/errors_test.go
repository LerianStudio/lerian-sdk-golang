package matcher

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// ---------------------------------------------------------------------------
// ParseError — valid JSON responses
// ---------------------------------------------------------------------------

func TestParseError_ValidJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantCat    sdkerrors.ErrorCategory
		wantCode   sdkerrors.ErrorCode
		wantMsg    string
		wantRes    string
	}{
		{
			name:       "400 validation with entity type",
			statusCode: 400,
			body:       `{"code":"INVALID_EXPRESSION","message":"rule expression is malformed","entityType":"Rule"}`,
			wantCat:    sdkerrors.CategoryValidation,
			wantCode:   "INVALID_EXPRESSION",
			wantMsg:    "rule expression is malformed",
			wantRes:    "Rule",
		},
		{
			name:       "401 authentication",
			statusCode: 401,
			body:       `{"code":"AUTH_REQUIRED","message":"API key is missing"}`,
			wantCat:    sdkerrors.CategoryAuthentication,
			wantCode:   "AUTH_REQUIRED",
			wantMsg:    "API key is missing",
			wantRes:    "",
		},
		{
			name:       "403 authorization",
			statusCode: 403,
			body:       `{"code":"FORBIDDEN","message":"insufficient permissions"}`,
			wantCat:    sdkerrors.CategoryAuthorization,
			wantCode:   "FORBIDDEN",
			wantMsg:    "insufficient permissions",
			wantRes:    "",
		},
		{
			name:       "404 not found with entity type",
			statusCode: 404,
			body:       `{"code":"NOT_FOUND","message":"context not found","entityType":"Context"}`,
			wantCat:    sdkerrors.CategoryNotFound,
			wantCode:   "NOT_FOUND",
			wantMsg:    "context not found",
			wantRes:    "Context",
		},
		{
			name:       "409 conflict",
			statusCode: 409,
			body:       `{"code":"DUPLICATE_NAME","message":"rule name already exists","entityType":"Rule"}`,
			wantCat:    sdkerrors.CategoryConflict,
			wantCode:   "DUPLICATE_NAME",
			wantMsg:    "rule name already exists",
			wantRes:    "Rule",
		},
		{
			name:       "429 rate limit",
			statusCode: 429,
			body:       `{"code":"RATE_LIMITED","message":"too many requests"}`,
			wantCat:    sdkerrors.CategoryRateLimit,
			wantCode:   "RATE_LIMITED",
			wantMsg:    "too many requests",
			wantRes:    "",
		},
		{
			name:       "500 internal server error",
			statusCode: 500,
			body:       `{"code":"INTERNAL","message":"unexpected error"}`,
			wantCat:    sdkerrors.CategoryInternal,
			wantCode:   "INTERNAL",
			wantMsg:    "unexpected error",
			wantRes:    "",
		},
		{
			name:       "502 bad gateway",
			statusCode: 502,
			body:       `{"code":"UPSTREAM","message":"upstream unavailable"}`,
			wantCat:    sdkerrors.CategoryInternal,
			wantCode:   "UPSTREAM",
			wantMsg:    "upstream unavailable",
			wantRes:    "",
		},
		{
			name:       "no entity type field",
			statusCode: 400,
			body:       `{"code":"BAD_REQUEST","message":"invalid input"}`,
			wantCat:    sdkerrors.CategoryValidation,
			wantCode:   "BAD_REQUEST",
			wantMsg:    "invalid input",
			wantRes:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseError(tc.statusCode, []byte(tc.body))

			require.NotNil(t, got)
			assert.Equal(t, "matcher", got.Product)
			assert.Equal(t, tc.wantCat, got.Category)
			assert.Equal(t, tc.wantCode, got.Code)
			assert.Equal(t, tc.wantMsg, got.Message)
			assert.Equal(t, tc.wantRes, got.Resource)
			assert.Equal(t, tc.statusCode, got.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseError — invalid JSON falls back to raw body
// ---------------------------------------------------------------------------

func TestParseError_InvalidJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantCat    sdkerrors.ErrorCategory
	}{
		{
			name:       "plain text body",
			statusCode: 500,
			body:       "Internal Server Error",
			wantCat:    sdkerrors.CategoryInternal,
		},
		{
			name:       "HTML error page",
			statusCode: 502,
			body:       "<html><body>Bad Gateway</body></html>",
			wantCat:    sdkerrors.CategoryInternal,
		},
		{
			name:       "empty body",
			statusCode: 404,
			body:       "",
			wantCat:    sdkerrors.CategoryNotFound,
		},
		{
			name:       "malformed JSON",
			statusCode: 400,
			body:       `{"code": "INVALID"`,
			wantCat:    sdkerrors.CategoryValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseError(tc.statusCode, []byte(tc.body))

			require.NotNil(t, got)
			assert.Equal(t, "matcher", got.Product)
			assert.Equal(t, tc.wantCat, got.Category)
			assert.Equal(t, tc.body, got.Message)
			assert.Equal(t, sdkerrors.ErrorCode(""), got.Code)
			assert.Equal(t, "", got.Resource)
			assert.Equal(t, tc.statusCode, got.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseError — sentinel matching via errors.Is
// ---------------------------------------------------------------------------

func TestParseError_SentinelMatching(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		sentinel   *sdkerrors.Error
	}{
		{name: "400 matches ErrValidation", statusCode: 400, sentinel: sdkerrors.ErrValidation},
		{name: "401 matches ErrAuthentication", statusCode: 401, sentinel: sdkerrors.ErrAuthentication},
		{name: "403 matches ErrAuthorization", statusCode: 403, sentinel: sdkerrors.ErrAuthorization},
		{name: "404 matches ErrNotFound", statusCode: 404, sentinel: sdkerrors.ErrNotFound},
		{name: "409 matches ErrConflict", statusCode: 409, sentinel: sdkerrors.ErrConflict},
		{name: "429 matches ErrRateLimit", statusCode: 429, sentinel: sdkerrors.ErrRateLimit},
		{name: "500 matches ErrInternal", statusCode: 500, sentinel: sdkerrors.ErrInternal},
	}

	body := `{"code":"TEST","message":"test error"}`

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseError(tc.statusCode, []byte(body))
			assert.True(t, stderrors.Is(got, tc.sentinel),
				"expected errors.Is(err, %v) to be true", tc.sentinel.Category)
		})
	}
}

// ---------------------------------------------------------------------------
// categoryFromStatus — edge cases
// ---------------------------------------------------------------------------

func TestCategoryFromStatus(t *testing.T) {
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
		// Unmapped status codes default to internal.
		{418, sdkerrors.CategoryInternal},
		{422, sdkerrors.CategoryInternal},
	}

	for _, tc := range tests {
		got := sdkerrors.CategoryFromStatus(tc.status)
		assert.Equal(t, tc.want, got, "status %d", tc.status)
	}
}

// ---------------------------------------------------------------------------
// ErrorParser — convenience accessor
// ---------------------------------------------------------------------------

func TestErrorParser(t *testing.T) {
	parser := ErrorParser()
	require.NotNil(t, parser)

	got := parser(404, []byte(`{"code":"NOT_FOUND","message":"gone","entityType":"Schedule"}`))
	require.NotNil(t, got)
	assert.Equal(t, "matcher", got.Product)
	assert.Equal(t, sdkerrors.CategoryNotFound, got.Category)
	assert.Equal(t, sdkerrors.ErrorCode("NOT_FOUND"), got.Code)
	assert.Equal(t, "gone", got.Message)
	assert.Equal(t, "Schedule", got.Resource)
}
