package reporter

import "time"

// ---------------------------------------------------------------------------
// Domain models
// ---------------------------------------------------------------------------

// DataSource represents a data source connected to the reporting system.
// Data sources define the upstream systems (databases, APIs, streams) from
// which reports pull their data.
type DataSource struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	Config    map[string]any `json:"config,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// Report represents a generated report. Reports are produced by applying
// a template to a data source with specific parameters, yielding a file
// in the requested format (PDF, CSV, XLSX, etc.).
type Report struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  *string        `json:"description,omitempty"`
	Status       string         `json:"status"`
	Format       string         `json:"format"` // "pdf", "csv", "xlsx"
	TemplateID   *string        `json:"templateId,omitempty"`
	DataSourceID *string        `json:"dataSourceId,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
	GeneratedAt  *time.Time     `json:"generatedAt,omitempty"`
	ExpiresAt    *time.Time     `json:"expiresAt,omitempty"`
	FileSize     *int64         `json:"fileSize,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

// Template represents a report template. Templates define the layout,
// formatting rules, and file type for generated reports.
type Template struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Format      string         `json:"format"`
	FileType    string         `json:"fileType"`
	Version     int            `json:"version"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// ---------------------------------------------------------------------------
// Input types
// ---------------------------------------------------------------------------

// CreateReportInput is the input for creating a new report.
type CreateReportInput struct {
	Name         string         `json:"name"`
	Description  *string        `json:"description,omitempty"`
	Format       string         `json:"format"`
	TemplateID   *string        `json:"templateId,omitempty"`
	DataSourceID *string        `json:"dataSourceId,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// UpdateReportInput is the input for updating a report.
type UpdateReportInput struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CreateTemplateInput is the input for uploading a new template.
// The file content is provided separately via io.Reader.
type CreateTemplateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Format      string  `json:"format"`
}
