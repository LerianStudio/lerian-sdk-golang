package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultOAuthTimeout = 10 * time.Second

// Authenticator enriches an outbound HTTP request with credentials.
// Implementations must be safe for concurrent use.
type Authenticator interface {
	// Enrich adds authentication information (headers, query params, etc.)
	// to the given request. It returns an error if the enrichment fails
	// (e.g. a token refresh encounters a network error).
	Enrich(ctx context.Context, req *http.Request) error
}

// ---------------------------------------------------------------------------
// NoAuth
// ---------------------------------------------------------------------------

// NoAuth is a no-op authenticator that leaves the request untouched.
// It is useful as a default or for unauthenticated endpoints.
type NoAuth struct{}

// NewNoAuth creates a NoAuth authenticator.
func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

// Enrich does nothing and always returns nil.
func (n *NoAuth) Enrich(_ context.Context, _ *http.Request) error {
	return nil
}

// ---------------------------------------------------------------------------
// OAuth2 (client-credentials flow)
// ---------------------------------------------------------------------------

// tokenResponse represents the JSON payload returned by an OAuth2 token
// endpoint for the client_credentials grant type.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// expiryBuffer is subtracted from the reported token lifetime so that the
// SDK refreshes slightly before the real expiry, avoiding edge-case 401s.
const expiryBuffer = 30 * time.Second

// OAuth2 authenticates requests using the OAuth2 client-credentials flow.
// It automatically fetches, caches, and refreshes access tokens.
//
// The struct is safe for concurrent use; a sync.Mutex serializes token
// refresh operations while allowing concurrent reads of a valid cached token.
type OAuth2 struct {
	// ClientID is the OAuth2 client identifier.
	ClientID string
	// ClientSecret is the OAuth2 client secret.
	ClientSecret string
	// TokenURL is the full URL of the token endpoint.
	TokenURL string
	// Scopes is the set of OAuth2 scopes to request.
	Scopes []string

	mu     sync.Mutex
	token  string
	expiry time.Time

	// httpClient is the HTTP client used for token requests.
	// If nil at construction time, a default client with a 10-second timeout
	// is used.
	httpClient *http.Client

	// nowFunc is an internal hook for testing time-dependent behavior.
	// In production it defaults to time.Now.
	nowFunc func() time.Time
}

// NewOAuth2 creates an OAuth2 authenticator for the client-credentials flow.
// It uses a default HTTP client with a 10-second timeout.
func NewOAuth2(clientID, clientSecret, tokenURL string, scopes []string) *OAuth2 {
	return NewOAuth2WithHTTPClient(clientID, clientSecret, tokenURL, scopes, nil)
}

// NewOAuth2WithHTTPClient creates an OAuth2 authenticator for the
// client-credentials flow using the provided HTTP client for token requests.
//
// If httpClient is nil, a default client with a 10-second timeout is used.
func NewOAuth2WithHTTPClient(clientID, clientSecret, tokenURL string, scopes []string, httpClient *http.Client) *OAuth2 {
	oauth := &OAuth2{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       append([]string(nil), scopes...),
		httpClient:   httpClient,
		nowFunc:      time.Now,
	}
	oauth.ensureDefaultsLocked()

	return oauth
}

func defaultOAuthHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultOAuthTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("auth/oauth2: stopped after 10 redirects")
			}

			if len(via) == 0 {
				return nil
			}

			if req.URL.Host != via[0].URL.Host {
				return fmt.Errorf(
					"auth/oauth2: refusing cross-host redirect from %q to %q",
					via[0].URL.Host,
					req.URL.Host,
				)
			}

			return nil
		},
	}
}

func (o *OAuth2) ensureDefaultsLocked() {
	if o.httpClient == nil {
		o.httpClient = defaultOAuthHTTPClient()
	}

	if o.nowFunc == nil {
		o.nowFunc = time.Now
	}
}

// Enrich obtains a valid access token (from cache or by refreshing) and sets
// the Authorization header to "Bearer <token>".
func (o *OAuth2) Enrich(ctx context.Context, req *http.Request) error {
	token, err := o.validToken(ctx)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	return nil
}

// validToken returns a cached token if still valid, or fetches a new one.
func (o *OAuth2) validToken(ctx context.Context) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.ensureDefaultsLocked()

	now := o.nowFunc()
	if o.token != "" && now.Before(o.expiry) {
		return o.token, nil
	}

	return o.refreshToken(ctx)
}

// refreshToken performs the client-credentials POST and caches the result.
// It must be called with o.mu held.
func (o *OAuth2) refreshToken(ctx context.Context) (string, error) {
	o.ensureDefaultsLocked()

	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {o.ClientID},
		"client_secret": {o.ClientSecret},
	}

	if len(o.Scopes) > 0 {
		data.Set("scope", strings.Join(o.Scopes, " "))
	}

	tokenReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		o.TokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("auth/oauth2: building token request: %w", err)
	}

	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.httpClient.Do(tokenReq)
	if err != nil {
		return "", fmt.Errorf("auth/oauth2: executing token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB limit
	if err != nil {
		return "", fmt.Errorf("auth/oauth2: reading token response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"auth/oauth2: token endpoint returned %d: %s",
			resp.StatusCode, string(body),
		)
	}

	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("auth/oauth2: decoding token response: %w", err)
	}

	if tok.AccessToken == "" {
		return "", fmt.Errorf("auth/oauth2: token response missing access_token")
	}

	o.token = tok.AccessToken

	tokenDuration := time.Duration(tok.ExpiresIn) * time.Second
	cacheDuration := tokenDuration - expiryBuffer

	if cacheDuration <= 0 {
		// For very short-lived tokens, cache for half the token lifetime
		// with a minimum of 1 second to prevent token endpoint flooding.
		cacheDuration = tokenDuration / 2
		if cacheDuration < time.Second {
			cacheDuration = time.Second
		}
	}

	o.expiry = o.nowFunc().Add(cacheDuration)

	return o.token, nil
}

// String implements fmt.Stringer to prevent credential leakage in logs.
func (o *OAuth2) String() string {
	return fmt.Sprintf("OAuth2{ClientID: %q, TokenURL: %q, ClientSecret: [REDACTED]}", o.ClientID, o.TokenURL)
}

// MarshalJSON prevents credential leakage during JSON serialization.
func (o *OAuth2) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(
		`{"type":"oauth2","client_id":%q,"token_url":%q,"client_secret":"[REDACTED]"}`,
		o.ClientID, o.TokenURL,
	)), nil
}
