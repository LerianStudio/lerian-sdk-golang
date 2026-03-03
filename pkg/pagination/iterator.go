package pagination

import (
	"context"
	"iter"
	"sync"
)

// PageFetcher defines the function signature for fetching a single page of results.
// It receives a context and a cursor string (empty for the first page), and returns:
//   - items: the slice of results for this page
//   - nextCursor: the cursor for the next page (empty string when no more pages)
//   - err: any error encountered during the fetch
type PageFetcher[T any] func(ctx context.Context, cursor string) (items []T, nextCursor string, err error)

// Iterator provides lazy, single-pass iteration over paginated API results.
// It fetches pages on demand as items are consumed, keeping at most one page
// of results in memory at any time.
//
// Usage with Next/Item loop:
//
//	it := pagination.NewIterator(myFetcher)
//	for it.Next(ctx) {
//	    item := it.Item()
//	    // process item
//	}
//	if err := it.Err(); err != nil {
//	    // handle error
//	}
//
// Usage with Go 1.23+ range-over-function:
//
//	for item, err := range it.All() {
//	    if err != nil {
//	        // handle error
//	        break
//	    }
//	    // process item
//	}
type Iterator[T any] struct {
	fetcher PageFetcher[T]
	items   []T    // current page buffer
	index   int    // position within current page (points to current item after Next returns true)
	cursor  string // cursor for the next page fetch
	done    bool   // true when all pages have been consumed
	err     error  // first error encountered, sticky
	started bool   // true after the first fetch has been issued
}

// NewIterator creates a new Iterator backed by the given page-fetching function.
// The fetcher is called with a cursor (empty string for the first page) and returns
// a slice of items, the next cursor (empty when done), and any error.
//
// The iterator does not fetch any data until Next() is called for the first time,
// making construction essentially free.
func NewIterator[T any](fetcher PageFetcher[T]) *Iterator[T] {
	return &Iterator[T]{fetcher: fetcher}
}

// Next advances the iterator to the next item, fetching a new page from the
// underlying fetcher if the current page buffer is exhausted.
//
// It returns true if an item is available (retrieve it with Item()), or false
// when iteration is complete or an error occurred (check with Err()).
//
// Next respects context cancellation: if the context is cancelled before or
// during a page fetch, the fetcher's context-aware error will be propagated
// through Err().
//
// The state machine works as follows:
//  1. If done or err is set, return false immediately (terminal state).
//  2. If there are unconsumed items in the current page, advance the index
//     and return true.
//  3. If more pages are available (not started yet, or cursor is non-empty),
//     fetch the next page. On success, set the buffer and return true for
//     the first item. On error or empty terminal page, transition to the
//     appropriate terminal state.
//  4. Otherwise, mark done and return false.
func (it *Iterator[T]) Next(ctx context.Context) bool {
	// Terminal states: once done or errored, iteration is over.
	if it.done || it.err != nil {
		return false
	}

	// Case 1: We have more items in the current page buffer.
	// After the initial fetch, index starts at 0 and we already returned true
	// for item[0], so subsequent calls advance to the next item.
	if it.started && it.index+1 < len(it.items) {
		it.index++
		return true
	}

	// Case 2: Current page exhausted. Do we have more pages to fetch?
	if !it.started || it.cursor != "" {
		return it.fetchNextPage(ctx)
	}

	// Case 3: No more pages, no more items. We're done.
	it.done = true

	return false
}

// fetchNextPage requests the next page from the fetcher and updates the
// iterator's internal state. Returns true if at least one item is available.
func (it *Iterator[T]) fetchNextPage(ctx context.Context) bool {
	items, nextCursor, err := it.fetcher(ctx, it.cursor)
	it.started = true

	if err != nil {
		it.err = err
		return false
	}

	if len(items) == 0 {
		// The fetcher returned an empty page with no error and (presumably)
		// no next cursor. This signals end-of-stream.
		it.done = true
		return false
	}

	it.items = items
	it.index = 0
	it.cursor = nextCursor

	return true
}

// Item returns the current item. It must only be called after Next() has
// returned true. Calling Item() without a preceding successful Next() call
// or after Next() returned false produces undefined behavior.
func (it *Iterator[T]) Item() T {
	return it.items[it.index]
}

