// Package state manages traefik dynamic configurations per environment,
// merging local files and agent submissions into a single master config.
package state

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"go.yaml.in/yaml/v3"
)

type Environment struct {
	Master *dynamic.Configuration
	Agents map[string]*dynamic.Configuration
	Local  *dynamic.Configuration
}

// State holds traefik configurations grouped by environment.
type State struct {
	mu          sync.RWMutex
	Envs        map[string]*Environment
	subscribers map[string][]chan *dynamic.Configuration
	watchers    []*FileWatcher
}

// FileWatcher tracks a single config file for live reloads.
type FileWatcher struct {
	file    string
	env     string
	watcher *fsnotify.Watcher
}

func New() *State {
	return &State{
		Envs:        make(map[string]*Environment),
		subscribers: make(map[string][]chan *dynamic.Configuration),
		watchers:    make([]*FileWatcher, 0),
	}
}

// getEnv retrieves or creates an environment entry.
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

// GetEnvNames returns all registered environment names.
func (s *State) GetEnvNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.Envs))
	for name := range s.Envs {
		names = append(names, name)
	}
	return names
}

// GetMaster returns the merged configuration for the given environment.
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

// UpdateAgent replaces an agent's config and rebuilds the merged master.
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

// parseFile reads a .yaml, .yml, or .json config file.
func parseFile(path string) (*dynamic.Configuration, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read local config: %w", err)
	}

	cfg := &dynamic.Configuration{}
	ext := filepath.Ext(path)

	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse yaml: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse json: %w", err)
		}
	}
	return cfg, nil
}

// LoadLocalFile reads a config file, stores it, and watches for live changes.
func (s *State) LoadLocalFile(ctx context.Context, env, path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	slog.Info("Loading local configuration file", "path", path)

	cfg, err := parseFile(path)
	if err != nil {
		return err
	}

	s.mu.Lock()
	envs := s.getEnv(env)
	envs.Local = cfg
	s.rebuildMaster(env)
	s.mu.Unlock()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	fw := &FileWatcher{
		file:    path,
		env:     env,
		watcher: watcher,
	}

	s.watchers = append(s.watchers, fw)

	go func() {
		if err := watcher.Add(path); err != nil {
			slog.Error("Failed to watch config file", "path", path, "error", err)
			return
		}

		defer func() { _ = watcher.Close() }()

		for {
			select {
			case <-ctx.Done():
				slog.Debug("Stopping config file watcher", "path", path)
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Chmod) {
					slog.Debug("Config file changed, reloading", "path", path)
					cfg, err := parseFile(path)
					if err != nil {
						slog.Error("Failed to reload config", "path", path, "error", err)
						continue
					}

					s.mu.Lock()
					envs.Local = cfg
					s.rebuildMaster(env)
					s.mu.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("Config watcher error", "path", path, "error", err)
			}
		}
	}()

	return nil
}

// rebuildMaster merges local and agent configs into a fresh master, then broadcasts.
func (s *State) rebuildMaster(env string) {
	envs := s.getEnv(env)
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

// Subscribe returns a channel that receives config updates for the environment.
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

// Unsubscribe removes and closes a subscription channel.
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
