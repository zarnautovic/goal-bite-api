package handlers

import (
	"encoding/json"
	"net/http"

	"goal-bite-api/internal/http/dto"
	"goal-bite-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// CreateMeal godoc
// @Summary Create meal
// @Tags meals
// @Accept json
// @Produce json
// @Param payload body dto.CreateMealRequest true "Meal payload"
// @Success 201 {object} MealResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /meals [post]
func (h *Handler) CreateMeal(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.CreateMealRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload")
		return
	}

	in, err := req.ToServiceInput(authUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload")
		return
	}

	value, err := h.mealService.Create(r.Context(), in)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload"),
		mapServiceError(service.ErrInvalidMealType, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload"),
		mapServiceError(service.ErrInvalidEatenAt, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload"),
		mapServiceError(service.ErrInvalidItemSource, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload"),
		mapServiceError(service.ErrInvalidItemWeight, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload"),
		mapServiceError(service.ErrFoodNotFound, http.StatusBadRequest, "food_not_found", "food not found"),
		mapServiceError(service.ErrRecipeSourceNotFound, http.StatusBadRequest, "recipe_not_found", "recipe not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusCreated, toMealResponse(value))
}

// GetMealByID godoc
// @Summary Get meal by ID
// @Tags meals
// @Produce json
// @Param id path int true "Meal ID"
// @Success 200 {object} MealResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /meals/{id} [get]
func (h *Handler) GetMealByID(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	id, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_meal_id", "invalid meal id")
		return
	}

	value, err := h.mealService.GetByID(r.Context(), authUserID, id)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrMealNotFound, http.StatusNotFound, "meal_not_found", "meal not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, toMealResponse(value))
}

// ListMeals godoc
// @Summary List meals by user and date
// @Tags meals
// @Produce json
// @Param date query string true "Date (YYYY-MM-DD)"
// @Param limit query int false "Page size (default 20, max 100)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {array} MealResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /meals [get]
func (h *Handler) ListMeals(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	date := r.URL.Query().Get("date")
	limit, offset, ok := parsePagination(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_pagination", "invalid pagination")
		return
	}

	query := dto.ListMealsQuery{Date: date, Limit: limit, Offset: offset}
	if err := query.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_meal_query", "invalid meal query")
		return
	}

	values, err := h.mealService.List(r.Context(), query.ToServiceInput(authUserID))
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusBadRequest, "invalid_meal_query", "invalid meal query"),
		mapServiceError(service.ErrInvalidDate, http.StatusBadRequest, "invalid_meal_query", "invalid meal query"),
		mapServiceError(service.ErrInvalidPagination, http.StatusBadRequest, "invalid_meal_query", "invalid meal query"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, toMealResponses(values))
}

// AddMealItem godoc
// @Summary Add meal item
// @Tags meals
// @Accept json
// @Produce json
// @Param id path int true "Meal ID"
// @Param payload body dto.AddMealItemRequest true "Meal item payload"
// @Success 201 {object} MealItemResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /meals/{id}/items [post]
func (h *Handler) AddMealItem(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	mealID, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_meal_id", "invalid meal id")
		return
	}

	var req dto.AddMealItemRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_meal_item_payload", "invalid meal item payload")
		return
	}

	value, err := h.mealService.AddItem(r.Context(), authUserID, mealID, req.ToServiceInput())
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidItemSource, http.StatusBadRequest, "invalid_meal_item_payload", "invalid meal item payload"),
		mapServiceError(service.ErrInvalidItemWeight, http.StatusBadRequest, "invalid_meal_item_payload", "invalid meal item payload"),
		mapServiceError(service.ErrMealNotFound, http.StatusNotFound, "meal_not_found", "meal not found"),
		mapServiceError(service.ErrFoodNotFound, http.StatusBadRequest, "food_not_found", "food not found"),
		mapServiceError(service.ErrRecipeSourceNotFound, http.StatusBadRequest, "recipe_not_found", "recipe not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusCreated, value)
}

