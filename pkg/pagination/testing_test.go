package pagination

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIteratorFromSlice(t *testing.T) {
	t.Parallel()

	it := NewIteratorFromSlice([]string{"a", "b", "c"})
	ctx := context.Background()

	items, err := it.Collect(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, items)
}

func TestNewIteratorFromSliceEmpty(t *testing.T) {
	t.Parallel()

	it := NewIteratorFromSlice[string](nil)
	ctx := context.Background()

	assert.False(t, it.Next(ctx))
	assert.Nil(t, it.Err())
}

func TestNewIteratorFromSliceSingleItem(t *testing.T) {
	t.Parallel()

	it := NewIteratorFromSlice([]int{42})
	ctx := context.Background()

	items, err := it.Collect(ctx)
	require.NoError(t, err)
	assert.Equal(t, []int{42}, items)
}

func TestNewIteratorFromSliceIterateManually(t *testing.T) {
	t.Parallel()

	it := NewIteratorFromSlice([]string{"x", "y"})
	ctx := context.Background()

	require.True(t, it.Next(ctx))
	assert.Equal(t, "x", it.Item())
	require.True(t, it.Next(ctx))
	assert.Equal(t, "y", it.Item())
	assert.False(t, it.Next(ctx))
	assert.Nil(t, it.Err())
}

func TestNewErrorIterator(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection refused")
	it := NewErrorIterator[string](expectedErr)
	ctx := context.Background()

	assert.False(t, it.Next(ctx))
	assert.ErrorIs(t, it.Err(), expectedErr)
}

func TestNewErrorIteratorCollect(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("timeout")
	it := NewErrorIterator[int](expectedErr)
	ctx := context.Background()

	items, err := it.Collect(ctx)
	assert.ErrorIs(t, err, expectedErr)
	assert.Empty(t, items)
}

func TestNewIteratorFromSliceCollectN(t *testing.T) {
	t.Parallel()

	it := NewIteratorFromSlice([]int{1, 2, 3, 4, 5})
	ctx := context.Background()

	items, err := it.CollectN(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, items)
}

func TestNewIteratorFromSliceWithAll(t *testing.T) {
	t.Parallel()

	it := NewIteratorFromSlice([]string{"a", "b"})

	collected := make([]string, 0, 2)

	for item, err := range it.All() {
		require.NoError(t, err)

		collected = append(collected, item)
	}

	assert.Equal(t, []string{"a", "b"}, collected)
}
