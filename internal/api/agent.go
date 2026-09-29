package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/mizuchilabs/tether/internal/state"
)

const pingInterval = 15 * time.Second

type updateRequest struct {
	Env    string          `json:"env"`
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

// agentWS accepts WebSocket connections from agents pushing config updates.
func agentWS(st *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = c.CloseNow() }()
		c.SetReadLimit(8 << 20)

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		go keepAlive(ctx, c)

		addr := middleware.GetClientIP(r.Context())
		if addr == "" {
			addr, _, _ = net.SplitHostPort(r.RemoteAddr)
		}

		var env, name string
		defer func() {
			if name != "" {
				st.AgentDisconnected(env, name)
			}
		}()

		for {
			var req updateRequest
			if err := wsjson.Read(ctx, c, &req); err != nil {
				slog.Debug("Agent connection closed", "agent", name, "error", err)
				return
			}
			if req.Name == "" || len(req.Config) == 0 {
				continue
			}

			var cfg state.Config
			if err := json.Unmarshal(req.Config, &cfg); err != nil {
				slog.Error("Invalid agent config", "agent", req.Name, "error", err)
				continue
			}

			if req.Env != env || req.Name != name {
				if name != "" {
					st.AgentDisconnected(env, name)
				}
				env, name = req.Env, req.Name
				st.AgentConnected(env, name, addr)
			}
			st.UpdateAgent(env, name, cfg)
		}
	}
}

// keepAlive pings the agent so dead connections are dropped instead of hanging forever.
func keepAlive(ctx context.Context, c *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingInterval)
			err := c.Ping(pingCtx)
			cancel()
			if err != nil {
				_ = c.CloseNow()
				return
			}
		}
	}
}
