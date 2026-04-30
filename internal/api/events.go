package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mizuchilabs/tether/internal/state"
)

func EventStream(ctx context.Context, state *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		rc := http.NewResponseController(w)

		// Clear the global write deadline for this long-lived connection
		if err := rc.SetWriteDeadline(time.Time{}); err != nil {
			http.Error(w, "Failed to configure SSE connection", http.StatusInternalServerError)
			return
		}

		env := r.URL.Query().Get("env")

		updateCh := state.Subscribe(env)
		defer state.Unsubscribe(env, updateCh)

		ping := time.NewTicker(15 * time.Second)
		defer ping.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ctx.Done():
				return
			case <-ping.C: // Send a keep-alive ping
				_, _ = fmt.Fprintf(w, ": ping\n\n")
				_ = rc.Flush()
			case newConfig := <-updateCh:
				data, _ := json.Marshal(newConfig)
				_, _ = fmt.Fprintf(w, "data: %s\n\n", data)

				if err := rc.Flush(); err != nil {
					// Client disconnected or connection dropped
					return
				}
			}
		}
	}
}
