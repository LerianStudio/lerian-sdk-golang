// Package tracer provides the client for the Tracer audit-trail and
// compliance service.
//
// Tracer records and queries audit events, manages compliance rules and
// validation limits, and supports verification workflows. It provides a
// tamper-evident log of all financial operations for regulatory compliance
// and internal auditing.
//
// # Usage
//
// Access Tracer services through the umbrella client:
//
//	client, _ := lerian.New(
//	    lerian.WithTracer(
//	        tracer.WithBaseURL("http://localhost:3003/v1"),
//	        tracer.WithAPIKey("my-api-key"),
//	    ),
//	)
//
//	events, err := client.Tracer.Audit.ListEvents(ctx, orgID, ledgerID, nil)
//
// # Available Services
//
//   - Audit -- query audit events and verification records
//   - Rules -- compliance rule definitions
//   - Limits -- transaction and operation limits
//   - Validations -- validation rule management
package tracer
