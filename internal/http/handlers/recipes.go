package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"goal-bite-api/internal/domain/recipe"
	"goal-bite-api/internal/http/dto"
	"goal-bite-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// CreateRecipe godoc
// @Summary Create recipe
// @Tags recipes
// @Accept json
// @Produce json
// @Param payload body dto.CreateRecipeRequest true "Recipe payload"
// @Success 201 {object} RecipeResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /recipes [post]
func (h *Handler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	var req dto.CreateRecipeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload")
		return
	}

	value, err := h.recipeService.Create(r.Context(), userID, req.ToServiceInput())
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusUnauthorized, "unauthorized", "unauthorized"),
		mapServiceError(service.ErrInvalidRecipeName, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload"),
		mapServiceError(service.ErrInvalidYieldWeight, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload"),
		mapServiceError(service.ErrInvalidRecipeIngredients, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload"),
		mapServiceError(service.ErrIngredientFoodNotFound, http.StatusBadRequest, "ingredient_food_not_found", "ingredient food not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusCreated, value)
}

// GetRecipeByID godoc
// @Summary Get recipe by ID
// @Tags recipes
// @Produce json
// @Param id path int true "Recipe ID"
// @Success 200 {object} RecipeResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /recipes/{id} [get]
func (h *Handler) GetRecipeByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_recipe_id", "invalid recipe id")
		return
	}

	value, err := h.recipeService.GetByID(r.Context(), id)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrRecipeNotFound, http.StatusNotFound, "recipe_not_found", "recipe not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// ListRecipes godoc
// @Summary List recipes
// @Tags recipes
// @Produce json
// @Param q query string false "Search by recipe name (case-insensitive, partial match)"
// @Param limit query int false "Page size (default 20, max 100)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {array} RecipeResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /recipes [get]
func (h *Handler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := parsePagination(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_pagination", "invalid pagination")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	var (
		values []recipe.Recipe
		err    error
	)
	if query == "" {
		values, err = h.recipeService.List(r.Context(), limit, offset)
	} else {
		values, err = h.recipeService.Search(r.Context(), query, limit, offset)
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

// UpdateRecipe godoc
// @Summary Update recipe
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path int true "Recipe ID"
// @Param payload body dto.UpdateRecipeRequest true "Recipe update payload"
// @Success 200 {object} RecipeResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /recipes/{id} [patch]
func (h *Handler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	id, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_recipe_id", "invalid recipe id")
		return
	}

	var req dto.UpdateRecipeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload")
		return
	}

	value, err := h.recipeService.Update(r.Context(), userID, id, req.ToServiceInput())
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusUnauthorized, "unauthorized", "unauthorized"),
		mapServiceError(service.ErrRecipeNotFound, http.StatusNotFound, "recipe_not_found", "recipe not found"),
		mapServiceError(service.ErrRecipeForbidden, http.StatusForbidden, "forbidden", "forbidden"),
		mapServiceError(service.ErrInvalidRecipeName, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload"),
		mapServiceError(service.ErrInvalidYieldWeight, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload"),
		mapServiceError(service.ErrInvalidRecipeIngredients, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload"),
		mapServiceError(service.ErrNoFieldsToUpdate, http.StatusBadRequest, "invalid_recipe_payload", "invalid recipe payload"),
		mapServiceError(service.ErrIngredientFoodNotFound, http.StatusBadRequest, "ingredient_food_not_found", "ingredient food not found"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// DeleteRecipe godoc
// @Summary Delete recipe
// @Tags recipes
// @Produce json
// @Param id path int true "Recipe ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 403 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /recipes/{id} [delete]
func (h *Handler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuthUserID(w, r)
	if !ok {
		return
	}

	id, ok := parseIDFromPath(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_recipe_id", "invalid recipe id")
		return
	}

	err := h.recipeService.Delete(r.Context(), userID, id)
	if writeMappedServiceError(w, err,
		mapServiceError(service.ErrInvalidUserID, http.StatusUnauthorized, "unauthorized", "unauthorized"),
		mapServiceError(service.ErrRecipeNotFound, http.StatusNotFound, "recipe_not_found", "recipe not found"),
		mapServiceError(service.ErrRecipeForbidden, http.StatusForbidden, "forbidden", "forbidden"),
	) {
		return
	}
	if err != nil {
		writeDatabaseError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
