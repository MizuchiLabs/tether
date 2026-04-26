package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/mizuchilabs/tether/internal/config"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func WithLogger(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.statusCode),
			slog.Duration("duration", duration),
		}

		switch {
		case rw.statusCode >= 500:
			slog.LogAttrs(r.Context(), slog.LevelError, "http", attrs...)
		case rw.statusCode >= 400:
			slog.LogAttrs(r.Context(), slog.LevelWarn, "http", attrs...)
		case cfg.Debug:
			slog.LogAttrs(r.Context(), slog.LevelDebug, "http", attrs...)
		}
	})
}
