package lerian

import sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"

// Error is the SDK's structured error type. Use errors.Is() and errors.As()
// to inspect SDK errors.
type Error = sdkerrors.Error

// ErrorCategory represents the class of error.
type ErrorCategory = sdkerrors.ErrorCategory

// ErrorCode represents a product-specific error code.
type ErrorCode = sdkerrors.ErrorCode

// Sentinel errors for use with errors.Is().
// These match errors by category, regardless of the specific product or message.
//
//	if errors.Is(err, lerian.ErrNotFound) {
//	    // handle not-found for any product
//	}
var (
	ErrValidation     = sdkerrors.ErrValidation
	ErrNotFound       = sdkerrors.ErrNotFound
	ErrAuthentication = sdkerrors.ErrAuthentication
	ErrAuthorization  = sdkerrors.ErrAuthorization
	ErrConflict       = sdkerrors.ErrConflict
	ErrRateLimit      = sdkerrors.ErrRateLimit
	ErrNetwork        = sdkerrors.ErrNetwork
	ErrTimeout        = sdkerrors.ErrTimeout
	ErrCancellation   = sdkerrors.ErrCancellation
	ErrInternal       = sdkerrors.ErrInternal
)

// Re-export error categories for convenience.
const (
	CategoryValidation     = sdkerrors.CategoryValidation
	CategoryNotFound       = sdkerrors.CategoryNotFound
	CategoryAuthentication = sdkerrors.CategoryAuthentication
	CategoryAuthorization  = sdkerrors.CategoryAuthorization
	CategoryConflict       = sdkerrors.CategoryConflict
	CategoryRateLimit      = sdkerrors.CategoryRateLimit
	CategoryNetwork        = sdkerrors.CategoryNetwork
	CategoryTimeout        = sdkerrors.CategoryTimeout
	CategoryCancellation   = sdkerrors.CategoryCancellation
	CategoryInternal       = sdkerrors.CategoryInternal
)
