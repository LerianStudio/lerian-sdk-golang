package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/textproto"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

var (
	// ErrNilService is returned when a nil *BaseService is passed to a
	// generic CRUD function.
	ErrNilService = sdkerrors.NewInternal("sdk", "BaseService", "service is nil", nil)

	// ErrNilBackend is returned when a BaseService has no Backend configured.
	// This typically means the service was constructed without wiring it to
	// an HTTP transport (e.g., via NewBackendImpl).
	ErrNilBackend = sdkerrors.NewInternal("sdk", "BaseService", "backend is not configured", nil)
)

// BaseService provides shared infrastructure for all product services.
// Product services embed this struct and delegate HTTP operations to the
// generic package-level functions in this package, including CRUD helpers,
// custom actions, header-aware actions, upserts, lists, and counts.
//
// Example usage in a product service:
//
//	type OrganizationService struct {
//	    core.BaseService
//	}
//
//	func (s *OrganizationService) Get(ctx context.Context, id string) (*models.Organization, error) {
//	    return core.Get[models.Organization](ctx, &s.BaseService, "/organizations/"+id)
//	}
type BaseService struct {
	Backend Backend
}

// ResolveBackend validates that the service and its backend are configured
// and returns the backend for use by higher-level helpers.
func ResolveBackend(s *BaseService) (Backend, error) {
	if s == nil {
		return nil, ErrNilService
	}

	if isNilInterface(s.Backend) {
		return nil, ErrNilBackend
	}

	return s.Backend, nil
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// Get retrieves a single resource by path.
// The type parameter T is the expected response type.
func Get[T any](ctx context.Context, s *BaseService, path string) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "GET", Path: path})
}

// Create creates a new resource by POSTing the input to the given path.
// The type parameter T is the response type; I is the input/request type.
func Create[T any, I any](ctx context.Context, s *BaseService, path string, input *I) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "POST", Path: path, Body: input})
}

// Update modifies an existing resource using PATCH.
// The type parameter T is the response type; I is the input/request type.
func Update[T any, I any](ctx context.Context, s *BaseService, path string, input *I) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "PATCH", Path: path, Body: input})
}

// Upsert creates or replaces a resource using PUT.
// The type parameter T is the response type; I is the input/request type.
// This is used for endpoints with create-or-update (idempotent PUT) semantics.
func Upsert[T any, I any](ctx context.Context, s *BaseService, path string, input *I) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "PUT", Path: path, Body: input})
}

// Delete removes a resource by path. It sends a DELETE request and expects
// no response body (result is nil).
func Delete(ctx context.Context, s *BaseService, path string) error {
	backend, err := ResolveBackend(s)
	if err != nil {
		return err
	}

	_, err = backend.Do(ctx, Request{Method: "DELETE", Path: path, ExpectNoResponse: true})

	return err
}

// Action performs a state-transition action on a resource (e.g., commit,
// cancel, activate). It sends a POST request to the given path with an
// optional input body and returns the resulting resource state.
func Action[T any](ctx context.Context, s *BaseService, path string, input any) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "POST", Path: path, Body: input})
}

// ActionWithHeaders performs a POST action on a resource while allowing
// callers to provide per-request headers such as multipart content types.
func ActionWithHeaders[T any](ctx context.Context, s *BaseService, path string, headers map[string]string, input any) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "POST", Path: path, Headers: headers, Body: input})
}

// List returns a paginated [pagination.Iterator] over resources at the
// given path. Query parameters are built from [models.CursorListOptions].
// The iterator fetches pages lazily as Next() is called.
//
// If the BaseService or its Backend is nil, List returns a "poisoned"
// iterator whose first Next() call yields the appropriate error via Err().
func List[T any](ctx context.Context, s *BaseService, path string, opts *models.CursorListOptions) *pagination.Iterator[T] {
	backend, err := ResolveBackend(s)
	if err != nil {
		return pagination.NewIterator[T](func(_ context.Context, _ string) ([]T, string, error) {
			return nil, "", err
		})
	}

	fetcher := func(fetchCtx context.Context, cursor string) ([]T, string, error) {
		queryPath := buildListPath(path, opts, cursor)

		resp, err := doJSON[models.ListResponse[T]](fetchCtx, backend, Request{Method: "GET", Path: queryPath})
		if err != nil {
			return nil, "", err
		}

		return resp.Items, resp.Pagination.NextCursor, nil
	}

	return pagination.NewIterator[T](fetcher)
}

// Count sends a HEAD request to a metrics/count endpoint and returns the
// resource count from the X-Total-Count response header.
func Count(ctx context.Context, s *BaseService, path string) (int, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return 0, err
	}

	res, err := backend.Do(ctx, Request{Method: "HEAD", Path: path})
	if err != nil {
		return 0, err
	}

	if res == nil {
		return 0, sdkerrors.NewInternal("sdk", "Count", "backend returned nil response", nil)
	}

	var countValues []string

	for key, values := range res.Headers {
		if textproto.CanonicalMIMEHeaderKey(key) == "X-Total-Count" {
			countValues = values
			break
		}
	}

	if len(countValues) == 0 {
		return 0, sdkerrors.NewInternal("sdk", "Count", "missing X-Total-Count header", nil)
	}

	if len(countValues) > 1 {
		return 0, sdkerrors.NewInternal("sdk", "Count", "multiple X-Total-Count header values", nil)
	}

	count, err := strconv.Atoi(strings.TrimSpace(countValues[0]))
	if err != nil || count < 0 {
		return 0, sdkerrors.NewInternal("sdk", "Count", fmt.Sprintf("invalid X-Total-Count header value: %q", countValues[0]), err)
	}

	return count, nil
}

func doJSON[T any](ctx context.Context, backend Backend, req Request) (*T, error) {
	res, err := backend.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, sdkerrors.NewInternal("sdk", req.Method+" "+req.Path, "backend returned nil response", nil)
	}

	var result T
	if err := json.Unmarshal(res.Body, &result); err != nil {
		return nil, sdkerrors.NewInternal("sdk", req.Method+" "+req.Path, "failed to unmarshal response body", err)
	}

	return &result, nil
}

// buildListPath appends query parameters from CursorListOptions and cursor to the
// base path. It handles cursor precedence: a non-empty cursor argument (from
// pagination) takes priority over the initial cursor in CursorListOptions.
func buildListPath(path string, opts *models.CursorListOptions, cursor string) string {
	params := url.Values{}

	if cursor != "" {
		params.Set("cursor", cursor)
	}

	if opts != nil {
		if opts.Limit > 0 {
			params.Set("limit", strconv.Itoa(opts.Limit))
		}

		if opts.Cursor != "" && cursor == "" {
			params.Set("cursor", opts.Cursor)
		}

		if opts.SortBy != "" {
			params.Set("sortBy", opts.SortBy)
		}

		if opts.SortOrder != "" {
			params.Set("sortOrder", opts.SortOrder)
		}

		if opts.StartDate != nil {
			params.Set("startDate", opts.StartDate.Format("2006-01-02T15:04:05Z"))
		}

		if opts.EndDate != nil {
			params.Set("endDate", opts.EndDate.Format("2006-01-02T15:04:05Z"))
		}

		for k, v := range opts.Filters {
			params.Set(fmt.Sprintf("filter[%s]", k), v)
		}
	}

	if len(params) == 0 {
		return path
	}

	return path + "?" + params.Encode()
}
