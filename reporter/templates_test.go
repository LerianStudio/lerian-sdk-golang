package reporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// templatesServiceAPI.Create tests (multipart/form-data)
// ---------------------------------------------------------------------------

func TestTemplatesCreate(t *testing.T) {
	t.Parallel()

	fileContent := "<!DOCTYPE html><html><body>{{title}}</body></html>"

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string,
			headers map[string]string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/templates", path)

			// Verify Content-Type is multipart/form-data with boundary.
			contentType, ok := headers["Content-Type"]
			require.True(t, ok, "Content-Type header must be set")
			assert.Contains(t, contentType, "multipart/form-data")

			rawBody, ok := body.([]byte)
			require.True(t, ok, "body must be []byte")
			require.NotEmpty(t, rawBody)

			// Parse the multipart body to verify fields and file content.
			_, params, err := mime.ParseMediaType(contentType)
			require.NoError(t, err)

			boundary := params["boundary"]
			require.NotEmpty(t, boundary)

			reader := multipart.NewReader(bytes.NewReader(rawBody), boundary)

			fields := make(map[string]string)

			var fileData []byte

			for {
				part, partErr := reader.NextPart()
				if partErr == io.EOF {
					break
				}

				require.NoError(t, partErr)

				partBytes, readErr := io.ReadAll(part)
				require.NoError(t, readErr)

				if part.FileName() != "" {
					// This is the file part.
					fileData = partBytes
				} else {
					// This is a form field.
					fields[part.FormName()] = string(partBytes)
				}
			}

			// Verify form fields.
			assert.Equal(t, "Sales Report", fields["name"])
			assert.Equal(t, "html", fields["format"])
			assert.Equal(t, "Monthly sales template", fields["description"])

			// Verify file content.
			assert.Equal(t, fileContent, string(fileData))

			// Simulate response.
			return unmarshalInto(Template{
				ID:      "tpl-1",
				Name:    "Sales Report",
				Format:  "html",
				Version: 1,
			}, result)
		},
	}

	svc := newTemplatesService(mock)

	desc := "Monthly sales template"
	input := &CreateTemplateInput{
		Name:        "Sales Report",
		Format:      "html",
		Description: &desc,
	}

	tpl, err := svc.Create(context.Background(), input, strings.NewReader(fileContent))

	require.NoError(t, err)
	require.NotNil(t, tpl)
	assert.Equal(t, "tpl-1", tpl.ID)
	assert.Equal(t, "Sales Report", tpl.Name)
	assert.Equal(t, "html", tpl.Format)
	assert.Equal(t, 1, tpl.Version)
}

func TestTemplatesCreateWithoutDescription(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, _, _ string,
			headers map[string]string, body, result any) error {
			// Parse the multipart body and verify no description field.
			contentType := headers["Content-Type"]
			_, params, err := mime.ParseMediaType(contentType)
			require.NoError(t, err)

			payload, ok := body.([]byte)
			require.True(t, ok, "body must be []byte")

			reader := multipart.NewReader(bytes.NewReader(payload), params["boundary"])

			fields := make(map[string]string)

			for {
				part, partErr := reader.NextPart()
				if partErr == io.EOF {
					break
				}

				require.NoError(t, partErr)

				if part.FileName() == "" {
					partBytes, _ := io.ReadAll(part)
					fields[part.FormName()] = string(partBytes)
				}
			}

			assert.Equal(t, "Basic Template", fields["name"])
			assert.Equal(t, "csv", fields["format"])
			_, hasDesc := fields["description"]
			assert.False(t, hasDesc, "description field should not be present")

			return unmarshalInto(Template{ID: "tpl-2", Name: "Basic Template", Format: "csv"}, result)
		},
	}

	svc := newTemplatesService(mock)
	input := &CreateTemplateInput{Name: "Basic Template", Format: "csv"}
	tpl, err := svc.Create(context.Background(), input, strings.NewReader("col1,col2\n"))

	require.NoError(t, err)
	require.NotNil(t, tpl)
	assert.Equal(t, "tpl-2", tpl.ID)
}

func TestTemplatesCreateNilBackendUsesCoreError(t *testing.T) {
	t.Parallel()

	svc := newTemplatesService(nil)
	tpl, err := svc.Create(context.Background(), &CreateTemplateInput{Name: "Sales Report", Format: "html"}, strings.NewReader("<html></html>"))

	require.Error(t, err)
	assert.Nil(t, tpl)
	assert.ErrorIs(t, err, core.ErrNilBackend)
}

func TestTemplatesCreateTypedNilBackendUsesCoreError(t *testing.T) {
	t.Parallel()

	var mb *mockBackend

	svc := newTemplatesService(mb)
	tpl, err := svc.Create(context.Background(), &CreateTemplateInput{Name: "Sales Report", Format: "html"}, strings.NewReader("<html></html>"))

	require.Error(t, err)
	assert.Nil(t, tpl)
	assert.ErrorIs(t, err, core.ErrNilBackend)
}

