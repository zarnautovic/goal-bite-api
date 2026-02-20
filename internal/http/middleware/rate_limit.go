package httpmiddleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type rateLimitBucket struct {
	count   int
	resetAt time.Time
}

type IPRateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	store  map[string]rateLimitBucket
}

func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &IPRateLimiter{
		limit:  limit,
		window: window,
		now:    time.Now,
		store:  make(map[string]rateLimitBucket),
	}
}

func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + ":" + r.URL.Path + ":" + requestIP(r)
		allowed, retryAfter := l.allow(key)
		if !allowed {
			seconds := int(retryAfter.Seconds())
			if seconds <= 0 {
				seconds = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":       "rate_limited",
					"message":    "too many requests",
					"request_id": w.Header().Get(chimw.RequestIDHeader),
				},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *IPRateLimiter) allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.store[key]
	if !ok || now.After(b.resetAt) {
		l.store[key] = rateLimitBucket{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true, 0
	}
	if b.count >= l.limit {
		return false, b.resetAt.Sub(now)
	}

	b.count++
	l.store[key] = b
	return true, 0
}

func requestIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		return r.RemoteAddr
	}
	return host
}
