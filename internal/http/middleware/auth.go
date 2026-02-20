package httpmiddleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"goal-bite-api/internal/auth"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const userIDContextKey contextKey = "auth_user_id"

func WithUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (uint, bool) {
	v := ctx.Value(userIDContextKey)
	id, ok := v.(uint)
	return id, ok && id > 0
}

func RequireAuth(jwtManager *auth.JWTManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := strings.TrimSpace(r.Header.Get("Authorization"))
			if raw == "" || !strings.HasPrefix(raw, "Bearer ") {
				writeUnauthorized(w)
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
			userID, err := jwtManager.Parse(token)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":       "unauthorized",
			"message":    "unauthorized",
			"request_id": w.Header().Get(chimw.RequestIDHeader),
		},
	})
}
