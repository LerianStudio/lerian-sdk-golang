package pagination

import (
	"context"
	"strconv"
)

// NumberedPageFetcher defines a page-number-based fetch contract.
// It returns the items for the requested page plus enough metadata to decide
// whether another page exists.
type NumberedPageFetcher[T any] func(ctx context.Context, page int) (items []T, total int, pageSize int, actualPage int, err error)

// NewPageIterator adapts a page-number-based fetcher to the generic Iterator.
// The iterator internally carries the next page number as its cursor string.
func NewPageIterator[T any](initialPage int, fetcher NumberedPageFetcher[T]) *Iterator[T] {
	return NewIterator[T](func(ctx context.Context, cursor string) ([]T, string, error) {
		page := initialPage

		if cursor != "" {
			parsed, err := parsePageCursor(cursor)
			if err != nil {
				return nil, "", err
			}

			page = parsed
		}

		items, total, pageSize, actualPage, err := fetcher(ctx, page)
		if err != nil {
			return nil, "", err
		}

		nextPage := nextPageCursor(total, pageSize, actualPage, len(items))

		return items, nextPage, nil
	})
}

func parsePageCursor(cursor string) (int, error) {
	return strconv.Atoi(cursor)
}

func nextPageCursor(total, pageSize, actualPage, itemsCount int) string {
	if itemsCount == 0 || total <= 0 {
		return ""
	}

	if actualPage <= 0 {
		actualPage = 1
	}

	if pageSize <= 0 {
		pageSize = itemsCount
	}

	if pageSize <= 0 {
		return ""
	}

	if actualPage*pageSize >= total {
		return ""
	}

	return strconv.Itoa(actualPage + 1)
}
