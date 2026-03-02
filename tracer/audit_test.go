package tracer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleAuditEvent returns a test AuditEvent with sensible defaults.
func sampleAuditEvent(id string) AuditEvent {
	return AuditEvent{
		ID:         id,
		Type:       "mutation",
		Action:     "create",
		Actor:      "user-123",
		Resource:   "Rule",
		ResourceID: "rule-456",
		Details:    map[string]any{"field": "status", "from": "DRAFT", "to": "ACTIVE"},
		Timestamp:  time.Now(),
		CreatedAt:  time.Now(),
	}
}

// sampleAuditVerification returns a test AuditVerification.
func sampleAuditVerification(eventID string, valid bool) AuditVerification {
	return AuditVerification{
		ID:         "ver-" + eventID,
		EventID:    eventID,
		Valid:      valid,
		Hash:       "sha256:abc123def456",
		VerifiedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestAuditEventsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/audit-events/evt-42", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleAuditEvent("evt-42"), result)
		},
	}

	svc := newAuditEventsService(mock)
	event, err := svc.Get(context.Background(), "evt-42")

	require.NoError(t, err)
	require.NotNil(t, event)
	assert.Equal(t, "evt-42", event.ID)
	assert.Equal(t, "mutation", event.Type)
	assert.Equal(t, "create", event.Action)
	assert.Equal(t, "user-123", event.Actor)
}

func TestAuditEventsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAuditEventsService(&mockBackend{})
	event, err := svc.Get(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, event)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAuditEventsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAuditEventsService(mock)
	event, err := svc.Get(context.Background(), "evt-missing")

	require.Error(t, err)
	assert.Nil(t, event)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestAuditEventsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/audit-events")
			assert.Nil(t, body)

			resp := models.ListResponse[AuditEvent]{
				Items: []AuditEvent{
					sampleAuditEvent("evt-1"),
					sampleAuditEvent("evt-2"),
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newAuditEventsService(mock)
	iter := svc.List(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "evt-1", items[0].ID)
	assert.Equal(t, "evt-2", items[1].ID)
}

func TestAuditEventsListWithCursor(t *testing.T) {
	t.Parallel()

	callCount := 0

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			callCount++
			assert.Equal(t, "GET", method)
			assert.Nil(t, body)

			if callCount == 1 {
				assert.Equal(t, "/audit-events", path)

				resp := models.ListResponse[AuditEvent]{
					Items: []AuditEvent{sampleAuditEvent("evt-1")},
					Pagination: models.Pagination{
						Total:      2,
						Limit:      1,
						NextCursor: "cursor-page-2",
					},
				}

				return unmarshalInto(resp, result)
			}

			// Second page -- cursor should be in the path.
			assert.Contains(t, path, "cursor=cursor-page-2")

			resp := models.ListResponse[AuditEvent]{
				Items: []AuditEvent{sampleAuditEvent("evt-2")},
				Pagination: models.Pagination{
					Total: 2,
					Limit: 1,
				},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newAuditEventsService(mock)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "evt-1", items[0].ID)
	assert.Equal(t, "evt-2", items[1].ID)
	assert.Equal(t, 2, callCount)
}

func TestAuditEventsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, _ string, path string, _ any, result any) error {
			receivedPath = path

			resp := models.ListResponse[AuditEvent]{
				Items:      []AuditEvent{sampleAuditEvent("evt-1")},
				Pagination: models.Pagination{Total: 1, Limit: 25},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newAuditEventsService(mock)
	opts := &models.ListOptions{
		Limit:     25,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	iter := svc.List(context.Background(), opts)

	ctx := context.Background()
	require.True(t, iter.Next(ctx))

	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortBy=timestamp")
	assert.Contains(t, receivedPath, "sortOrder=desc")
}

func TestAuditEventsListEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			resp := models.ListResponse[AuditEvent]{
				Items:      []AuditEvent{},
				Pagination: models.Pagination{Total: 0, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newAuditEventsService(mock)
	iter := svc.List(context.Background(), nil)

	ctx := context.Background()
	assert.False(t, iter.Next(ctx))
	require.NoError(t, iter.Err())
}

func TestAuditEventsListBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: internal server error")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAuditEventsService(mock)
	iter := svc.List(context.Background(), nil)

	ctx := context.Background()
	assert.False(t, iter.Next(ctx))
	assert.Equal(t, expectedErr, iter.Err())
}

// ---------------------------------------------------------------------------
// Verify
// ---------------------------------------------------------------------------

func TestAuditEventsVerify(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/audit-events/evt-1/verify", path)
			assert.Nil(t, body)

			return unmarshalInto(sampleAuditVerification("evt-1", true), result)
		},
	}

	svc := newAuditEventsService(mock)
	ver, err := svc.Verify(context.Background(), "evt-1")

	require.NoError(t, err)
	require.NotNil(t, ver)
	assert.Equal(t, "evt-1", ver.EventID)
	assert.True(t, ver.Valid)
	assert.Equal(t, "sha256:abc123def456", ver.Hash)
}

func TestAuditEventsVerifyEmptyID(t *testing.T) {
	t.Parallel()

	svc := newAuditEventsService(&mockBackend{})
	ver, err := svc.Verify(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, ver)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestAuditEventsVerifyBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend: verification failed")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newAuditEventsService(mock)
	ver, err := svc.Verify(context.Background(), "evt-1")

	require.Error(t, err)
	assert.Nil(t, ver)
	assert.Equal(t, expectedErr, err)
}

func TestAuditEventsVerifyInvalid(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/audit-events/evt-tampered/verify", path)

			return unmarshalInto(sampleAuditVerification("evt-tampered", false), result)
		},
	}

	svc := newAuditEventsService(mock)
	ver, err := svc.Verify(context.Background(), "evt-tampered")

	require.NoError(t, err)
	require.NotNil(t, ver)
	assert.Equal(t, "evt-tampered", ver.EventID)
	assert.False(t, ver.Valid)
}
