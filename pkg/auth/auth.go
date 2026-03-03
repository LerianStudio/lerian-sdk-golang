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

// Authenticator enriches an outbound HTTP request with credentials.
// Implementations must be safe for concurrent use.
type Authenticator interface {
	// Enrich adds authentication information (headers, query params, etc.)
	// to the given request. It returns an error if the enrichment fails
	// (e.g. a token refresh encounters a network error).
	Enrich(ctx context.Context, req *http.Request) error
}

// ---------------------------------------------------------------------------
// BearerToken
// ---------------------------------------------------------------------------

// BearerToken authenticates requests with a static bearer token in the
// Authorization header.
type BearerToken struct {
	// Token is the raw bearer token value (without the "Bearer " prefix).
	Token string
}

// NewBearerToken creates a BearerToken authenticator with the given token.
func NewBearerToken(token string) *BearerToken {
	return &BearerToken{Token: token}
}

// Enrich sets the Authorization header to "Bearer <token>".
func (b *BearerToken) Enrich(_ context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+b.Token)
	return nil
}

// String implements fmt.Stringer to prevent credential leakage in logs.
func (b BearerToken) String() string {
	return "BearerToken{Token: [REDACTED]}"
}

// MarshalJSON prevents credential leakage during JSON serialization.
func (b BearerToken) MarshalJSON() ([]byte, error) {
	return []byte(`{"type":"bearer","token":"[REDACTED]"}`), nil
}

// ---------------------------------------------------------------------------
// APIKey
// ---------------------------------------------------------------------------

// APIKey authenticates requests by setting an arbitrary header to the
// concatenation of a prefix and a key. This covers multiple conventions:
//
//   - Matcher pattern:  Header="Authorization", Prefix="ApiKey ", Key="..."
//   - Tracer pattern:   Header="X-API-Key",     Prefix="",       Key="..."
//   - Custom prefix:    Header="Authorization", Prefix="Token ",  Key="..."
type APIKey struct {
	// Header is the HTTP header name where the key is placed.
	Header string
	// Prefix is prepended to Key when setting the header value.
	// It may include a trailing space if the convention requires one.
	Prefix string
	// Key is the API key value.
	Key string
}

// NewAPIKey creates an APIKey authenticator. The resulting header value will
// be Prefix+Key (e.g. "ApiKey abc123" or just "abc123" when Prefix is empty).
func NewAPIKey(header, prefix, key string) *APIKey {
	return &APIKey{
		Header: header,
		Prefix: prefix,
		Key:    key,
	}
}

// Enrich sets the configured header to Prefix+Key.
func (a *APIKey) Enrich(_ context.Context, req *http.Request) error {
	req.Header.Set(a.Header, a.Prefix+a.Key)
	return nil
}

// String implements fmt.Stringer to prevent credential leakage in logs.
func (a APIKey) String() string {
	return fmt.Sprintf("APIKey{Header: %q, Prefix: %q, Key: [REDACTED]}", a.Header, a.Prefix)
}

// MarshalJSON prevents credential leakage during JSON serialization.
func (a APIKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(
		`{"type":"apikey","header":%q,"prefix":%q,"key":"[REDACTED]"}`,
		a.Header, a.Prefix,
	)), nil
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
// If no custom HTTP client is needed, pass nil and a default client with a
// 10-second timeout will be used.
func NewOAuth2(clientID, clientSecret, tokenURL string, scopes []string) *OAuth2 {
	return &OAuth2{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       scopes,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}

				if len(via) > 0 && req.URL.Host != via[0].URL.Host {
					req.Header.Del("Authorization")
				}

				return nil
			},
		},
		nowFunc: time.Now,
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

	now := o.nowFunc()
	if o.token != "" && now.Before(o.expiry) {
		return o.token, nil
	}

	return o.refreshToken(ctx)
}

// refreshToken performs the client-credentials POST and caches the result.
// It must be called with o.mu held.
func (o *OAuth2) refreshToken(ctx context.Context) (string, error) {
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
