package handlers

import (
	"encoding/json"
	"net/http"

	"nutrition/internal/http/dto"
	"nutrition/internal/service"
)

// UpsertUserGoals godoc
// @Summary Create or update user goals
// @Tags user-goals
// @Accept json
// @Produce json
// @Param payload body dto.UpsertUserGoalRequest true "User goals payload"
// @Success 200 {object} UserGoalResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /user-goals [put]
func (h *Handler) UpsertUserGoals(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.UpsertUserGoalRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_goals_payload", "invalid user goals payload")
		return
	}

	value, err := h.userGoalService.Upsert(r.Context(), req.ToServiceInput(authUserID))
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusBadRequest, "invalid_user_goals_payload", "invalid user goals payload"),
		mapServiceError(service.ErrInvalidTargetKcal, http.StatusBadRequest, "invalid_user_goals_payload", "invalid user goals payload"),
		mapServiceError(service.ErrInvalidTargetProteinG, http.StatusBadRequest, "invalid_user_goals_payload", "invalid user goals payload"),
		mapServiceError(service.ErrInvalidTargetCarbsG, http.StatusBadRequest, "invalid_user_goals_payload", "invalid user goals payload"),
		mapServiceError(service.ErrInvalidTargetFatG, http.StatusBadRequest, "invalid_user_goals_payload", "invalid user goals payload"),
		mapServiceError(service.ErrInvalidWeightGoalKG, http.StatusBadRequest, "invalid_user_goals_payload", "invalid user goals payload"),
		mapServiceError(service.ErrInvalidActivityLevel, http.StatusBadRequest, "invalid_user_goals_payload", "invalid user goals payload"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// GetUserGoals godoc
// @Summary Get current user goals
// @Tags user-goals
// @Produce json
// @Success 200 {object} UserGoalResponse
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /user-goals [get]
func (h *Handler) GetUserGoals(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	value, err := h.userGoalService.GetByUserID(r.Context(), authUserID)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrUserGoalNotFound, http.StatusNotFound, "user_goals_not_found", "user goals not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// GetDailyProgress godoc
// @Summary Get daily progress against user goals
// @Tags user-goals
// @Produce json
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} DailyProgressResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /progress/daily [get]
func (h *Handler) GetDailyProgress(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}
	date := r.URL.Query().Get("date")

	value, err := h.userGoalService.GetDailyProgress(r.Context(), authUserID, date)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidProgressDate, http.StatusBadRequest, "invalid_daily_progress_query", "invalid daily progress query"),
		mapServiceError(service.ErrUserGoalNotFound, http.StatusNotFound, "user_goals_not_found", "user goals not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}
