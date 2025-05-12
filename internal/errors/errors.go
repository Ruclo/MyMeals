package errors

import (
	"errors"
	"net/http"
)

type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

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

func NewNotFoundErr(message string, err error) *AppError {
	return new(err, message, http.StatusNotFound)
}
func IsNotFoundErr(err error) bool {
	return statusEquals(err, http.StatusNotFound)
}

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

func NewUnauthorizedErr(message string, err error) *AppError {
	return new(err, message, http.StatusUnauthorized)
}
func IsUnauthorizedErr(err error) bool {
	return statusEquals(err, http.StatusUnauthorized)
}

func NewForbiddenErr(message string, err error) *AppError {
	return new(err, message, http.StatusForbidden)
}
func IsForbiddenErr(err error) bool {
	return statusEquals(err, http.StatusForbidden)
}

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
