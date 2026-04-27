package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/mizuchilabs/tether/internal/config"
)

// responseWriter wraps http.ResponseWriter to capture the status code and size.
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	size        int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
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

		// Skip logging for high-frequency endpoints unless debugging
		if !cfg.Debug && (r.URL.Path == "/healthz" || r.URL.Path == "/api/heartbeat") {
			return
		}

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.statusCode),
			slog.Duration("duration", duration),
			slog.Int("size", rw.size),
			slog.String("ip", r.RemoteAddr),
		}

		// Log based on status code severity
		switch {
		case rw.statusCode >= 500:
			slog.LogAttrs(r.Context(), slog.LevelError, "http_request", attrs...)
		case rw.statusCode >= 400:
			slog.LogAttrs(r.Context(), slog.LevelWarn, "http_request", attrs...)
		default:
			// Treat 2xx and 3xx as Info, or Debug based on your config
			level := slog.LevelInfo
			if cfg.Debug {
				level = slog.LevelDebug
			}
			slog.LogAttrs(r.Context(), level, "http_request", attrs...)
		}
	})
}
