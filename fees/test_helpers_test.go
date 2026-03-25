package fees

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

type mockBackend struct {
	callFn func(ctx context.Context, method, path string, body, result any) error
}

func (m *mockBackend) Do(ctx context.Context, req core.Request) (*core.Response, error) {
	if m.callFn == nil {
		return &core.Response{}, nil
	}

	var result any

	resultArg := any(&result)
	if req.ExpectNoResponse {
		resultArg = nil
	}

	if err := m.callFn(ctx, req.Method, req.Path, reqBody(req), resultArg); err != nil {
		return nil, err
	}

	if req.ExpectNoResponse {
		return &core.Response{}, nil
	}

	return jsonResponse(result)
}

var _ core.Backend = (*mockBackend)(nil)

func jsonInto(src, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("mock marshal: %w", err)
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

func strPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }
