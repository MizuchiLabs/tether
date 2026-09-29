package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/mizuchilabs/kata/buildinfo"
	"github.com/mizuchilabs/kata/logx"
	"github.com/mizuchilabs/kata/sigx"
	"github.com/urfave/cli/v3"

	"github.com/mizuchilabs/tether/internal/api"
	"github.com/mizuchilabs/tether/internal/state"
)

func main() {
	cmd := &cli.Command{
		EnableShellCompletion: true,
		Suggest:               true,
		Name:                  "tether",
		Version:               buildinfo.String(),
		Usage:                 "traefik center",
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			logx.Init(cmd.Bool("debug"))
			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.String("token") == "" {
				slog.Warn("Authentication is disabled")
			}
			st, err := state.New(ctx, cmd.String("config"))
			if err != nil {
				return err
			}
			return api.Serve(ctx, st, api.Config{
				Port:           cmd.String("port"),
				Token:          cmd.String("token"),
				NoWeb:          cmd.Bool("no-web"),
				TrustedProxies: cmd.String("trusted-proxies"),
			})
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable debug logging",
				Sources: cli.EnvVars("TETHER_DEBUG"),
			},
			&cli.BoolFlag{
				Name:    "no-web",
				Usage:   "Disable the web UI",
				Sources: cli.EnvVars("TETHER_NO_WEB"),
			},
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Port to listen on",
				Value:   "3000",
				Sources: cli.EnvVars("TETHER_PORT"),
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Local configuration file",
				Value:   "/data/dynamic.yml",
				Sources: cli.EnvVars("TETHER_CONFIG"),
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "Shared secret token for agent authentication",
				Sources: cli.EnvVars("TETHER_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "trusted-proxies",
				Usage:   "Where to read the client IP for rate limiting: direct, cloudflare, traefik, or CIDRs",
				Sources: cli.EnvVars("TETHER_TRUSTED_PROXIES"),
			},
		},
	}

	if err := cmd.Run(sigx.NotifyContext(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmd.Name, err)
		os.Exit(1)
	}
}
