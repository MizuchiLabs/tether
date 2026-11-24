package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mizuchilabs/nushi"

	"github.com/mizuchilabs/tether/internal/state"
)

func TestRoutesNeedSignIn(t *testing.T) {
	t.Parallel()
	const token = "0123456789abcdef0123456789abcdef"
	st, err := state.New(t.Context(), "")
	require.NoError(t, err)
	guard, err := nushi.New("tether", nushi.Config{Password: "correct horse battery", APITokens: []string{token}})
	require.NoError(t, err)
	h := routes(t.Context(), st, guard, true)

	tests := []struct {
		name string
		path string
		auth string
		want int
	}{
		{"health is open", "/healthz", "", http.StatusOK},
		{"session is open", "/api/session", "", http.StatusOK},
		{"config without a token", "/config", "", http.StatusUnauthorized},
		{"config with the token reaches the warmup", "/config", "Bearer " + token, http.StatusServiceUnavailable},
		{"envs with a wrong token", "/api/envs", "Bearer nope", http.StatusUnauthorized},
		{"envs with the token", "/api/envs", "Bearer " + token, http.StatusOK},
		{"events without a token", "/api/events", "", http.StatusUnauthorized},
		{"agent socket without a token", "/api/ws", "", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.auth != "" {
				r.Header.Set("Authorization", tt.auth)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestPublishConfigWaitsForWarmup(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		st, err := state.New(t.Context(), "")
		require.NoError(t, err)
		h := publishConfig(st)

		w := httptest.NewRecorder()
		h(w, httptest.NewRequest(http.MethodGet, "/config", nil))
		assert.Equal(t, http.StatusServiceUnavailable, w.Code, "traefik should keep its cached config during warmup")

		time.Sleep(warmup)
		w = httptest.NewRecorder()
		h(w, httptest.NewRequest(http.MethodGet, "/config", nil))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, "{}", w.Body.String())
	})
}