// Err returns the error that terminated iteration, or nil if iteration
// completed successfully (all pages consumed without error).
//
// Err should be checked after the Next() loop exits to distinguish between
// normal completion and error-terminated iteration.
func (it *Iterator[T]) Err() error {
	return it.err
}

// AllCtx returns a Go 1.23+ range-over-function iterator (iter.Seq2) that
// yields each item paired with a nil error, then-if the underlying fetcher
// failed-yields a single (zero, error) sentinel at the end.
//
// The provided context is threaded through every page fetch, enabling
// timeout propagation, cancellation support, and context value access.
//
// This enables idiomatic usage with for-range:
//
//	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
//	defer cancel()
//
//	for item, err := range it.AllCtx(ctx) {
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println(item)
//	}
//
// Breaking out of the loop early is safe and stops further page fetching.
func (it *Iterator[T]) AllCtx(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for it.Next(ctx) {
			if !yield(it.Item(), nil) {
				return
			}
		}

		if it.Err() != nil {
			var zero T
			yield(zero, it.Err())
		}
	}
}

// All returns a Go 1.23+ range-over-function iterator (iter.Seq2) that
// yields each item paired with a nil error, then-if the underlying fetcher
// failed-yields a single (zero, error) sentinel at the end.
//
// This enables idiomatic usage with for-range:
//
//	for item, err := range it.All() {
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println(item)
//	}
//
// Breaking out of the loop early is safe and stops further page fetching.
// The iterator uses context.Background() internally; for context-aware
// iteration, prefer AllCtx(ctx) or the Next()/Item() loop directly.
func (it *Iterator[T]) All() iter.Seq2[T, error] {
	return it.AllCtx(context.Background())
}

// Collect drains all remaining items from the iterator into a slice.
// It fetches pages until the stream is exhausted or an error occurs.
//
// If an error occurs mid-stream, Collect returns all items successfully
// retrieved so far along with the error. On success, the error is nil.
//
// This is a convenience method for cases where all results fit comfortably
// in memory. For large result sets, prefer the Next()/Item() loop or
// ForEachConcurrent.
func (it *Iterator[T]) Collect(ctx context.Context) ([]T, error) {
	var all []T
	for it.Next(ctx) {
		all = append(all, it.Item())
	}

	return all, it.Err()
}

// CollectN drains at most n items from the iterator into a slice.
// Pagination stops as soon as n items have been collected, even if
// more pages are available-no unnecessary fetches are made.
//
// If fewer than n items exist in the stream, all available items are returned.
// If an error occurs before n items are collected, the partial result is
// returned alongside the error.
func (it *Iterator[T]) CollectN(ctx context.Context, n int) ([]T, error) {
	if n <= 0 {
		return nil, nil
	}

	var all []T
	for it.Next(ctx) {
		all = append(all, it.Item())
		if len(all) >= n {
			break
		}
	}

	if it.Err() != nil {
		return all, it.Err()
	}

	return all, nil
}

// ForEachConcurrent processes items from the iterator using a pool of
// concurrent workers. It drains the iterator sequentially (maintaining
// single-pass semantics) but dispatches each item to a goroutine from
// the worker pool for parallel processing.
//
// Parameters:
//   - ctx: context for both iteration and worker functions
//   - it: the iterator to drain
//   - workers: maximum number of concurrent goroutines processing items
//   - fn: the function to call for each item; errors are collected
//
// Returns a slice of all errors returned by fn calls, plus any iterator
// error appended at the end. Returns nil if everything succeeded.
//
// The function blocks until all items have been processed and all workers
// have completed. The semaphore pattern ensures that at most `workers`
// goroutines are active simultaneously, providing backpressure to the
// iteration loop.
func ForEachConcurrent[T any](
	ctx context.Context,
	it *Iterator[T],
	workers int,
	fn func(ctx context.Context, item T) error,
) []error {
	if workers <= 0 {
		workers = 1
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
		sem  = make(chan struct{}, workers)
	)

	for it.Next(ctx) {
		item := it.Item()

		sem <- struct{}{}

		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			if err := fn(ctx, item); err != nil {
				mu.Lock()

				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if it.Err() != nil {
		errs = append(errs, it.Err())
	}

	return errs
}
