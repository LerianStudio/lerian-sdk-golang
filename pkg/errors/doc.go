// Package errors provides the SDK error taxonomy and structured error types.
//
// Every error returned by the SDK is an [*Error] carrying an [ErrorCategory]
// that classifies the failure (validation, not-found, authentication, etc.)
// and an optional [ErrorCode] for product-specific sub-classification.
//
// # Category Matching with errors.Is
//
// Sentinel errors ([ErrValidation], [ErrNotFound], [ErrAuthentication], etc.)
// are provided so callers can use [errors.Is] for category-level matching:
//
//	if errors.Is(err, sdkerrors.ErrNotFound) {
//	    // handle missing resource
//	}
//
// # Rich Inspection with errors.As
//
// For richer inspection, use [errors.As] to extract the full [*Error]:
//
//	var sdkErr *sdkerrors.Error
//	if errors.As(err, &sdkErr) {
//	    log.Printf("category=%s operation=%s resource=%s",
//	        sdkErr.Category, sdkErr.Operation, sdkErr.Resource)
//	}
//
// # Factory Functions
//
// Factory functions ([NewValidation], [NewNotFound], [NewAuthentication],
// etc.) produce errors pre-filled with the correct category, HTTP status
// code, and product label. Product-specific service implementations use
// these to return consistently structured errors.
package errors