// UpdateMeal godoc
// @Summary Update meal
// @Tags meals
// @Accept json
// @Produce json
// @Param id path int true "Meal ID"
// @Param payload body dto.UpdateMealRequest true "Meal update payload"
// @Success 200 {object} MealResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /meals/{id} [patch]
func (h *Handler) UpdateMeal(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	mealID, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_meal_id", "invalid meal id")
		return
	}

	var req dto.UpdateMealRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload")
		return
	}

	in, err := req.ToServiceInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload")
		return
	}

	value, err := h.mealService.Update(r.Context(), authUserID, mealID, in)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrNoFieldsToUpdate, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload"),
		mapServiceError(service.ErrInvalidMealType, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload"),
		mapServiceError(service.ErrInvalidEatenAt, http.StatusBadRequest, "invalid_meal_payload", "invalid meal payload"),
		mapServiceError(service.ErrMealNotFound, http.StatusNotFound, "meal_not_found", "meal not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, toMealResponse(value))
}

// DeleteMeal godoc
// @Summary Delete meal
// @Tags meals
// @Param id path int true "Meal ID"
// @Success 204
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /meals/{id} [delete]
func (h *Handler) DeleteMeal(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	mealID, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_meal_id", "invalid meal id")
		return
	}

	err := h.mealService.Delete(r.Context(), authUserID, mealID)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrMealNotFound, http.StatusNotFound, "meal_not_found", "meal not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateMealItem godoc
// @Summary Update meal item
// @Tags meals
// @Accept json
// @Produce json
// @Param meal_id path int true "Meal ID"
// @Param item_id path int true "Meal Item ID"
// @Param payload body dto.UpdateMealItemRequest true "Meal item update payload"
// @Success 200 {object} MealItemResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /meals/{meal_id}/items/{item_id} [patch]
func (h *Handler) UpdateMealItem(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	mealID, ok := parseIDFromPath(chi.URLParam(r, "meal_id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_meal_id", "invalid meal id")
		return
	}
	itemID, ok := parseIDFromPath(chi.URLParam(r, "item_id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_meal_item_id", "invalid meal item id")
		return
	}

	var req dto.UpdateMealItemRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_meal_item_payload", "invalid meal item payload")
		return
	}

	value, err := h.mealService.UpdateItem(r.Context(), authUserID, mealID, itemID, req.ToServiceInput())
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrNoFieldsToUpdate, http.StatusBadRequest, "invalid_meal_item_payload", "invalid meal item payload"),
		mapServiceError(service.ErrInvalidItemSource, http.StatusBadRequest, "invalid_meal_item_payload", "invalid meal item payload"),
		mapServiceError(service.ErrInvalidItemWeight, http.StatusBadRequest, "invalid_meal_item_payload", "invalid meal item payload"),
		mapServiceError(service.ErrMealItemNotFound, http.StatusNotFound, "meal_item_not_found", "meal item not found"),
		mapServiceError(service.ErrFoodNotFound, http.StatusBadRequest, "food_not_found", "food not found"),
		mapServiceError(service.ErrRecipeSourceNotFound, http.StatusBadRequest, "recipe_not_found", "recipe not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// DeleteMealItem godoc
// @Summary Delete meal item
// @Tags meals
// @Param meal_id path int true "Meal ID"
// @Param item_id path int true "Meal Item ID"
// @Success 204
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /meals/{meal_id}/items/{item_id} [delete]
func (h *Handler) DeleteMealItem(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	mealID, ok := parseIDFromPath(chi.URLParam(r, "meal_id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_meal_id", "invalid meal id")
		return
	}
	itemID, ok := parseIDFromPath(chi.URLParam(r, "item_id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_meal_item_id", "invalid meal item id")
		return
	}

	err := h.mealService.DeleteItem(r.Context(), authUserID, mealID, itemID)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrMealItemNotFound, http.StatusNotFound, "meal_item_not_found", "meal item not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDailyTotals godoc
// @Summary Get daily nutrition totals
// @Tags meals
// @Produce json
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} DailyTotalsResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /daily-totals [get]
func (h *Handler) GetDailyTotals(w http.ResponseWriter, r *http.Request) {
	authUserID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}
	date := r.URL.Query().Get("date")

	value, err := h.mealService.GetDailyTotals(r.Context(), authUserID, date)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusBadRequest, "invalid_daily_totals_query", "invalid daily totals query"),
		mapServiceError(service.ErrInvalidDate, http.StatusBadRequest, "invalid_daily_totals_query", "invalid daily totals query"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}
