package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewBadRequest(t *testing.T) {
	err := NewBadRequest("invalid input")

	if err.ErrorCode != CodeBadRequest {
		t.Errorf("expected error code %s, got %s", CodeBadRequest, err.ErrorCode)
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected HTTP status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}
	if err.Message != "invalid input" {
		t.Errorf("expected message 'invalid input', got '%s'", err.Message)
	}
}

func TestNewUnauthorized(t *testing.T) {
	err := NewUnauthorized("authentication required")

	if err.ErrorCode != CodeUnauthorized {
		t.Errorf("expected error code %s, got %s", CodeUnauthorized, err.ErrorCode)
	}
	if err.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("expected HTTP status %d, got %d", http.StatusUnauthorized, err.HTTPStatus)
	}
}

func TestNewForbidden(t *testing.T) {
	err := NewForbidden("insufficient permissions")

	if err.ErrorCode != CodeForbidden {
		t.Errorf("expected error code %s, got %s", CodeForbidden, err.ErrorCode)
	}
	if err.HTTPStatus != http.StatusForbidden {
		t.Errorf("expected HTTP status %d, got %d", http.StatusForbidden, err.HTTPStatus)
	}
}

func TestNewNotFound(t *testing.T) {
	err := NewNotFound("resource not found")

	if err.ErrorCode != CodeNotFound {
		t.Errorf("expected error code %s, got %s", CodeNotFound, err.ErrorCode)
	}
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}
}

func TestNewConflict(t *testing.T) {
	err := NewConflict("resource already exists")

	if err.ErrorCode != CodeConflict {
		t.Errorf("expected error code %s, got %s", CodeConflict, err.ErrorCode)
	}
	if err.HTTPStatus != http.StatusConflict {
		t.Errorf("expected HTTP status %d, got %d", http.StatusConflict, err.HTTPStatus)
	}
}

func TestNewTooManyRequests(t *testing.T) {
	err := NewTooManyRequests("rate limit exceeded")

	if err.ErrorCode != CodeTooManyRequests {
		t.Errorf("expected error code %s, got %s", CodeTooManyRequests, err.ErrorCode)
	}
	if err.HTTPStatus != http.StatusTooManyRequests {
		t.Errorf("expected HTTP status %d, got %d", http.StatusTooManyRequests, err.HTTPStatus)
	}
}

func TestNewInternalServerError(t *testing.T) {
	err := NewInternalServerError("database connection failed")

	if err.ErrorCode != CodeInternalServerError {
		t.Errorf("expected error code %s, got %s", CodeInternalServerError, err.ErrorCode)
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected HTTP status %d, got %d", http.StatusInternalServerError, err.HTTPStatus)
	}
}

func TestCustomError_ErrorAs(t *testing.T) {
	err := NewBadRequest("test error")

	var customErr *CustomError
	if !errors.As(err, &customErr) {
		t.Error("errors.As should work with CustomError")
	}

	if customErr.ErrorCode != CodeBadRequest {
		t.Errorf("expected error code %s, got %s", CodeBadRequest, customErr.ErrorCode)
	}
}

func TestCustomError_Error(t *testing.T) {
	err := NewBadRequest("validation failed")

	expected := "validation failed"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestCustomError_ErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := &CustomError{
		ErrorCode:  CodeInternalServerError,
		Message:    "operation failed",
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
	}

	expected := "operation failed: underlying error"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestCustomError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &CustomError{
		ErrorCode:  CodeInternalServerError,
		Message:    "operation failed",
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap should return the cause error")
	}
}

func TestWrap_WithCustomError(t *testing.T) {
	original := NewBadRequest("original error")
	wrapped := Wrap(original, "wrapped message")

	if wrapped.ErrorCode != CodeBadRequest {
		t.Errorf("expected error code %s, got %s", CodeBadRequest, wrapped.ErrorCode)
	}
	if wrapped.Message != "wrapped message" {
		t.Errorf("expected message 'wrapped message', got '%s'", wrapped.Message)
	}
	if wrapped.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected HTTP status %d, got %d", http.StatusBadRequest, wrapped.HTTPStatus)
	}
}

func TestWrap_WithStandardError(t *testing.T) {
	original := errors.New("standard error")
	wrapped := Wrap(original, "wrapped message")

	if wrapped.ErrorCode != CodeInternalServerError {
		t.Errorf("expected error code %s, got %s", CodeInternalServerError, wrapped.ErrorCode)
	}
	if wrapped.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected HTTP status %d, got %d", http.StatusInternalServerError, wrapped.HTTPStatus)
	}
}

func TestWrap_WithNil(t *testing.T) {
	wrapped := Wrap(nil, "test message")

	if wrapped != nil {
		t.Error("Wrap should return nil when given nil error")
	}
}

func TestToHTTPResponse_WithCustomError(t *testing.T) {
	err := NewBadRequest("validation failed")
	response := ToHTTPResponse(err)

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
	if response.ErrorCode != CodeBadRequest {
		t.Errorf("expected error code %s, got %s", CodeBadRequest, response.ErrorCode)
	}
	if response.Message != "validation failed" {
		t.Errorf("expected message 'validation failed', got '%s'", response.Message)
	}
}

func TestToHTTPResponse_WithStandardError(t *testing.T) {
	err := errors.New("standard error")
	response := ToHTTPResponse(err)

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, response.StatusCode)
	}
	if response.ErrorCode != CodeInternalServerError {
		t.Errorf("expected error code %s, got %s", CodeInternalServerError, response.ErrorCode)
	}
}
