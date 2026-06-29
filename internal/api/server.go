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
	"net/http/pprof"
	"time"

	"github.com/mizuchilabs/tether/internal/config"
	"github.com/mizuchilabs/tether/web"
	"github.com/vearutop/statigz"
)

type Server struct {
	mux *http.ServeMux
	cfg *config.Config
}

func New(cfg *config.Config) *Server {
	return &Server{
		mux: http.NewServeMux(),
		cfg: cfg,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.registerServices(ctx)

	chain := NewChain(
		s.WithLogger,
		WithRateLimit,
		WithBodyLimit,
		WithSecurityHeaders,
	)
	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", s.cfg.Port),
		Handler:           chain.Then(s.mux),
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

func (s *Server) registerServices(ctx context.Context) {
	protec := NewChain(s.WithAuth)

	s.mux.Handle("POST /api/login", Login(s.cfg.Token))
	s.mux.Handle("POST /api/logout", Logout())
	s.mux.Handle("GET /api/ws", protec.ThenFunc(AgentWS(s.cfg.State)))
	s.mux.Handle("GET /api/events", protec.ThenFunc(EventStream(ctx, s.cfg.State)))
	s.mux.Handle("GET /api/envs", protec.ThenFunc(PublishEnvs(s.cfg.State)))
	s.mux.Handle("GET /config", protec.ThenFunc(PublishConfig(s.cfg.State)))

	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !s.cfg.NoWeb {
		s.mux.Handle("/", statigz.FileServer(web.StaticFS, statigz.FSPrefix("build")))
	}

	if s.cfg.Debug {
		s.mux.HandleFunc("/debug/pprof/", pprof.Index)
		s.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		s.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		s.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		s.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
}
