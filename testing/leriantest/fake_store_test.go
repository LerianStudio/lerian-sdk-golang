package leriantest

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testItem is a simple struct used as the type parameter for fakeStore tests.
type testItem struct {
	ID   string
	Name string
}

// ---------------------------------------------------------------------------
// TestFakeStore_SetAndGet — basic store-and-retrieve round-trip.
// ---------------------------------------------------------------------------

func TestFakeStore_SetAndGet(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()
	item := testItem{ID: "1", Name: "Alice"}

	s.Set("1", item)

	got, ok := s.Get("1")
	require.True(t, ok)
	assert.Equal(t, item, got)
}

// ---------------------------------------------------------------------------
// TestFakeStore_GetNotFound — Get on a nonexistent key returns zero value.
// ---------------------------------------------------------------------------

func TestFakeStore_GetNotFound(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()

	got, ok := s.Get("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, testItem{}, got)
}

// ---------------------------------------------------------------------------
// TestFakeStore_Delete — Set, Delete, then verify Get returns not-found.
// ---------------------------------------------------------------------------

func TestFakeStore_Delete(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()
	s.Set("1", testItem{ID: "1", Name: "Alice"})

	s.Delete("1")

	got, ok := s.Get("1")
	assert.False(t, ok)
	assert.Equal(t, testItem{}, got)
}

// ---------------------------------------------------------------------------
// TestFakeStore_DeleteNotFound — Delete on a nonexistent key does not panic.
// ---------------------------------------------------------------------------

func TestFakeStore_DeleteNotFound(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()

	// Must not panic.
	assert.NotPanics(t, func() {
		s.Delete("nonexistent")
	})

	// Store should still be empty.
	assert.Equal(t, 0, s.Len())
}

// ---------------------------------------------------------------------------
// TestFakeStore_List — multiple items returned in insertion order.
// ---------------------------------------------------------------------------

func TestFakeStore_List(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()
	s.Set("a", testItem{ID: "a", Name: "Alpha"})
	s.Set("b", testItem{ID: "b", Name: "Bravo"})
	s.Set("c", testItem{ID: "c", Name: "Charlie"})

	items := s.List()
	require.Len(t, items, 3)
	assert.Equal(t, "a", items[0].ID)
	assert.Equal(t, "b", items[1].ID)
	assert.Equal(t, "c", items[2].ID)
}

// ---------------------------------------------------------------------------
// TestFakeStore_ListEmpty — empty store returns empty slice, not nil.
// ---------------------------------------------------------------------------

func TestFakeStore_ListEmpty(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()

	items := s.List()
	assert.NotNil(t, items, "List should return non-nil empty slice")
	assert.Empty(t, items)
}

// ---------------------------------------------------------------------------
// TestFakeStore_Len — verify count after various operations.
// ---------------------------------------------------------------------------

func TestFakeStore_Len(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()
	assert.Equal(t, 0, s.Len())

	s.Set("1", testItem{ID: "1", Name: "One"})
	assert.Equal(t, 1, s.Len())

	s.Set("2", testItem{ID: "2", Name: "Two"})
	assert.Equal(t, 2, s.Len())

	s.Delete("1")
	assert.Equal(t, 1, s.Len())

	s.Delete("2")
	assert.Equal(t, 0, s.Len())
}

// ---------------------------------------------------------------------------
// TestFakeStore_SetOverwrite — overwriting a key updates the value but does
// not duplicate the key in the order slice.
// ---------------------------------------------------------------------------

func TestFakeStore_SetOverwrite(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()
	s.Set("1", testItem{ID: "1", Name: "Original"})
	s.Set("2", testItem{ID: "2", Name: "Second"})
	s.Set("1", testItem{ID: "1", Name: "Updated"})

	// Value should be the latest.
	got, ok := s.Get("1")
	require.True(t, ok)
	assert.Equal(t, "Updated", got.Name)

	// Length should not have increased.
	assert.Equal(t, 2, s.Len())

	// List should not contain duplicates and should preserve original order.
	items := s.List()
	require.Len(t, items, 2)
	assert.Equal(t, "1", items[0].ID)
	assert.Equal(t, "Updated", items[0].Name)
	assert.Equal(t, "2", items[1].ID)
}

// ---------------------------------------------------------------------------
// TestFakeStore_ConcurrentAccess — multiple goroutines hammering the store
// concurrently must not trigger data races.
//
// IMPORTANT: we use assert (not require) inside goroutines, because require
// calls t.FailNow() which panics in non-test goroutines.
// ---------------------------------------------------------------------------

func TestFakeStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()

	const goroutines = 20
	const opsPerGoroutine = 50

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()

			for j := range opsPerGoroutine {
				key := fmt.Sprintf("key-%d-%d", idx, j)
				item := testItem{ID: key, Name: fmt.Sprintf("item-%d-%d", idx, j)}

				s.Set(key, item)

				got, ok := s.Get(key)
				// Use assert, NOT require, inside goroutines.
				assert.True(t, ok, "expected key %s to exist", key)
				assert.Equal(t, key, got.ID)

				// Delete every other item to exercise the Delete path.
				if j%2 == 0 {
					s.Delete(key)
				}
			}
		}(i)
	}

	wg.Wait()

	// After the dust settles, List and Len should agree.
	items := s.List()
	assert.Equal(t, s.Len(), len(items))
}

// ---------------------------------------------------------------------------
// TestFakeStore_InsertionOrder — verify that insertion order is preserved
// correctly through Set, Delete, and re-Set operations.
// ---------------------------------------------------------------------------

func TestFakeStore_InsertionOrder(t *testing.T) {
	t.Parallel()

	s := newFakeStore[testItem]()

	// Insert A, B, C in order.
	s.Set("a", testItem{ID: "a", Name: "Alpha"})
	s.Set("b", testItem{ID: "b", Name: "Bravo"})
	s.Set("c", testItem{ID: "c", Name: "Charlie"})

	// Delete B.
	s.Delete("b")

	// Re-insert B — it should now appear at the end.
	s.Set("b", testItem{ID: "b", Name: "Bravo-v2"})

	items := s.List()
	require.Len(t, items, 3)

	// Expected order: A, C, B (because B was deleted and re-inserted).
	assert.Equal(t, "a", items[0].ID)
	assert.Equal(t, "c", items[1].ID)
	assert.Equal(t, "b", items[2].ID)
	assert.Equal(t, "Bravo-v2", items[2].Name)
}
