package api

import (
	"encoding/json"
	"net/http"

	"github.com/mizuchilabs/tether/internal/state"
)

func PublishEnvs(state *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		envs := state.GetEnvNames()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(envs)
	}
}
