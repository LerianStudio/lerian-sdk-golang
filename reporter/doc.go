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
//   - Reports -- report generation and retrieval
//   - Templates -- reusable report templates
//   - DataSources -- data source configuration for reports
package reporter
