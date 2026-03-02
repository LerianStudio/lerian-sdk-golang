// Package performance provides JSON encoding pools for efficient
// serialisation and deserialisation of request/response bodies.
//
// By reusing encoder and decoder buffers through [sync.Pool], the SDK
// avoids per-request heap allocations on the hot path, which materially
// reduces GC pressure under high-throughput workloads.
//
// # Usage
//
// The [JSONPool] is typically created once and shared across all backends:
//
//	pool := performance.NewJSONPool()
//	body, err := pool.Marshal(input)
//	// ...
//	err = pool.Unmarshal(resp.Body, &output)
//
// Product clients do not interact with the pool directly; it is wired
// into [core.Backend] during client initialisation.
package performance
