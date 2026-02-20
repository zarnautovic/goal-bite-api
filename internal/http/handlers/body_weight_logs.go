package handlers

import (
	"encoding/json"
	"net/http"

	"goal-bite-api/internal/http/dto"
	"goal-bite-api/internal/service"
)

// CreateBodyWeightLog godoc
// @Summary Create body weight log
// @Tags body-weight-logs
// @Accept json
// @Produce json
// @Param payload body dto.CreateBodyWeightLogRequest true "Body weight payload"
// @Success 201 {object} BodyWeightLogResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /body-weight-logs [post]
func (h *Handler) CreateBodyWeightLog(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.CreateBodyWeightLogRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body_weight_payload", "invalid body weight payload")
		return
	}

	in, err := req.ToServiceInput(authUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body_weight_payload", "invalid body weight payload")
		return
	}

	value, err := h.bodyWeightLogService.Create(r.Context(), in)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusBadRequest, "invalid_body_weight_payload", "invalid body weight payload"),
		mapServiceError(service.ErrInvalidWeightKG, http.StatusBadRequest, "invalid_body_weight_payload", "invalid body weight payload"),
		mapServiceError(service.ErrInvalidEatenAt, http.StatusBadRequest, "invalid_body_weight_payload", "invalid body weight payload"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusCreated, value)
}

// ListBodyWeightLogs godoc
// @Summary List body weight logs by date range
// @Tags body-weight-logs
// @Produce json
// @Param from query string true "From date (YYYY-MM-DD)"
// @Param to query string true "To date (YYYY-MM-DD)"
// @Param limit query int false "Page size (default 20, max 100)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {array} BodyWeightLogResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /body-weight-logs [get]
func (h *Handler) ListBodyWeightLogs(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	limit, offset, ok := parsePagination(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_body_weight_query", "invalid body weight query")
		return
	}

	query := dto.ListBodyWeightLogsQuery{From: from, To: to, Limit: limit, Offset: offset}
	if err := query.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body_weight_query", "invalid body weight query")
		return
	}

	values, err := h.bodyWeightLogService.List(r.Context(), query.ToServiceInput(authUserID))
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusBadRequest, "invalid_body_weight_query", "invalid body weight query"),
		mapServiceError(service.ErrInvalidDateRange, http.StatusBadRequest, "invalid_body_weight_query", "invalid body weight query"),
		mapServiceError(service.ErrInvalidPagination, http.StatusBadRequest, "invalid_body_weight_query", "invalid body weight query"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, values)
}

// GetLatestBodyWeightLog godoc
// @Summary Get latest body weight log
// @Tags body-weight-logs
// @Produce json
// @Success 200 {object} BodyWeightLogResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /body-weight-logs/latest [get]
func (h *Handler) GetLatestBodyWeightLog(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	value, err := h.bodyWeightLogService.GetLatest(r.Context(), authUserID)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusBadRequest, "invalid_body_weight_query", "invalid body weight query"),
		mapServiceError(service.ErrBodyWeightLogNotFound, http.StatusNotFound, "body_weight_log_not_found", "body weight log not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}
