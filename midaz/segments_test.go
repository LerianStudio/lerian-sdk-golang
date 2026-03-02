package midaz

import (
	"context"
	"errors"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Segments — Create
// ---------------------------------------------------------------------------

func TestSegmentsCreate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/segments", path)
			assert.NotNil(t, body)

			input, ok := body.(*CreateSegmentInput)
			require.True(t, ok)
			assert.Equal(t, "Retail", input.Name)

			return unmarshalInto(Segment{
				ID:             "seg-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				Name:           "Retail",
			}, result)
		},
	}

	svc := newSegmentsService(mock)
	seg, err := svc.Create(context.Background(), "org-1", "led-1", &CreateSegmentInput{
		Name: "Retail",
	})

	require.NoError(t, err)
	require.NotNil(t, seg)
	assert.Equal(t, "seg-1", seg.ID)
	assert.Equal(t, "Retail", seg.Name)
}

func TestSegmentsCreateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Create(context.Background(), "", "led-1", &CreateSegmentInput{Name: "X"})

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsCreateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Create(context.Background(), "org-1", "", &CreateSegmentInput{Name: "X"})

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Create(context.Background(), "org-1", "led-1", nil)

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newSegmentsService(mock)
	seg, err := svc.Create(context.Background(), "org-1", "led-1", &CreateSegmentInput{Name: "X"})

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Segments — Get
// ---------------------------------------------------------------------------

func TestSegmentsGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/segments/seg-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Segment{
				ID:             "seg-1",
				OrganizationID: "org-1",
				LedgerID:       "led-1",
				Name:           "Retail",
			}, result)
		},
	}

	svc := newSegmentsService(mock)
	seg, err := svc.Get(context.Background(), "org-1", "led-1", "seg-1")

	require.NoError(t, err)
	require.NotNil(t, seg)
	assert.Equal(t, "seg-1", seg.ID)
	assert.Equal(t, "Retail", seg.Name)
}

func TestSegmentsGetEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Get(context.Background(), "", "led-1", "seg-1")

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsGetEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Get(context.Background(), "org-1", "", "seg-1")

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Get(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newSegmentsService(mock)
	seg, err := svc.Get(context.Background(), "org-1", "led-1", "seg-1")

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Segments — List
// ---------------------------------------------------------------------------

func TestSegmentsList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Contains(t, path, "/organizations/org-1/ledgers/led-1/segments")
			assert.Nil(t, body)

			resp := models.ListResponse[Segment]{
				Items: []Segment{
					{ID: "seg-1", Name: "Retail"},
					{ID: "seg-2", Name: "Wholesale"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newSegmentsService(mock)
	iter := svc.List(context.Background(), "org-1", "led-1", nil)

	require.NotNil(t, iter)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "seg-1", items[0].ID)
	assert.Equal(t, "seg-2", items[1].ID)
}

func TestSegmentsListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			receivedPath = path

			resp := models.ListResponse[Segment]{
				Items:      []Segment{{ID: "seg-1", Name: "Found"}},
				Pagination: models.Pagination{Total: 1, Limit: 25},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newSegmentsService(mock)
	opts := &models.ListOptions{
		Limit:     25,
		SortBy:    "name",
		SortOrder: "asc",
	}

	iter := svc.List(context.Background(), "org-1", "led-1", opts)

	ctx := context.Background()
	require.True(t, iter.Next(ctx))

	assert.Contains(t, receivedPath, "limit=25")
	assert.Contains(t, receivedPath, "sortBy=name")
	assert.Contains(t, receivedPath, "sortOrder=asc")
}

func TestSegmentsListEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	iter := svc.List(context.Background(), "", "led-1", nil)

	require.NotNil(t, iter)
	assert.False(t, iter.Next(context.Background()))
	require.Error(t, iter.Err())
	assert.True(t, errors.Is(iter.Err(), sdkerrors.ErrValidation))
}

func TestSegmentsListEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	iter := svc.List(context.Background(), "org-1", "", nil)

	require.NotNil(t, iter)
	assert.False(t, iter.Next(context.Background()))
	require.Error(t, iter.Err())
	assert.True(t, errors.Is(iter.Err(), sdkerrors.ErrValidation))
}

// ---------------------------------------------------------------------------
// Segments — Update
// ---------------------------------------------------------------------------

func TestSegmentsUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/segments/seg-1", path)
			assert.NotNil(t, body)

			return unmarshalInto(Segment{
				ID:   "seg-1",
				Name: "Updated Segment",
			}, result)
		},
	}

	svc := newSegmentsService(mock)
	newName := "Updated Segment"
	seg, err := svc.Update(context.Background(), "org-1", "led-1", "seg-1", &UpdateSegmentInput{
		Name: &newName,
	})

	require.NoError(t, err)
	require.NotNil(t, seg)
	assert.Equal(t, "seg-1", seg.ID)
	assert.Equal(t, "Updated Segment", seg.Name)
}

func TestSegmentsUpdateEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Update(context.Background(), "", "led-1", "seg-1", &UpdateSegmentInput{})

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsUpdateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Update(context.Background(), "org-1", "", "seg-1", &UpdateSegmentInput{})

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsUpdateEmptyID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Update(context.Background(), "org-1", "led-1", "", &UpdateSegmentInput{})

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsUpdateNilInput(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	seg, err := svc.Update(context.Background(), "org-1", "led-1", "seg-1", nil)

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsUpdateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: conflict")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newSegmentsService(mock)
	newName := "X"
	seg, err := svc.Update(context.Background(), "org-1", "led-1", "seg-1", &UpdateSegmentInput{
		Name: &newName,
	})

	require.Error(t, err)
	assert.Nil(t, seg)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Segments — Delete
// ---------------------------------------------------------------------------

func TestSegmentsDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/segments/seg-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newSegmentsService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "seg-1")

	require.NoError(t, err)
}

func TestSegmentsDeleteEmptyOrgID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	err := svc.Delete(context.Background(), "", "led-1", "seg-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsDeleteEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "", "seg-1")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newSegmentsService(&mockBackend{})
	err := svc.Delete(context.Background(), "org-1", "led-1", "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestSegmentsDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("backend error: forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newSegmentsService(mock)
	err := svc.Delete(context.Background(), "org-1", "led-1", "seg-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// Segments — compile-time interface check
// ---------------------------------------------------------------------------

func TestSegmentsServiceInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ SegmentsService = (*segmentsService)(nil)
	t.Log("segmentsService implements SegmentsService")
}
