package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nutrition/internal/domain/user"
	"nutrition/internal/http/handlers"
	"nutrition/internal/service"
)

type fakeHealthUserService struct{}

func (f fakeHealthUserService) GetByID(_ context.Context, _ uint) (user.User, error) {
	return user.User{}, nil
}

func (f fakeHealthUserService) Update(_ context.Context, _ uint, _ service.UpdateUserInput) (user.User, error) {
	return user.User{}, nil
}

func TestHealthHandler(t *testing.T) {
	t.Run("liveness returns ok", func(t *testing.T) {
		h := handlers.New(fakeHealthUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
		rec := httptest.NewRecorder()

		h.HealthLive(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var payload map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if payload["status"] != "ok" {
			t.Fatalf("expected status ok, got %q", payload["status"])
		}
		if payload["message"] == "" {
			t.Fatal("expected message in response")
		}
	})

	t.Run("readiness error returns 503", func(t *testing.T) {
		h := handlers.New(
			fakeHealthUserService{},
			fakeAuthService{},
			fakeFoodService{},
			fakeRecipeService{},
			fakeMealService{},
			fakeBodyWeightLogService{},
			fakeReadinessChecker{readyFn: func(context.Context) error { return errors.New("db down") }},
		)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
		rec := httptest.NewRecorder()

		h.HealthReady(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
		}
		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if payload.Error.Code != "service_unavailable" {
			t.Fatalf("expected service_unavailable, got %q", payload.Error.Code)
		}
	})
}

type fakeReadinessChecker struct {
	readyFn func(ctx context.Context) error
}

func (f fakeReadinessChecker) Ready(ctx context.Context) error {
	if f.readyFn == nil {
		return nil
	}
	return f.readyFn(ctx)
}
