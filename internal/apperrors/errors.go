package apperrors

import (
	"errors"
	"net/http"
)

// AppError is a wrapper for errors. Every error gets associated with a status code and a message. This status code and
// error message is then sent to the client in the response. Internal server error messages are not sent in the response.
// Every error should get wrapped in this.
type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error wrapped by the AppError instance.
func (e *AppError) Unwrap() error {
	return e.Err
}

func new(err error, message string, status int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: status,
	}
}

// NewNotFoundErr creates a new AppError with a status code of 404.
func NewNotFoundErr(message string, err error) *AppError {
	return new(err, message, http.StatusNotFound)
}
func IsNotFoundErr(err error) bool {
	return statusEquals(err, http.StatusNotFound)
}

// NewValidationErr creates a new AppError with a status code of 400.
// The wrapped validation error gets processed and the invalid fields get included in the response.
func NewValidationErr(message string, err error) *AppError {
	return new(err, message, http.StatusBadRequest)
}
func IsValidationErr(err error) bool {
	return statusEquals(err, http.StatusBadRequest)
}

func NewInternalServerErr(message string, err error) *AppError {
	return new(err, message, http.StatusInternalServerError)
}
func IsInternalServerErr(err error) bool {
	return statusEquals(err, http.StatusInternalServerError)
}

// NewUnauthorizedErr creates a new AppError with a status code of 401.
func NewUnauthorizedErr(message string, err error) *AppError {
	return new(err, message, http.StatusUnauthorized)
}
func IsUnauthorizedErr(err error) bool {
	return statusEquals(err, http.StatusUnauthorized)
}

// NewForbiddenErr creates a new AppError with a status code of 403.
func NewForbiddenErr(message string, err error) *AppError {
	return new(err, message, http.StatusForbidden)
}
func IsForbiddenErr(err error) bool {
	return statusEquals(err, http.StatusForbidden)
}

// NewAlreadyExistsErr creates a new AppError with a status code of 409.
func NewAlreadyExistsErr(message string, err error) *AppError {
	return new(err, message, http.StatusConflict)
}
func IsAlreadyExistsErr(err error) bool {
	return statusEquals(err, http.StatusConflict)
}

func statusEquals(err error, status int) bool {
	var appError *AppError
	if errors.As(err, &appError) {
		return appError.StatusCode == status
	}
	return false
}
