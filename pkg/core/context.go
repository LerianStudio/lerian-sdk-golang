package core

import "context"

// contextKey is an unexported type used for context value keys to avoid
// collisions with keys defined in other packages.
type contextKey string

const idempotencyKeyCtx contextKey = "idempotency_key"

// WithIdempotencyKey returns a derived context that carries an idempotency key.
// When [BackendImpl] encounters this value, it injects an X-Idempotency-Key
// header into the outbound HTTP request. This enables safe retries for
// non-idempotent operations (e.g. creating a transaction).
//
// Usage:
//
//	ctx := core.WithIdempotencyKey(ctx, "unique-request-id-123")
//	err := client.Transactions.Create(ctx, input)
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyCtx, key)
}

// idempotencyKeyFromContext extracts the idempotency key from the context,
// returning the key and true if present, or "" and false if absent.
func idempotencyKeyFromContext(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(idempotencyKeyCtx).(string)
	return key, ok && key != ""
}