func TestTemplatesCreateNilInput(t *testing.T) {
	t.Parallel()

	svc := newTemplatesService(&mockBackend{})
	tpl, err := svc.Create(context.Background(), nil, strings.NewReader("data"))

	require.Error(t, err)
	assert.Nil(t, tpl)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Contains(t, sdkErr.Message, "input is required")
}

func TestTemplatesCreateNilFile(t *testing.T) {
	t.Parallel()

	svc := newTemplatesService(&mockBackend{})
	tpl, err := svc.Create(context.Background(), &CreateTemplateInput{Name: "x", Format: "pdf"}, nil)

	require.Error(t, err)
	assert.Nil(t, tpl)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Contains(t, sdkErr.Message, "file is required")
}

func TestTemplatesCreateBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("upload failed")
	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, _, _ string,
			_ map[string]string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTemplatesService(mock)
	input := &CreateTemplateInput{Name: "x", Format: "pdf"}
	tpl, err := svc.Create(context.Background(), input, strings.NewReader("data"))

	require.Error(t, err)
	assert.Nil(t, tpl)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// templatesServiceAPI.Get tests
// ---------------------------------------------------------------------------

func TestTemplatesGet(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "GET", method)
			assert.Equal(t, "/templates/tpl-1", path)
			assert.Nil(t, body)

			return unmarshalInto(Template{
				ID:       "tpl-1",
				Name:     "Invoice Template",
				Format:   "docx",
				FileType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				Version:  3,
			}, result)
		},
	}

	svc := newTemplatesService(mock)
	tpl, err := svc.Get(context.Background(), "tpl-1")

	require.NoError(t, err)
	require.NotNil(t, tpl)
	assert.Equal(t, "tpl-1", tpl.ID)
	assert.Equal(t, "Invoice Template", tpl.Name)
	assert.Equal(t, 3, tpl.Version)
}

func TestTemplatesGetEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTemplatesService(&mockBackend{})
	tpl, err := svc.Get(context.Background(), "")

	require.Error(t, err)
	assert.Nil(t, tpl)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTemplatesGetBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("not found")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTemplatesService(mock)
	tpl, err := svc.Get(context.Background(), "tpl-missing")

	require.Error(t, err)
	assert.Nil(t, tpl)
	assert.Equal(t, expectedErr, err)
}

// ---------------------------------------------------------------------------
// templatesServiceAPI.List tests
// ---------------------------------------------------------------------------

func TestTemplatesList(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, _ string, _, result any) error {
			assert.Equal(t, "GET", method)

			resp := models.ListResponse[Template]{
				Items: []Template{
					{ID: "tpl-1", Name: "Template A"},
					{ID: "tpl-2", Name: "Template B"},
				},
				Pagination: models.Pagination{Total: 2, Limit: 10},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newTemplatesService(mock)
	iter := svc.List(context.Background(), nil)

	items, err := iter.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "tpl-1", items[0].ID)
	assert.Equal(t, "tpl-2", items[1].ID)
}

func TestTemplatesListWithOptions(t *testing.T) {
	t.Parallel()

	var receivedPath string

	mock := &mockBackend{
		callFn: func(_ context.Context, _, path string, _, result any) error {
			receivedPath = path

			resp := models.ListResponse[Template]{
				Items:      []Template{{ID: "tpl-1"}},
				Pagination: models.Pagination{Total: 1, Limit: 50},
			}

			return unmarshalInto(resp, result)
		},
	}

	svc := newTemplatesService(mock)
	opts := &models.CursorListOptions{Limit: 50, SortBy: "createdAt", SortOrder: "desc"}
	iter := svc.List(context.Background(), opts)

	require.True(t, iter.Next(context.Background()))
	assert.Contains(t, receivedPath, "limit=50")
	assert.Contains(t, receivedPath, "sortBy=createdAt")
	assert.Contains(t, receivedPath, "sortOrder=desc")
}

func TestTemplatesListError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("list error")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTemplatesService(mock)
	iter := svc.List(context.Background(), nil)

	assert.False(t, iter.Next(context.Background()))
	assert.Equal(t, expectedErr, iter.Err())
}

// ---------------------------------------------------------------------------
// templatesServiceAPI.Delete tests
// ---------------------------------------------------------------------------

func TestTemplatesDelete(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "DELETE", method)
			assert.Equal(t, "/templates/tpl-1", path)
			assert.Nil(t, body)
			assert.Nil(t, result)

			return nil
		},
	}

	svc := newTemplatesService(mock)
	err := svc.Delete(context.Background(), "tpl-1")

	require.NoError(t, err)
}

func TestTemplatesDeleteEmptyID(t *testing.T) {
	t.Parallel()

	svc := newTemplatesService(&mockBackend{})
	err := svc.Delete(context.Background(), "")

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTemplatesDeleteBackendError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("forbidden")
	mock := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return expectedErr
		},
	}

	svc := newTemplatesService(mock)
	err := svc.Delete(context.Background(), "tpl-1")

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
