package api

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// responseWriter captures status code and response size for logging.
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

// WithLogger logs request method, path, status, duration, and client IP.
func (s *Server) WithLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		path := r.URL.Path

		if strings.HasPrefix(path, "/_app/") {
			return
		}

		if !s.cfg.Debug {
			isSmallGet := r.Method == http.MethodGet && rw.statusCode < 400 && rw.size < 1024
			if path == "/healthz" || (path == "/config" && rw.statusCode == http.StatusOK) || isSmallGet {
				return
			}
		}

		var level slog.Level
		switch {
		case rw.statusCode >= 500:
			level = slog.LevelError
		case rw.statusCode >= 400:
			level = slog.LevelWarn
		default:
			level = slog.LevelInfo
			if s.cfg.Debug {
				level = slog.LevelDebug
			}
		}

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", path),
			slog.Int("status", rw.statusCode),
			slog.Duration("duration", duration),
			slog.String("ip", r.RemoteAddr),
		}

		slog.LogAttrs(r.Context(), level, "http", attrs...)
	})
}
