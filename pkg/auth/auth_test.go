package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Compile-time interface checks
// ---------------------------------------------------------------------------

var (
	_ Authenticator = (*NoAuth)(nil)
	_ Authenticator = (*OAuth2)(nil)
)

// ---------------------------------------------------------------------------
// NoAuth
// ---------------------------------------------------------------------------

func TestNoAuthImplementsAuthenticator(t *testing.T) {
	t.Parallel()

	na := NewNoAuth()
	assert.NotNil(t, na)
}

func TestNoAuthEnrich(t *testing.T) {
	t.Parallel()

	na := NewNoAuth()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	// Pre-set an Authorization header to prove NoAuth leaves it untouched.
	req.Header.Set("Authorization", "Bearer existing-token")

	err := na.Enrich(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "Bearer existing-token", req.Header.Get("Authorization"),
		"NoAuth must not modify existing headers")
}

// ---------------------------------------------------------------------------
// OAuth2 helpers
// ---------------------------------------------------------------------------

// newOAuth2TestServer returns an httptest.Server that acts as a token endpoint.
// Each call increments the atomic counter. The server responds with a valid
// token whose expires_in is configurable.
func newOAuth2TestServer(t *testing.T, counter *atomic.Int64, expiresIn int64) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)

		// Validate the request basics.
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		require.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.NotEmpty(t, r.FormValue("client_id"))
		assert.NotEmpty(t, r.FormValue("client_secret"))

		resp := tokenResponse{
			AccessToken: "tok-123",
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
}

// ---------------------------------------------------------------------------
// OAuth2 tests
// ---------------------------------------------------------------------------

func TestOAuth2Enrich(t *testing.T) {
	t.Parallel()

	var counter atomic.Int64

	srv := newOAuth2TestServer(t, &counter, 3600)
	defer srv.Close()

	oauth := NewOAuth2("cid", "csecret", srv.URL, []string{"read", "write"})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api", nil)

	err := oauth.Enrich(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "Bearer tok-123", req.Header.Get("Authorization"))
	assert.Equal(t, int64(1), counter.Load(), "should have made exactly 1 token request")
}

func TestOAuth2TokenCaching(t *testing.T) {
	t.Parallel()

	var counter atomic.Int64

	srv := newOAuth2TestServer(t, &counter, 3600)
	defer srv.Close()

	oauth := NewOAuth2("cid", "csecret", srv.URL, []string{"read"})

	// First call — fetches token from server.
	req1 := httptest.NewRequest(http.MethodGet, "http://example.com/1", nil)
	require.NoError(t, oauth.Enrich(context.Background(), req1))
	assert.Equal(t, "Bearer tok-123", req1.Header.Get("Authorization"))

	// Second call — should use cached token (no additional server request).
	req2 := httptest.NewRequest(http.MethodGet, "http://example.com/2", nil)
	require.NoError(t, oauth.Enrich(context.Background(), req2))
	assert.Equal(t, "Bearer tok-123", req2.Header.Get("Authorization"))

	assert.Equal(t, int64(1), counter.Load(),
		"second Enrich must reuse the cached token, not hit the server again")
}

func TestOAuth2TokenRefresh(t *testing.T) {
	t.Parallel()

	var counter atomic.Int64

	srv := newOAuth2TestServer(t, &counter, 3600)
	defer srv.Close()

	oauth := NewOAuth2("cid", "csecret", srv.URL, nil)

	// Simulate time progression using nowFunc.
	now := time.Now()
	oauth.nowFunc = func() time.Time { return now }

	// First call — fetches token.
	req1 := httptest.NewRequest(http.MethodGet, "http://example.com/1", nil)
	require.NoError(t, oauth.Enrich(context.Background(), req1))
	assert.Equal(t, int64(1), counter.Load())

	// Advance time past the token expiry (3600s - 30s buffer = 3570s valid).
	now = now.Add(3571 * time.Second)

	// Second call — token expired, must re-fetch.
	req2 := httptest.NewRequest(http.MethodGet, "http://example.com/2", nil)
	require.NoError(t, oauth.Enrich(context.Background(), req2))
	assert.Equal(t, int64(2), counter.Load(),
		"expired token must trigger a second token request")
}

