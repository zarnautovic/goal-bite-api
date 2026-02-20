package handlers

import (
	"encoding/json"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type AppError struct {
	Status  int
	Code    string
	Message string
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

type APIError struct {
	// Stable machine-readable error code.
	Code string `json:"code" example:"invalid_food_payload"`
	// Human-readable error message.
	Message string `json:"message" example:"invalid food payload"`
	// Request ID for log correlation.
	RequestID string `json:"request_id,omitempty" example:"f328f2f67798/SrQv8xTxwA-000001"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func WriteErrorResponse(w http.ResponseWriter, appErr AppError) {
	requestID := w.Header().Get(chimw.RequestIDHeader)
	writeJSON(w, appErr.Status, ErrorEnvelope{
		Error: APIError{
			Code:      appErr.Code,
			Message:   appErr.Message,
			RequestID: requestID,
		},
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	WriteErrorResponse(w, AppError{
		Status:  status,
		Code:    code,
		Message: message,
	})
}
