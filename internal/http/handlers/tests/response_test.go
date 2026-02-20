package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nutrition/internal/http/handlers"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func TestWriteErrorResponseIncludesRequestID(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set(chimw.RequestIDHeader, "req-123")

	handlers.WriteErrorResponse(rec, handlers.AppError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_request_body",
		Message: "invalid request body",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var out handlers.ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Error.RequestID != "req-123" {
		t.Fatalf("expected request_id req-123, got %q", out.Error.RequestID)
	}
}