func TestOAuth2ShortLivedTokenCaching(t *testing.T) {
	t.Parallel()

	// When ExpiresIn < expiryBuffer (30s), the cache duration should still be
	// positive to prevent flooding the token endpoint on every API call.
	tests := []struct {
		name             string
		expiresIn        int64
		wantMinCacheSecs float64
		wantMaxCacheSecs float64
	}{
		{
			name:             "ExpiresIn=5s caches for 2.5s",
			expiresIn:        5,
			wantMinCacheSecs: 2,
			wantMaxCacheSecs: 3,
		},
		{
			name:             "ExpiresIn=1s caches for 1s floor",
			expiresIn:        1,
			wantMinCacheSecs: 1,
			wantMaxCacheSecs: 1,
		},
		{
			name:             "ExpiresIn=0s caches for 1s floor",
			expiresIn:        0,
			wantMinCacheSecs: 1,
			wantMaxCacheSecs: 1,
		},
		{
			name:             "ExpiresIn=30s (exactly expiryBuffer) caches for 15s",
			expiresIn:        30,
			wantMinCacheSecs: 14,
			wantMaxCacheSecs: 16,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var counter atomic.Int64

			srv := newOAuth2TestServer(t, &counter, tc.expiresIn)
			defer srv.Close()

			now := time.Now()
			oauth := NewOAuth2("cid", "csecret", srv.URL, nil)
			oauth.nowFunc = func() time.Time { return now }

			// First call — fetches token from server.
			req1 := httptest.NewRequest(http.MethodGet, "http://example.com/1", nil)
			require.NoError(t, oauth.Enrich(context.Background(), req1))
			assert.Equal(t, int64(1), counter.Load(), "first call must fetch token")

			// Second call at the same instant — must use the cached token.
			req2 := httptest.NewRequest(http.MethodGet, "http://example.com/2", nil)
			require.NoError(t, oauth.Enrich(context.Background(), req2))
			assert.Equal(t, int64(1), counter.Load(),
				"second call at same time must reuse cached token, not flood the endpoint")

			// Verify expiry is within expected range by checking it's still cached
			// just before the minimum, and expired after the maximum.
			oauth.mu.Lock()
			expiry := oauth.expiry
			oauth.mu.Unlock()

			cacheDuration := expiry.Sub(now)
			assert.GreaterOrEqual(t, cacheDuration.Seconds(), tc.wantMinCacheSecs,
				"cache duration should be at least %v seconds", tc.wantMinCacheSecs)
			assert.LessOrEqual(t, cacheDuration.Seconds(), tc.wantMaxCacheSecs,
				"cache duration should be at most %v seconds", tc.wantMaxCacheSecs)
		})
	}
}

func TestOAuth2ConcurrentSafety(t *testing.T) {
	t.Parallel()

	var counter atomic.Int64

	srv := newOAuth2TestServer(t, &counter, 3600)
	defer srv.Close()

	oauth := NewOAuth2("cid", "csecret", srv.URL, []string{"scope1"})

	const goroutines = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errs := make(chan error, goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

			if err := oauth.Enrich(context.Background(), req); err != nil {
				errs <- err
				return
			}

			if got := req.Header.Get("Authorization"); got != "Bearer tok-123" {
				errs <- &unexpectedHeaderError{got: got}
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent Enrich error: %v", err)
	}

	// With the mutex, all goroutines should share a single token fetch.
	assert.Equal(t, int64(1), counter.Load(),
		"all concurrent Enrich calls should share one token fetch")
}

// unexpectedHeaderError is a small error type for concurrent test reporting.
type unexpectedHeaderError struct {
	got string
}

func (e *unexpectedHeaderError) Error() string {
	return "unexpected Authorization header: " + e.got
}

// ---------------------------------------------------------------------------
// OAuth2 error scenarios
// ---------------------------------------------------------------------------

func TestOAuth2TokenEndpointError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"error": "invalid_client"}))
	}))
	defer srv.Close()

	oauth := NewOAuth2("bad-id", "bad-secret", srv.URL, nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	err := oauth.Enrich(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "invalid_client")
}

