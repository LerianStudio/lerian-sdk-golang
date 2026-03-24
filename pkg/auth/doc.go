// Package auth provides authenticator implementations for SDK transport.
//
// Each authenticator satisfies the [Authenticator] interface that the
// transport layer calls before every outbound request to inject the
// appropriate credentials. Implementations must be safe for concurrent use.
//
// # Strategies
//
// The following authentication strategies are available:
//
//   - [NoAuth] -- passthrough that adds no credentials
//   - [OAuth2] -- OAuth2 client-credentials flow with automatic token caching
//     and refresh
//
// # Usage
//
// Authenticators are typically created by product option functions (e.g.
// [midaz.WithClientCredentials]), but can also be constructed directly:
//
//	authn := auth.NewOAuth2("client-id", "client-secret", "https://auth.example.com/token")
//	authn.Enrich(ctx, req) // sets Authorization: Bearer <token>
package auth
