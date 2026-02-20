package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"goal-bite-api/internal/http/dto"
	"goal-bite-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// GetUserByID godoc
// @Summary Get user by ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 403 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /users/{id} [get]
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	idPart := chi.URLParam(r, "id")
	if idPart == "" {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "invalid user id")
		return
	}

	id, err := strconv.ParseUint(idPart, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "invalid user id")
		return
	}
	if uint(id) != authUserID {
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	u, err := h.userService.GetByID(r.Context(), uint(id))
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrUserNotFound, http.StatusNotFound, "user_not_found", "user not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, u)
}

// Me godoc
// @Summary Get current authenticated user
// @Tags auth
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	u, err := h.userService.GetByID(r.Context(), authUserID)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrUserNotFound, http.StatusNotFound, "user_not_found", "user not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, u)
}

// UpdateMe godoc
// @Summary Update current authenticated user profile
// @Tags users
// @Accept json
// @Produce json
// @Param payload body dto.UpdateMeRequest true "Update me payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /users/me [patch]
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.UpdateMeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_payload", "invalid user payload")
		return
	}

	in, err := req.ToServiceInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_payload", "invalid user payload")
		return
	}

	u, err := h.userService.Update(r.Context(), authUserID, in)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrNoFieldsToUpdate, http.StatusBadRequest, "invalid_user_payload", "invalid user payload"),
		mapServiceError(service.ErrInvalidUserProfile, http.StatusBadRequest, "invalid_user_payload", "invalid user payload"),
		mapServiceError(service.ErrUserNotFound, http.StatusNotFound, "user_not_found", "user not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, u)
}
