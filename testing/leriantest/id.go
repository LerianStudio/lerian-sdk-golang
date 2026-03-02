package leriantest

import (
	"fmt"
	"sync/atomic"
)

// idCounter is a monotonically increasing counter used to generate unique
// identifiers for fake resources. Using atomic operations makes ID generation
// safe for concurrent use across goroutines.
var idCounter atomic.Int64

// generateID returns a deterministic, unique identifier with the given prefix.
// The format is "<prefix>-<n>" where n starts at 1 and increases monotonically
// for the lifetime of the process.
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, idCounter.Add(1))
}
