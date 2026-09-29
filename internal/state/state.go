// Package state manages traefik dynamic configurations per environment,
// merging local files and agent submissions into a single master config.
package state

import (
	"cmp"
	"context"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"
)

const (
	defaultEnv = "default"

	// expireAfter is how long a disconnected agent keeps its routes. Covers
	// agent restarts and short network blips without dropping traffic.
	expireAfter = 30 * time.Second
)

// Agent is a tetherd instance pushing its container config.
type Agent struct {
	Name      string    `json:"name"`
	Addr      string    `json:"addr"`
	Connected bool      `json:"connected"`
	Since     time.Time `json:"since"`
	Updated   time.Time `json:"updated"`
	Routers   int       `json:"routers"`
	Services  int       `json:"services"`

	conns  int
	config Config
}

// Snapshot is the merged view of one environment. Never mutated after creation.
type Snapshot struct {
	Config     Config          `json:"config"`
	Agents     []Agent         `json:"agents"`
	Collisions []Collision     `json:"collisions"`
	Shared     []SharedService `json:"shared"`
}

type environment struct {
	local    Config
	agents   map[string]*Agent
	snapshot *Snapshot
}

// State holds traefik configurations grouped by environment.
type State struct {
	mu          sync.RWMutex
	envs        map[string]*environment
	subscribers map[string][]chan *Snapshot
}

// New creates the state and, if localFile is set, loads and watches it until ctx is done.
func New(ctx context.Context, localFile string) (*State, error) {
	s := &State{
		envs:        make(map[string]*environment),
		subscribers: make(map[string][]chan *Snapshot),
	}
	if localFile != "" {
		if err := s.watchLocalFile(ctx, localFile); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// EnvNames returns all registered environment names, sorted.
func (s *State) EnvNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Sorted(maps.Keys(s.envs))
}

// Snapshot returns the current merged view of the environment.
func (s *State) Snapshot(env string) *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.envs[envName(env)]; ok {
		return e.snapshot
	}
	return &Snapshot{Config: Config{}}
}

// AgentConnected registers a new connection for the agent.
func (s *State) AgentConnected(env, name, addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.env(env).agents[name]
	if !ok {
		a = &Agent{Name: name}
		s.env(env).agents[name] = a
	}
	if a.conns > 0 {
		slog.Warn("Multiple agents share the same name", "agent", name, "env", envName(env), "addr", addr)
	}
	a.conns++
	a.Connected = true
	a.Addr = addr
	a.Since = time.Now()
	slog.Info("Agent connected", "agent", name, "env", envName(env), "addr", addr)
}

// UpdateAgent replaces an agent's config and rebuilds the environment.
func (s *State) UpdateAgent(env, name string, cfg Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.env(env).agents[name]
	if !ok {
		return
	}
	a.config = cfg
	a.Updated = time.Now()
	a.Routers, a.Services = cfg.count()
	s.rebuild(env)
}

// AgentDisconnected drops a connection. Once the agent has no connections left
// its routes are kept for expireAfter, then removed unless it reconnects.
func (s *State) AgentDisconnected(env, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.env(env).agents[name]
	if !ok {
		return
	}
	a.conns--
	if a.conns > 0 {
		return
	}
	a.Connected = false
	a.Since = time.Now()
	since := a.Since
	slog.Info("Agent disconnected", "agent", name, "env", envName(env), "expires_in", expireAfter.String())
	s.rebuild(env)

	time.AfterFunc(expireAfter, func() { s.expireAgent(env, name, since) })
}

// Subscribe returns a channel that receives snapshots for the environment.
func (s *State) Subscribe(env string) chan *Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *Snapshot, 1)
	s.subscribers[envName(env)] = append(s.subscribers[envName(env)], ch)
	return ch
}

// Unsubscribe removes a subscription channel.
func (s *State) Unsubscribe(env string, ch chan *Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribers[envName(env)] = slices.DeleteFunc(s.subscribers[envName(env)], func(c chan *Snapshot) bool {
		return c == ch
	})
}

func (s *State) expireAgent(env, name string, since time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.envs[envName(env)]
	if !ok {
		return
	}
	a, ok := e.agents[name]
	if !ok || a.Connected || !a.Since.Equal(since) {
		return
	}
	delete(e.agents, name)
	slog.Warn("Agent expired, removing its routes", "agent", name, "env", envName(env))
	s.rebuild(env)
	if len(e.agents) == 0 && e.local == nil {
		delete(s.envs, envName(env))
	}
}

func (s *State) setLocal(env string, cfg Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.env(env).local = cfg
	s.rebuild(env)
}

// env returns the environment, creating it if needed. Caller holds the lock.
func (s *State) env(name string) *environment {
	name = envName(name)
	if e, ok := s.envs[name]; ok {
		return e
	}
	e := &environment{
		agents:   make(map[string]*Agent),
		snapshot: &Snapshot{Config: Config{}},
	}
	s.envs[name] = e
	return e
}

// rebuild merges local and agent configs into a fresh snapshot, then broadcasts. Caller holds the lock.
func (s *State) rebuild(env string) {
	e := s.env(env)
	prev := e.snapshot

	m := newMerger()
	m.add(localSource, e.local, true)
	names := slices.Sorted(maps.Keys(e.agents))
	agents := make([]Agent, 0, len(names))
	for _, name := range names {
		agents = append(agents, *e.agents[name])
	}
	// Online agents first, so a shared service never depends on an offline one.
	for _, online := range []bool{true, false} {
		for _, name := range names {
			if a := e.agents[name]; a.Connected == online {
				m.add(name, a.config, online)
			}
		}
	}

	slices.SortFunc(m.collisions, func(a, b Collision) int {
		return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.Name, b.Name), cmp.Compare(a.Source, b.Source))
	})
	for _, c := range m.collisions {
		if !slices.Contains(prev.Collisions, c) {
			slog.Warn("Collision detected, skipping", "env", envName(env), "kind", c.Kind, "name", c.Name,
				"agent", c.Source, "owner", c.Owner)
		}
	}

	shared := m.sharedServices()
	for _, sh := range shared {
		if !slices.ContainsFunc(prev.Shared, func(p SharedService) bool {
			return p.Name == sh.Name && slices.Equal(p.Agents, sh.Agents)
		}) {
			slog.Info("Load balancing service across agents", "env", envName(env), "service", sh.Name,
				"agents", sh.Agents)
		}
	}

	e.snapshot = &Snapshot{Config: m.result(), Agents: agents, Collisions: m.collisions, Shared: shared}
	for _, ch := range s.subscribers[envName(env)] {
		// Drop the stale snapshot so slow clients always get the latest.
		select {
		case <-ch:
		default:
		}
		ch <- e.snapshot
	}
}

func envName(env string) string {
	if env == "" {
		return defaultEnv
	}
	return env
}
