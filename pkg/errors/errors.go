package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// ErrorCategory classifies the broad kind of failure.
type ErrorCategory string

const (
	// CategoryValidation indicates the request failed local validation
	// (bad input, missing required field, constraint violation).
	CategoryValidation ErrorCategory = "validation"

	// CategoryNotFound indicates the requested resource does not exist.
	CategoryNotFound ErrorCategory = "not_found"

	// CategoryAuthentication indicates missing or invalid credentials.
	CategoryAuthentication ErrorCategory = "authentication"

	// CategoryAuthorization indicates the caller lacks permission.
	CategoryAuthorization ErrorCategory = "authorization"

	// CategoryConflict indicates a state conflict (duplicate, version mismatch).
	CategoryConflict ErrorCategory = "conflict"

	// CategoryRateLimit indicates the caller exceeded API rate limits.
	CategoryRateLimit ErrorCategory = "rate_limit"

	// CategoryNetwork indicates a transport-level failure (DNS, TCP, TLS).
	CategoryNetwork ErrorCategory = "network"

	// CategoryTimeout indicates the operation exceeded its deadline.
	CategoryTimeout ErrorCategory = "timeout"

	// CategoryCancellation indicates the operation was cancelled by the caller.
	CategoryCancellation ErrorCategory = "cancellation"

	// CategoryInternal indicates an unexpected internal error.
	CategoryInternal ErrorCategory = "internal"
)

// ErrorCode carries a product-specific sub-code within a category.
// For example, a validation category might have codes like "invalid_email"
// or "missing_field".
type ErrorCode string

// ---------------------------------------------------------------------------
// Error struct
// ---------------------------------------------------------------------------

// Error is the canonical SDK error type. It carries structured metadata
// that enables programmatic inspection, logging, and user-facing messages.
type Error struct {
	// Product identifies which Lerian product originated the error
	// (e.g. "ledger", "transaction", "sdk").
	Product string `json:"product"`

	// Category is the broad classification of the error.
	Category ErrorCategory `json:"category"`

	// Code is an optional product-specific sub-code within the category.
	Code ErrorCode `json:"code,omitempty"`

	// Message is a human-readable description of what went wrong.
	Message string `json:"message"`

	// Operation is the SDK operation that was being performed
	// (e.g. "accounts.Create", "transactions.List").
	Operation string `json:"operation"`

	// Resource is the type of resource involved (e.g. "Account", "Ledger").
	Resource string `json:"resource,omitempty"`

	// ResourceID is the identifier of the specific resource, when known.
	ResourceID string `json:"resourceId,omitempty"`

	// StatusCode is the HTTP status code from the API response, if applicable.
	StatusCode int `json:"statusCode,omitempty"`

	// RequestID is the server-assigned request identifier for correlation.
	RequestID string `json:"requestId,omitempty"`

	// HelpURL points to documentation about this error, if available.
	HelpURL string `json:"helpUrl,omitempty"`

	// Err is the underlying error that caused this one, if any.
	// It is excluded from JSON serialization but participates in
	// the standard Unwrap chain.
	Err error `json:"-"`
}

// ---------------------------------------------------------------------------
// error interface + Is / Unwrap
// ---------------------------------------------------------------------------

// Error returns a formatted string representation of the error.
//
// With a RequestID:
//
//	[PRODUCT] category during Operation: message (request_id: xxx)
//
// Without a RequestID:
//
//	[PRODUCT] category during Operation: message
func (e *Error) Error() string {
	var b strings.Builder

	product := e.Product
	if product != "" {
		product = strings.ToUpper(product)
	}

	// [PRODUCT] category during Operation: message
	fmt.Fprintf(&b, "[%s] %s during %s: %s", product, string(e.Category), e.Operation, e.Message)

	if e.RequestID != "" {
		fmt.Fprintf(&b, " (request_id: %s)", e.RequestID)
	}

	return b.String()
}

// Is reports whether this error matches target for [errors.Is] purposes.
// Two SDK errors match when they share the same Category. If the target
// also specifies a Code, the codes must match as well.
func (e *Error) Is(target error) bool {
	var t *Error
	if !stderrors.As(target, &t) {
		return false
	}

	if e.Category != t.Category {
		return false
	}

	if t.Code != "" && e.Code != t.Code {
		return false
	}

	return true
}

// Unwrap returns the underlying error, enabling [errors.Unwrap] chains.
func (e *Error) Unwrap() error {
	return e.Err
}

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

