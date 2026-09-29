package state

import (
	"cmp"
	"maps"
	"reflect"
	"slices"
)

const localSource = "local"

// Collision is an entry skipped because an earlier source already defined it differently.
type Collision struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Source string `json:"source"`
	Owner  string `json:"owner"`
}

// SharedService is an HTTP service load balanced across several agents.
type SharedService struct {
	Name    string   `json:"name"`
	Agents  []string `json:"agents"`
	Servers []string `json:"servers"`
}

// merger builds a new config without mutating the sources, they are shared with published snapshots.
type merger struct {
	cfg        Config
	owners     map[string]string
	sources    map[string]Config
	collisions []Collision
	shared     map[string]*SharedService
}

func newMerger() *merger {
	return &merger{
		cfg:     Config{},
		owners:  make(map[string]string),
		sources: make(map[string]Config),
		shared:  make(map[string]*SharedService),
	}
}

// add merges src into the result. Named entries (routers, services, ...) already defined by an
// earlier source win, except HTTP services that can be shared, see shareHTTP. Lists like
// tls.certificates are appended. Offline sources only fill gaps.
func (m *merger) add(source string, src Config, online bool) {
	m.sources[source] = src
	for name, v := range src {
		sec := asMap(v)
		if sec == nil {
			continue
		}
		dst := asMap(m.cfg[name])
		if dst == nil {
			dst = make(map[string]any)
			m.cfg[name] = dst
		}

		for kind, v := range sec {
			switch v := v.(type) {
			case map[string]any:
				if name == "http" && kind == "services" && source != localSource {
					v = m.shareHTTP(source, sec, v, online)
				}
				dst[kind] = m.merge(source, name+"."+kind, asMap(dst[kind]), v)
			case []any:
				existing, _ := dst[kind].([]any)
				dst[kind] = slices.Concat(existing, v)
			default:
				if _, ok := dst[kind]; !ok {
					dst[kind] = v
				}
			}
		}
	}
}

// shareHTTP load balances services another agent already defines, when both agents route to
// them with identical routers and the services match apart from their servers. Same routers
// means both claim the same domain the same way, so it is the same app on several hosts.
// It returns the services left for the regular merge.
func (m *merger) shareHTTP(source string, http, services map[string]any, online bool) map[string]any {
	rest := maps.Clone(services)
services:
	for name, svc := range services {
		owner := m.owners["http.services/"+name]
		if owner == "" || owner == localSource {
			continue
		}
		routers := routersFor(http, name)
		ownerRouters := routersFor(m.sources[owner].section("http"), name)
		if len(routers) == 0 || !reflect.DeepEqual(routers, ownerRouters) {
			continue
		}
		for r := range ownerRouters {
			if m.owners["http.routers/"+r] != owner {
				continue services // the owner lost its router to an earlier source
			}
		}

		current := asMap(m.cfg.section("http")["services"])
		merged, ok := mergeService(asMap(current[name]), asMap(svc))
		if !ok {
			continue
		}
		// An offline agent adds nothing to a service others still serve.
		delete(rest, name)
		if !online {
			continue
		}
		current[name] = merged

		s, ok := m.shared[name]
		if !ok {
			s = &SharedService{Name: name, Agents: []string{owner}}
			m.shared[name] = s
		}
		s.Agents = append(s.Agents, source)
		s.Servers = s.Servers[:0]
		servers, _ := asMap(merged["loadBalancer"])["servers"].([]any)
		for _, srv := range servers {
			if url, ok := asMap(srv)["url"].(string); ok {
				s.Servers = append(s.Servers, url)
			}
		}
	}
	return rest
}

// merge copies src entries into dst. A name another source already owns is skipped, and
// recorded as a collision unless both definitions are identical.
func (m *merger) merge(source, kind string, dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = make(map[string]any, len(src))
	}
	for name, v := range src {
		key := kind + "/" + name
		if owner, taken := m.owners[key]; taken {
			if !reflect.DeepEqual(dst[name], v) {
				m.collisions = append(m.collisions, Collision{Kind: kind, Name: name, Source: source, Owner: owner})
			}
			continue
		}
		m.owners[key] = source
		dst[name] = v
	}
	return dst
}

// result returns the merged config with empty entries and sections dropped.
func (m *merger) result() Config {
	for name, v := range m.cfg {
		sec := asMap(v)
		maps.DeleteFunc(sec, func(_ string, v any) bool { return isEmpty(v) })
		if len(sec) == 0 {
			delete(m.cfg, name)
		}
	}
	return m.cfg
}

func (m *merger) sharedServices() []SharedService {
	out := make([]SharedService, 0, len(m.shared))
	for _, s := range m.shared {
		out = append(out, *s)
	}
	slices.SortFunc(out, func(a, b SharedService) int { return cmp.Compare(a.Name, b.Name) })
	return out
}

func isEmpty(v any) bool {
	switch v := v.(type) {
	case map[string]any:
		return len(v) == 0
	case []any:
		return len(v) == 0
	}
	return v == nil
}

func routersFor(http map[string]any, service string) map[string]any {
	out := make(map[string]any)
	for name, r := range asMap(http["routers"]) {
		if asMap(r)["service"] == service {
			out[name] = r
		}
	}
	return out
}

// mergeService returns a new service with the servers of both, or false if they differ in
// anything but their servers.
func mergeService(a, b map[string]any) (map[string]any, bool) {
	lbA, lbB := asMap(a["loadBalancer"]), asMap(b["loadBalancer"])
	if lbA == nil || lbB == nil {
		return nil, false
	}
	if !reflect.DeepEqual(without(a, "loadBalancer"), without(b, "loadBalancer")) ||
		!reflect.DeepEqual(without(lbA, "servers"), without(lbB, "servers")) {
		return nil, false
	}

	serversA, _ := lbA["servers"].([]any)
	serversB, _ := lbB["servers"].([]any)
	servers := slices.Clone(serversA)
	for _, srv := range serversB {
		if !slices.ContainsFunc(servers, func(s any) bool { return asMap(s)["url"] == asMap(srv)["url"] }) {
			servers = append(servers, srv)
		}
	}

	lb := maps.Clone(lbA)
	lb["servers"] = servers
	out := maps.Clone(a)
	out["loadBalancer"] = lb
	return out, true
}

func without(m map[string]any, key string) map[string]any {
	c := maps.Clone(m)
	delete(c, key)
	return c
}
