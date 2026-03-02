package pagination

import "context"

// NewIteratorFromSlice creates an Iterator backed by a static slice.
// Useful in tests to create iterators without a real Backend.
// All items are returned on the first page; subsequent calls signal done.
func NewIteratorFromSlice[T any](items []T) *Iterator[T] {
	called := false

	return NewIterator[T](func(ctx context.Context, cursor string) ([]T, string, error) {
		if called {
			return nil, "", nil
		}

		called = true

		return items, "", nil
	})
}

// NewErrorIterator creates an Iterator that immediately fails with the given error.
// Useful in tests to simulate backend failures.
func NewErrorIterator[T any](err error) *Iterator[T] {
	return NewIterator[T](func(ctx context.Context, cursor string) ([]T, string, error) {
		return nil, "", err
	})
}
