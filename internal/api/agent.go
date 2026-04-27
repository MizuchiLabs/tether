package api

import (
	"encoding/json"
	"net/http"

	"github.com/mizuchilabs/tether/internal/state"
)

type HeartbeatRequest struct {
	Env    string          `json:"env"`
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

func Heartbeat(state *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req HeartbeatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		state.UpdateAgent(req.Env, req.Name, req.Config)
		w.WriteHeader(http.StatusOK)
	}
}
