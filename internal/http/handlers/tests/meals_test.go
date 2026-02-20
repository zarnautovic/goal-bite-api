package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"goal-bite-api/internal/domain/meal"
	"goal-bite-api/internal/domain/mealitem"
	"goal-bite-api/internal/http/handlers"
	httpmiddleware "goal-bite-api/internal/http/middleware"
	"goal-bite-api/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestMealHandlers(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Post("/api/v1/meals", h.CreateMeal)
		r.Get("/api/v1/meals", h.ListMeals)
		r.Get("/api/v1/meals/{id}", h.GetMealByID)
		r.Patch("/api/v1/meals/{id}", h.UpdateMeal)
		r.Delete("/api/v1/meals/{id}", h.DeleteMeal)
		r.Post("/api/v1/meals/{id}/items", h.AddMealItem)
		r.Patch("/api/v1/meals/{meal_id}/items/{item_id}", h.UpdateMealItem)
		r.Delete("/api/v1/meals/{meal_id}/items/{item_id}", h.DeleteMealItem)
		return r
	}

	t.Run("create meal returns 201", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{createFn: func(_ context.Context, _ service.CreateMealInput) (meal.Meal, error) {
			return meal.Meal{ID: 1, UserID: 1, MealType: meal.MealTypeLunch, EatenAt: now}, nil
		}}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/meals", strings.NewReader(`{"meal_type":"lunch","eaten_at":"2026-02-17T12:00:00Z"}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, rec.Code)
		}
	})

	t.Run("add meal item returns 201", func(t *testing.T) {
		fid := uint(1)
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{addItemFn: func(_ context.Context, _ uint, _ uint, _ service.AddMealItemInput) (mealitem.MealItem, error) {
			return mealitem.MealItem{ID: 1, MealID: 1, FoodID: &fid, WeightG: 200}, nil
		}}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/meals/1/items", strings.NewReader(`{"food_id":1,"weight_g":200}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, rec.Code)
		}
	})

	t.Run("create meal with invalid nested item returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/meals", strings.NewReader(`{"meal_type":"lunch","eaten_at":"2026-02-17T12:00:00Z","items":[{"weight_g":200}]}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("list meals invalid query returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/meals?date=2026-02-17&limit=bad", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("daily totals returns 200", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{
			dailyFn: func(_ context.Context, _ uint, _ string) (service.DailyTotalsOutput, error) {
				return service.DailyTotalsOutput{
					Date:          "2026-02-17",
					TotalKcal:     650,
					TotalProteinG: 40,
					TotalCarbsG:   75,
					TotalFatG:     20,
				}, nil
			},
		}, fakeBodyWeightLogService{})
		r := chi.NewRouter()
		r.Get("/api/v1/daily-totals", h.GetDailyTotals)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-totals?date=2026-02-17", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("update meal returns 200", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{
			updateFn: func(_ context.Context, _ uint, _ uint, _ service.UpdateMealInput) (meal.Meal, error) {
				return meal.Meal{ID: 1, UserID: 1, MealType: meal.MealTypeDinner, EatenAt: now}, nil
			},
		}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/meals/1", strings.NewReader(`{"meal_type":"dinner"}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("delete meal returns 204", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{
			deleteFn: func(_ context.Context, _ uint, _ uint) error { return nil },
		}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/meals/1", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected %d, got %d", http.StatusNoContent, rec.Code)
		}
	})

	t.Run("update meal item returns 200", func(t *testing.T) {
		fid := uint(1)
		w := 180.0
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{
			updateItemFn: func(_ context.Context, _ uint, _ uint, _ uint, _ service.UpdateMealItemInput) (mealitem.MealItem, error) {
				return mealitem.MealItem{ID: 2, MealID: 1, FoodID: &fid, WeightG: w}, nil
			},
		}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/meals/1/items/2", strings.NewReader(`{"weight_g":180}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("delete meal item returns 204", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{
			deleteItemFn: func(_ context.Context, _ uint, _ uint, _ uint) error { return nil },
		}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/meals/1/items/2", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}
