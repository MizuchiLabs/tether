package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mizuchilabs/tether/internal/state"
	"sigs.k8s.io/yaml"
)

func PublishConfig(state *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")
		accept := r.Header.Get("Accept")
		envs := state.GetEnv(r.URL.Query().Get("env"))

		// Determine response format: prefer query param over header
		if format == "yaml" || (format == "" && strings.Contains(accept, "yaml")) {
			w.Header().Set("Content-Type", "application/x-yaml")
			yamlBytes, err := yaml.Marshal(envs.Master)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := w.Write(yamlBytes); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		// Default to JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(envs.Master); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
