package api

import (
	"net/http"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/mizuchilabs/tether/internal/state"
)

// warmup gives agents time to reconnect after a restart. Until then Traefik gets a 503
// and keeps its last known config instead of dropping every agent route.
const warmup = 15 * time.Second

// publishConfig returns the merged traefik config as JSON or YAML.
func publishConfig(st *state.State) http.HandlerFunc {
	readyAt := time.Now().Add(warmup)
	return func(w http.ResponseWriter, r *http.Request) {
		if time.Now().Before(readyAt) {
			w.Header().Set("Retry-After", "5")
			http.Error(w, "Waiting for agents to reconnect", http.StatusServiceUnavailable)
			return
		}

		cfg := st.Snapshot(r.URL.Query().Get("env")).Config
		format := r.URL.Query().Get("format")
		if format == "yaml" || (format == "" && strings.Contains(r.Header.Get("Accept"), "yaml")) {
			out, err := yaml.Marshal(cfg)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/x-yaml")
			// #nosec G705 -- Content-Type is explicitly set to application/x-yaml to prevent XSS
			_, _ = w.Write(out)
			return
		}
		writeJSON(w, cfg)
	}
}
