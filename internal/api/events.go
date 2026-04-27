package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mizuchilabs/tether/internal/state"
)

func EventStream(state *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		rc := http.NewResponseController(w)
		env := r.URL.Query().Get("env")

		updateCh := state.Subscribe(env)
		defer state.Unsubscribe(env, updateCh)

		for {
			select {
			case <-r.Context().Done():
				return
			case newConfig := <-updateCh:
				data, _ := json.Marshal(newConfig)
				_, _ = fmt.Fprintf(w, "data: %s\n\n", data)

				err := rc.Flush()
				if err != nil {
					// Client disconnected or connection dropped
					return
				}
			}
		}
	}
}
