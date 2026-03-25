package leriantest

import (
	"context"
	"strconv"
	"sync"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// fakeStore is a generic in-memory key-value store for fake services.
// It is safe for concurrent use and maintains insertion order for List.
type fakeStore[T any] struct {
	mu    sync.RWMutex
	items map[string]T
	order []string // insertion order for List
}

func newFakeStore[T any]() *fakeStore[T] {
	return &fakeStore[T]{items: make(map[string]T)}
}

// Set stores an item with the given id. If the id already exists the item is
// replaced in-place without changing the insertion order.
func (s *fakeStore[T]) Set(id string, item T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[id]; !exists {
		s.order = append(s.order, id)
	}

	s.items[id] = item
}

// Get returns the item with the given id and true, or the zero value and
// false if the id is not in the store.
func (s *fakeStore[T]) Get(id string) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]

	return item, ok
}

// Delete removes the item with the given id from the store.
func (s *fakeStore[T]) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, id)

	for i, k := range s.order {
		if k == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
}

// List returns all items in insertion order.
func (s *fakeStore[T]) List() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]T, 0, len(s.order))
	for _, id := range s.order {
		if item, ok := s.items[id]; ok {
			result = append(result, item)
		}
	}

	return result
}

// Len returns the number of items in the store.
func (s *fakeStore[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.items)
}

// Iterator returns a paginated Iterator backed by the store contents.
// All items are returned on a single page. This is kept for backward
// compatibility and internal helpers like seed-data methods.
func (s *fakeStore[T]) Iterator() *pagination.Iterator[T] {
	items := s.List()
	return pagination.NewIteratorFromSlice(items)
}

// defaultFakePageSize is the page size used when ListOptions is nil or has a
// zero/negative Limit.
const defaultFakePageSize = 10

// PaginatedIterator returns a paginated Iterator that respects the Limit and
// Cursor fields of the supplied ListOptions, simulating real API pagination.
//
// Cursor semantics: the cursor is the string-encoded integer index into the
// ordered items slice. An empty cursor starts at index 0. For example, a
// cursor of "5" means "start returning items from index 5 onward".
//
// When opts is nil or opts.Limit <= 0, the default page size is used.
func (s *fakeStore[T]) PaginatedIterator(opts *models.CursorListOptions) *pagination.Iterator[T] {
	pageSize := defaultFakePageSize
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}

	// Capture the snapshot once so the fetcher closes over a stable slice.
	snapshot := s.List()

	// If opts carries a starting cursor, honour it for the very first fetch.
	initialCursor := ""
	if opts != nil && opts.Cursor != "" {
		initialCursor = opts.Cursor
	}

	firstCall := true

	return pagination.NewIterator[T](func(_ context.Context, cursor string) ([]T, string, error) {
		// On the very first invocation the pagination.Iterator always passes
		// an empty cursor. If the caller specified a starting cursor via
		// ListOptions, inject it here.
		if firstCall {
			firstCall = false
			if cursor == "" && initialCursor != "" {
				cursor = initialCursor
			}
		}

		start := 0
		if cursor != "" {
			idx, err := strconv.Atoi(cursor)
			if err != nil {
				return nil, "", err
			}

			start = idx
		}

		// Past the end — nothing left.
		if start >= len(snapshot) {
			return nil, "", nil
		}

		end := start + pageSize
		if end > len(snapshot) {
			end = len(snapshot)
		}

		page := snapshot[start:end]

		nextCursor := ""
		if end < len(snapshot) {
			nextCursor = strconv.Itoa(end)
		}

		return page, nextCursor, nil
	})
}
