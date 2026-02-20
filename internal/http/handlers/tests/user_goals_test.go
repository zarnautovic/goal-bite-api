package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"goal-bite-api/internal/domain/usergoal"
	"goal-bite-api/internal/http/handlers"
	httpmiddleware "goal-bite-api/internal/http/middleware"
	"goal-bite-api/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestUserGoalHandlers(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Put("/api/v1/user-goals", h.UpsertUserGoals)
		r.Get("/api/v1/user-goals", h.GetUserGoals)
		r.Get("/api/v1/progress/daily", h.GetDailyProgress)
		return r
	}

	t.Run("upsert returns 200", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{}, fakeUserGoalService{
			upsertFn: func(_ context.Context, _ service.UpsertUserGoalInput) (usergoal.UserGoal, error) {
				return usergoal.UserGoal{
					ID:             1,
					UserID:         1,
					TargetKcal:     2200,
					TargetProteinG: 150,
					TargetCarbsG:   220,
					TargetFatG:     70,
					CreatedAt:      now,
					UpdatedAt:      now,
				}, nil
			},
		})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/user-goals", strings.NewReader(`{"target_kcal":2200,"target_protein_g":150,"target_carbs_g":220,"target_fat_g":70}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("get goals returns 404 when missing", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{}, fakeUserGoalService{
			getFn: func(_ context.Context, _ uint) (usergoal.UserGoal, error) {
				return usergoal.UserGoal{}, service.ErrUserGoalNotFound
			},
		})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user-goals", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("daily progress returns 200", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{}, fakeUserGoalService{
			progressFn: func(_ context.Context, _ uint, _ string) (service.DailyProgressOutput, error) {
				return service.DailyProgressOutput{
					Date:              "2026-02-17",
					TotalKcal:         1800,
					TargetKcal:        2200,
					RemainingKcal:     400,
					TotalProteinG:     120,
					TargetProteinG:    150,
					RemainingProteinG: 30,
				}, nil
			},
		})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/daily?date=2026-02-17", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
