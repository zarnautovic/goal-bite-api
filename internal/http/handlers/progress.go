package handlers

import (
	"net/http"
	"time"

	"nutrition/internal/service"
)

// GetEnergyProgress godoc
// @Summary Get observed/formula energy progress
// @Tags progress
// @Produce json
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Success 200 {object} EnergyProgressResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /progress/energy [get]
func (h *Handler) GetEnergyProgress(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		now := time.Now().UTC()
		if to == "" {
			to = now.Format("2006-01-02")
		}
		if from == "" {
			from = now.AddDate(0, 0, -27).Format("2006-01-02")
		}
	}

	value, err := h.energyService.GetProgress(r.Context(), service.EnergyProgressInput{
		UserID: authUserID,
		From:   from,
		To:     to,
	})
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidEnergyProgressQuery, http.StatusBadRequest, "invalid_energy_progress_query", "invalid energy progress query"),
		mapServiceError(service.ErrInsufficientWeightData, http.StatusBadRequest, "insufficient_weight_data", "insufficient weight data"),
		mapServiceError(service.ErrInsufficientIntakeData, http.StatusBadRequest, "insufficient_intake_data", "insufficient intake data"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}
