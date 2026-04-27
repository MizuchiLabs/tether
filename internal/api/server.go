// Package api contains the API server
package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/mizuchilabs/tether/internal/config"
	"github.com/mizuchilabs/tether/web"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
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
	s.registerServices()

	server := &http.Server{
		Addr:              ":" + s.cfg.Port,
		Handler:           WithLogger(s.cfg, s.mux),
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)

	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}

func (s *Server) registerServices() {
	authChain := NewChain(NewAuthInterceptor(s.cfg).WithAuth)

	// Agent Service
	agentService := NewAgentService(s.cfg)
	s.mux.Handle("POST /agent/heartbeat", authChain.ThenFunc(agentService.Heartbeat()))

	// Config
	s.mux.Handle("GET /config", authChain.ThenFunc(PublishConfig(s.cfg.State)))
	s.mux.Handle("GET /envs", authChain.ThenFunc(PublishEnvs(s.cfg.State)))

	// Health
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s.mux.Handle("/", statigz.FileServer(
		web.StaticFS,
		brotli.AddEncoding,
		statigz.FSPrefix("build"),
	))

	if s.cfg.Debug {
		s.mux.HandleFunc("/debug/pprof/", pprof.Index)
		s.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		s.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		s.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		s.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
}
