package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"nutrition/internal/http/handlers"
	httpmiddleware "nutrition/internal/http/middleware"
	"nutrition/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestProgressHandlers(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Get("/api/v1/progress/energy", h.GetEnergyProgress)
		return r
	}

	t.Run("energy progress returns 200", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{}, fakeUserGoalService{}, fakeEnergyService{
			progressFn: func(_ context.Context, _ service.EnergyProgressInput) (service.EnergyProgressOutput, error) {
				return service.EnergyProgressOutput{
					From:                "2026-01-22",
					To:                  "2026-02-18",
					AvgIntakeKcal:       2200,
					ObservedTDEEKcal:    2400,
					RecommendedTDEEKcal: 2350,
					DataQualityScore:    0.8,
				}, nil
			},
		})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/energy?from=2026-01-22&to=2026-02-18", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("energy progress invalid query returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{}, fakeUserGoalService{}, fakeEnergyService{
			progressFn: func(_ context.Context, _ service.EnergyProgressInput) (service.EnergyProgressOutput, error) {
				return service.EnergyProgressOutput{}, service.ErrInvalidEnergyProgressQuery
			},
		})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/energy?from=bad&to=2026-02-18", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}
