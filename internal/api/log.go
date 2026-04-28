package api

import (
	"log/slog"
	"net/http"
	"strings"
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

		var level slog.Level

		// Determine log severity based on status code
		switch {
		case rw.statusCode >= 500:
			level = slog.LevelError
		case rw.statusCode >= 400:
			level = slog.LevelWarn
		default:
			level = slog.LevelInfo
			if cfg.Debug {
				level = slog.LevelDebug
			}
			path := r.URL.Path
			if strings.HasPrefix(path, "/_app/") {
				return
			}

			// Filter out noisy successful requests (2xx/3xx) when not debugging
			if !cfg.Debug {
				if path == "/healthz" {
					return
				}
				if path == "/config" && rw.statusCode == http.StatusOK {
					return
				}

				// Skip static file spam
				isAPI := strings.HasPrefix(path, "/api/")
				if r.Method == http.MethodGet && !isAPI && rw.statusCode < 400 && rw.size < 1024 {
					return
				}
			}
		}

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.statusCode),
			slog.Duration("duration", duration),
			slog.String("ip", r.RemoteAddr),
		}

		slog.LogAttrs(r.Context(), level, "http", attrs...)
	})
}
