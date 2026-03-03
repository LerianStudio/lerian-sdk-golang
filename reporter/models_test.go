package reporter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// DataSource — JSON round-trip
// ---------------------------------------------------------------------------

func TestDataSourceJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	original := DataSource{
		ID:     "ds-001",
		Name:   "Main Database",
		Type:   "postgresql",
		Status: "active",
		Config: map[string]any{
			"host": "db.example.com",
			"port": float64(5432), // JSON numbers unmarshal as float64
		},
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded DataSource

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.Config["host"], decoded.Config["host"])
	assert.Equal(t, original.Config["port"], decoded.Config["port"])
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt))
	assert.True(t, original.UpdatedAt.Equal(decoded.UpdatedAt))
}

func TestDataSourceJSONOmitsEmptyConfig(t *testing.T) {
	t.Parallel()

	ds := DataSource{
		ID:        "ds-002",
		Name:      "Empty Config",
		Type:      "mysql",
		Status:    "inactive",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(ds)
	require.NoError(t, err)

	assert.NotContains(t, string(data), `"config"`)
}

// ---------------------------------------------------------------------------
// Report — JSON round-trip
// ---------------------------------------------------------------------------

func TestReportJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC)
	desc := "Monthly financial summary"
	tmplID := "tmpl-abc"
	dsID := "ds-001"
	generatedAt := now.Add(5 * time.Minute)
	expiresAt := now.Add(30 * 24 * time.Hour)
	fileSize := int64(1048576)

	original := Report{
		ID:           "rpt-001",
		Name:         "March 2026 Summary",
		Description:  &desc,
		Status:       "completed",
		Format:       "pdf",
		TemplateID:   &tmplID,
		DataSourceID: &dsID,
		Parameters: map[string]any{
			"month": "2026-03",
		},
		GeneratedAt: &generatedAt,
		ExpiresAt:   &expiresAt,
		FileSize:    &fileSize,
		Metadata: map[string]any{
			"author": "system",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Report

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.Format, decoded.Format)
	require.NotNil(t, decoded.TemplateID)
	assert.Equal(t, *original.TemplateID, *decoded.TemplateID)
	require.NotNil(t, decoded.DataSourceID)
	assert.Equal(t, *original.DataSourceID, *decoded.DataSourceID)
	assert.Equal(t, original.Parameters["month"], decoded.Parameters["month"])
	require.NotNil(t, decoded.GeneratedAt)
	assert.True(t, original.GeneratedAt.Equal(*decoded.GeneratedAt))
	require.NotNil(t, decoded.ExpiresAt)
	assert.True(t, original.ExpiresAt.Equal(*decoded.ExpiresAt))
	require.NotNil(t, decoded.FileSize)
	assert.Equal(t, *original.FileSize, *decoded.FileSize)
	assert.Equal(t, original.Metadata["author"], decoded.Metadata["author"])
}

func TestReportJSONOmitsOptionalNils(t *testing.T) {
	t.Parallel()

	rpt := Report{
		ID:        "rpt-002",
		Name:      "Minimal Report",
		Status:    "pending",
		Format:    "csv",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(rpt)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, `"description"`)
	assert.NotContains(t, raw, `"templateId"`)
	assert.NotContains(t, raw, `"dataSourceId"`)
	assert.NotContains(t, raw, `"parameters"`)
	assert.NotContains(t, raw, `"generatedAt"`)
	assert.NotContains(t, raw, `"expiresAt"`)
	assert.NotContains(t, raw, `"fileSize"`)
	assert.NotContains(t, raw, `"metadata"`)
}

// ---------------------------------------------------------------------------
// Template — JSON round-trip
// ---------------------------------------------------------------------------

func TestTemplateJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)
	desc := "Standard monthly report layout"

	original := Template{
		ID:          "tmpl-abc",
		Name:        "Monthly Template",
		Description: &desc,
		Format:      "pdf",
		FileType:    "jrxml",
		Version:     3,
		Metadata: map[string]any{
			"orientation": "landscape",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Template

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, original.Format, decoded.Format)
	assert.Equal(t, original.FileType, decoded.FileType)
	assert.Equal(t, original.Version, decoded.Version)
	assert.Equal(t, original.Metadata["orientation"], decoded.Metadata["orientation"])
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt))
}

