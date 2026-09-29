package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mizuchilabs/tether/internal/state"
)

// eventStream pushes environment snapshots to the UI over SSE, starting with the current one.
func eventStream(ctx context.Context, st *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		if err := rc.SetWriteDeadline(time.Time{}); err != nil {
			http.Error(w, "Failed to configure SSE connection", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		env := r.URL.Query().Get("env")
		updates := st.Subscribe(env)
		defer st.Unsubscribe(env, updates)

		send := func(s *state.Snapshot) error {
			data, err := json.Marshal(s)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return err
			}
			return rc.Flush()
		}
		if err := send(st.Snapshot(env)); err != nil {
			return
		}

		ping := time.NewTicker(15 * time.Second)
		defer ping.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ctx.Done():
				return
			case <-ping.C:
				_, _ = fmt.Fprint(w, ": ping\n\n")
				_ = rc.Flush()
			case s := <-updates:
				if err := send(s); err != nil {
					return
				}
			}
		}
	}
}
