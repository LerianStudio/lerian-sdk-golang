package matcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

// mockBackend is a test double for [core.Backend] that captures the HTTP
// method, path, and body sent by service methods. The callFn field allows
// each test to define custom verification and response logic. The callRawFn
// field supports services that return raw byte downloads.
type mockBackend struct {
	callFn         func(ctx context.Context, method, path string, body, result any) error
	callWithHdrsFn func(ctx context.Context, method, path string, headers map[string]string, body, result any) error
	callRawFn      func(ctx context.Context, method, path string, body any) ([]byte, error)
}

func (m *mockBackend) Do(ctx context.Context, req core.Request) (*core.Response, error) {
	headers := reqHeaders(req)
	if len(headers) > 0 && m.callWithHdrsFn != nil {
		return m.doJSONWithHeaders(ctx, req, headers)
	}

	return m.doWithoutHeaders(ctx, req)
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
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}

func jsonResponse(result any) (*core.Response, error) {
	if result == nil {
		return &core.Response{}, nil
	}

	b, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &core.Response{Body: b}, nil
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
