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

	"nutrition/internal/domain/user"
	"nutrition/internal/http/handlers"
	httpmiddleware "nutrition/internal/http/middleware"
	"nutrition/internal/service"

	"github.com/go-chi/chi/v5"
)

type fakeUserService struct {
	result   user.User
	err      error
	updateFn func(ctx context.Context, id uint, in service.UpdateUserInput) (user.User, error)
}

func (f fakeUserService) GetByID(_ context.Context, _ uint) (user.User, error) {
	return f.result, f.err
}

func (f fakeUserService) Update(ctx context.Context, id uint, in service.UpdateUserInput) (user.User, error) {
	if f.updateFn == nil {
		return f.result, f.err
	}
	return f.updateFn(ctx, id, in)
}

func TestGetUserByIDHandler(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Get("/api/v1/users/{id}", h.GetUserByID)
		return r
	}

	t.Run("returns 400 for invalid id", func(t *testing.T) {
		h := handlers.New(fakeUserService{}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/not-a-number", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if payload.Error.Code != "invalid_user_id" {
			t.Fatalf("expected invalid_user_id, got %q", payload.Error.Code)
		}
	})

	t.Run("returns 404 when user does not exist", func(t *testing.T) {
		h := handlers.New(fakeUserService{err: service.ErrUserNotFound}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/999", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 999))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}

		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if payload.Error.Code != "user_not_found" {
			t.Fatalf("expected user_not_found, got %q", payload.Error.Code)
		}
	})

	t.Run("returns 500 on unexpected service error", func(t *testing.T) {
		h := handlers.New(fakeUserService{err: errors.New("db timeout")}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		var payload handlers.ErrorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if payload.Error.Code != "database_error" {
			t.Fatalf("expected database_error, got %q", payload.Error.Code)
		}
	})

	t.Run("returns 200 with user payload", func(t *testing.T) {
		now := time.Now().UTC()
		expected := user.User{ID: 1, Name: "Test User", CreatedAt: now, UpdatedAt: now}
		h := handlers.New(fakeUserService{result: expected}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var got user.User
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if got.ID != expected.ID || got.Name != expected.Name {
			t.Fatalf("unexpected user payload: got %+v expected %+v", got, expected)
		}
	})
}

func TestMeHandler(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Get("/api/v1/auth/me", h.Me)
		return r
	}

	t.Run("returns 200 with current user payload", func(t *testing.T) {
		now := time.Now().UTC()
		sex := "male"
		height := 178.0
		activity := "moderate"
		expected := user.User{
			ID:            1,
			Name:          "Test User",
			Email:         "test@example.com",
			Sex:           &sex,
			HeightCM:      &height,
			ActivityLevel: &activity,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		h := handlers.New(fakeUserService{result: expected}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var got user.User
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if got.ID != expected.ID || got.Email != expected.Email {
			t.Fatalf("unexpected user payload: got %+v expected %+v", got, expected)
		}
		if got.Sex == nil || *got.Sex != sex {
			t.Fatalf("expected sex %q, got %+v", sex, got.Sex)
		}
	})
}

func TestUpdateMeHandler(t *testing.T) {
	newRouter := func(h *handlers.Handler) http.Handler {
		r := chi.NewRouter()
		r.Patch("/api/v1/users/me", h.UpdateMe)
		return r
	}

	t.Run("returns 200 when profile is updated", func(t *testing.T) {
		activity := "active"
		height := 180.0
		h := handlers.New(fakeUserService{
			updateFn: func(_ context.Context, id uint, _ service.UpdateUserInput) (user.User, error) {
				return user.User{ID: id, Name: "Updated", Email: "test@example.com", ActivityLevel: &activity, HeightCM: &height}, nil
			},
		}, fakeAuthService{}, fakeFoodService{}, fakeRecipeService{}, fakeMealService{}, fakeBodyWeightLogService{})
		router := newRouter(h)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", strings.NewReader(`{"activity_level":"active","height_cm":180}`))
		req = req.WithContext(httpmiddleware.WithUserID(req.Context(), 1))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
