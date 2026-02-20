package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"goal-bite-api/internal/domain/recipe"
	"goal-bite-api/internal/domain/recipeingredient"
	"goal-bite-api/internal/http/handlers"
	httpmiddleware "goal-bite-api/internal/http/middleware"
	"goal-bite-api/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestRecipeHandlers(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Post("/api/v1/recipes", h.CreateRecipe)
		r.Get("/api/v1/recipes", h.ListRecipes)
		r.Get("/api/v1/recipes/{id}", h.GetRecipeByID)
		r.Patch("/api/v1/recipes/{id}", h.UpdateRecipe)
		r.Delete("/api/v1/recipes/{id}", h.DeleteRecipe)
		return r
	}

	t.Run("create recipe returns 201", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{createFn: func(_ context.Context, userID uint, in service.CreateRecipeInput) (recipe.Recipe, error) {
			if userID != 7 {
				return recipe.Recipe{}, errors.New("unexpected user id")
			}
			return recipe.Recipe{ID: 1, Name: in.Name, YieldWeightG: in.YieldWeightG, KcalPer100g: 120, ProteinPer100g: 8, CarbsPer100g: 10, FatPer100g: 5, CreatedAt: now, UpdatedAt: now}, nil
		}}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)

		body := `{"name":"Goulash","yield_weight_g":1200,"ingredients":[{"food_id":1,"raw_weight_g":500}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(body))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 7))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	})

	t.Run("ingredient food missing returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{createFn: func(_ context.Context, _ uint, _ service.CreateRecipeInput) (recipe.Recipe, error) {
			return recipe.Recipe{}, service.ErrIngredientFoodNotFound
		}}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)

		body := `{"name":"Goulash","yield_weight_g":1200,"ingredients":[{"food_id":999,"raw_weight_g":500}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(body))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 7))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("get recipe returns 200", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{getFn: func(_ context.Context, _ uint) (recipe.Recipe, error) {
			return recipe.Recipe{ID: 1, Name: "Goulash", YieldWeightG: 1200, KcalPer100g: 120, ProteinPer100g: 8, CarbsPer100g: 10, FatPer100g: 5, Ingredients: []recipeingredient.RecipeIngredient{{ID: 1, RecipeID: 1, FoodID: 1, RawWeightG: 500}}, CreatedAt: now, UpdatedAt: now}, nil
		}}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/1", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var payload recipe.Recipe
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if payload.ID != 1 || payload.Name == "" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
	})

	t.Run("list recipes with q uses search", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{searchFn: func(_ context.Context, query string, limit, offset int) ([]recipe.Recipe, error) {
			if query != "gou" {
				return nil, errors.New("unexpected query")
			}
			if limit != 20 || offset != 0 {
				return nil, errors.New("unexpected pagination")
			}
			return []recipe.Recipe{{ID: 1, Name: "Goulash", CreatedAt: now, UpdatedAt: now}}, nil
		}}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=gou", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("update recipe forbidden returns 403", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{updateFn: func(_ context.Context, _, _ uint, _ service.UpdateRecipeInput) (recipe.Recipe, error) {
			return recipe.Recipe{}, service.ErrRecipeForbidden
		}}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/recipes/1", strings.NewReader(`{"name":"x"}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 7))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
		}
	})

	t.Run("delete recipe returns 204", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{deleteFn: func(_ context.Context, userID, id uint) error {
			if userID != 7 || id != 1 {
				return errors.New("unexpected args")
			}
			return nil
		}}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/recipes/1", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 7))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}
