package core

import (
	"context"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// GetWithHeaders executes a GET request with extra headers using the service backend.
func GetWithHeaders[T any](ctx context.Context, s *BaseService, path string, headers map[string]string) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "GET", Path: path, Headers: headers})
}

// CreateWithHeaders executes a POST request with extra headers using the service backend.
func CreateWithHeaders[T any, I any](ctx context.Context, s *BaseService, path string, headers map[string]string, input *I) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "POST", Path: path, Headers: headers, Body: input})
}

// UpdateWithHeaders executes a PATCH request with extra headers using the service backend.
func UpdateWithHeaders[T any, I any](ctx context.Context, s *BaseService, path string, headers map[string]string, input *I) (*T, error) {
	backend, err := ResolveBackend(s)
	if err != nil {
		return nil, err
	}

	return doJSON[T](ctx, backend, Request{Method: "PATCH", Path: path, Headers: headers, Body: input})
}

// DeleteWithHeaders executes a DELETE request with extra headers using the service backend.
func DeleteWithHeaders(ctx context.Context, s *BaseService, path string, headers map[string]string) error {
	backend, err := ResolveBackend(s)
	if err != nil {
		return err
	}

	_, err = backend.Do(ctx, Request{Method: "DELETE", Path: path, Headers: headers, ExpectNoResponse: true})

	return err
}

// ListPageWithHeaders builds a paginated iterator that sends extra headers on each request.
func ListPageWithHeaders[T any](ctx context.Context, s *BaseService, headers map[string]string, initialPage int, buildPath func(page int) string) *pagination.Iterator[T] {
	backend, err := ResolveBackend(s)
	if err != nil {
		return pagination.NewErrorIterator[T](err)
	}

	return pagination.NewPageIterator[T](initialPage, func(fetchCtx context.Context, page int) ([]T, int, int, int, error) {
		resp, err := doJSON[models.ListResponse[T]](fetchCtx, backend, Request{Method: "GET", Path: buildPath(page), Headers: headers})
		if err != nil {
			return nil, 0, 0, 0, err
		}

		return resp.Items, resp.Pagination.Total, resp.Pagination.Limit, resp.Pagination.Page, nil
	})
}
