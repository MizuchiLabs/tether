package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parse(t *testing.T, s string) Config {
	t.Helper()
	var cfg Config
	require.NoError(t, json.Unmarshal([]byte(s), &cfg), "invalid test config")
	return cfg
}

// httpConfig has a router per name, routing Host(`name`) to service "app".
func httpConfig(t *testing.T, routers ...string) Config {
	t.Helper()
	entries := make([]string, 0, len(routers))
	for _, r := range routers {
		entries = append(entries, fmt.Sprintf("%q: {\"rule\": \"Host(`%s`)\", \"service\": \"app\"}", r, r))
	}
	return parse(t, `{"http": {"routers": {`+strings.Join(entries, ",")+`}}}`)
}

func routers(s *State) map[string]any {
	return asMap(s.Snapshot("").Config.section("http")["routers"])
}

func newState(t *testing.T) *State {
	t.Helper()
	s, err := New(t.Context(), "")
	require.NoError(t, err)
	return s
}

func connect(s *State, env, name string, cfg Config) {
	s.AgentConnected(env, name, "10.0.0.1")
	s.UpdateAgent(env, name, cfg)
}

func TestMergeFirstSourceWinsDeterministically(t *testing.T) {
	t.Parallel()
	s := newState(t)
	s.setLocal("", httpConfig(t, "shared-local"))
	connect(s, "", "b", parse(t, `{"http": {"routers": {
		"shared": {"rule": "Host(`+"`b`"+`)"},
		"only-b": {"rule": "Host(`+"`only-b`"+`)"}
	}}}`))
	connect(s, "", "a", parse(t, `{"http": {"routers": {
		"shared": {"rule": "Host(`+"`a`"+`)"},
		"only-a": {"rule": "Host(`+"`only-a`"+`)"},
		"shared-local": {"rule": "Host(`+"`a`"+`)"}
	}}}`))

	assert.Len(t, routers(s), 4)
	assert.Equal(t, map[string]any{"rule": "Host(`a`)"}, routers(s)["shared"], "agent a sorts first")
	assert.Equal(t, []Collision{
		{Kind: "http.routers", Name: "shared", Source: "b", Owner: "a"},
		{Kind: "http.routers", Name: "shared-local", Source: "a", Owner: localSource},
	}, s.Snapshot("").Collisions)
}

func TestIdenticalEntriesAreNotCollisions(t *testing.T) {
	t.Parallel()
	s := newState(t)
	connect(s, "", "a", httpConfig(t, "app"))
	connect(s, "", "b", httpConfig(t, "app"))
	assert.Empty(t, s.Snapshot("").Collisions)
}

func TestUnknownFieldsPassThrough(t *testing.T) {
	t.Parallel()
	s := newState(t)
	cfg := `{
		"http": {
			"routers": {"app": {"rule": "Host(` + "`app`" + `)", "someNewOption": {"x": 1}}},
			"futureKind": {"thing": {"enabled": true}}
		},
		"tls": {"certificates": [{"certFile": "/a.crt"}]},
		"newSection": {"entries": {"e": {}}}
	}`
	connect(s, "", "a", parse(t, cfg))

	out, err := json.Marshal(s.Snapshot("").Config)
	require.NoError(t, err)
	assert.JSONEq(t, cfg, string(out), "options tether doesn't know must reach traefik unchanged")
}

func TestListsAreAppended(t *testing.T) {
	t.Parallel()
	s := newState(t)
	connect(s, "", "a", parse(t, `{"tls": {"certificates": [{"certFile": "/a.crt"}]}}`))
	connect(s, "", "b", parse(t, `{"tls": {"certificates": [{"certFile": "/b.crt"}]}}`))

	certs := s.Snapshot("").Config.section("tls")["certificates"]
	assert.Len(t, certs, 2)
}

func TestMergeDropsEmptySections(t *testing.T) {
	t.Parallel()
	s := newState(t)
	connect(s, "", "a", parse(t, `{"http": {"routers": {}}, "tls": {}}`))
	assert.Empty(t, s.Snapshot("").Config)
}

