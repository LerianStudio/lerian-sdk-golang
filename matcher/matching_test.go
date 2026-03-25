package matcher

import (
	"context"
	"errors"
	"testing"
	"time"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Shared fixture
// ---------------------------------------------------------------------------

var testMatchResult = MatchResult{
	ID:             "mr-1",
	ContextID:      "ctx-1",
	Status:         "completed",
	MatchedCount:   950,
	UnmatchedCount: 30,
	ExceptionCount: 20,
	Duration:       1500,
	CreatedAt:      time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
}

var testAdjustment = Adjustment{
	ID:        "adj-1",
	ContextID: "ctx-1",
	Type:      "credit",
	Amount:    5000,
	Reason:    "Rounding correction",
	Status:    "applied",
	CreatedAt: time.Date(2026, 1, 15, 13, 0, 0, 0, time.UTC),
}

// ---------------------------------------------------------------------------
// Run
// ---------------------------------------------------------------------------

func TestMatchingRun(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/matching/run", path)
			assert.NotNil(t, body)

			// Verify the body is a *matchRunInput with the correct contextId.
			input, ok := body.(*matchRunInput)
			require.True(t, ok, "body should be *matchRunInput")
			assert.Equal(t, "ctx-1", input.ContextID)

			return unmarshalInto(testMatchResult, result)
		}}

		svc := newMatchingService(mb)
		got, err := svc.Run(context.Background(), "ctx-1")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "mr-1", got.ID)
		assert.Equal(t, "ctx-1", got.ContextID)
		assert.Equal(t, "completed", got.Status)
		assert.Equal(t, 950, got.MatchedCount)
		assert.Equal(t, 30, got.UnmatchedCount)
		assert.Equal(t, 20, got.ExceptionCount)
		assert.Equal(t, int64(1500), got.Duration)
	})

	t.Run("empty contextID", func(t *testing.T) {
		t.Parallel()

		svc := newMatchingService(&mockBackend{callFn: noopCallFn})
		got, err := svc.Run(context.Background(), "")

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

		var sdkErr *sdkerrors.Error
		require.True(t, errors.As(err, &sdkErr))
		assert.Equal(t, "Matching.Run", sdkErr.Operation)
		assert.Contains(t, sdkErr.Message, "contextID is required")
	})

	t.Run("backend error", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return errors.New("context not found")
		}}

		svc := newMatchingService(mb)
		got, err := svc.Run(context.Background(), "ctx-missing")

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "context not found")
	})
}

// ---------------------------------------------------------------------------
// Manual
// ---------------------------------------------------------------------------

func TestMatchingManual(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/matching/manual", path)
			assert.NotNil(t, body)

			input, ok := body.(*ManualMatchInput)
			require.True(t, ok, "body should be *ManualMatchInput")
			assert.Equal(t, "ctx-1", input.ContextID)
			assert.Equal(t, []string{"src-1", "src-2"}, input.SourceRecordIDs)
			assert.Equal(t, []string{"tgt-1", "tgt-2"}, input.TargetRecordIDs)

			return unmarshalInto(testMatchResult, result)
		}}

		svc := newMatchingService(mb)
		got, err := svc.Manual(context.Background(), &ManualMatchInput{
			ContextID:       "ctx-1",
			SourceRecordIDs: []string{"src-1", "src-2"},
			TargetRecordIDs: []string{"tgt-1", "tgt-2"},
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "mr-1", got.ID)
		assert.Equal(t, "completed", got.Status)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newMatchingService(&mockBackend{callFn: noopCallFn})
		got, err := svc.Manual(context.Background(), nil)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

		var sdkErr *sdkerrors.Error
		require.True(t, errors.As(err, &sdkErr))
		assert.Equal(t, "Matching.Manual", sdkErr.Operation)
		assert.Contains(t, sdkErr.Message, "input is required")
	})

	t.Run("backend error", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return errors.New("internal server error")
		}}

		svc := newMatchingService(mb)
		got, err := svc.Manual(context.Background(), &ManualMatchInput{
			ContextID:       "ctx-1",
			SourceRecordIDs: []string{"src-1"},
			TargetRecordIDs: []string{"tgt-1"},
		})

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "internal server error")
	})
}

// ---------------------------------------------------------------------------
// Adjust
// ---------------------------------------------------------------------------

func TestMatchingAdjust(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/matching/adjust", path)
			assert.NotNil(t, body)

			input, ok := body.(*AdjustmentInput)
			require.True(t, ok, "body should be *AdjustmentInput")
			assert.Equal(t, "ctx-1", input.ContextID)
			assert.Equal(t, "credit", input.Type)
			assert.Equal(t, int64(5000), input.Amount)
			assert.Equal(t, "Rounding correction", input.Reason)

			return unmarshalInto(testAdjustment, result)
		}}

		svc := newMatchingService(mb)
		got, err := svc.Adjust(context.Background(), &AdjustmentInput{
			ContextID: "ctx-1",
			Type:      "credit",
			Amount:    5000,
			Reason:    "Rounding correction",
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "adj-1", got.ID)
		assert.Equal(t, "ctx-1", got.ContextID)
		assert.Equal(t, "credit", got.Type)
		assert.Equal(t, int64(5000), got.Amount)
		assert.Equal(t, "Rounding correction", got.Reason)
		assert.Equal(t, "applied", got.Status)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		svc := newMatchingService(&mockBackend{callFn: noopCallFn})
		got, err := svc.Adjust(context.Background(), nil)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

		var sdkErr *sdkerrors.Error
		require.True(t, errors.As(err, &sdkErr))
		assert.Equal(t, "Matching.Adjust", sdkErr.Operation)
		assert.Contains(t, sdkErr.Message, "input is required")
	})

	t.Run("backend error", func(t *testing.T) {
		t.Parallel()

		mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return errors.New("conflict: adjustment already exists")
		}}

		svc := newMatchingService(mb)
		got, err := svc.Adjust(context.Background(), &AdjustmentInput{
			ContextID: "ctx-1",
			Type:      "debit",
			Amount:    1000,
			Reason:    "Test",
		})

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "conflict")
	})
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ matchingServiceAPI = (*matchingService)(nil)
