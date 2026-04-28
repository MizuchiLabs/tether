// Package state contains the traefik configuration by environment
package state

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"go.yaml.in/yaml/v3"
)

type Environment struct {
	Master *dynamic.Configuration
	Agents map[string]*dynamic.Configuration
	Local  *dynamic.Configuration
}

type State struct {
	mu          sync.RWMutex
	Envs        map[string]*Environment // Map of group names to their environments
	subscribers map[string][]chan *dynamic.Configuration
}

func New() *State {
	return &State{
		Envs:        make(map[string]*Environment),
		subscribers: make(map[string][]chan *dynamic.Configuration),
	}
}

// getEnv safely retrieves an environment or creates it if it doesn't exist.
func (s *State) getEnv(env string) *Environment {
	// Default to a common pool if no group is specified
	if env == "" {
		env = "default"
	}

	if e, exists := s.Envs[env]; exists {
		return e
	}

	newEnv := &Environment{
		Master: &dynamic.Configuration{},
		Agents: make(map[string]*dynamic.Configuration),
	}
	s.Envs[env] = newEnv
	return newEnv
}

// GetEnvNames returns a list of all environment names.
func (s *State) GetEnvNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.Envs))
	for name := range s.Envs {
		names = append(names, name)
	}
	return names
}

// GetMaster safely returns a snapshot of the master configuration for the given environment.
func (s *State) GetMaster(env string) *dynamic.Configuration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if env == "" {
		env = "default"
	}
	if e, exists := s.Envs[env]; exists {
		return e.Master
	}
	return &dynamic.Configuration{}
}

// UpdateAgent completely replaces the state for a specific agent and rebuilds the Master.
func (s *State) UpdateAgent(env, name string, data []byte) {
	cfg := &dynamic.Configuration{}
	if err := json.Unmarshal(data, cfg); err != nil {
		slog.Error("Failed to unmarshal agent config", "agent", name, "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	envs := s.getEnv(env)
	envs.Agents[name] = cfg
	s.rebuildMaster(env)
}

// LoadLocalFile reads a JSON or YAML dynamic config and stores it
func (s *State) LoadLocalFile(env, path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Clean(path)); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to read local config: %w", err)
	}
	slog.Info("Loading local configuration file", "path", path)

	cfg := &dynamic.Configuration{}
	ext := filepath.Ext(path)

	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse yaml: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse json: %w", err)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	envs := s.getEnv(env)
	envs.Local = cfg
	s.rebuildMaster(env)
	return nil
}

// rebuildMaster loops through all known agents and merges them into a brand new Master.
func (s *State) rebuildMaster(env string) {
	envs := s.Envs[env]
	newMaster := &dynamic.Configuration{}
	if envs.Local != nil {
		newMaster.HTTP = mergeHTTP(newMaster.HTTP, envs.Local.HTTP)
		newMaster.TCP = mergeTCP(newMaster.TCP, envs.Local.TCP)
		newMaster.UDP = mergeUDP(newMaster.UDP, envs.Local.UDP)
		newMaster.TLS = mergeTLS(newMaster.TLS, envs.Local.TLS)
	}

	for _, agentCfg := range envs.Agents {
		newMaster.HTTP = mergeHTTP(newMaster.HTTP, agentCfg.HTTP)
		newMaster.TCP = mergeTCP(newMaster.TCP, agentCfg.TCP)
		newMaster.UDP = mergeUDP(newMaster.UDP, agentCfg.UDP)
		newMaster.TLS = mergeTLS(newMaster.TLS, agentCfg.TLS)
	}
	envs.Master = newMaster

	// Broadcast to listeners
	if subs, exists := s.subscribers[env]; exists {
		for _, ch := range subs {
			// Non-blocking send: if a client is slow/hung, we drop the event
			select {
			case ch <- newMaster:
			default:
			}
		}
	}
}

// Subscribe creates a channel that receives updates for a specific environment.
func (s *State) Subscribe(env string) chan *dynamic.Configuration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if env == "" {
		env = "default"
	}

	// Use a buffered channel (size 1) so broadcasting doesn't block
	ch := make(chan *dynamic.Configuration, 1)
	s.subscribers[env] = append(s.subscribers[env], ch)
	return ch
}

// Unsubscribe removes a channel from the pool and closes it.
func (s *State) Unsubscribe(env string, ch chan *dynamic.Configuration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if env == "" {
		env = "default"
	}

	subs := s.subscribers[env]
	for i, sub := range subs {
		if sub == ch {
			// Remove the channel from the slice
			s.subscribers[env] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}
