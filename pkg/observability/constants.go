package observability

// ---------------------------------------------------------------------------
// Attribute keys
// ---------------------------------------------------------------------------

// Attribute key constants for OTel spans and metrics.
//
// All keys use the "lerian." namespace to ensure a consistent, product-wide
// naming convention across all SDKs and instrumentation.
const (
	// KeySDKVersion records the SDK version producing the telemetry.
	KeySDKVersion = "lerian.sdk.version"

	// KeySDKLanguage records the programming language of the SDK (e.g. "go").
	KeySDKLanguage = "lerian.sdk.language"

	// KeyProduct identifies the Lerian product being called (e.g. "ledger").
	KeyProduct = "lerian.product"

	// KeyOperationName records the logical operation name (e.g. "CreateAccount").
	KeyOperationName = "lerian.operation.name"

	// KeyOperationType classifies the operation kind (e.g. "create", "read").
	KeyOperationType = "lerian.operation.type"

	// KeyResourceType records the API resource type (e.g. "Account", "Ledger").
	KeyResourceType = "lerian.resource.type"

	// KeyResourceID records the identifier of the specific resource involved.
	KeyResourceID = "lerian.resource.id"
)

// ---------------------------------------------------------------------------
// HTTP attribute keys
// ---------------------------------------------------------------------------

// HTTP span and metric attribute keys used by [core.BackendImpl] to annotate
// outbound HTTP request telemetry with standard semantics.
const (
	// AttrHTTPMethod records the HTTP method (GET, POST, PATCH, DELETE, …).
	AttrHTTPMethod = "http.method"

	// AttrHTTPURL records the full request URL.
	AttrHTTPURL = "http.url"

	// AttrHTTPStatusCode records the HTTP response status code.
	AttrHTTPStatusCode = "http.status_code"

	// AttrHTTPRetries records the number of retries performed for a request.
	AttrHTTPRetries = "http.retries"
)

// ---------------------------------------------------------------------------
// Metric names
// ---------------------------------------------------------------------------

// Metric name constants for SDK-level request instrumentation.
//
// All metric names use the "lerian.sdk." prefix so they are easily
// discoverable and will not collide with application-level metrics.
const (
	// MetricRequestTotal counts the total number of outbound SDK requests.
	MetricRequestTotal = "lerian.sdk.request.total"

	// MetricRequestDuration records the duration of outbound SDK requests
	// (in milliseconds, as a histogram).
	MetricRequestDuration = "lerian.sdk.request.duration"

	// MetricRequestErrorTotal counts the total number of failed outbound
	// SDK requests (non-2xx or transport-level errors).
	MetricRequestErrorTotal = "lerian.sdk.request.error.total"
)

// ---------------------------------------------------------------------------
// HTTP metric names
// ---------------------------------------------------------------------------

// HTTP-specific metric names used by [core.BackendImpl] to record per-request
// telemetry at the transport layer.
const (
	// MetricHTTPRequestDuration is the histogram for HTTP request duration
	// (in milliseconds). Recorded for every outbound HTTP round-trip.
	MetricHTTPRequestDuration = "lerian.sdk.http.request.duration"

	// MetricHTTPRequestTotal is the counter for total HTTP requests.
	// Incremented once per outbound HTTP round-trip (including retries).
	MetricHTTPRequestTotal = "lerian.sdk.http.request.total"
)
