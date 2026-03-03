package matcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

// mockBackend is a test double for [core.Backend] that captures the HTTP
// method, path, and body sent by service methods. The callFn field allows
// each test to define custom verification and response logic. The callRawFn
// field supports services that use CallRaw (e.g., file downloads).
type mockBackend struct {
	callFn         func(ctx context.Context, method, path string, body, result any) error
	callWithHdrsFn func(ctx context.Context, method, path string, headers map[string]string, body, result any) error
	callRawFn      func(ctx context.Context, method, path string, body any) ([]byte, error)
}

func (m *mockBackend) Call(ctx context.Context, method, path string, body, result any) error {
	if m.callFn != nil {
		return m.callFn(ctx, method, path, body, result)
	}

	return fmt.Errorf("mockBackend.Call not configured")
}

func (m *mockBackend) CallWithHeaders(ctx context.Context, method, path string,
	headers map[string]string, body, result any) error {
	if m.callWithHdrsFn != nil {
		return m.callWithHdrsFn(ctx, method, path, headers, body, result)
	}

	return fmt.Errorf("mockBackend.CallWithHeaders not configured")
}

func (m *mockBackend) CallRaw(ctx context.Context, method, path string, body any) ([]byte, error) {
	if m.callRawFn != nil {
		return m.callRawFn(ctx, method, path, body)
	}

	return nil, fmt.Errorf("mockBackend.CallRaw not configured")
}

// Compile-time interface compliance check.
var _ core.Backend = (*mockBackend)(nil)

// unmarshalInto uses JSON round-trip to populate the result pointer from
// a source value. This simulates what a real backend does when deserializing
// API responses.
func unmarshalInto(src any, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}
