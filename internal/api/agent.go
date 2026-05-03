package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/mizuchilabs/tether/internal/state"
)

type UpdateRequest struct {
	Env    string          `json:"env"`
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

func AgentWS(state *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = c.CloseNow() }()

		for {
			var req UpdateRequest
			if err := wsjson.Read(r.Context(), c, &req); err != nil {
				slog.Info("Agent disconnected", "error", err)
				return // Drop connection on read error or close
			}

			if req.Name == "" {
				continue
			}

			state.UpdateAgent(req.Env, req.Name, req.Config)
		}
	}
}
