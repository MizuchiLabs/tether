// Package config contains the application configuration
package config

import (
	"context"
	"log/slog"

	"github.com/mizuchilabs/tether/internal/state"
	"github.com/urfave/cli/v3"
)

type Config struct {
	Port    string
	Secret  string
	Version string
	Debug   bool

	State *state.State
}

// New loads configuration from environment variables
func New(ctx context.Context, cmd *cli.Command) (*Config, error) {
	cfg := Config{}

	cfg.State = state.New()
	cfg.Version = cmd.Root().Version
	cfg.Debug = cmd.Bool("debug")
	cfg.Port = cmd.String("port")
	cfg.Secret = cmd.String("secret")
	if cfg.Secret == "" {
		slog.Warn("Authentication is disabled")
	}

	local := cmd.String("config")
	if local != "" {
		if err := cfg.State.LoadLocalFile("default", local); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}