// Sentinel errors for category-level matching with [errors.Is].
//
//	if errors.Is(err, errors.ErrNotFound) { ... }
var (
	ErrValidation     = &Error{Category: CategoryValidation}
	ErrNotFound       = &Error{Category: CategoryNotFound}
	ErrAuthentication = &Error{Category: CategoryAuthentication}
	ErrAuthorization  = &Error{Category: CategoryAuthorization}
	ErrConflict       = &Error{Category: CategoryConflict}
	ErrRateLimit      = &Error{Category: CategoryRateLimit}
	ErrNetwork        = &Error{Category: CategoryNetwork}
	ErrTimeout        = &Error{Category: CategoryTimeout}
	ErrCancellation   = &Error{Category: CategoryCancellation}
	ErrInternal       = &Error{Category: CategoryInternal}
)

// ---------------------------------------------------------------------------
// Factory functions
// ---------------------------------------------------------------------------

// NewValidation creates a validation error for SDK-side input failures.
// Product is always "sdk" because validation happens locally.
func NewValidation(operation, resource, message string) *Error {
	return &Error{
		Product:   "sdk",
		Category:  CategoryValidation,
		Operation: operation,
		Resource:  resource,
		Message:   message,
	}
}

// NewNotFound creates a not-found error with a 404 status code.
// Product is "sdk" and the message is derived from the resource type and ID.
func NewNotFound(operation, resource, id string) *Error {
	return &Error{
		Product:    "sdk",
		Category:   CategoryNotFound,
		Operation:  operation,
		Resource:   resource,
		ResourceID: id,
		StatusCode: 404,
		Message:    resource + " not found: " + id,
	}
}

// NewAuthentication creates an authentication error with a 401 status code.
func NewAuthentication(product, operation, message string) *Error {
	return &Error{
		Product:    product,
		Category:   CategoryAuthentication,
		Operation:  operation,
		StatusCode: 401,
		Message:    message,
	}
}

// NewAuthorization creates an authorization error with a 403 status code.
func NewAuthorization(product, operation, message string) *Error {
	return &Error{
		Product:    product,
		Category:   CategoryAuthorization,
		Operation:  operation,
		StatusCode: 403,
		Message:    message,
	}
}

// NewConflict creates a conflict error with a 409 status code.
func NewConflict(product, operation, resource, message string) *Error {
	return &Error{
		Product:    product,
		Category:   CategoryConflict,
		Operation:  operation,
		Resource:   resource,
		StatusCode: 409,
		Message:    message,
	}
}

// NewNetwork creates a network error wrapping the underlying transport failure.
func NewNetwork(product, operation string, err error) *Error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	return &Error{
		Product:   product,
		Category:  CategoryNetwork,
		Operation: operation,
		Message:   msg,
		Err:       err,
	}
}

// NewTimeout creates a timeout error wrapping the underlying deadline/context error.
func NewTimeout(product, operation string, err error) *Error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	return &Error{
		Product:   product,
		Category:  CategoryTimeout,
		Operation: operation,
		Message:   msg,
		Err:       err,
	}
}

// NewCancellation creates a cancellation error wrapping the context cancellation.
func NewCancellation(product, operation string, err error) *Error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	return &Error{
		Product:   product,
		Category:  CategoryCancellation,
		Operation: operation,
		Message:   msg,
		Err:       err,
	}
}

// NewInternal creates an internal error with a 500 status code,
// wrapping an optional underlying cause.
func NewInternal(product, operation, message string, err error) *Error {
	return &Error{
		Product:    product,
		Category:   CategoryInternal,
		Operation:  operation,
		StatusCode: 500,
		Message:    message,
		Err:        err,
	}
}

// ---------------------------------------------------------------------------
// Body truncation
// ---------------------------------------------------------------------------

// MaxErrorBodyBytes is the maximum number of response body bytes included
// in error messages. Bodies exceeding this limit are truncated with a
// "... [truncated]" suffix.
const MaxErrorBodyBytes = 512

// TruncateBody converts a raw response body to a string, capping it at
// [MaxErrorBodyBytes]. This prevents enormous server responses from
// polluting logs and error messages.
func TruncateBody(body []byte) string {
	s := string(body)
	if len(s) > MaxErrorBodyBytes {
		return s[:MaxErrorBodyBytes] + "... [truncated]"
	}

	return s
}

// ---------------------------------------------------------------------------
// Status-to-category mapping
// ---------------------------------------------------------------------------

// CategoryFromStatus maps an HTTP status code to the appropriate SDK error category.
func CategoryFromStatus(status int) ErrorCategory {
	switch {
	case status == 400:
		return CategoryValidation
	case status == 401:
		return CategoryAuthentication
	case status == 403:
		return CategoryAuthorization
	case status == 404:
		return CategoryNotFound
	case status == 409:
		return CategoryConflict
	case status == 429:
		return CategoryRateLimit
	case status >= 500:
		return CategoryInternal
	default:
		return CategoryInternal
	}
}
