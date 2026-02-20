package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"goal-bite-api/internal/domain/bodyweightlog"
	"goal-bite-api/internal/http/handlers"
	httpmiddleware "goal-bite-api/internal/http/middleware"
	"goal-bite-api/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestBodyWeightLogHandlers(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Post("/api/v1/body-weight-logs", h.CreateBodyWeightLog)
		r.Get("/api/v1/body-weight-logs", h.ListBodyWeightLogs)
		r.Get("/api/v1/body-weight-logs/latest", h.GetLatestBodyWeightLog)
		return r
	}

	t.Run("create body weight log returns 201", func(t *testing.T) {
		now := time.Now().UTC()
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{
			createFn: func(_ context.Context, _ service.CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error) {
				return bodyweightlog.BodyWeightLog{ID: 1, UserID: 1, WeightKG: 85.2, LoggedAt: now}, nil
			},
		})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/body-weight-logs", strings.NewReader(`{"weight_kg":85.2,"logged_at":"2026-02-17T08:00:00Z"}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, rec.Code)
		}
	})

	t.Run("latest returns 404 when missing", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{
			latestFn: func(_ context.Context, _ uint) (bodyweightlog.BodyWeightLog, error) {
				return bodyweightlog.BodyWeightLog{}, service.ErrBodyWeightLogNotFound
			},
		})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/body-weight-logs/latest", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("list invalid query returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/body-weight-logs?from=2026-02-01&to=2026-02-17&limit=bad", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}
