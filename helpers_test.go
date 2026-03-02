package lerian

import (
	"errors"
	"testing"
	"time"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringHelper(t *testing.T) {
	s := String("hello")
	require.NotNil(t, s)
	assert.Equal(t, "hello", *s)
}

func TestIntHelper(t *testing.T) {
	i := Int(42)
	require.NotNil(t, i)
	assert.Equal(t, 42, *i)
}

func TestInt64Helper(t *testing.T) {
	i := Int64(100)
	require.NotNil(t, i)
	assert.Equal(t, int64(100), *i)
}

func TestBoolHelper(t *testing.T) {
	b := Bool(true)
	require.NotNil(t, b)
	assert.True(t, *b)

	b2 := Bool(false)
	assert.False(t, *b2)
}

func TestFloat64Helper(t *testing.T) {
	f := Float64(3.14)
	require.NotNil(t, f)
	assert.InDelta(t, 3.14, *f, 0.001)
}

func TestTimeHelper(t *testing.T) {
	now := time.Now()
	tp := Time(now)
	require.NotNil(t, tp)
	assert.Equal(t, now, *tp)
}

func TestSentinelReExports(t *testing.T) {
	tests := []struct {
		name     string
		factory  func() error
		sentinel error
	}{
		{"ErrNotFound", func() error { return sdkerrors.NewNotFound("Get", "Org", "org-1") }, ErrNotFound},
		{"ErrValidation", func() error { return sdkerrors.NewValidation("Create", "Org", "bad") }, ErrValidation},
		{"ErrAuthentication", func() error { return sdkerrors.NewAuthentication("sdk", "Get", "bad token") }, ErrAuthentication},
		{"ErrAuthorization", func() error { return sdkerrors.NewAuthorization("sdk", "Delete", "denied") }, ErrAuthorization},
		{"ErrConflict", func() error { return sdkerrors.NewConflict("sdk", "Create", "Org", "exists") }, ErrConflict},
		{"ErrNetwork", func() error { return sdkerrors.NewNetwork("sdk", "Get", errors.New("refused")) }, ErrNetwork},
		{"ErrTimeout", func() error { return sdkerrors.NewTimeout("sdk", "Get", errors.New("timed out")) }, ErrTimeout},
		{"ErrCancellation", func() error { return sdkerrors.NewCancellation("sdk", "Get", errors.New("cancelled")) }, ErrCancellation},
		{"ErrInternal", func() error { return sdkerrors.NewInternal("sdk", "Get", "oops", errors.New("boom")) }, ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.factory()
			assert.True(t, errors.Is(err, tt.sentinel), "expected errors.Is to match %s", tt.name)
		})
	}
}

func TestSentinelCrossMatch(t *testing.T) {
	// NotFound should NOT match Validation
	err := sdkerrors.NewNotFound("Get", "Org", "org-1")
	assert.False(t, errors.Is(err, ErrValidation))
	assert.False(t, errors.Is(err, ErrAuthentication))
	assert.False(t, errors.Is(err, ErrInternal))
}

func TestErrorTypeAlias(t *testing.T) {
	// Verify Error type alias works with errors.As
	err := sdkerrors.NewNotFound("Get", "Org", "org-1")
	var sdkErr *Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, 404, sdkErr.StatusCode)
	assert.Equal(t, "sdk", sdkErr.Product)
}

func TestCategoryReExports(t *testing.T) {
	assert.Equal(t, sdkerrors.CategoryNotFound, CategoryNotFound)
	assert.Equal(t, sdkerrors.CategoryValidation, CategoryValidation)
	assert.Equal(t, sdkerrors.CategoryAuthentication, CategoryAuthentication)
	assert.Equal(t, sdkerrors.CategoryAuthorization, CategoryAuthorization)
	assert.Equal(t, sdkerrors.CategoryConflict, CategoryConflict)
	assert.Equal(t, sdkerrors.CategoryRateLimit, CategoryRateLimit)
	assert.Equal(t, sdkerrors.CategoryNetwork, CategoryNetwork)
	assert.Equal(t, sdkerrors.CategoryTimeout, CategoryTimeout)
	assert.Equal(t, sdkerrors.CategoryCancellation, CategoryCancellation)
	assert.Equal(t, sdkerrors.CategoryInternal, CategoryInternal)
}
