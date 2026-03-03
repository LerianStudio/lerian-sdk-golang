package performance

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestJSONPoolRoundTrip(t *testing.T) {
	t.Parallel()

	pool := NewJSONPool()

	tests := []struct {
		name  string
		input any
		into  func() any // factory for unmarshal target
		check func(t *testing.T, result any)
	}{
		{
			name:  "struct",
			input: testStruct{Name: "test", Age: 30},
			into:  func() any { return &testStruct{} },
			check: func(t *testing.T, result any) {
				t.Helper()

				r := result.(*testStruct)
				assert.Equal(t, "test", r.Name)
				assert.Equal(t, 30, r.Age)
			},
		},
		{
			name:  "slice",
			input: []int{1, 2, 3},
			into:  func() any { return &[]int{} },
			check: func(t *testing.T, result any) {
				t.Helper()

				r := result.(*[]int)
				assert.Equal(t, []int{1, 2, 3}, *r)
			},
		},
		{
			name:  "map",
			input: map[string]string{"key": "value"},
			into:  func() any { return &map[string]string{} },
			check: func(t *testing.T, result any) {
				t.Helper()

				r := result.(*map[string]string)
				assert.Equal(t, "value", (*r)["key"])
			},
		},
		{
			name:  "nil produces null",
			input: nil,
			into:  func() any { var v any; return &v },
			check: func(t *testing.T, result any) {
				t.Helper()

				v := result.(*any)
				assert.Nil(t, *v)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := pool.Marshal(tt.input)
			require.NoError(t, err)

			target := tt.into()
			err = pool.Unmarshal(data, target)
			require.NoError(t, err)

			tt.check(t, target)
		})
	}
}

func TestJSONPoolMatchesStdlib(t *testing.T) {
	t.Parallel()

	pool := NewJSONPool()
	input := testStruct{Name: "compare", Age: 42}

	poolBytes, err := pool.Marshal(input)
	require.NoError(t, err)

	stdBytes, err := json.Marshal(input)
	require.NoError(t, err)

	assert.Equal(t, string(stdBytes), string(poolBytes))
}

func TestJSONPoolConcurrent(t *testing.T) {
	t.Parallel()

	pool := NewJSONPool()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			input := testStruct{Name: "goroutine", Age: n}

			data, err := pool.Marshal(input)
			require.NoError(t, err)

			var result testStruct

			err = pool.Unmarshal(data, &result)
			require.NoError(t, err)
			assert.Equal(t, n, result.Age)
		}(i)
	}

	wg.Wait()
}

func BenchmarkJSONPoolMarshal(b *testing.B) {
	pool := NewJSONPool()
	input := testStruct{Name: "bench", Age: 25}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = pool.Marshal(input)
	}
}

func BenchmarkStdlibMarshal(b *testing.B) {
	input := testStruct{Name: "bench", Age: 25}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(input) //nolint:errchkjson // benchmark: error deliberately ignored
	}
}
