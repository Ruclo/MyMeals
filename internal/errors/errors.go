package errors

import (
	"errors"
	"net/http"
)

const (
	postgresUniqueViolation = "23505"
	postgresForeignKeyViol  = "23503"
	postgresNotNullViol     = "23502"
	postgresCheckViolation  = "23514"
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

func NewDuplicateErr(message string, err error) *AppError {
	return new(err, message, http.StatusConflict)
}

func NewForeignKeyErr(message string, err error) *AppError {
	return new(err, message, http.StatusBadRequest)
}

func NewValidationErr(message string, err error) *AppError {
	return new(err, message, http.StatusBadRequest)
}

func NewInternalServerErr(message string, err error) *AppError {
	return new(err, message, http.StatusInternalServerError)
}

func NewUnauthorizedErr(message string, err error) *AppError {
	return new(err, message, http.StatusUnauthorized)
}
func IsUnauthorizedErr(err error) bool {
	return statusEquals(err, http.StatusUnauthorized)
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

/*func classifyError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewNotFoundErr("Resource not found", err)
	}

	fmt.Printf("%T", err)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case postgresUniqueViolation:
			return NewDuplicateErr(fmt.Sprintf("%s already exists", pgErr.ColumnName), err)
		case postgresForeignKeyViol:
			return NewForeignKeyErr(fmt.Sprintf("Referenced %s does not exist", pgErr.ColumnName), err)
		case postgresNotNullViol:
			return NewValidationErr(fmt.Sprintf("Required field missing: %s", pgErr.ColumnName), err)
		case postgresCheckViolation:
			return NewValidationErr(fmt.Sprintf("Value violates check constraint: %s", pgErr.Detail), err)
		}
	}

	var appError *AppError
	if errors.As(err, &appError) {
		return err
	}

	log.Println("Internal server error occured:", err.Error())

	return NewInternalServerErr("Internal Server Error", err)
}*/
