package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"goal-bite-api/internal/http/dto"
	"goal-bite-api/internal/service"
)

// Register godoc
// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body dto.RegisterRequest true "Registration payload"
// @Success 201 {object} service.AuthResult
// @Failure 400 {object} ErrorEnvelope
// @Failure 409 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		if errors.Is(err, dto.ErrInvalidPassword) {
			writeError(w, http.StatusBadRequest, "invalid_password_policy", "password does not meet policy")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_register_payload", "invalid register payload")
		return
	}

	in, err := req.ToServiceInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_register_payload", "invalid register payload")
		return
	}

	result, err := h.authService.Register(r.Context(), in)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidName, http.StatusBadRequest, "invalid_register_payload", "invalid register payload"),
		mapServiceError(service.ErrInvalidEmail, http.StatusBadRequest, "invalid_register_payload", "invalid register payload"),
		mapServiceError(service.ErrInvalidPassword, http.StatusBadRequest, "invalid_password_policy", "password does not meet policy"),
		mapServiceError(service.ErrInvalidProfile, http.StatusBadRequest, "invalid_register_payload", "invalid register payload"),
		mapServiceError(service.ErrEmailAlreadyExists, http.StatusConflict, "email_already_exists", "email already exists"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body dto.LoginRequest true "Login payload"
// @Success 200 {object} service.AuthResult
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 429 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_login_payload", "invalid login payload")
		return
	}

	result, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrTooManyLoginAttempts, http.StatusTooManyRequests, "too_many_login_attempts", "too many login attempts"),
		mapServiceError(service.ErrInvalidCredentials, http.StatusUnauthorized, "invalid_credentials", "invalid credentials"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Refresh godoc
// @Summary Refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body dto.RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} service.AuthResult
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_refresh_payload", "invalid refresh payload")
		return
	}

	result, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidRefreshToken, http.StatusUnauthorized, "invalid_refresh_token", "invalid refresh token"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Logout godoc
// @Summary Logout session
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body dto.RefreshTokenRequest true "Refresh token payload"
// @Success 204
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_logout_payload", "invalid logout payload")
		return
	}

	err := h.authService.Logout(r.Context(), req.RefreshToken)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidRefreshToken, http.StatusUnauthorized, "invalid_refresh_token", "invalid refresh token"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
