// Package matcher provides the client for the Matcher reconciliation service.
//
// Matcher automates the reconciliation of financial transactions across
// multiple data sources. It supports context management, rule-based
// matching, scheduling, source configuration, exception handling, dispute
// resolution, fee schedules, field mappings, data export/import, governance,
// and analytics.
//
// # Usage
//
// Access Matcher services through the umbrella client:
//
//	client, _ := lerian.New(
//	    lerian.WithMatcher(
//	        matcher.WithBaseURL("http://localhost:3002/v1"),
//	    ),
//	)
//	// Optional OAuth2 credentials can be loaded from the matching LERIAN_MATCHER_* env vars.
//
//	ctx, err := client.Matcher.Contexts.Create(ctx, &matcher.CreateContextInput{
//	    Name: "Monthly Bank Reconciliation",
//	})
//
// # Available Services
//
// The Matcher client exposes the following services:
//
//   - Contexts -- top-level reconciliation contexts
//   - Rules -- matching rule definitions ordered by priority
//   - Schedules -- cron-based recurring reconciliation triggers
//   - Sources -- data source connections (bank feeds, ERP, etc.)
//   - FieldMaps -- cross-source data alignment definitions
//   - SourceFieldMaps -- per-source field-to-schema mappings
//   - FeeSchedules -- fee calculation rules for matched records
//   - Exceptions -- anomalies requiring human review
//   - Disputes -- formal challenges against reconciliation results
//   - ExportJobs -- downloadable file generation (CSV, XLSX, etc.)
//   - Imports -- data import job management
//   - Matching -- RPC-style automated/manual matching operations
//   - Reports -- analytics and reporting endpoints
//   - Governance -- audit logs and archive operations
package matcher
