// Package api provides the HTTP server, middleware, and request handlers.
package api

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/rs/cors"

	"github.com/mizuchilabs/kata/logx"

	"github.com/mizuchilabs/tether/internal/config"
	"github.com/mizuchilabs/tether/web"
)

type Server struct {
	mux *chi.Mux
	cfg *config.Config
}

func New(ctx context.Context, cfg *config.Config) *Server {
	mux := chi.NewRouter()

	if logx.IsTerminal() {
		mux.Use(middleware.Logger)
		mux.Use(middleware.Recoverer)
	} else {
		mux.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
			RecoverPanics: true,
			Schema:        httplog.SchemaOTEL,
			Skip: func(req *http.Request, respStatus int) bool {
				return respStatus < http.StatusBadRequest && req.URL.Path == "/healthz"
			},
		}))
	}
	mux.Use(cors.Default().Handler)
	mux.Use(middleware.RequestSize(1 << 20))
	mux.Use(securityHeaders())
	mux.Use(rateLimitAPI(100, time.Minute))
	mux.Use(middleware.CleanPath)

	mux.Group(func(r chi.Router) {
		r.Post("/api/login", Login(cfg.Token))
		r.Post("/api/logout", Logout())
		r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		if !cfg.NoWeb {
			r.Handle("/*", web.Handler())
		}
	})
	mux.Group(func(r chi.Router) {
		r.Use(WithAuth(cfg.Token))
		r.Get("/api/ws", AgentWS(cfg.State))
		r.Get("/api/events", EventStream(ctx, cfg.State))
		r.Get("/api/envs", PublishEnvs(cfg.State))
		r.Get("/config", PublishConfig(cfg.State))
	})

	return &Server{
		mux: mux,
		cfg: cfg,
	}
}

func (s *Server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:              net.JoinHostPort("0.0.0.0", s.cfg.Port),
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MiB
		TLSConfig:         &tls.Config{MinVersion: tls.VersionTLS13},
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Server listening on", "port", s.cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down server...")
		errShutdown := errors.New("shutdown timeout")
		shutdownCtx, cancel := context.WithTimeoutCause(context.Background(), 3*time.Second, errShutdown)
		defer cancel()
		return server.Shutdown(shutdownCtx)

	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}
