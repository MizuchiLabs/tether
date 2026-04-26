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
	mu   sync.RWMutex
	Envs map[string]*Environment // Map of group names to their environments
}

func New() *State {
	return &State{Envs: make(map[string]*Environment)}
}

// GetEnv safely retrieves an environment or creates it if it doesn't exist
func (s *State) GetEnv(env string) *Environment {
	// Default to a common pool if no group is specified
	if env == "" {
		env = "default"
	}

	if env, exists := s.Envs[env]; exists {
		return env
	}

	newEnv := &Environment{
		Master: &dynamic.Configuration{},
		Agents: make(map[string]*dynamic.Configuration),
	}
	s.Envs[env] = newEnv
	return newEnv
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

	envs := s.GetEnv(env)
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

	envs := s.GetEnv(env)
	envs.Local = cfg
	return nil
}

// rebuildMaster loops through all known agents and merges them into a brand new Master.
// Note: Must be called while holding the write lock (s.mu.Lock()).
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
}
