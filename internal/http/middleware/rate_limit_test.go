package httpmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIPRateLimiterBlocksAfterLimit(t *testing.T) {
	limiter := NewIPRateLimiter(2, time.Minute)
	limiter.now = func() time.Time { return time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC) }

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "203.0.113.10:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d: expected 204, got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatalf("expected Retry-After header")
	}
}

func TestIPRateLimiterResetsWindow(t *testing.T) {
	now := time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC)
	limiter := NewIPRateLimiter(1, time.Minute)
	limiter.now = func() time.Time { return now }
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req1.RemoteAddr = "203.0.113.20:10001"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req2.RemoteAddr = "203.0.113.20:10001"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}

	now = now.Add(61 * time.Second)
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req3.RemoteAddr = "203.0.113.20:10001"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusNoContent {
		t.Fatalf("expected 204 after reset, got %d", rec3.Code)
	}
}
