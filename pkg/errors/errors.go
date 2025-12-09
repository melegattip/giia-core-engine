package errors

import (
	"fmt"
	"net/http"
)

type CustomError struct {
	ErrorCode  string
	Message    string
	HTTPStatus int
	Cause      error
}

func (e *CustomError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *CustomError) Unwrap() error {
	return e.Cause
}

func NewBadRequest(message string) *CustomError {
	return &CustomError{
		ErrorCode:  "BAD_REQUEST",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewUnauthorized(message string) *CustomError {
	return &CustomError{
		ErrorCode:  "UNAUTHORIZED",
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func NewForbidden(message string) *CustomError {
	return &CustomError{
		ErrorCode:  "FORBIDDEN",
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

func NewNotFound(message string) *CustomError {
	return &CustomError{
		ErrorCode:  "NOT_FOUND",
		Message:    message,
		HTTPStatus: http.StatusNotFound,
	}
}

func NewInternalServerError(message string) *CustomError {
	return &CustomError{
		ErrorCode:  "INTERNAL_SERVER_ERROR",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func NewServiceUnavailable(message string) *CustomError {
	return &CustomError{
		ErrorCode:  "SERVICE_UNAVAILABLE",
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

func Wrap(err error, message string) *CustomError {
	if err == nil {
		return nil
	}

	if customErr, ok := err.(*CustomError); ok {
		return &CustomError{
			ErrorCode:  customErr.ErrorCode,
			Message:    message,
			HTTPStatus: customErr.HTTPStatus,
			Cause:      err,
		}
	}

	return &CustomError{
		ErrorCode:  "INTERNAL_SERVER_ERROR",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Cause:      err,
	}
}