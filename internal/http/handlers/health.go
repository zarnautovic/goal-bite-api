package handlers

import "net/http"

// HealthLive godoc
// @Summary Liveness check
// @Description Returns process liveness status
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health/live [get]
func (h *Handler) HealthLive(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "nutrition api is running"})
}

// HealthReady godoc
// @Summary Readiness check
// @Description Returns dependency readiness status
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} ErrorEnvelope
// @Router /health/ready [get]
func (h *Handler) HealthReady(w http.ResponseWriter, r *http.Request) {
	if err := h.readinessChecker.Ready(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "service_unavailable", "service unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "nutrition api is running"})
}

// Health godoc
// @Summary Health check
// @Description Backward-compatible alias for readiness
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} ErrorEnvelope
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.HealthReady(w, r)
}
