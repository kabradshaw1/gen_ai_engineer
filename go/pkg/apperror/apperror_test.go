package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestConstructors(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(string, string) *AppError
		wantStatus int
	}{
		{"NotFound", NotFound, http.StatusNotFound},
		{"BadRequest", BadRequest, http.StatusBadRequest},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized},
		{"Forbidden", Forbidden, http.StatusForbidden},
		{"Conflict", Conflict, http.StatusConflict},
		{"Internal", Internal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ae := tt.fn("TEST_CODE", "test message")
			if ae.Code != "TEST_CODE" {
				t.Errorf("Code = %q, want TEST_CODE", ae.Code)
			}
			if ae.Message != "test message" {
				t.Errorf("Message = %q, want test message", ae.Message)
			}
			if ae.HTTPStatus != tt.wantStatus {
				t.Errorf("HTTPStatus = %d, want %d", ae.HTTPStatus, tt.wantStatus)
			}
		})
	}
}

func TestError(t *testing.T) {
	ae := NotFound("X", "not here")
	if ae.Error() != "not here" {
		t.Errorf("Error() = %q", ae.Error())
	}
}

func TestUnwrap(t *testing.T) {
	cause := fmt.Errorf("db timeout")
	ae := Wrap(cause, "DB_ERROR", "database error", http.StatusInternalServerError)
	if ae.Unwrap() != cause {
		t.Errorf("Unwrap returned %v, want %v", ae.Unwrap(), cause)
	}
}

func TestUnwrap_Nil(t *testing.T) {
	ae := NotFound("X", "msg")
	if ae.Unwrap() != nil {
		t.Errorf("Unwrap returned %v, want nil", ae.Unwrap())
	}
}

func TestIs(t *testing.T) {
	ae := NotFound("NF", "not found")
	got, ok := Is(ae)
	if !ok {
		t.Fatal("Is returned false for *AppError")
	}
	if got.Code != "NF" {
		t.Errorf("Code = %q", got.Code)
	}
}

func TestIs_Wrapped(t *testing.T) {
	ae := NotFound("NF", "not found")
	wrapped := fmt.Errorf("handler: %w", ae)
	got, ok := Is(wrapped)
	if !ok {
		t.Fatal("Is returned false for wrapped *AppError")
	}
	if got.Code != "NF" {
		t.Errorf("Code = %q", got.Code)
	}
}

func TestIs_NotAppError(t *testing.T) {
	_, ok := Is(errors.New("plain error"))
	if ok {
		t.Error("Is returned true for plain error")
	}
}

func TestErrorsIs_WithAppError(t *testing.T) {
	sentinel := NotFound("NF", "not found")
	if !errors.Is(sentinel, sentinel) {
		t.Error("errors.Is failed for same *AppError")
	}
}

func TestErrorsAs_ExtractsAppError(t *testing.T) {
	ae := Conflict("DUP", "duplicate")
	wrapped := fmt.Errorf("repo: %w", ae)
	var target *AppError
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As failed")
	}
	if target.Code != "DUP" {
		t.Errorf("Code = %q", target.Code)
	}
}
