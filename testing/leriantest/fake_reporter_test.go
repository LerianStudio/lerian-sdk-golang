package leriantest_test

import (
	"context"
	"strings"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/reporter"
	"github.com/LerianStudio/lerian-sdk-golang/testing/leriantest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 1. DataSources -- read-only (Get + List)
// ---------------------------------------------------------------------------

func TestFakeReporterDataSourcesReadOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Stores start empty -- List should return zero items.
	iter := client.Reporter.DataSources.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)

	// Get on empty store -- not found
	_, err = client.Reporter.DataSources.Get(ctx, "nonexistent-ds-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeReporterDataSourcesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	_, err := client.Reporter.DataSources.Get(ctx, "ghost-datasource")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// 2. Reports -- CRUD + Download
// ---------------------------------------------------------------------------

func TestFakeReporterReportsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Reporter.Reports.Create(ctx, &reporter.CreateReportInput{
		Name:   "Monthly Revenue",
		Format: "pdf",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Monthly Revenue", created.Name)
	assert.Equal(t, "pending", created.Status)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Reporter.Reports.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Monthly Revenue", got.Name)

	// Update
	updated, err := client.Reporter.Reports.Update(ctx, created.ID, &reporter.UpdateReportInput{})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Reporter.Reports.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Download
	data, err := client.Reporter.Reports.Download(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, []byte("fake-report-data"), data)

	// Delete
	err = client.Reporter.Reports.Delete(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Reporter.Reports.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeReporterReportsNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-report-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Reporter.Reports.Get(ctx, ghost); return err }},
		{"Update", func() error {
			_, err := client.Reporter.Reports.Update(ctx, ghost, &reporter.UpdateReportInput{})
			return err
		}},
		{"Delete", func() error { return client.Reporter.Reports.Delete(ctx, ghost) }},
		{"Download", func() error { _, err := client.Reporter.Reports.Download(ctx, ghost); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Templates -- Create (with io.Reader) + Get + List + Delete
// ---------------------------------------------------------------------------

func TestFakeReporterTemplatesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create with an io.Reader for the template file
	fileContent := strings.NewReader("<html><body>{{.Title}}</body></html>")
	created, err := client.Reporter.Templates.Create(ctx, &reporter.CreateTemplateInput{
		Name:   "Invoice Template",
		Format: "html",
	}, fileContent)
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Invoice Template", created.Name)
	assert.Equal(t, "html", created.Format)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Reporter.Templates.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Invoice Template", got.Name)
	assert.Equal(t, "html", got.Format)

	// Create a second template
	fileContent2 := strings.NewReader("col1,col2,col3\n{{range .Rows}}{{.Col1}},{{.Col2}},{{.Col3}}\n{{end}}")
	created2, err := client.Reporter.Templates.Create(ctx, &reporter.CreateTemplateInput{
		Name:   "CSV Export Template",
		Format: "csv",
	}, fileContent2)
	require.NoError(t, err)
	require.NotEmpty(t, created2.ID)
	assert.NotEqual(t, created.ID, created2.ID)

	// List -- should have 2 items
	iter := client.Reporter.Templates.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// Delete the first
	err = client.Reporter.Templates.Delete(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Reporter.Templates.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Second template should still exist
	_, err = client.Reporter.Templates.Get(ctx, created2.ID)
	require.NoError(t, err)
}

func TestFakeReporterTemplatesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-template-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Reporter.Templates.Get(ctx, ghost); return err }},
		{"Delete", func() error { return client.Reporter.Templates.Delete(ctx, ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

// ---------------------------------------------------------------------------
// Supplemental: List with multiple items
// ---------------------------------------------------------------------------

func TestFakeReporterListMultipleItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create 3 reports
	for i := 0; i < 3; i++ {
		_, err := client.Reporter.Reports.Create(ctx, &reporter.CreateReportInput{
			Name:   "Report",
			Format: "csv",
		})
		require.NoError(t, err)
	}

	reportIter := client.Reporter.Reports.List(ctx, nil)
	reports, err := reportIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, reports, 3)

	// Create 3 templates
	for i := 0; i < 3; i++ {
		_, err := client.Reporter.Templates.Create(ctx, &reporter.CreateTemplateInput{
			Name:   "Template",
			Format: "pdf",
		}, strings.NewReader("content"))
		require.NoError(t, err)
	}

	tmplIter := client.Reporter.Templates.List(ctx, nil)
	templates, err := tmplIter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, templates, 3)
}

// ---------------------------------------------------------------------------
// Supplemental: Error injection
// ---------------------------------------------------------------------------

func TestFakeReporterReportsErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("reporter.Reports.Create", injectedErr),
	)

	_, err := client.Reporter.Reports.Create(ctx, &reporter.CreateReportInput{
		Name:   "Should Fail",
		Format: "pdf",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)

	// Non-injected template creation still works
	_, err = client.Reporter.Templates.Create(ctx, &reporter.CreateTemplateInput{
		Name:   "Works Fine",
		Format: "html",
	}, strings.NewReader("template-data"))
	require.NoError(t, err)
}

func TestFakeReporterTemplatesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("reporter.Templates.Create", injectedErr),
	)

	_, err := client.Reporter.Templates.Create(ctx, &reporter.CreateTemplateInput{
		Name:   "Should Fail",
		Format: "pdf",
	}, strings.NewReader("template-data"))
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)

	// Non-injected report creation still works
	_, err = client.Reporter.Reports.Create(ctx, &reporter.CreateReportInput{
		Name:   "Works Fine",
		Format: "csv",
	})
	require.NoError(t, err)
}
