package pagination

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// stringFetcher builds a PageFetcher[string] that serves the given pages
// sequentially, using the page index as the cursor ("1", "2", etc.).
// An atomic counter tracks how many times the fetcher was called.
func stringFetcher(pages [][]string, counter *atomic.Int32) PageFetcher[string] {
	return func(_ context.Context, cursor string) ([]string, string, error) {
		counter.Add(1)

		idx := 0

		if cursor != "" {
			// Parse the cursor to determine the page index.
			for i, c := range "0123456789" {
				if cursor == string(c) {
					idx = i
					break
				}
			}
		}

		if idx >= len(pages) {
			return nil, "", nil
		}

		nextCursor := ""
		if idx+1 < len(pages) {
			nextCursor = fmt.Sprintf("%d", idx+1)
		}

		return pages[idx], nextCursor, nil
	}
}

// intFetcher builds a PageFetcher[int] that serves the given pages
// sequentially with cursor-based pagination.
func intFetcher(pages [][]int, counter *atomic.Int32) PageFetcher[int] {
	return func(_ context.Context, cursor string) ([]int, string, error) {
		counter.Add(1)

		idx := 0

		if cursor != "" {
			for i, c := range "0123456789" {
				if cursor == string(c) {
					idx = i
					break
				}
			}
		}

		if idx >= len(pages) {
			return nil, "", nil
		}

		nextCursor := ""
		if idx+1 < len(pages) {
			nextCursor = fmt.Sprintf("%d", idx+1)
		}

		return pages[idx], nextCursor, nil
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestSinglePageIteration(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	fetcher := stringFetcher([][]string{{"a", "b", "c"}}, &calls)
	it := NewIterator(fetcher)
	ctx := context.Background()

	// Should yield exactly 3 items.
	var items []string
	for it.Next(ctx) {
		items = append(items, it.Item())
	}

	assert.Equal(t, []string{"a", "b", "c"}, items, "should yield all 3 items")
	assert.NoError(t, it.Err(), "should complete without error")
	assert.Equal(t, int32(1), calls.Load(), "fetcher should be called exactly once")

	// One more Next should still be false.
	assert.False(t, it.Next(ctx), "Next after exhaustion should return false")
}

func TestMultiPageIteration(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	// 3 pages: [1,2] -> [3,4] -> [5]
	pages := [][]int{{1, 2}, {3, 4}, {5}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)
	ctx := context.Background()

	var items []int
	for it.Next(ctx) {
		items = append(items, it.Item())
	}

	assert.Equal(t, []int{1, 2, 3, 4, 5}, items, "should yield all 5 items in order across 3 pages")
	assert.NoError(t, it.Err(), "should complete without error")
	assert.Equal(t, int32(3), calls.Load(), "fetcher should be called once per page")
}

func TestEmptyResult(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	// Fetcher returns empty page immediately.
	fetcher := func(_ context.Context, _ string) ([]string, string, error) {
		calls.Add(1)
		return nil, "", nil
	}

	it := NewIterator(fetcher)
	ctx := context.Background()

	assert.False(t, it.Next(ctx), "Next should return false for empty result")
	assert.NoError(t, it.Err(), "error should be nil for empty result")
	assert.Equal(t, int32(1), calls.Load(), "fetcher should be called once")
}

func TestErrorOnFirstPage(t *testing.T) {
	t.Parallel()

	errFetch := errors.New("network timeout")

	fetcher := func(_ context.Context, _ string) ([]string, string, error) {
		return nil, "", errFetch
	}

	it := NewIterator(fetcher)
	ctx := context.Background()

	assert.False(t, it.Next(ctx), "Next should return false when fetcher errors")
	assert.ErrorIs(t, it.Err(), errFetch, "Err should return the fetcher error")
}

func TestErrorOnSecondPage(t *testing.T) {
	t.Parallel()

	errFetch := errors.New("server error on page 2")

	var calls atomic.Int32

	fetcher := func(_ context.Context, cursor string) ([]string, string, error) {
		calls.Add(1)

		if cursor == "" {
			// First page succeeds.
			return []string{"x", "y"}, "page2", nil
		}
		// Second page fails.
		return nil, "", errFetch
	}

	it := NewIterator(fetcher)
	ctx := context.Background()

	// First page items should be yielded successfully.
	var items []string
	for it.Next(ctx) {
		items = append(items, it.Item())
	}

	assert.Equal(t, []string{"x", "y"}, items, "should yield first page items before error")
	assert.ErrorIs(t, it.Err(), errFetch, "Err should return the second page error")
	assert.Equal(t, int32(2), calls.Load(), "fetcher should be called twice")
}

func TestCollect(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	// 3 pages totaling 7 items.
	pages := [][]int{{1, 2, 3}, {4, 5}, {6, 7}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	items, err := it.Collect(context.Background())

	assert.NoError(t, err, "Collect should not error")
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, items, "Collect should return all 7 items")
	assert.Equal(t, int32(3), calls.Load(), "fetcher should be called 3 times")
}

func TestCollectWithError(t *testing.T) {
	t.Parallel()

	errFetch := errors.New("fetch failed")

	fetcher := func(_ context.Context, cursor string) ([]int, string, error) {
		if cursor == "" {
			return []int{10, 20, 30}, "next", nil
		}

		return nil, "", errFetch
	}

	it := NewIterator(fetcher)
	items, err := it.Collect(context.Background())

	assert.ErrorIs(t, err, errFetch, "Collect should return the fetch error")
	assert.Equal(t, []int{10, 20, 30}, items, "Collect should return items from successful pages")
}

func TestCollectN(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	// 3 pages of 5 items each = 15 items total.
	pages := [][]int{
		{1, 2, 3, 4, 5},
		{6, 7, 8, 9, 10},
		{11, 12, 13, 14, 15},
	}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	items, err := it.CollectN(context.Background(), 7)

	assert.NoError(t, err, "CollectN should not error")
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, items, "CollectN(7) should return exactly 7 items")

	// Only 2 pages should have been fetched: page 1 (items 1-5) and page 2
	// (to get items 6-7). Page 3 should never be requested.
	assert.Equal(t, int32(2), calls.Load(),
		"fetcher should only be called twice (no unnecessary page 3 fetch)")
}

func TestCollectNExceedsTotal(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	// Single page with 3 items.
	pages := [][]string{{"alpha", "beta", "gamma"}}
	fetcher := stringFetcher(pages, &calls)
	it := NewIterator(fetcher)

	items, err := it.CollectN(context.Background(), 10)

	assert.NoError(t, err, "CollectN should not error when n > total")
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, items,
		"CollectN(10) should return all 3 available items")
	assert.Equal(t, int32(1), calls.Load(), "fetcher should be called once")
}

