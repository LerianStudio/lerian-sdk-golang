package midaz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	callHeadFn     func(ctx context.Context, path string) (map[string][]string, error)
}

func (m *mockBackend) Do(ctx context.Context, req core.Request) (*core.Response, error) {
	if req.Method == http.MethodHead {
		return m.doHead(ctx, req.Path)
	}

	headers := reqHeaders(req)
	if len(headers) > 0 && m.callWithHdrsFn != nil {
		return m.doJSONWithHeaders(ctx, req, headers)
	}

	return m.doWithoutHeaders(ctx, req)
}

func (m *mockBackend) doHead(ctx context.Context, path string) (*core.Response, error) {
	if m.callHeadFn == nil {
		return nil, fmt.Errorf("mockBackend.head not configured")
	}

	headers, err := m.callHeadFn(ctx, path)
	if err != nil {
		return nil, err
	}

	return &core.Response{Headers: headers}, nil
}

func (m *mockBackend) doWithoutHeaders(ctx context.Context, req core.Request) (*core.Response, error) {
	if m.callRawFn != nil && strings.Contains(req.Path, "/download") {
		return m.doRaw(ctx, req)
	}

	if m.callFn != nil {
		return m.doJSON(ctx, req)
	}

	if m.callRawFn != nil {
		return m.doRaw(ctx, req)
	}

	return nil, fmt.Errorf("mockBackend.Do not configured")
}

func (m *mockBackend) doJSONWithHeaders(ctx context.Context, req core.Request, headers map[string]string) (*core.Response, error) {
	var result any

	resultArg := any(&result)
	if req.ExpectNoResponse {
		resultArg = nil
	}

	if err := m.callWithHdrsFn(ctx, req.Method, req.Path, headers, reqBody(req), resultArg); err != nil {
		return nil, err
	}

	return responseForRequest(req, result)
}

func (m *mockBackend) doJSON(ctx context.Context, req core.Request) (*core.Response, error) {
	var result any

	resultArg := any(&result)
	if req.ExpectNoResponse {
		resultArg = nil
	}

	if err := m.callFn(ctx, req.Method, req.Path, reqBody(req), resultArg); err != nil {
		return nil, err
	}

	return responseForRequest(req, result)
}

func (m *mockBackend) doRaw(ctx context.Context, req core.Request) (*core.Response, error) {
	body, err := m.callRawFn(ctx, req.Method, req.Path, reqBody(req))
	if err != nil {
		return nil, err
	}

	return &core.Response{Body: body}, nil
}

func responseForRequest(req core.Request, result any) (*core.Response, error) {
	if req.ExpectNoResponse {
		return &core.Response{}, nil
	}

	return jsonResponse(result)
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

func jsonResponse(result any) (*core.Response, error) {
	if result == nil {
		return &core.Response{}, nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &core.Response{Body: data}, nil
}

func reqBody(req core.Request) any {
	if len(req.BodyBytes) > 0 {
		return req.BodyBytes
	}

	return req.Body
}

func reqHeaders(req core.Request) map[string]string {
	if len(req.Headers) == 0 && req.ContentType == "" {
		return nil
	}

	headers := map[string]string{}
	for k, v := range req.Headers {
		headers[k] = v
	}

	if req.ContentType != "" {
		headers["Content-Type"] = req.ContentType
	}

	return headers
}