func TestOAuth2MalformedJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`not json`))
		require.NoError(t, err)
	}))
	defer srv.Close()

	oauth := NewOAuth2("cid", "csecret", srv.URL, nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	err := oauth.Enrich(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding token response")
}

func TestOAuth2EmptyAccessToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(tokenResponse{TokenType: "Bearer", ExpiresIn: 3600}))
	}))
	defer srv.Close()

	oauth := NewOAuth2("cid", "csecret", srv.URL, nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	err := oauth.Enrich(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing access_token")
}

func TestOAuth2UnreachableServer(t *testing.T) {
	t.Parallel()

	// Point to a server that doesn't exist.
	oauth := NewOAuth2("cid", "csecret", "http://127.0.0.1:1", nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	err := oauth.Enrich(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "executing token request")
}

func TestOAuth2ScopeJoining(t *testing.T) {
	t.Parallel()

	var receivedScope string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		receivedScope = r.FormValue("scope")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := tokenResponse{
			AccessToken: "tok-scoped",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}

		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer srv.Close()

	oauth := NewOAuth2("cid", "csecret", srv.URL, []string{"read", "write", "admin"})
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	require.NoError(t, oauth.Enrich(context.Background(), req))
	assert.Equal(t, "read write admin", receivedScope,
		"scopes must be space-joined in the token request")
}

func TestOAuth2NoScopes(t *testing.T) {
	t.Parallel()

	var scopePresent bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		_, scopePresent = r.Form["scope"]

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := tokenResponse{
			AccessToken: "tok-noscope",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}

		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer srv.Close()

	oauth := NewOAuth2("cid", "csecret", srv.URL, nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	require.NoError(t, oauth.Enrich(context.Background(), req))
	assert.False(t, scopePresent,
		"when no scopes are configured, the scope field must be omitted from the request")
}

// ---------------------------------------------------------------------------
// Credential redaction tests (Issue #12)
// ---------------------------------------------------------------------------

func TestOAuth2StringRedaction(t *testing.T) {
	t.Parallel()

	oauth := NewOAuth2("my-client-id", "my-client-secret", "https://auth.example.com/token", nil)
	s := oauth.String()

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "my-client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.NotContains(t, s, "my-client-secret",
		"String() must not contain the client secret")
}

func TestOAuth2MarshalJSONRedaction(t *testing.T) {
	t.Parallel()

	oauth := NewOAuth2("my-client-id", "my-client-secret", "https://auth.example.com/token", nil)
	data, err := json.Marshal(oauth)

	require.NoError(t, err)
	assert.Contains(t, string(data), "[REDACTED]")
	assert.Contains(t, string(data), "my-client-id", "ClientID should be visible")
	assert.Contains(t, string(data), "https://auth.example.com/token", "TokenURL should be visible")
	assert.NotContains(t, string(data), "my-client-secret",
		"MarshalJSON must not contain the client secret")
}

// ---------------------------------------------------------------------------
// OAuth2 redirect and zero-value safety tests
// ---------------------------------------------------------------------------

func TestOAuth2RedirectRejectsCrossHost(t *testing.T) {
	t.Parallel()

	var foreignHits atomic.Int64

	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		foreignHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "tok-redirect",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}))
	}))
	defer foreign.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, foreign.URL+"/token", http.StatusTemporaryRedirect)
	}))
	defer origin.Close()

	oauth := NewOAuth2("cid", "csecret", origin.URL, nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	err := oauth.Enrich(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "refusing cross-host redirect")
	assert.Equal(t, int64(0), foreignHits.Load(),
		"token requests must not follow cross-host redirects")
}

func TestOAuth2RedirectAllowsSameHost(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			http.Redirect(w, r, "/token", http.StatusTemporaryRedirect)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "tok-same-host",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}))
	}))
	defer ts.Close()

	oauth := NewOAuth2("cid", "csecret", ts.URL+"/start", nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	require.NoError(t, oauth.Enrich(context.Background(), req))
	assert.Equal(t, "Bearer tok-same-host", req.Header.Get("Authorization"))
	assert.Equal(t, int64(2), requestCount.Load())
}

func TestOAuth2ZeroValueUsesDefaults(t *testing.T) {
	t.Parallel()

	var counter atomic.Int64

	srv := newOAuth2TestServer(t, &counter, 3600)
	defer srv.Close()

	oauth := &OAuth2{
		ClientID:     "cid",
		ClientSecret: "csecret",
		TokenURL:     srv.URL,
	}
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	require.NoError(t, oauth.Enrich(context.Background(), req))
	assert.Equal(t, "Bearer tok-123", req.Header.Get("Authorization"))
	assert.Equal(t, int64(1), counter.Load())
	assert.NotNil(t, oauth.httpClient)
	assert.NotNil(t, oauth.nowFunc)
	assert.False(t, oauth.expiry.IsZero())
}
