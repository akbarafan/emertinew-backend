package utils

import (
	"errors"
	"net/http"
)

// Common error types
var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInternalServer   = errors.New("internal server error")
	ErrBadRequest       = errors.New("bad request")
	ErrNotCreatedYet    = errors.New("not created yet")
	ErrValidationFailed = errors.New("validation failed")
	ErrRecordNotFound   = errors.New("record not found")
	ErrDuplicateRecord  = errors.New("duplicate record")
)

// ErrorCode represents custom error codes
type ErrorCode struct {
	Code    string
	Message string
	Status  int
}

// Common error codes
var (
	CodeNotFound = ErrorCode{
		Code:    "NOT_FOUND",
		Message: "Resource not found",
		Status:  http.StatusNotFound,
	}
	CodeAlreadyExists = ErrorCode{
		Code:    "ALREADY_EXISTS",
		Message: "Resource already exists",
		Status:  http.StatusConflict,
	}
	CodeInvalidInput = ErrorCode{
		Code:    "INVALID_INPUT",
		Message: "Invalid input provided",
		Status:  http.StatusBadRequest,
	}
	CodeUnauthorized = ErrorCode{
		Code:    "UNAUTHORIZED",
		Message: "Authentication required",
		Status:  http.StatusUnauthorized,
	}
	CodeForbidden = ErrorCode{
		Code:    "FORBIDDEN",
		Message: "Access forbidden",
		Status:  http.StatusForbidden,
	}
	CodeInternalServer = ErrorCode{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "Internal server error",
		Status:  http.StatusInternalServerError,
	}
	CodeNotCreatedYet = ErrorCode{
		Code:    "NOT_CREATED_YET",
		Message: "Resource not created yet",
		Status:  http.StatusNotFound,
	}
	CodeValidationFailed = ErrorCode{
		Code:    "VALIDATION_FAILED",
		Message: "Validation failed",
		Status:  http.StatusBadRequest,
	}
)

// NewError creates a new error with message
func NewError(message string) error {
	return errors.New(message)
}

// NewNotFoundError creates a not found error with custom message
func NewNotFoundError(resource string) error {
	return errors.New(resource + " not found")
}

// NewAlreadyExistsError creates an already exists error with custom message
func NewAlreadyExistsError(resource string) error {
	return errors.New(resource + " already exists")
}

// NewNotCreatedYetError creates a not created yet error with custom message
func NewNotCreatedYetError(resource string) error {
	return errors.New(resource + " not created yet")
}

// NewValidationError creates a validation error with custom message
func NewValidationError(message string) error {
	return errors.New("validation failed: " + message)
}
