package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mizuchilabs/tether/internal/state"
	"go.yaml.in/yaml/v3"
)

func PublishConfig(state *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")
		accept := r.Header.Get("Accept")
		master := state.GetMaster(r.URL.Query().Get("env"))

		// Determine response format: prefer query param over header
		if format == "yaml" || (format == "" && strings.Contains(accept, "yaml")) {
			w.Header().Set("Content-Type", "application/x-yaml")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			yamlBytes, err := yaml.Marshal(master)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// #nosec G705 -- Content-Type is explicitly set to application/x-yaml to prevent XSS
			if _, err := w.Write(yamlBytes); err != nil {
				return
			}
			return
		}

		// Default to JSON
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if err := json.NewEncoder(w).Encode(master); err != nil {
			return
		}
	}
}