func TestEnvironmentsAreIsolated(t *testing.T) {
	t.Parallel()
	s := newState(t)
	connect(s, "", "a", httpConfig(t, "one"))
	connect(s, "prod", "b", httpConfig(t, "two"))

	assert.Equal(t, []string{"default", "prod"}, s.EnvNames())
	assert.Contains(t, routers(s), "one")
	assert.NotContains(t, routers(s), "two")
	assert.Contains(t, asMap(s.Snapshot("prod").Config.section("http")["routers"]), "two")
}

func TestAgentCounts(t *testing.T) {
	t.Parallel()
	s := newState(t)
	connect(s, "", "a", parse(t, `{
		"http": {"routers": {"a": {}, "b": {}}, "services": {"a": {}}},
		"tcp": {"routers": {"c": {}}, "services": {"c": {}}}
	}`))
	agent := s.Snapshot("").Agents[0]
	assert.Equal(t, 3, agent.Routers)
	assert.Equal(t, 2, agent.Services)
}

func TestDisconnectedAgentExpires(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		s := newState(t)
		connect(s, "", "a", httpConfig(t, "app"))
		s.AgentDisconnected("", "a")

		snap := s.Snapshot("")
		require.Len(t, snap.Agents, 1)
		assert.False(t, snap.Agents[0].Connected, "agent should be marked disconnected")
		assert.Contains(t, routers(s), "app", "routes should survive the grace period")

		time.Sleep(expireAfter + time.Second)
		synctest.Wait()
		assert.Empty(t, s.Snapshot("").Agents)
		assert.Empty(t, s.Snapshot("").Config)
		assert.Empty(t, s.EnvNames(), "empty env should be removed")
	})
}

func TestReconnectCancelsExpiry(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		s := newState(t)
		connect(s, "", "a", httpConfig(t, "app"))
		s.AgentDisconnected("", "a")
		time.Sleep(expireAfter / 2)
		connect(s, "", "a", httpConfig(t, "app"))

		time.Sleep(expireAfter)
		synctest.Wait()
		assert.Contains(t, routers(s), "app")
	})
}

func TestSecondConnectionKeepsAgentAlive(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		s := newState(t)
		connect(s, "", "a", httpConfig(t, "app"))
		connect(s, "", "a", httpConfig(t, "app"))
		s.AgentDisconnected("", "a")

		time.Sleep(expireAfter + time.Second)
		synctest.Wait()
		snap := s.Snapshot("")
		require.Len(t, snap.Agents, 1)
		assert.True(t, snap.Agents[0].Connected)
	})
}

func TestSubscribeGetsLatestSnapshot(t *testing.T) {
	t.Parallel()
	s := newState(t)
	ch := s.Subscribe("")
	connect(s, "", "a", httpConfig(t, "one"))
	connect(s, "", "a", httpConfig(t, "two"))

	snap := <-ch
	assert.Contains(t, asMap(snap.Config.section("http")["routers"]), "two",
		"slow subscriber should only see the latest snapshot")
	s.Unsubscribe("", ch)
}

func TestLocalFileReloads(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "dynamic.yml")
	write := func(router string) {
		body := "http:\n  routers:\n    " + router + ":\n      rule: Host(`x`)\n      service: x\n      priority: 10\n"
		// Write and rename like editors do.
		tmp := path + ".tmp"
		require.NoError(t, os.WriteFile(tmp, []byte(body), 0o600))
		require.NoError(t, os.Rename(tmp, path))
	}

	s, err := New(t.Context(), path)
	require.NoError(t, err, "missing file should not fail startup")
	assert.Empty(t, s.Snapshot("").Config)

	write("first")
	require.Eventually(t, func() bool { return routers(s)["first"] != nil }, 2*time.Second, 10*time.Millisecond,
		"file created after startup should load")
	assert.Equal(t, map[string]any{"rule": "Host(`x`)", "service": "x", "priority": float64(10)},
		routers(s)["first"], "yaml values should match what agents send as JSON")

	write("second")
	require.Eventually(t, func() bool { return routers(s)["second"] != nil }, 2*time.Second, 10*time.Millisecond,
		"watch should survive an atomic replace")
}
