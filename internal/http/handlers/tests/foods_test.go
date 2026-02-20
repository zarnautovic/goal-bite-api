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

	"nutrition/internal/domain/food"
	"nutrition/internal/domain/user"
	"nutrition/internal/http/handlers"
	httpmiddleware "nutrition/internal/http/middleware"
	"nutrition/internal/service"

	"github.com/go-chi/chi/v5"
)

type noopUserService struct{}

func (n noopUserService) GetByID(_ context.Context, _ uint) (user.User, error) {
	return user.User{}, nil
}

func (n noopUserService) Update(_ context.Context, _ uint, _ service.UpdateUserInput) (user.User, error) {
	return user.User{}, nil
}

func TestFoodHandlers(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Post("/api/v1/foods", h.CreateFood)
		r.Get("/api/v1/foods", h.ListFoods)
		r.Get("/api/v1/foods/by-barcode/{barcode}", h.GetFoodByBarcode)
		r.Get("/api/v1/foods/{id}", h.GetFoodByID)
		r.Patch("/api/v1/foods/{id}", h.UpdateFood)
		r.Delete("/api/v1/foods/{id}", h.DeleteFood)
		return r
	}

	t.Run("create food returns 201", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{createFn: func(_ context.Context, userID uint, in service.CreateFoodInput) (food.Food, error) {
			if userID != 7 {
				return food.Food{}, errors.New("unexpected user id")
			}
			if in.BrandName == nil || *in.BrandName != "Fage" {
				return food.Food{}, errors.New("unexpected brand name")
			}
			return food.Food{ID: 1, Name: in.Name, BrandName: in.BrandName, KcalPer100g: in.KcalPer100g, ProteinPer100g: in.ProteinPer100g, CarbsPer100g: in.CarbsPer100g, FatPer100g: in.FatPer100g, CreatedAt: now, UpdatedAt: now}, nil
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)

		body := `{"name":"Rice","brand_name":"Fage","kcal_per_100g":130,"protein_per_100g":2.7,"carbs_per_100g":28,"fat_per_100g":0.3}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/foods", strings.NewReader(body))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 7))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	})

	t.Run("list foods invalid pagination returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/foods?limit=bad", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("get food not found returns 404", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{}, service.ErrFoodNotFound
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/foods/10", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("get food by barcode returns 200", func(t *testing.T) {
		now := time.Now().UTC()
		barcode := "5901234123457"
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{getBarcodeFn: func(_ context.Context, in string) (food.Food, error) {
			if in != barcode {
				return food.Food{}, errors.New("unexpected barcode")
			}
			return food.Food{ID: 2, Name: "Yogurt", Barcode: &barcode, KcalPer100g: 80, ProteinPer100g: 4, CarbsPer100g: 10, FatPer100g: 2, CreatedAt: now, UpdatedAt: now}, nil
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/foods/by-barcode/5901234123457", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("get food by barcode not found returns 404", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{getBarcodeFn: func(_ context.Context, _ string) (food.Food, error) {
			return food.Food{}, service.ErrFoodBarcodeNotFound
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/foods/by-barcode/5901234123457", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("patch food with no fields returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{updateFn: func(_ context.Context, _, _ uint, _ service.UpdateFoodInput) (food.Food, error) {
			return food.Food{}, service.ErrNoFieldsToUpdate
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/foods/1", strings.NewReader(`{}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 7))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("update food forbidden returns 403", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{updateFn: func(_ context.Context, _ uint, _ uint, _ service.UpdateFoodInput) (food.Food, error) {
			return food.Food{}, service.ErrFoodForbidden
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/foods/1", strings.NewReader(`{"name":"x"}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 7))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
		}
	})

	t.Run("delete food returns 204", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{deleteFn: func(_ context.Context, userID, id uint) error {
			if userID != 7 || id != 1 {
				return errors.New("unexpected args")
			}
			return nil
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/foods/1", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 7))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})

	t.Run("list foods returns 200 with payload", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{listFn: func(_ context.Context, limit, offset int) ([]food.Food, error) {
			if limit != 20 || offset != 0 {
				return nil, errors.New("unexpected pagination")
			}
			return []food.Food{{ID: 1, Name: "Egg", KcalPer100g: 155, ProteinPer100g: 13, CarbsPer100g: 1.1, FatPer100g: 11, CreatedAt: now, UpdatedAt: now}}, nil
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/foods", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var payload []food.Food
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(payload) != 1 || payload[0].Name != "Egg" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
	})

	t.Run("list foods with q uses search", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{searchFn: func(_ context.Context, query string, limit, offset int) ([]food.Food, error) {
			if query != "egg" {
				return nil, errors.New("unexpected query")
			}
			if limit != 20 || offset != 0 {
				return nil, errors.New("unexpected pagination")
			}
			return []food.Food{{ID: 1, Name: "Egg", CreatedAt: now, UpdatedAt: now}}, nil
		}}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/foods?q=egg", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
