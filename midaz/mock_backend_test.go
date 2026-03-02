package midaz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

// ---------------------------------------------------------------------------
// Mock Backend — shared test helper
// ---------------------------------------------------------------------------

// mockBackend is a hand-written mock that implements the [core.Backend]
// interface. Each function field can be overridden per-test to control
// behaviour. This mirrors the pattern used in pkg/core/service_test.go.
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
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dst)
}
