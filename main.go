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
			token, err := readToken(cmd.String("token"), cmd.String("token-file"))
			if err != nil {
				return err
			}
			if token == "" {
				slog.Warn("Authentication is disabled")
			}
			warnSignInChange(token, cmd.String("trusted-proxies"))
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
			return api.Serve(ctx, st, api.Config{
				Port:           cmd.String("port"),
				Token:          token,
				NoWeb:          cmd.Bool("no-web"),
				TrustedProxies: cmd.String("trusted-proxies"),
			})
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
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "Shared secret token for agent authentication",
				Sources: cli.EnvVars("TETHER_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "token-file",
				Usage:   "File holding the token, such as a docker secret",
				Sources: cli.EnvVars("TETHER_TOKEN_FILE"),
			},
			&cli.StringSliceFlag{
				Name:    "upstream",
				Usage:   "Other traefik http provider to merge in, as name=url (repeatable)",
				Sources: cli.EnvVars("TETHER_UPSTREAMS"),
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

func warnSignInChange(token, trustedProxies string) {
	slog.Warn(
		"v0.8.0 replaces TETHER_TOKEN with TETHER_PASSWORD for the web UI and TETHER_API_TOKEN for agents and Traefik, pin the image to 0.7.0 until you switch",
	)
	switch {
	case token == "":
		slog.Warn("From v0.8.0 Tether without TETHER_PASSWORD only answers on localhost")
	case len(token) < 32:
		slog.Warn(
			"From v0.8.0 the agent token needs at least 32 characters, set a longer TETHER_TOKEN on Tether and its agents now",
		)
	}
	switch strings.ToLower(strings.TrimSpace(trustedProxies)) {
	case "direct", "none", "cloudflare", "traefik", "nginx", "standard":
		slog.Warn(
			"From v0.8.0 TETHER_TRUSTED_PROXIES only takes IPs and ranges, such as 172.18.0.0/16",
			"value",
			trustedProxies,
		)
	}
}

func readToken(token, path string) (string, error) {
	if path == "" {
		return token, nil
	}
	if token != "" {
		return "", errors.New("set either --token or --token-file, not both")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading token file: %w", err)
	}
	token = strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file %s is empty", path)
	}
	return token, nil
}
