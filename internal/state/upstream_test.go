package state

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamActsLikeAgent(t *testing.T) {
	t.Parallel()
	var body atomic.Pointer[string]
	body.Store(new(`{"http": {"routers": {"first": {"rule": "Host(` + "`x`" + `)"}}}}`))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		b := body.Load()
		if b == nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(*b))
	}))
	t.Cleanup(srv.Close)

	s := newState(t)
	s.watchUpstream(t.Context(), "pangolin", srv.URL, 10*time.Millisecond)
	require.Eventually(t, func() bool { return routers(s)["first"] != nil }, 2*time.Second, 10*time.Millisecond,
		"first poll should load the upstream config")
	snap := s.Snapshot("")
	require.Len(t, snap.Agents, 1)
	assert.Equal(t, "pangolin", snap.Agents[0].Name)

	body.Store(new(`{"http": {"routers": {"second": {"rule": "Host(` + "`x`" + `)"}}}}`))
	require.Eventually(t, func() bool { return routers(s)["second"] != nil }, 2*time.Second, 10*time.Millisecond,
		"changed upstream config should replace the old one")

	body.Store(nil)
	require.Eventually(t, func() bool {
		agents := s.Snapshot("").Agents
		return len(agents) == 1 && !agents[0].Connected
	}, 2*time.Second, 10*time.Millisecond, "failing upstream should disconnect")
	assert.NotNil(t, routers(s)["second"], "routes should survive until the agent expires")

	body.Store(new(`{"http": {"routers": {"second": {"rule": "Host(` + "`x`" + `)"}}}}`))
	require.Eventually(t, func() bool {
		agents := s.Snapshot("").Agents
		return len(agents) == 1 && agents[0].Connected
	}, 2*time.Second, 10*time.Millisecond, "recovered upstream should reconnect even with the same config")
}
