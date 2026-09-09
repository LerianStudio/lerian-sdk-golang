// Package reporter provides the client for the Reporter analytics service.
//
// Reporter generates, schedules, and retrieves reports over financial data.
// It supports configurable data sources, reusable report templates, and
// on-demand or scheduled report generation with multiple output formats.
//
// # Usage
//
// Access Reporter services through the umbrella client:
//
//	client, _ := lerian.New(lerian.Config{
//	    Reporter: &reporter.Config{
//	        BaseURL:        "http://localhost:3004/v1",
//	        OrganizationID: "org-uuid",
//	    },
//	})
//
//	report, err := client.Reporter.Reports.Create(ctx, &reporter.CreateReportInput{
//	    Name: "Monthly Summary",
//	})
//
// # Available Services
//
//   - Reports -- report creation, retrieval and download
//   - Templates -- template upload, retrieval and removal
//   - DataSources -- read access to the configured data sources
//
// # Coverage
//
// This package covers part of the Reporter API. Reporter's REST API is the
// complete surface: a capability absent here is not a capability the product
// lacks. Reach for the REST API directly for anything this package does not
// carry.
//
// Areas Reporter serves that this package does not reach include report
// deadlines and their notifications, service metrics, the template builder
// (block configuration, filters, code generation, preview and validation),
// template updates and template content retrieval, the streaming event
// manifest, and data-source creation, update, removal and diagnostics.
package reporter
