package reporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// TemplatesService provides access to Reporter template endpoints.
// Templates define the layout, formatting rules, and file type for generated
// reports. Template creation requires uploading a file via multipart/form-data.
type TemplatesService interface {
	// Create uploads a new template with metadata and file content.
	// The file parameter provides the template file bytes (e.g. a .docx or
	// .html template). The input provides the metadata fields (name, format,
	// optional description).
	Create(ctx context.Context, input *CreateTemplateInput, file io.Reader) (*Template, error)

	// Get retrieves a single template by its unique identifier.
	Get(ctx context.Context, id string) (*Template, error)

	// List returns a paginated iterator over all available templates.
	List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Template]

	// Delete removes a template by ID.
	Delete(ctx context.Context, id string) error
}

// templatesService is the concrete implementation of [TemplatesService].
type templatesService struct {
	core.BaseService
}

// Compile-time interface compliance check.
var _ TemplatesService = (*templatesService)(nil)

// newTemplatesService constructs a [TemplatesService] backed by the given
// [core.Backend].
func newTemplatesService(backend core.Backend) TemplatesService {
	return &templatesService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Create uploads a new template via multipart/form-data.
//
// This method builds a multipart request body containing the metadata fields
// (name, format, description) as form fields and the template file as a file
// part. It uses [core.RawBody] to bypass JSON marshaling in the Backend,
// allowing the pre-built multipart body to be sent verbatim with the correct
// Content-Type header.
func (s *templatesService) Create(ctx context.Context, input *CreateTemplateInput, file io.Reader) (*Template, error) {
	const operation = "Templates.Create"

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, "Template", "input is required")
	}

	if file == nil {
		return nil, sdkerrors.NewValidation(operation, "Template", "file is required")
	}

	// Build the multipart request body.
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	// Write metadata form fields.
	if err := writer.WriteField("name", input.Name); err != nil {
		return nil, fmt.Errorf("reporter: write name field: %w", err)
	}

	if err := writer.WriteField("format", input.Format); err != nil {
		return nil, fmt.Errorf("reporter: write format field: %w", err)
	}

	if input.Description != nil {
		if err := writer.WriteField("description", *input.Description); err != nil {
			return nil, fmt.Errorf("reporter: write description field: %w", err)
		}
	}

	// Write the template file part.
	part, err := writer.CreateFormFile("file", "template")
	if err != nil {
		return nil, fmt.Errorf("reporter: create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("reporter: copy file content: %w", err)
	}

	// Close the writer to finalize the multipart boundary.
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("reporter: close multipart writer: %w", err)
	}

	// Send via CallWithHeaders with RawBody to bypass JSON marshaling
	// and override Content-Type with the multipart boundary.
	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	var result Template

	err = s.Backend.CallWithHeaders(ctx, "POST", "/templates", headers,
		core.RawBody{Data: buf.Bytes()}, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a single template by ID.
func (s *templatesService) Get(ctx context.Context, id string) (*Template, error) {
	const operation = "Templates.Get"

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, "Template", "id is required")
	}

	return core.Get[Template](ctx, &s.BaseService, "/templates/"+url.PathEscape(id))
}

// List returns a paginated iterator over templates.
func (s *templatesService) List(ctx context.Context, opts *models.ListOptions) *pagination.Iterator[Template] {
	return core.List[Template](ctx, &s.BaseService, "/templates", opts)
}

// Delete removes a template by ID.
func (s *templatesService) Delete(ctx context.Context, id string) error {
	const operation = "Templates.Delete"

	if id == "" {
		return sdkerrors.NewValidation(operation, "Template", "id is required")
	}

	return core.Delete(ctx, &s.BaseService, "/templates/"+url.PathEscape(id))
}
