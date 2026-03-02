package core

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

var (
	// ErrNilService is returned when a nil *BaseService is passed to a
	// generic CRUD function.
	ErrNilService = errors.New("service is nil")

	// ErrNilBackend is returned when a BaseService has no Backend configured.
	// This typically means the service was constructed without wiring it to
	// an HTTP transport (e.g., via NewBackendImpl).
	ErrNilBackend = errors.New("backend is not configured")
)

// BaseService provides shared infrastructure for all product services.
// Product services embed this struct and delegate HTTP operations to the
// generic package-level functions (Get, Create, Update, Delete, Action, List).
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

// Get retrieves a single resource by path.
// The type parameter T is the expected response type.
func Get[T any](ctx context.Context, s *BaseService, path string) (*T, error) {
	if s == nil {
		return nil, ErrNilService
	}

	if s.Backend == nil {
		return nil, ErrNilBackend
	}

	var result T
	if err := s.Backend.Call(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Create creates a new resource by POSTing the input to the given path.
// The type parameter T is the response type; I is the input/request type.
func Create[T any, I any](ctx context.Context, s *BaseService, path string, input *I) (*T, error) {
	if s == nil {
		return nil, ErrNilService
	}

	if s.Backend == nil {
		return nil, ErrNilBackend
	}

	var result T
	if err := s.Backend.Call(ctx, "POST", path, input, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update modifies an existing resource using PATCH.
// The type parameter T is the response type; I is the input/request type.
func Update[T any, I any](ctx context.Context, s *BaseService, path string, input *I) (*T, error) {
	if s == nil {
		return nil, ErrNilService
	}

	if s.Backend == nil {
		return nil, ErrNilBackend
	}

	var result T
	if err := s.Backend.Call(ctx, "PATCH", path, input, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete removes a resource by path. It sends a DELETE request and expects
// no response body (result is nil).
func Delete(ctx context.Context, s *BaseService, path string) error {
	if s == nil {
		return ErrNilService
	}

	if s.Backend == nil {
		return ErrNilBackend
	}

	return s.Backend.Call(ctx, "DELETE", path, nil, nil)
}

// Action performs a state-transition action on a resource (e.g., commit,
// cancel, activate). It sends a POST request to the given path with an
// optional input body and returns the resulting resource state.
func Action[T any](ctx context.Context, s *BaseService, path string, input any) (*T, error) {
	if s == nil {
		return nil, ErrNilService
	}

	if s.Backend == nil {
		return nil, ErrNilBackend
	}

	var result T
	if err := s.Backend.Call(ctx, "POST", path, input, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List returns a paginated [pagination.Iterator] over resources at the
// given path. Query parameters are built from [models.ListOptions].
// The iterator fetches pages lazily as Next() is called.
//
// If the BaseService or its Backend is nil, List returns a "poisoned"
// iterator whose first Next() call yields the appropriate error via Err().
func List[T any](ctx context.Context, s *BaseService, path string, opts *models.ListOptions) *pagination.Iterator[T] {
	if s == nil {
		return pagination.NewIterator[T](func(_ context.Context, _ string) ([]T, string, error) {
			return nil, "", ErrNilService
		})
	}

	if s.Backend == nil {
		return pagination.NewIterator[T](func(_ context.Context, _ string) ([]T, string, error) {
			return nil, "", ErrNilBackend
		})
	}

	fetcher := func(fetchCtx context.Context, cursor string) ([]T, string, error) {
		var resp models.ListResponse[T]

		queryPath := buildListPath(path, opts, cursor)
		if err := s.Backend.Call(fetchCtx, "GET", queryPath, nil, &resp); err != nil {
			return nil, "", err
		}

		return resp.Items, resp.Pagination.NextCursor, nil
	}

	return pagination.NewIterator[T](fetcher)
}

// buildListPath appends query parameters from ListOptions and cursor to the
// base path. It handles cursor precedence: a non-empty cursor argument (from
// pagination) takes priority over the initial cursor in ListOptions.
func buildListPath(path string, opts *models.ListOptions, cursor string) string {
	params := url.Values{}

	if cursor != "" {
		params.Set("cursor", cursor)
	}

	if opts != nil {
		if opts.Limit > 0 {
			params.Set("limit", strconv.Itoa(opts.Limit))
		}

		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
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
