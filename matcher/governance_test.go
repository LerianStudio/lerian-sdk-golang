package matcher

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

// ---------------------------------------------------------------------------
// ListArchives
// ---------------------------------------------------------------------------

func TestGovernanceListArchives(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/archives")

		resp := models.ListResponse[Archive]{
			Items: []Archive{
				{ID: "arc-1", ContextID: "ctx-1", Type: "monthly", RecordCount: 500},
				{ID: "arc-2", ContextID: "ctx-1", Type: "quarterly", RecordCount: 1500},
			},
			Pagination: models.Pagination{Total: 2, Limit: 10},
		}

		return unmarshalInto(resp, result)
	}}

	svc := newGovernanceService(mb)
	iter := svc.ListArchives(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "arc-1", items[0].ID)
	assert.Equal(t, "arc-2", items[1].ID)
}

func TestGovernanceListArchivesWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mb := &mockBackend{callFn: func(_ context.Context, _, path string, _, result any) error {
		receivedPath = path

		resp := models.ListResponse[Archive]{
			Items:      []Archive{{ID: "arc-1"}},
			Pagination: models.Pagination{Total: 1, Limit: 5},
		}

		return unmarshalInto(resp, result)
	}}

	svc := newGovernanceService(mb)
	opts := &models.ListOptions{Limit: 5, SortBy: "createdAt"}
	iter := svc.ListArchives(context.Background(), opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=5")
	assert.Contains(t, receivedPath, "sortBy=createdAt")
}

// ---------------------------------------------------------------------------
// GetArchive
// ---------------------------------------------------------------------------

func TestGovernanceGetArchive(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
		assert.Equal(t, "GET", method)
		assert.Equal(t, "/archives/arc-1", path)
		assert.Nil(t, body)

		return unmarshalInto(Archive{
			ID:          "arc-1",
			ContextID:   "ctx-1",
			Type:        "monthly",
			RecordCount: 500,
			CreatedAt:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		}, result)
	}}

	svc := newGovernanceService(mb)
	got, err := svc.GetArchive(context.Background(), "arc-1")

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "arc-1", got.ID)
	assert.Equal(t, "ctx-1", got.ContextID)
	assert.Equal(t, "monthly", got.Type)
	assert.Equal(t, 500, got.RecordCount)
}

func TestGovernanceGetArchiveEmptyID(t *testing.T) {
	t.Parallel()

	svc := newGovernanceService(&mockBackend{callFn: noopCallFn})
	got, err := svc.GetArchive(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestGovernanceGetArchiveBackendError(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return errors.New("backend failure")
	}}

	svc := newGovernanceService(mb)
	got, err := svc.GetArchive(context.Background(), "arc-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "backend failure")
}

// ---------------------------------------------------------------------------
// ListAuditLogs
// ---------------------------------------------------------------------------

func TestGovernanceListAuditLogs(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, _, result any) error {
		assert.Equal(t, "GET", method)
		assert.Contains(t, path, "/audit-logs")

		resp := models.ListResponse[AuditLog]{
			Items: []AuditLog{
				{ID: "log-1", Action: "create", Actor: "user-1", Resource: "context"},
				{ID: "log-2", Action: "update", Actor: "user-2", Resource: "rule"},
			},
			Pagination: models.Pagination{Total: 2, Limit: 10},
		}

		return unmarshalInto(resp, result)
	}}

	svc := newGovernanceService(mb)
	iter := svc.ListAuditLogs(context.Background(), nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "log-1", items[0].ID)
	assert.Equal(t, "log-2", items[1].ID)
}

// ---------------------------------------------------------------------------
// GetAuditLog
// ---------------------------------------------------------------------------

func TestGovernanceGetAuditLog(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
		assert.Equal(t, "GET", method)
		assert.Equal(t, "/audit-logs/log-1", path)
		assert.Nil(t, body)

		return unmarshalInto(AuditLog{
			ID:         "log-1",
			ContextID:  "ctx-1",
			Action:     "create",
			Actor:      "user-1",
			Resource:   "context",
			ResourceID: "ctx-1",
		}, result)
	}}

	svc := newGovernanceService(mb)
	got, err := svc.GetAuditLog(context.Background(), "log-1")

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "log-1", got.ID)
	assert.Equal(t, "create", got.Action)
	assert.Equal(t, "user-1", got.Actor)
}

func TestGovernanceGetAuditLogEmptyID(t *testing.T) {
	t.Parallel()

	svc := newGovernanceService(&mockBackend{callFn: noopCallFn})
	got, err := svc.GetAuditLog(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestGovernanceGetAuditLogBackendError(t *testing.T) {
	t.Parallel()

	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return errors.New("service unavailable")
	}}

	svc := newGovernanceService(mb)
	got, err := svc.GetAuditLog(context.Background(), "log-1")

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "service unavailable")
}

// ---------------------------------------------------------------------------
// Governance — Additional Backend Error Propagation
// ---------------------------------------------------------------------------

func TestGovernanceListArchivesBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newGovernanceService(mb)
	iter := svc.ListArchives(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

func TestGovernanceListAuditLogsBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: internal")
	mb := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		return expectedErr
	}}

	svc := newGovernanceService(mb)
	iter := svc.ListAuditLogs(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ GovernanceService = (*governanceService)(nil)
