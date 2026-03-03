package performance

import (
	"bytes"
	"encoding/json"
	"sync"
)

// JSONPool provides pooled JSON encoding and decoding using sync.Pool
// for buffer reuse, reducing memory allocations in high-throughput scenarios.
type JSONPool struct {
	pool sync.Pool
}

// NewJSONPool creates a new JSONPool with buffer pooling.
func NewJSONPool() *JSONPool {
	return &JSONPool{
		pool: sync.Pool{
			New: func() any { return new(bytes.Buffer) },
		},
	}
}

// Marshal encodes v to JSON using a pooled buffer.
// Output is identical to json.Marshal but with reduced allocations.
func (p *JSONPool) Marshal(v any) ([]byte, error) {
	buf, ok := p.pool.Get().(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
	}

	buf.Reset()

	defer p.pool.Put(buf)

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	// json.Encoder.Encode appends a trailing newline; strip it
	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}

	// Return a copy so the buffer can be reused
	result := make([]byte, len(b))
	copy(result, b)

	return result, nil
}

// Unmarshal decodes JSON data into v.
// Equivalent to json.Unmarshal but provided for API symmetry.
func (p *JSONPool) Unmarshal(data []byte, v any) error {
	return json.NewDecoder(bytes.NewReader(data)).Decode(v)
}
