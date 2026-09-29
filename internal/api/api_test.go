package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mizuchilabs/tether/internal/state"
)

func TestWithAuth(t *testing.T) {
	t.Parallel()
	ok := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	tests := []struct {
		name   string
		token  string
		header string
		cookie string
		want   int
	}{
		{"auth disabled", "", "", "", http.StatusOK},
		{"bearer", "secret", "Bearer secret", "", http.StatusOK},
		{"bearer lowercase scheme", "secret", "bearer secret", "", http.StatusOK},
		{"cookie", "secret", "", "secret", http.StatusOK},
		{"wrong bearer", "secret", "Bearer nope", "", http.StatusUnauthorized},
		{"missing", "secret", "", "", http.StatusUnauthorized},
		{"basic scheme", "secret", "Basic secret", "", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, "/config", nil)
			if tt.header != "" {
				r.Header.Set("Authorization", tt.header)
			}
			if tt.cookie != "" {
				r.AddCookie(&http.Cookie{Name: accessCookie, Value: tt.cookie})
			}
			w := httptest.NewRecorder()
			withAuth(tt.token)(ok).ServeHTTP(w, r)
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
