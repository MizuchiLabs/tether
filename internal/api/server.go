// Package api provides the HTTP server, middleware, and request handlers.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"

	"github.com/mizuchilabs/kata/logx"

	"github.com/mizuchilabs/tether/internal/state"
	"github.com/mizuchilabs/tether/web"
)

// Config holds the server settings from the CLI.
type Config struct {
	Port           string
	Token          string
	NoWeb          bool
	TrustedProxies string
}

// Serve runs the HTTP server until ctx is done.
func Serve(ctx context.Context, st *state.State, cfg Config) error {
	mux := chi.NewRouter()
	mux.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
		RecoverPanics: true,
		Schema:        httplog.SchemaOTEL.Concise(logx.IsTerminal()),
		Skip:          quietRequest,
	}))
	mux.Use(middleware.RequestSize(1 << 20))
	mux.Use(securityHeaders())
	mux.Use(rateLimitAPI(cfg.TrustedProxies, 100, time.Minute))
	mux.Use(middleware.CleanPath)

	mux.Post("/api/login", login(cfg.Token))
	mux.Post("/api/logout", logout)
	mux.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	if !cfg.NoWeb {
		mux.Handle("/*", web.Handler())
	}
	mux.Group(func(r chi.Router) {
		r.Use(withAuth(cfg.Token))
		r.Get("/api/ws", agentWS(st))
		r.Get("/api/events", eventStream(ctx, st))
		r.Get("/api/envs", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, st.EnvNames())
		})
		r.Get("/config", publishConfig(st))
	})

	server := &http.Server{
		Addr:              net.JoinHostPort("0.0.0.0", cfg.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Server listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}

// quietRequest skips logging successful polling and stream traffic. Failures always log.
func quietRequest(r *http.Request, status int) bool {
	if status >= http.StatusBadRequest {
		return false
	}
	switch r.URL.Path {
	case "/healthz", "/config", "/api/envs", "/api/ws", "/api/events":
		return true
	}
	return false
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
