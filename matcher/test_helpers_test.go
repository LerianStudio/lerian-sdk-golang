package matcher

import (
	"context"
	"encoding/json"
	"testing"
)

// mockBackend is a test double for [core.Backend] that captures the HTTP
// method, path, and body sent by service methods. The callFn field allows
// each test to define custom verification and response logic. The callRawFn
// field supports services that use CallRaw (e.g., file downloads).
type mockBackend struct {
	t         *testing.T
	callFn    func(ctx context.Context, method, path string, body, result any) error
	callRawFn func(ctx context.Context, method, path string, body any) ([]byte, error)
}

func (m *mockBackend) Call(ctx context.Context, method, path string, body, result any) error {
	return m.callFn(ctx, method, path, body, result)
}

func (m *mockBackend) CallWithHeaders(ctx context.Context, method, path string, _ map[string]string, body, result any) error {
	return m.callFn(ctx, method, path, body, result)
}

func (m *mockBackend) CallRaw(ctx context.Context, method, path string, body any) ([]byte, error) {
	if m.callRawFn != nil {
		return m.callRawFn(ctx, method, path, body)
	}

	return nil, nil
}

// unmarshalInto marshals data to JSON and then unmarshals it into result.
// This simulates the JSON round-trip that the real backend performs when
// populating result pointers.
func unmarshalInto(t *testing.T, data any, result any) {
	t.Helper()

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(b, result); err != nil {
		t.Fatal(err)
	}
}
