package errors

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TestErrorCategoryConstants — all 10 constants have correct string values
// ---------------------------------------------------------------------------

func TestErrorCategoryConstants(t *testing.T) {
	tests := []struct {
		constant ErrorCategory
		want     string
	}{
		{CategoryValidation, "validation"},
		{CategoryNotFound, "not_found"},
		{CategoryAuthentication, "authentication"},
		{CategoryAuthorization, "authorization"},
		{CategoryConflict, "conflict"},
		{CategoryRateLimit, "rate_limit"},
		{CategoryNetwork, "network"},
		{CategoryTimeout, "timeout"},
		{CategoryCancellation, "cancellation"},
		{CategoryInternal, "internal"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, string(tc.constant))
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorStruct — all 11 fields accessible
// ---------------------------------------------------------------------------

func TestErrorStruct(t *testing.T) {
	inner := fmt.Errorf("disk full")

	e := &Error{
		Product:    "ledger",
		Category:   CategoryInternal,
		Code:       ErrorCode("disk_full"),
		Message:    "storage backend unavailable",
		Operation:  "accounts.Create",
		Resource:   "Account",
		ResourceID: "acc-123",
		StatusCode: 500,
		RequestID:  "req-abc",
		HelpURL:    "https://docs.lerian.io/errors/disk_full",
		Err:        inner,
	}

	assert.Equal(t, "ledger", e.Product)
	assert.Equal(t, CategoryInternal, e.Category)
	assert.Equal(t, ErrorCode("disk_full"), e.Code)
	assert.Equal(t, "storage backend unavailable", e.Message)
	assert.Equal(t, "accounts.Create", e.Operation)
	assert.Equal(t, "Account", e.Resource)
	assert.Equal(t, "acc-123", e.ResourceID)
	assert.Equal(t, 500, e.StatusCode)
	assert.Equal(t, "req-abc", e.RequestID)
	assert.Equal(t, "https://docs.lerian.io/errors/disk_full", e.HelpURL)
	assert.Equal(t, inner, e.Err)
}

// ---------------------------------------------------------------------------
// TestErrorString — table-driven: with RequestID, without, empty Product
// ---------------------------------------------------------------------------

func TestErrorString(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "with request ID",
			err: &Error{
				Product:   "ledger",
				Category:  CategoryNotFound,
				Operation: "accounts.Get",
				Message:   "Account not found: acc-123",
				RequestID: "req-xyz-789",
			},
			want: "[LEDGER] not_found during accounts.Get: Account not found: acc-123 (request_id: req-xyz-789)",
		},
		{
			name: "without request ID",
			err: &Error{
				Product:   "transaction",
				Category:  CategoryValidation,
				Operation: "transactions.Create",
				Message:   "amount must be positive",
			},
			want: "[TRANSACTION] validation during transactions.Create: amount must be positive",
		},
		{
			name: "empty product",
			err: &Error{
				Product:   "",
				Category:  CategoryNetwork,
				Operation: "health.Check",
				Message:   "connection refused",
			},
			want: "[] network during health.Check: connection refused",
		},
		{
			name: "all fields populated",
			err: &Error{
				Product:    "ledger",
				Category:   CategoryConflict,
				Code:       ErrorCode("duplicate_key"),
				Message:    "account already exists",
				Operation:  "accounts.Create",
				Resource:   "Account",
				ResourceID: "acc-dup",
				StatusCode: 409,
				RequestID:  "req-conflict",
				HelpURL:    "https://docs.lerian.io/errors",
			},
			want: "[LEDGER] conflict during accounts.Create: account already exists (request_id: req-conflict)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err.Error()
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorIs — category matching, code-level matching
// ---------------------------------------------------------------------------

func TestErrorIs(t *testing.T) {
	tests := []struct {
		name   string
		err    *Error
		target *Error
		want   bool
	}{
		{
			name:   "same category matches",
			err:    &Error{Category: CategoryNotFound, Code: "missing_account"},
			target: &Error{Category: CategoryNotFound},
			want:   true,
		},
		{
			name:   "different category does not match",
			err:    &Error{Category: CategoryNotFound},
			target: &Error{Category: CategoryValidation},
			want:   false,
		},
		{
			name:   "same category and same code matches",
			err:    &Error{Category: CategoryValidation, Code: "invalid_email"},
			target: &Error{Category: CategoryValidation, Code: "invalid_email"},
			want:   true,
		},
		{
			name:   "same category but different code does not match",
			err:    &Error{Category: CategoryValidation, Code: "invalid_email"},
			target: &Error{Category: CategoryValidation, Code: "missing_field"},
			want:   false,
		},
		{
			name:   "target without code matches any code in same category",
			err:    &Error{Category: CategoryValidation, Code: "any_code"},
			target: &Error{Category: CategoryValidation},
			want:   true,
		},
		{
			name:   "source without code and target with code does not match",
			err:    &Error{Category: CategoryValidation},
			target: &Error{Category: CategoryValidation, Code: "specific_code"},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err.Is(tc.target)
			assert.Equal(t, tc.want, got)
		})
	}

	// Also verify Is returns false for non-Error targets.
	t.Run("non-Error target returns false", func(t *testing.T) {
		e := &Error{Category: CategoryNetwork}
		got := e.Is(fmt.Errorf("plain error"))
		assert.False(t, got)
	})
}

// ---------------------------------------------------------------------------
// TestSentinelErrors — errors.Is works with all 10 sentinels
// ---------------------------------------------------------------------------

func TestSentinelErrors(t *testing.T) {
	sentinels := []struct {
		name     string
		sentinel *Error
		category ErrorCategory
	}{
		{"ErrValidation", ErrValidation, CategoryValidation},
		{"ErrNotFound", ErrNotFound, CategoryNotFound},
		{"ErrAuthentication", ErrAuthentication, CategoryAuthentication},
		{"ErrAuthorization", ErrAuthorization, CategoryAuthorization},
		{"ErrConflict", ErrConflict, CategoryConflict},
		{"ErrRateLimit", ErrRateLimit, CategoryRateLimit},
		{"ErrNetwork", ErrNetwork, CategoryNetwork},
		{"ErrTimeout", ErrTimeout, CategoryTimeout},
		{"ErrCancellation", ErrCancellation, CategoryCancellation},
		{"ErrInternal", ErrInternal, CategoryInternal},
	}

	for _, s := range sentinels {
		t.Run(s.name+"_positive", func(t *testing.T) {
			// A rich error with the same category should match the sentinel.
			rich := &Error{
				Product:   "ledger",
				Category:  s.category,
				Operation: "test.Op",
				Message:   "something happened",
			}
			assert.True(t, stderrors.Is(rich, s.sentinel),
				"errors.Is(rich, %s) should be true", s.name)
		})

		t.Run(s.name+"_negative", func(t *testing.T) {
			// Pick a category that is definitely different.
			differentCategory := CategoryInternal
			if s.category == CategoryInternal {
				differentCategory = CategoryValidation
			}

			other := &Error{
				Product:   "ledger",
				Category:  differentCategory,
				Operation: "test.Op",
				Message:   "unrelated",
			}
			assert.False(t, stderrors.Is(other, s.sentinel),
				"errors.Is(other, %s) should be false", s.name)
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorsAs — extract *Error details via errors.As
// ---------------------------------------------------------------------------

func TestErrorsAs(t *testing.T) {
	original := &Error{
		Product:    "ledger",
		Category:   CategoryNotFound,
		Operation:  "accounts.Get",
		Resource:   "Account",
		ResourceID: "acc-42",
		StatusCode: 404,
		Message:    "Account not found: acc-42",
	}

	// Wrap in a fmt.Errorf chain so errors.As has to unwrap.
	wrapped := fmt.Errorf("service layer: %w", original)

	var target *Error
	require.True(t, stderrors.As(wrapped, &target))

	assert.Equal(t, "ledger", target.Product)
	assert.Equal(t, CategoryNotFound, target.Category)
	assert.Equal(t, "accounts.Get", target.Operation)
	assert.Equal(t, "Account", target.Resource)
	assert.Equal(t, "acc-42", target.ResourceID)
	assert.Equal(t, 404, target.StatusCode)
	assert.Equal(t, "Account not found: acc-42", target.Message)
}

// ---------------------------------------------------------------------------
// TestFactoryFunctions — table-driven for all 9 factories
// ---------------------------------------------------------------------------

func TestFactoryFunctions(t *testing.T) {
	innerErr := fmt.Errorf("connection reset by peer")

	t.Run("NewValidation", func(t *testing.T) {
		e := NewValidation("accounts.Create", "Account", "name is required")
		assert.Equal(t, "sdk", e.Product)
		assert.Equal(t, CategoryValidation, e.Category)
		assert.Equal(t, "accounts.Create", e.Operation)
		assert.Equal(t, "Account", e.Resource)
		assert.Equal(t, "name is required", e.Message)
		assert.Equal(t, 0, e.StatusCode)
		assert.Nil(t, e.Err)
		// Must match sentinel.
		assert.True(t, stderrors.Is(e, ErrValidation))
	})

	t.Run("NewNotFound", func(t *testing.T) {
		e := NewNotFound("accounts.Get", "Account", "acc-99")
		assert.Equal(t, "sdk", e.Product)
		assert.Equal(t, CategoryNotFound, e.Category)
		assert.Equal(t, "accounts.Get", e.Operation)
		assert.Equal(t, "Account", e.Resource)
		assert.Equal(t, "acc-99", e.ResourceID)
		assert.Equal(t, 404, e.StatusCode)
		assert.Equal(t, "Account not found: acc-99", e.Message)
		assert.Nil(t, e.Err)
		assert.True(t, stderrors.Is(e, ErrNotFound))
	})

	t.Run("NewAuthentication", func(t *testing.T) {
		e := NewAuthentication("ledger", "auth.Login", "invalid token")
		assert.Equal(t, "ledger", e.Product)
		assert.Equal(t, CategoryAuthentication, e.Category)
		assert.Equal(t, "auth.Login", e.Operation)
		assert.Equal(t, 401, e.StatusCode)
		assert.Equal(t, "invalid token", e.Message)
		assert.Nil(t, e.Err)
		assert.True(t, stderrors.Is(e, ErrAuthentication))
	})

	t.Run("NewAuthorization", func(t *testing.T) {
		e := NewAuthorization("ledger", "accounts.Delete", "insufficient privileges")
		assert.Equal(t, "ledger", e.Product)
		assert.Equal(t, CategoryAuthorization, e.Category)
		assert.Equal(t, "accounts.Delete", e.Operation)
		assert.Equal(t, 403, e.StatusCode)
		assert.Equal(t, "insufficient privileges", e.Message)
		assert.Nil(t, e.Err)
		assert.True(t, stderrors.Is(e, ErrAuthorization))
	})

	t.Run("NewConflict", func(t *testing.T) {
		e := NewConflict("ledger", "accounts.Create", "Account", "duplicate account name")
		assert.Equal(t, "ledger", e.Product)
		assert.Equal(t, CategoryConflict, e.Category)
		assert.Equal(t, "accounts.Create", e.Operation)
		assert.Equal(t, "Account", e.Resource)
		assert.Equal(t, 409, e.StatusCode)
		assert.Equal(t, "duplicate account name", e.Message)
		assert.Nil(t, e.Err)
		assert.True(t, stderrors.Is(e, ErrConflict))
	})

	t.Run("NewNetwork", func(t *testing.T) {
		e := NewNetwork("ledger", "health.Check", innerErr)
		assert.Equal(t, "ledger", e.Product)
		assert.Equal(t, CategoryNetwork, e.Category)
		assert.Equal(t, "health.Check", e.Operation)
		assert.Equal(t, 0, e.StatusCode)
		assert.Equal(t, innerErr.Error(), e.Message)
		assert.Equal(t, innerErr, e.Err)
		assert.True(t, stderrors.Is(e, ErrNetwork))
		// Unwrap chain reaches inner error.
		assert.True(t, stderrors.Is(e, innerErr))
	})

	t.Run("NewTimeout", func(t *testing.T) {
		timeoutErr := fmt.Errorf("context deadline exceeded")
		e := NewTimeout("transaction", "transactions.Commit", timeoutErr)
		assert.Equal(t, "transaction", e.Product)
		assert.Equal(t, CategoryTimeout, e.Category)
		assert.Equal(t, "transactions.Commit", e.Operation)
		assert.Equal(t, 0, e.StatusCode)
		assert.Equal(t, timeoutErr.Error(), e.Message)
		assert.Equal(t, timeoutErr, e.Err)
		assert.True(t, stderrors.Is(e, ErrTimeout))
		assert.True(t, stderrors.Is(e, timeoutErr))
	})

	t.Run("NewCancellation", func(t *testing.T) {
		cancelErr := fmt.Errorf("context canceled")
		e := NewCancellation("sdk", "batch.Run", cancelErr)
		assert.Equal(t, "sdk", e.Product)
		assert.Equal(t, CategoryCancellation, e.Category)
		assert.Equal(t, "batch.Run", e.Operation)
		assert.Equal(t, 0, e.StatusCode)
		assert.Equal(t, cancelErr.Error(), e.Message)
		assert.Equal(t, cancelErr, e.Err)
		assert.True(t, stderrors.Is(e, ErrCancellation))
		assert.True(t, stderrors.Is(e, cancelErr))
	})

	t.Run("NewInternal", func(t *testing.T) {
		e := NewInternal("ledger", "accounts.Sync", "unexpected nil pointer", innerErr)
		assert.Equal(t, "ledger", e.Product)
		assert.Equal(t, CategoryInternal, e.Category)
		assert.Equal(t, "accounts.Sync", e.Operation)
		assert.Equal(t, 500, e.StatusCode)
		assert.Equal(t, "unexpected nil pointer", e.Message)
		assert.Equal(t, innerErr, e.Err)
		assert.True(t, stderrors.Is(e, ErrInternal))
		assert.True(t, stderrors.Is(e, innerErr))
	})
}

// ---------------------------------------------------------------------------
// TestUnwrapChain — Unwrap returns inner Err, nil when no Err
// ---------------------------------------------------------------------------

func TestUnwrapChain(t *testing.T) {
	t.Run("unwrap returns inner error", func(t *testing.T) {
		inner := fmt.Errorf("root cause")
		e := &Error{
			Category:  CategoryNetwork,
			Operation: "test",
			Message:   "wrapped",
			Err:       inner,
		}
		assert.Equal(t, inner, e.Unwrap())
	})

	t.Run("unwrap returns nil when no inner error", func(t *testing.T) {
		e := &Error{
			Category:  CategoryValidation,
			Operation: "test",
			Message:   "no cause",
		}
		assert.Nil(t, e.Unwrap())
	})

	t.Run("deep unwrap chain", func(t *testing.T) {
		root := fmt.Errorf("disk I/O error")
		mid := fmt.Errorf("storage layer: %w", root)
		top := NewInternal("ledger", "persist", "write failed", mid)

		// errors.Is should see through the full chain.
		assert.True(t, stderrors.Is(top, mid))
		assert.True(t, stderrors.Is(top, root))
	})
}

// ---------------------------------------------------------------------------
// TestErrorNilFields — Error with only Category doesn't panic
// ---------------------------------------------------------------------------

func TestErrorNilFields(t *testing.T) {
	t.Run("minimal error does not panic", func(t *testing.T) {
		e := &Error{Category: CategoryInternal}
		// Should not panic. Output: "[] internal during : "
		require.NotPanics(t, func() {
			_ = e.Error()
		})
		got := e.Error()
		assert.Contains(t, got, "internal")
	})

	t.Run("zero-value Error does not panic", func(t *testing.T) {
		e := &Error{}
		require.NotPanics(t, func() {
			_ = e.Error()
		})
	})

	t.Run("sentinel error string does not panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			_ = ErrNotFound.Error()
		})
		got := ErrNotFound.Error()
		assert.Contains(t, got, "not_found")
	})
}

// ---------------------------------------------------------------------------
// TestErrorCodeType — ErrorCode can be used as a string
// ---------------------------------------------------------------------------

func TestErrorCodeType(t *testing.T) {
	code := ErrorCode("invalid_amount")
	assert.Equal(t, "invalid_amount", string(code))

	// ErrorCode("") is the zero value.
	var zero ErrorCode
	assert.Equal(t, "", string(zero))
}

// ---------------------------------------------------------------------------
// TestErrorJSONTags — verify JSON tag presence (structural, not serialization)
// ---------------------------------------------------------------------------

func TestErrorImplementsErrorInterface(t *testing.T) {
	// Compile-time check that *Error satisfies error.
	var _ error = (*Error)(nil)
}

// ---------------------------------------------------------------------------
// TestFactoryNilError — NewNetwork/NewTimeout/NewCancellation with nil err
// ---------------------------------------------------------------------------

func TestFactoryNilError(t *testing.T) {
	t.Run("NewNetwork with nil err does not panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			e := NewNetwork("ledger", "health.Check", nil)
			assert.Equal(t, "ledger", e.Product)
			assert.Equal(t, CategoryNetwork, e.Category)
			assert.Equal(t, "health.Check", e.Operation)
			assert.Equal(t, "", e.Message)
			assert.Nil(t, e.Err)
			assert.True(t, stderrors.Is(e, ErrNetwork))
		})
	})

	t.Run("NewTimeout with nil err does not panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			e := NewTimeout("transaction", "transactions.Commit", nil)
			assert.Equal(t, "transaction", e.Product)
			assert.Equal(t, CategoryTimeout, e.Category)
			assert.Equal(t, "transactions.Commit", e.Operation)
			assert.Equal(t, "", e.Message)
			assert.Nil(t, e.Err)
			assert.True(t, stderrors.Is(e, ErrTimeout))
		})
	})

	t.Run("NewCancellation with nil err does not panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			e := NewCancellation("sdk", "batch.Run", nil)
			assert.Equal(t, "sdk", e.Product)
			assert.Equal(t, CategoryCancellation, e.Category)
			assert.Equal(t, "batch.Run", e.Operation)
			assert.Equal(t, "", e.Message)
			assert.Nil(t, e.Err)
			assert.True(t, stderrors.Is(e, ErrCancellation))
		})
	})
}

