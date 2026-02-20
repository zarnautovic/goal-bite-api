package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nutrition/internal/http/handlers"
	"nutrition/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestAuthHandlers(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Post("/api/v1/auth/register", h.Register)
		r.Post("/api/v1/auth/login", h.Login)
		r.Post("/api/v1/auth/refresh", h.Refresh)
		r.Post("/api/v1/auth/logout", h.Logout)
		return r
	}

	t.Run("register weak password returns password policy error", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"name":"A","email":"a@example.com","password":"short"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.Error.Code != "invalid_password_policy" {
			t.Fatalf("expected invalid_password_policy, got %q", payload.Error.Code)
		}
	})

	t.Run("refresh returns 200", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{
			refreshFn: func(_ context.Context, _ string) (service.AuthResult, error) {
				return service.AuthResult{Token: "a", AccessToken: "a", RefreshToken: "r"}, nil
			},
		}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(`{"refresh_token":"abc"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("login blocked returns 429", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{
			loginFn: func(_ context.Context, _, _ string) (service.AuthResult, error) {
				return service.AuthResult{}, service.ErrTooManyLoginAttempts
			},
		}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"a@example.com","password":"Pass1234!"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("expected %d, got %d", http.StatusTooManyRequests, rec.Code)
		}
		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.Error.Code != "too_many_login_attempts" {
			t.Fatalf("expected too_many_login_attempts, got %q", payload.Error.Code)
		}
	})

	t.Run("refresh invalid payload returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.Error.Code != "invalid_refresh_payload" {
			t.Fatalf("expected invalid_refresh_payload, got %q", payload.Error.Code)
		}
	})

	t.Run("refresh invalid token returns 401", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{
			refreshFn: func(_ context.Context, _ string) (service.AuthResult, error) {
				return service.AuthResult{}, service.ErrInvalidRefreshToken
			},
		}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(`{"refresh_token":"bad"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
		}
		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.Error.Code != "invalid_refresh_token" {
			t.Fatalf("expected invalid_refresh_token, got %q", payload.Error.Code)
		}
	})

	t.Run("logout returns 204", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{
			logoutFn: func(_ context.Context, _ string) error { return nil },
		}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{"refresh_token":"abc"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected %d, got %d", http.StatusNoContent, rec.Code)
		}
	})

	t.Run("logout invalid payload returns 400", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
		}
		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.Error.Code != "invalid_logout_payload" {
			t.Fatalf("expected invalid_logout_payload, got %q", payload.Error.Code)
		}
	})

	t.Run("logout invalid token returns 401", func(t *testing.T) {
		h := handlers.New(noopUserService{}, fakeAuthService{
			logoutFn: func(_ context.Context, _ string) error {
				return service.ErrInvalidRefreshToken
			},
		}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		r := newRouter(h)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{"refresh_token":"bad"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
		}
		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.Error.Code != "invalid_refresh_token" {
			t.Fatalf("expected invalid_refresh_token, got %q", payload.Error.Code)
		}
	})
}
