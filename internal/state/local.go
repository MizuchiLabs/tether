package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.yaml.in/yaml/v3"
)

// watchLocalFile loads the file into the default environment and reloads it on change.
// A missing file is fine, it gets picked up once created.
func (s *State) watchLocalFile(ctx context.Context, path string) error {
	path = filepath.Clean(path)
	if err := s.loadLocalFile(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}
	// Watch the directory, editors replace files on save which kills a watch on the file itself.
	if err := watcher.Add(filepath.Dir(path)); err != nil {
		_ = watcher.Close()
		slog.Debug("Not watching local config", "path", path, "error", err)
		return nil
	}

	go func() {
		defer func() { _ = watcher.Close() }()
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-watcher.Events:
				if filepath.Clean(event.Name) != path || !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
					continue
				}
				if err := s.loadLocalFile(path); err != nil {
					slog.Error("Failed to reload local config", "path", path, "error", err)
				}
			case err := <-watcher.Errors:
				slog.Error("Local config watcher error", "path", path, "error", err)
			}
		}
	}()
	return nil
}

func (s *State) loadLocalFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading local config: %w", err)
	}

	if ext := filepath.Ext(path); ext == ".yaml" || ext == ".yml" {
		// Go through JSON so values have the same types as agent configs.
		var v any
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("parsing local config %s: %w", path, err)
		}
		if data, err = json.Marshal(v); err != nil {
			return fmt.Errorf("converting local config %s: %w", path, err)
		}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parsing local config %s: %w", path, err)
	}

	slog.Info("Loaded local config", "path", path)
	s.setLocal(defaultEnv, cfg)
	return nil
}
