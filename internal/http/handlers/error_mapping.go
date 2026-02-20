package handlers

import (
	"errors"
	"net/http"
)

type ServiceErrorMapping struct {
	Err    error
	Status int
	Code   string
	Text   string
}

func mapServiceError(err error, status int, code, text string) ServiceErrorMapping {
	return ServiceErrorMapping{
		Err:    err,
		Status: status,
		Code:   code,
		Text:   text,
	}
}

func writeMappedServiceError(w http.ResponseWriter, err error, mappings ...ServiceErrorMapping) bool {
	for _, m := range mappings {
		if errors.Is(err, m.Err) {
			WriteErrorResponse(w, AppError{
				Status:  m.Status,
				Code:    m.Code,
				Message: m.Text,
			})
			return true
		}
	}
	return false
}

func writeDatabaseError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "database_error", "database error")
}
