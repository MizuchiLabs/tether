package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/mizuchilabs/kata/buildinfo"
	"github.com/mizuchilabs/kata/logx"
	"github.com/mizuchilabs/kata/sigx"
	"github.com/mizuchilabs/nushi"

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
			if os.Getenv("TETHER_TOKEN") != "" || os.Getenv("TETHER_TOKEN_FILE") != "" {
				return errors.New(
					"TETHER_TOKEN is gone since v0.8.0, set TETHER_PASSWORD for the web UI and TETHER_API_TOKEN for agents and Traefik",
				)
			}
			guard, err := nushi.FromEnv("tether")
			if err != nil {
				return err
			}
			if guard.Mode() == nushi.ModeNone {
				slog.Warn("No sign in is set, Tether only answers on localhost")
			}
			st, err := state.New(ctx, cmd.String("config"))
			if err != nil {
				return err
			}
			for _, u := range cmd.StringSlice("upstream") {
				name, url, ok := strings.Cut(u, "=")
				if !ok || name == "" || url == "" {
					return fmt.Errorf("invalid upstream %q, want name=url", u)
				}
				st.WatchUpstream(ctx, name, url)
			}
			return api.Serve(ctx, st, guard, cmd.String("port"), cmd.Bool("no-web"))
		},
		Commands: []*cli.Command{{
			Name:  "health",
			Usage: "Exit 0 if the running server answers, for container healthchecks",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return api.Healthy(ctx, cmd.String("port"))
			},
		}},
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
			&cli.StringSliceFlag{
				Name:    "upstream",
				Usage:   "Other traefik http provider to merge in, as name=url (repeatable)",
				Sources: cli.EnvVars("TETHER_UPSTREAMS"),
			},
		},
	}

	if err := cmd.Run(sigx.NotifyContext(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmd.Name, err)
		os.Exit(1)
	}
}