// ---------------------------------------------------------------------------
// TestCategoryFromStatus — maps HTTP status codes to error categories
// ---------------------------------------------------------------------------

func TestCategoryFromStatus(t *testing.T) {
	tests := []struct {
		status int
		want   ErrorCategory
	}{
		{400, CategoryValidation},
		{401, CategoryAuthentication},
		{403, CategoryAuthorization},
		{404, CategoryNotFound},
		{409, CategoryConflict},
		{429, CategoryRateLimit},
		{500, CategoryInternal},
		{503, CategoryInternal},
		{200, CategoryInternal},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("status_%d", tc.status), func(t *testing.T) {
			got := CategoryFromStatus(tc.status)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TruncateBody
// ---------------------------------------------------------------------------

func TestTruncateBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     []byte
		wantLen   int
		wantTrunc bool
	}{
		{
			name:      "empty body",
			input:     nil,
			wantLen:   0,
			wantTrunc: false,
		},
		{
			name:      "short body not truncated",
			input:     []byte("short error"),
			wantLen:   len("short error"),
			wantTrunc: false,
		},
		{
			name:      "exactly at limit not truncated",
			input:     make([]byte, MaxErrorBodyBytes),
			wantLen:   MaxErrorBodyBytes,
			wantTrunc: false,
		},
		{
			name:      "one byte over limit is truncated",
			input:     make([]byte, MaxErrorBodyBytes+1),
			wantLen:   MaxErrorBodyBytes + len("... [truncated]"),
			wantTrunc: true,
		},
		{
			name:      "large body is truncated",
			input:     make([]byte, 4096),
			wantLen:   MaxErrorBodyBytes + len("... [truncated]"),
			wantTrunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Fill non-nil slices with a visible character for readability.
			for i := range tt.input {
				tt.input[i] = 'A'
			}

			result := TruncateBody(tt.input)
			assert.Len(t, result, tt.wantLen)

			if tt.wantTrunc {
				assert.Contains(t, result, "... [truncated]")
			} else {
				assert.NotContains(t, result, "... [truncated]")
			}
		})
	}
}
