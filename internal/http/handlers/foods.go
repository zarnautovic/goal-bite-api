package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"nutrition/internal/domain/food"
	"nutrition/internal/http/dto"
	"nutrition/internal/service"

	"github.com/go-chi/chi/v5"
)

// CreateFood godoc
// @Summary Create food
// @Tags foods
// @Accept json
// @Produce json
// @Param payload body dto.CreateFoodRequest true "Food payload"
// @Success 201 {object} FoodResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 409 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /foods [post]
func (h *Handler) CreateFood(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.CreateFoodRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_food_payload", "invalid food payload")
		return
	}

	value, err := h.foodService.Create(r.Context(), userID, req.ToServiceInput())
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusUnauthorized, "unauthorized", "unauthorized"),
		mapServiceError(service.ErrInvalidFoodName, http.StatusBadRequest, "invalid_food_payload", "invalid food payload"),
		mapServiceError(service.ErrInvalidFoodBarcode, http.StatusBadRequest, "invalid_food_payload", "invalid food payload"),
		mapServiceError(service.ErrFoodBarcodeExists, http.StatusConflict, "food_barcode_already_exists", "food barcode already exists"),
		mapServiceError(service.ErrInvalidNutritionData, http.StatusBadRequest, "invalid_food_payload", "invalid food payload"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusCreated, value)
}

// GetFoodByID godoc
// @Summary Get food by ID
// @Tags foods
// @Produce json
// @Param id path int true "Food ID"
// @Success 200 {object} FoodResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /foods/{id} [get]
func (h *Handler) GetFoodByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_food_id", "invalid food id")
		return
	}

	value, err := h.foodService.GetByID(r.Context(), id)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrFoodNotFound, http.StatusNotFound, "food_not_found", "food not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// GetFoodByBarcode godoc
// @Summary Get food by barcode
// @Tags foods
// @Produce json
// @Param barcode path string true "Food barcode"
// @Success 200 {object} FoodResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /foods/by-barcode/{barcode} [get]
func (h *Handler) GetFoodByBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := strings.TrimSpace(chi.URLParam(r, "barcode"))
	if barcode == "" {
		writeError(w, http.StatusBadRequest, "invalid_food_barcode", "invalid food barcode")
		return
	}

	value, err := h.foodService.GetByBarcode(r.Context(), barcode)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidFoodBarcode, http.StatusBadRequest, "invalid_food_barcode", "invalid food barcode"),
		mapServiceError(service.ErrFoodBarcodeNotFound, http.StatusNotFound, "food_barcode_not_found", "food barcode not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// ListFoods godoc
// @Summary List foods
// @Tags foods
// @Produce json
// @Param q query string false "Search by food name (case-insensitive, partial match)"
// @Param limit query int false "Page size (default 20, max 100)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {array} FoodResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /foods [get]
func (h *Handler) ListFoods(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := parsePagination(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_pagination", "invalid pagination")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	var (
		values []food.Food
		err    error
	)
	if query == "" {
		values, err = h.foodService.List(r.Context(), limit, offset)
	} else {
		values, err = h.foodService.Search(r.Context(), query, limit, offset)
	}
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidPagination, http.StatusBadRequest, "invalid_pagination", "invalid pagination"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, values)
}

// UpdateFood godoc
// @Summary Update food
// @Tags foods
// @Accept json
// @Produce json
// @Param id path int true "Food ID"
// @Param payload body dto.UpdateFoodRequest true "Food update payload"
// @Success 200 {object} FoodResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 409 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /foods/{id} [patch]
func (h *Handler) UpdateFood(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	id, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_food_id", "invalid food id")
		return
	}

	var req dto.UpdateFoodRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_food_payload", "invalid food payload")
		return
	}

	value, err := h.foodService.Update(r.Context(), userID, id, req.ToServiceInput())
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusUnauthorized, "unauthorized", "unauthorized"),
		mapServiceError(service.ErrFoodNotFound, http.StatusNotFound, "food_not_found", "food not found"),
		mapServiceError(service.ErrFoodForbidden, http.StatusForbidden, "forbidden", "forbidden"),
		mapServiceError(service.ErrInvalidFoodBarcode, http.StatusBadRequest, "invalid_food_payload", "invalid food payload"),
		mapServiceError(service.ErrFoodBarcodeExists, http.StatusConflict, "food_barcode_already_exists", "food barcode already exists"),
		mapServiceError(service.ErrInvalidFoodName, http.StatusBadRequest, "invalid_food_payload", "invalid food payload"),
		mapServiceError(service.ErrInvalidNutritionData, http.StatusBadRequest, "invalid_food_payload", "invalid food payload"),
		mapServiceError(service.ErrNoFieldsToUpdate, http.StatusBadRequest, "invalid_food_payload", "invalid food payload"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// DeleteFood godoc
// @Summary Delete food
// @Tags foods
// @Produce json
// @Param id path int true "Food ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 403 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /foods/{id} [delete]
func (h *Handler) DeleteFood(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	id, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_food_id", "invalid food id")
		return
	}

	err := h.foodService.Delete(r.Context(), userID, id)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusUnauthorized, "unauthorized", "unauthorized"),
		mapServiceError(service.ErrFoodNotFound, http.StatusNotFound, "food_not_found", "food not found"),
		mapServiceError(service.ErrFoodForbidden, http.StatusForbidden, "forbidden", "forbidden"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIDFromPath(idPart string) (uint, bool) {
	idPart = strings.TrimSpace(idPart)
	if idPart == "" {
		return 0, false
	}

	id, err := strconv.ParseUint(idPart, 10, 64)
	if err != nil {
		return 0, false
	}

	return uint(id), true
}
