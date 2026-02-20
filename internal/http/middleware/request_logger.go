package httpmiddleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type slogFormatter struct {
	logger *slog.Logger
}

type slogEntry struct {
	logger *slog.Logger
	req    *http.Request
}

func NewSlogRequestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return chimw.RequestLogger(&slogFormatter{logger: logger})
}

func (f *slogFormatter) NewLogEntry(r *http.Request) chimw.LogEntry {
	return &slogEntry{logger: f.logger, req: r}
}

func (e *slogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ interface{}) {
	attrs := []any{
		"request_id", chimw.GetReqID(e.req.Context()),
		"method", e.req.Method,
		"path", e.req.URL.Path,
		"status", status,
		"bytes", bytes,
		"duration_ms", elapsed.Milliseconds(),
		"remote_ip", e.req.RemoteAddr,
		"user_agent", e.req.UserAgent(),
	}

	switch {
	case status >= 500:
		e.logger.Error("http_request", attrs...)
	case status >= 400:
		e.logger.Warn("http_request", attrs...)
	default:
		e.logger.Info("http_request", attrs...)
	}
}

func (e *slogEntry) Panic(v interface{}, stack []byte) {
	e.logger.Error("http_panic",
		"request_id", chimw.GetReqID(e.req.Context()),
		"method", e.req.Method,
		"path", e.req.URL.Path,
		"panic", v,
		"stack", string(stack),
	)
}
