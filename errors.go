// Package vandargo provides a secure integration with the Vandar payment gateway
// errors.go contains custom error types and error handling utilities
package vandargo

import (
	"errors"
	"fmt"
)

// Common error types for package users to check against
var (
	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrInvalidRequest is returned when a request is invalid
	ErrInvalidRequest = errors.New("invalid request parameters")

	// ErrAuthentication is returned for authentication failures
	ErrAuthentication = errors.New("authentication failed")

	// ErrPermission is returned for permission issues
	ErrPermission = errors.New("permission denied")

	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrPaymentFailed is returned when a payment fails
	ErrPaymentFailed = errors.New("payment failed")

	// ErrVerificationFailed is returned when verification fails
	ErrVerificationFailed = errors.New("verification failed")

	// ErrRefundFailed is returned when a refund fails
	ErrRefundFailed = errors.New("refund failed")

	// ErrNetworkFailure is returned for network-related issues
	ErrNetworkFailure = errors.New("network error")

	// ErrTimeout is returned when a request times out
	ErrTimeout = errors.New("request timed out")

	// ErrInternalError is returned for unexpected internal errors
	ErrInternalError = errors.New("internal error")
)

// ValidationError represents an error that occurred during validation
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation errors"
	}

	return fmt.Sprintf("validation errors (%d errors)", len(e))
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewValidationErrors creates a new validation errors list
func NewValidationErrors(errors []ValidationError) error {
	return ValidationErrors(errors)
}

// IsDomainError checks if an error is one of the domain errors
func IsDomainError(err error) bool {
	return errors.Is(err, ErrInvalidConfig) ||
		errors.Is(err, ErrInvalidRequest) ||
		errors.Is(err, ErrAuthentication) ||
		errors.Is(err, ErrPermission) ||
		errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrPaymentFailed) ||
		errors.Is(err, ErrVerificationFailed) ||
		errors.Is(err, ErrRefundFailed)
}

// IsNetworkError checks if an error is network-related
func IsNetworkError(err error) bool {
	return errors.Is(err, ErrNetworkFailure) ||
		errors.Is(err, ErrTimeout)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	var validationErrs ValidationErrors

	return errors.As(err, &validationErr) || errors.As(err, &validationErrs)
}

// ExtractValidationErrors extracts validation errors from an error
func ExtractValidationErrors(err error) []ValidationError {
	var validationErr *ValidationError
	var validationErrs ValidationErrors

	if errors.As(err, &validationErr) {
		return []ValidationError{*validationErr}
	}

	if errors.As(err, &validationErrs) {
		return validationErrs
	}

	return nil
}

// APIErrorResponse converts an error to a safe API response
func APIErrorResponse(err error) map[string]interface{} {
	if err == nil {
		return map[string]interface{}{
			"status":  false,
			"message": "Unknown error",
		}
	}

	response := map[string]interface{}{
		"status": false,
	}

	// Handle API errors
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		response["message"] = apiErr.Message
		if apiErr.Code != "" {
			response["code"] = apiErr.Code
		}
		if len(apiErr.Errors) > 0 {
			response["errors"] = apiErr.Errors
		}
		return response
	}

	// Handle validation errors
	if validationErrs := ExtractValidationErrors(err); len(validationErrs) > 0 {
		errorsMap := make(map[string]string)
		for _, ve := range validationErrs {
			errorsMap[ve.Field] = ve.Message
		}
		response["message"] = "Validation failed"
		response["errors"] = errorsMap
		return response
	}

	// Handle standard errors with safe messages
	if IsDomainError(err) {
		response["message"] = err.Error()
	} else if IsNetworkError(err) {
		response["message"] = "A network error occurred. Please try again."
	} else {
		// For unexpected errors, don't expose details
		response["message"] = "An unexpected error occurred. Please try again."
	}

	return response
}