func TestTemplateJSONOmitsOptionalNils(t *testing.T) {
	t.Parallel()

	tmpl := Template{
		ID:        "tmpl-min",
		Name:      "Bare Template",
		Format:    "csv",
		FileType:  "csv",
		Version:   1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(tmpl)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, `"description"`)
	assert.NotContains(t, raw, `"metadata"`)
}

// ---------------------------------------------------------------------------
// CreateReportInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestCreateReportInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	desc := "Quarterly revenue breakdown"
	tmplID := "tmpl-xyz"
	dsID := "ds-003"

	original := CreateReportInput{
		Name:         "Q1 Revenue",
		Description:  &desc,
		Format:       "xlsx",
		TemplateID:   &tmplID,
		DataSourceID: &dsID,
		Parameters: map[string]any{
			"quarter": "Q1",
			"year":    float64(2026),
		},
		Metadata: map[string]any{
			"requestedBy": "finance-team",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CreateReportInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Name, decoded.Name)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, original.Format, decoded.Format)
	require.NotNil(t, decoded.TemplateID)
	assert.Equal(t, *original.TemplateID, *decoded.TemplateID)
	require.NotNil(t, decoded.DataSourceID)
	assert.Equal(t, *original.DataSourceID, *decoded.DataSourceID)
	assert.Equal(t, original.Parameters["quarter"], decoded.Parameters["quarter"])
	assert.Equal(t, original.Parameters["year"], decoded.Parameters["year"])
	assert.Equal(t, original.Metadata["requestedBy"], decoded.Metadata["requestedBy"])
}

func TestCreateReportInputMinimal(t *testing.T) {
	t.Parallel()

	input := CreateReportInput{
		Name:   "Simple Report",
		Format: "csv",
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	raw := string(data)
	assert.Contains(t, raw, `"name":"Simple Report"`)
	assert.Contains(t, raw, `"format":"csv"`)
	assert.NotContains(t, raw, `"description"`)
	assert.NotContains(t, raw, `"templateId"`)
	assert.NotContains(t, raw, `"dataSourceId"`)
	assert.NotContains(t, raw, `"parameters"`)
	assert.NotContains(t, raw, `"metadata"`)
}

// ---------------------------------------------------------------------------
// UpdateReportInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestUpdateReportInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	name := "Updated Report Name"
	desc := "Updated description"

	original := UpdateReportInput{
		Name:        &name,
		Description: &desc,
		Parameters: map[string]any{
			"newParam": "value",
		},
		Metadata: map[string]any{
			"updatedBy": "admin",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded UpdateReportInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	require.NotNil(t, decoded.Name)
	assert.Equal(t, *original.Name, *decoded.Name)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, original.Parameters["newParam"], decoded.Parameters["newParam"])
	assert.Equal(t, original.Metadata["updatedBy"], decoded.Metadata["updatedBy"])
}

func TestUpdateReportInputEmpty(t *testing.T) {
	t.Parallel()

	input := UpdateReportInput{}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	// All fields are omitempty, so an empty input should produce minimal JSON.
	raw := string(data)
	assert.NotContains(t, raw, `"name"`)
	assert.NotContains(t, raw, `"description"`)
	assert.NotContains(t, raw, `"parameters"`)
	assert.NotContains(t, raw, `"metadata"`)
}

// ---------------------------------------------------------------------------
// CreateTemplateInput — JSON round-trip
// ---------------------------------------------------------------------------

func TestCreateTemplateInputJSONRoundTrip(t *testing.T) {
	t.Parallel()

	desc := "Standard layout for invoices"

	original := CreateTemplateInput{
		Name:        "Invoice Template",
		Description: &desc,
		Format:      "pdf",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CreateTemplateInput

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Name, decoded.Name)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, *original.Description, *decoded.Description)
	assert.Equal(t, original.Format, decoded.Format)
}
