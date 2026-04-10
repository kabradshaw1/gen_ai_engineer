package apperror

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestErrorHandler_AppError(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("requestId", "req-123")
		c.Next()
	})
	r.Use(ErrorHandler())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(NotFound("PRODUCT_NOT_FOUND", "product not found"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error.Code != "PRODUCT_NOT_FOUND" {
		t.Errorf("code = %q", resp.Error.Code)
	}
	if resp.Error.Message != "product not found" {
		t.Errorf("message = %q", resp.Error.Message)
	}
	if resp.Error.RequestID != "req-123" {
		t.Errorf("request_id = %q", resp.Error.RequestID)
	}
}

func TestErrorHandler_UnknownError(t *testing.T) {
	r := gin.New()
	r.Use(ErrorHandler())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(errors.New("db connection failed"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("code = %q", resp.Error.Code)
	}
	if resp.Error.Message != "an unexpected error occurred" {
		t.Errorf("message = %q, want hidden", resp.Error.Message)
	}
}

func TestErrorHandler_NoError(t *testing.T) {
	r := gin.New()
	r.Use(ErrorHandler())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestErrorHandler_MissingRequestID(t *testing.T) {
	r := gin.New()
	r.Use(ErrorHandler())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(BadRequest("VALIDATION_ERROR", "bad input"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error.RequestID != "" {
		t.Errorf("request_id = %q, want empty", resp.Error.RequestID)
	}
}
