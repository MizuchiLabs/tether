package state

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// webApp is what tetherd sends for the same labels on different hosts, only the server differs.
func webApp(t *testing.T, serverURL string) Config {
	t.Helper()
	return parse(t, `{"http": {
		"routers": {"web": {"rule": "Host(`+"`web.lan`"+`)", "service": "web", "entryPoints": ["websecure"]}},
		"services": {"web": {"loadBalancer": {"passHostHeader": true, "servers": [{"url": "`+serverURL+`"}]}}}
	}}`)
}

func webServers(t *testing.T, s *State) []string {
	t.Helper()
	svc := asMap(asMap(s.Snapshot("").Config.section("http")["services"])["web"])
	require.NotNil(t, svc, "service web missing")
	servers, _ := asMap(svc["loadBalancer"])["servers"].([]any)
	var urls []string
	for _, srv := range servers {
		urls = append(urls, asMap(srv)["url"].(string))
	}
	return urls
}

func TestSameAppOnSeveralAgentsIsLoadBalanced(t *testing.T) {
	t.Parallel()
	s := newState(t)
	a, b := webApp(t, "http://10.0.0.1:80"), webApp(t, "http://10.0.0.2:80")
	connect(s, "", "a", a)
	connect(s, "", "b", b)
	connect(s, "", "c", webApp(t, "http://10.0.0.3:80"))

	assert.Equal(t, []string{"http://10.0.0.1:80", "http://10.0.0.2:80", "http://10.0.0.3:80"}, webServers(t, s))
	snap := s.Snapshot("")
	assert.Empty(t, snap.Collisions)
	require.Len(t, snap.Shared, 1)
	assert.Equal(t, []string{"a", "b", "c"}, snap.Shared[0].Agents)

	// Rebuilding must not pile servers onto the agents' own configs.
	connect(s, "", "b", b)
	assert.Len(t, webServers(t, s), 3)
	assert.Equal(t, webApp(t, "http://10.0.0.1:80"), a, "agent config must not be mutated")
}

func TestSameServiceNameWithDifferentRoutersIsNotShared(t *testing.T) {
	t.Parallel()
	s := newState(t)
	other := webApp(t, "http://10.0.0.2:80")
	asMap(asMap(other.section("http")["routers"])["web"])["rule"] = "Host(`other.lan`)"
	connect(s, "", "a", webApp(t, "http://10.0.0.1:80"))
	connect(s, "", "b", other)

	assert.Equal(t, []string{"http://10.0.0.1:80"}, webServers(t, s), "unrelated apps must not be merged")
	snap := s.Snapshot("")
	assert.Empty(t, snap.Shared)
	assert.Len(t, snap.Collisions, 2, "router and service collide")
}

func TestServicesWithDifferentSettingsAreNotShared(t *testing.T) {
	t.Parallel()
	s := newState(t)
	sticky := webApp(t, "http://10.0.0.2:80")
	lb := asMap(asMap(asMap(sticky.section("http")["services"])["web"])["loadBalancer"])
	lb["sticky"] = map[string]any{"cookie": map[string]any{"name": "s"}}
	connect(s, "", "a", webApp(t, "http://10.0.0.1:80"))
	connect(s, "", "b", sticky)

	assert.Equal(t, []string{"http://10.0.0.1:80"}, webServers(t, s))
	assert.Equal(t, []Collision{{Kind: "http.services", Name: "web", Source: "b", Owner: "a"}},
		s.Snapshot("").Collisions)
}

func TestTCPIsNeverShared(t *testing.T) {
	t.Parallel()
	s := newState(t)
	db := func(addr string) Config {
		return parse(t, `{"tcp": {
			"routers": {"db": {"rule": "HostSNI(`+"`*`"+`)", "service": "db"}},
			"services": {"db": {"loadBalancer": {"servers": [{"address": "`+addr+`"}]}}}
		}}`)
	}
	connect(s, "", "a", db("10.0.0.1:5432"))
	connect(s, "", "b", db("10.0.0.2:5432"))

	svc := asMap(asMap(s.Snapshot("").Config.section("tcp")["services"])["db"])
	assert.Equal(t, []any{map[string]any{"address": "10.0.0.1:5432"}}, asMap(svc["loadBalancer"])["servers"])
}

func TestOfflineAgentLeavesSharedServiceImmediately(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		s := newState(t)
		connect(s, "", "a", webApp(t, "http://10.0.0.1:80"))
		connect(s, "", "b", webApp(t, "http://10.0.0.2:80"))

		s.AgentDisconnected("", "a")
		assert.Equal(t, []string{"http://10.0.0.2:80"}, webServers(t, s), "traffic must not go to the offline owner")
		assert.Contains(t, routers(s), "web")

		connect(s, "", "a", webApp(t, "http://10.0.0.1:80"))
		assert.Len(t, webServers(t, s), 2, "reconnected agent should rejoin")
	})
}

func TestOfflineAgentKeepsItsOwnRoutesDuringGrace(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		s := newState(t)
		connect(s, "", "a", webApp(t, "http://10.0.0.1:80"))
		s.AgentDisconnected("", "a")
		assert.Equal(t, []string{"http://10.0.0.1:80"}, webServers(t, s), "nobody else serves it, keep it until expiry")

		time.Sleep(expireAfter + time.Second)
		synctest.Wait()
		assert.Empty(t, s.Snapshot("").Config)
	})
}