func TestAll(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	// 2 pages: ["hello", "world"] and ["foo"]
	pages := [][]string{{"hello", "world"}, {"foo"}}
	fetcher := stringFetcher(pages, &calls)
	it := NewIterator(fetcher)

	items := make([]string, 0, 3)

	for item, err := range it.All() {
		require.NoError(t, err, "All should not yield errors on success")

		items = append(items, item)
	}

	assert.Equal(t, []string{"hello", "world", "foo"}, items,
		"All should yield all items from both pages")
}

func TestAllWithError(t *testing.T) {
	t.Parallel()

	errPage2 := errors.New("page 2 failed")

	fetcher := func(_ context.Context, cursor string) ([]string, string, error) {
		if cursor == "" {
			return []string{"ok1", "ok2"}, "next", nil
		}

		return nil, "", errPage2
	}

	it := NewIterator(fetcher)

	var items []string

	var gotErr error

	for item, err := range it.All() {
		if err != nil {
			gotErr = err
			break
		}

		items = append(items, item)
	}

	assert.Equal(t, []string{"ok1", "ok2"}, items,
		"should yield first-page items before error")
	assert.ErrorIs(t, gotErr, errPage2, "should yield the fetch error")
}

func TestForEachConcurrent(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	pages := [][]int{{1, 2, 3, 4, 5}, {6, 7, 8, 9, 10}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	var mu sync.Mutex

	var processed []int

	errs := ForEachConcurrent(context.Background(), it, 3,
		func(_ context.Context, item int) error {
			mu.Lock()

			processed = append(processed, item)
			mu.Unlock()

			return nil
		},
	)

	assert.Empty(t, errs, "should have no errors")

	// Items may arrive out of order due to concurrency, but all should be present.
	assert.Len(t, processed, 10, "all 10 items should be processed")

	for i := 1; i <= 10; i++ {
		assert.Contains(t, processed, i, "item %d should be in processed set", i)
	}
}

func TestForEachConcurrentErrors(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	pages := [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	errs := ForEachConcurrent(context.Background(), it, 3,
		func(_ context.Context, item int) error {
			if item%2 == 0 {
				return fmt.Errorf("error processing %d", item)
			}

			return nil
		},
	)

	// Even items: 2, 4, 6, 8, 10 = 5 errors.
	assert.Len(t, errs, 5, "should collect 5 errors (one per even item)")
}

func TestSinglePassSemantics(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	pages := [][]string{{"a", "b"}, {"c"}}
	fetcher := stringFetcher(pages, &calls)
	it := NewIterator(fetcher)
	ctx := context.Background()

	// First pass: consume everything.
	var firstPass []string
	for it.Next(ctx) {
		firstPass = append(firstPass, it.Item())
	}

	assert.Equal(t, []string{"a", "b", "c"}, firstPass, "first pass should yield all items")

	// Second pass: should yield nothing.
	secondStartCalls := calls.Load()

	var secondPass []string
	for it.Next(ctx) {
		secondPass = append(secondPass, it.Item())
	}

	assert.Empty(t, secondPass, "second iteration should yield no items (single-pass)")
	assert.Equal(t, secondStartCalls, calls.Load(),
		"fetcher should not be called again on second pass")
	assert.NoError(t, it.Err(), "error should still be nil after second pass attempt")
}

func TestContextCancellationDuringIteration(t *testing.T) {
	t.Parallel()

	// Create a fetcher where the second page respects context cancellation.
	fetcher := func(ctx context.Context, cursor string) ([]int, string, error) {
		if cursor == "" {
			return []int{1, 2, 3}, "page2", nil
		}

		// Simulate a fetch that checks context before doing work.
		if err := ctx.Err(); err != nil {
			return nil, "", fmt.Errorf("fetch cancelled: %w", err)
		}

		return []int{4, 5, 6}, "", nil
	}

	it := NewIterator(fetcher)

	// Create a context that we cancel after the first page.
	ctx, cancel := context.WithCancel(context.Background())

	var items []int
	for it.Next(ctx) {
		items = append(items, it.Item())

		// Cancel context after consuming item 3 (end of first page).
		// The next call to Next() will try to fetch page 2 with a cancelled context.
		if it.Item() == 3 {
			cancel()
		}
	}

	assert.Equal(t, []int{1, 2, 3}, items,
		"should yield all first-page items before cancellation takes effect")
	assert.Error(t, it.Err(), "should have an error from cancelled context")
	assert.ErrorIs(t, it.Err(), context.Canceled,
		"error should wrap context.Canceled")

	// Cleanup (no-op since we already cancelled, but good practice).
	cancel()
}

// ---------------------------------------------------------------------------
// Edge case tests
// ---------------------------------------------------------------------------

func TestIteratorWithSingleItem(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	pages := [][]string{{"only"}}
	fetcher := stringFetcher(pages, &calls)
	it := NewIterator(fetcher)
	ctx := context.Background()

	require.True(t, it.Next(ctx), "Next should return true for single item")
	assert.Equal(t, "only", it.Item(), "should yield the single item")
	assert.False(t, it.Next(ctx), "Next should return false after single item")
	assert.NoError(t, it.Err())
}

func TestCollectOnFreshIterator(t *testing.T) {
	t.Parallel()

	// Collect on an iterator that hasn't had Next called yet.
	var calls atomic.Int32

	pages := [][]int{{100, 200}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	items, err := it.Collect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, []int{100, 200}, items)
}

func TestForEachConcurrentWithIteratorError(t *testing.T) {
	t.Parallel()

	errPage := errors.New("page fetch error")

	callCount := 0
	fetcher := func(_ context.Context, cursor string) ([]int, string, error) {
		callCount++

		if cursor == "" {
			return []int{1, 2}, "next", nil
		}

		return nil, "", errPage
	}

	it := NewIterator(fetcher)

	errs := ForEachConcurrent(context.Background(), it, 2,
		func(_ context.Context, _ int) error {
			return nil
		},
	)

	// The iterator error should appear in the errors slice.
	require.Len(t, errs, 1, "should have one error (the iterator error)")
	assert.ErrorIs(t, errs[0], errPage, "the error should be the page fetch error")
}

func TestForEachConcurrentWithZeroWorkers(t *testing.T) {
	t.Parallel()

	// workers=0 previously caused a deadlock because make(chan struct{}, 0)
	// creates an unbuffered channel, blocking the first send forever.
	// The guard should clamp workers to 1, allowing normal execution.
	var calls atomic.Int32

	pages := [][]int{{1, 2, 3}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	var mu sync.Mutex

	var processed []int

	// Use a timeout to catch deadlocks: if this blocks for 5 seconds, the test fails.
	done := make(chan []error, 1)

	go func() {
		done <- ForEachConcurrent(context.Background(), it, 0,
			func(_ context.Context, item int) error {
				mu.Lock()

				processed = append(processed, item)
				mu.Unlock()

				return nil
			},
		)
	}()

	select {
	case errs := <-done:
		assert.Empty(t, errs, "should have no errors with workers=0")
		assert.Len(t, processed, 3, "all 3 items should be processed")

		for i := 1; i <= 3; i++ {
			assert.Contains(t, processed, i, "item %d should be in processed set", i)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ForEachConcurrent with workers=0 deadlocked")
	}
}

func TestForEachConcurrentWithNegativeWorkers(t *testing.T) {
	t.Parallel()

	// workers=-1 previously caused a panic because make(chan struct{}, -1)
	// panics at runtime. The guard should clamp workers to 1.
	var calls atomic.Int32

	pages := [][]int{{10, 20}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	var mu sync.Mutex

	var processed []int

	done := make(chan []error, 1)

	go func() {
		done <- ForEachConcurrent(context.Background(), it, -1,
			func(_ context.Context, item int) error {
				mu.Lock()

				processed = append(processed, item)
				mu.Unlock()

				return nil
			},
		)
	}()

	select {
	case errs := <-done:
		assert.Empty(t, errs, "should have no errors with workers=-1")
		assert.Len(t, processed, 2, "both items should be processed")
		assert.Contains(t, processed, 10)
		assert.Contains(t, processed, 20)
	case <-time.After(5 * time.Second):
		t.Fatal("ForEachConcurrent with workers=-1 deadlocked or panicked")
	}
}

func TestCollectN_Zero(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	pages := [][]int{{1, 2, 3}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	items, err := it.CollectN(context.Background(), 0)

	assert.NoError(t, err, "CollectN(0) should not error")
	assert.Empty(t, items, "CollectN(0) should return an empty slice")
	assert.Equal(t, int32(0), calls.Load(),
		"CollectN(0) should not invoke the fetcher at all")
}

func TestCollectN_Negative(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	pages := [][]int{{1, 2, 3}}
	fetcher := intFetcher(pages, &calls)
	it := NewIterator(fetcher)

	items, err := it.CollectN(context.Background(), -1)

	assert.NoError(t, err, "CollectN(-1) should not error")
	assert.Empty(t, items, "CollectN(-1) should return an empty slice")
	assert.Equal(t, int32(0), calls.Load(),
		"CollectN(-1) should not invoke the fetcher at all")
}

func TestAllCtx_CancelledContext(t *testing.T) {
	t.Parallel()

	// Fetcher where page 2 is slow enough that a cancelled context
	// will prevent it from completing.
	fetcher := func(ctx context.Context, cursor string) ([]int, string, error) {
		if cursor == "" {
			return []int{1, 2, 3}, "page2", nil
		}

		// Respect context cancellation before returning page 2.
		if err := ctx.Err(); err != nil {
			return nil, "", fmt.Errorf("fetch aborted: %w", err)
		}

		return []int{4, 5, 6}, "", nil
	}

	it := NewIterator(fetcher)

	ctx, cancel := context.WithCancel(context.Background())

	var items []int

	var gotErr error

	for item, err := range it.AllCtx(ctx) {
		if err != nil {
			gotErr = err
			break
		}

		items = append(items, item)

		// Cancel after consuming the last item of page 1.
		if item == 3 {
			cancel()
		}
	}

	assert.Equal(t, []int{1, 2, 3}, items,
		"should yield all first-page items before cancellation")
	assert.Error(t, gotErr, "should receive an error from cancelled context")
	assert.ErrorIs(t, gotErr, context.Canceled,
		"error should wrap context.Canceled")

	cancel() // idempotent cleanup
}

func TestAllCtx_WithValues(t *testing.T) {
	t.Parallel()

	type ctxKey string

	const testKey ctxKey = "test-key"

	// Fetcher that reads a value from the context and includes it in results.
	fetcher := func(ctx context.Context, cursor string) ([]string, string, error) {
		val, ok := ctx.Value(testKey).(string)
		if !ok {
			return nil, "", fmt.Errorf("expected context value for key %q", testKey)
		}

		if cursor == "" {
			return []string{val + "-page1"}, "next", nil
		}

		return []string{val + "-page2"}, "", nil
	}

	it := NewIterator(fetcher)

	ctx := context.WithValue(context.Background(), testKey, "hello")

	items := make([]string, 0, 2)

	for item, err := range it.AllCtx(ctx) {
		require.NoError(t, err, "AllCtx should not yield errors when context carries value")

		items = append(items, item)
	}

	assert.Equal(t, []string{"hello-page1", "hello-page2"}, items,
		"fetcher should receive context values through AllCtx")
}
