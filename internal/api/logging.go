package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// quietRequest reports whether a success on this path is routine polling or
// stream traffic (agents, UI, probes) that is not worth a log line. Failures always log.
func quietRequest(path string, status int) bool {
	if status >= http.StatusBadRequest {
		return false
	}
	switch path {
	case "/healthz", "/config", "/api/envs", "/api/ws", "/api/events":
		return true
	}
	return false
}

// terminalLogger is chi's request logger routed through slog. Colors off since slog escapes ANSI codes.
func terminalLogger() func(http.Handler) http.Handler {
	inner := &middleware.DefaultLogFormatter{
		Logger:  slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo),
		NoColor: true,
	}
	return middleware.RequestLogger(&quietLogFormatter{inner: inner})
}

// quietLogFormatter filters at the entry, chi's middleware.Logger has no skip hook.
type quietLogFormatter struct {
	inner middleware.LogFormatter
}

func (f *quietLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &quietLogEntry{inner: f.inner.NewLogEntry(r), path: r.URL.Path}
}

type quietLogEntry struct {
	inner middleware.LogEntry
	path  string
}

func (e *quietLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra any) {
	if quietRequest(e.path, status) {
		return
	}
	e.inner.Write(status, bytes, header, elapsed, extra)
}

func (e *quietLogEntry) Panic(v any, stack []byte) {
	e.inner.Panic(v, stack)
}
