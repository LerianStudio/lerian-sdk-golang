// Package pagination provides the [Iterator] for lazy paginated iteration
// over collections returned by Lerian APIs.
//
// [Iterator] fetches pages on demand, yielding items one at a time via
// Next/Item semantics. It handles cursor propagation, end-of-stream
// detection, and error surfacing so callers can range over large result
// sets without manual page management.
//
// The iterator is single-pass by design: once consumed, calling Next
// again always returns false. This matches the streaming nature of
// paginated API responses and avoids caching entire result sets in memory.
//
// # Basic Usage
//
//	it := pagination.NewIterator(myFetcher)
//	for it.Next(ctx) {
//	    process(it.Item())
//	}
//	if err := it.Err(); err != nil { ... }
//
// # Range-Over-Function (Go 1.23+)
//
//	for item, err := range it.All() { ... }
//
// # Convenience Methods
//
// [Iterator.Collect] and [Iterator.CollectN] gather results into a slice.
// The package-level [ForEachConcurrent] function processes items across
// multiple goroutines.
package pagination
