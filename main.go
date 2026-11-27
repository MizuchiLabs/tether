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
			ready, err := signInReady(token, cmd.String("trusted-proxies"))
			if err != nil {
				return err
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
			return api.Serve(ctx, st, api.Config{
				Port:           cmd.String("port"),
				Token:          token,
				NoWeb:          cmd.Bool("no-web"),
				TrustedProxies: cmd.String("trusted-proxies"),
				SignInReady:    ready,
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
				Sources: cli.EnvVars("TETHER_TOKEN", "TETHER_API_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "token-file",
				Usage:   "File holding the token, such as a docker secret",
				Sources: cli.EnvVars("TETHER_TOKEN_FILE", "TETHER_API_TOKEN_FILE"),
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

// signInReady logs what v0.8.0 still needs.
func signInReady(token, trustedProxies string) (bool, error) {
	oldToken := os.Getenv("TETHER_TOKEN") != "" || os.Getenv("TETHER_TOKEN_FILE") != ""
	apiToken := os.Getenv("TETHER_API_TOKEN") != "" || os.Getenv("TETHER_API_TOKEN_FILE") != ""
	if oldToken && apiToken {
		return false, errors.New("TETHER_TOKEN and TETHER_API_TOKEN are both set, keep TETHER_API_TOKEN")
	}
	if apiToken && len(token) < 32 {
		return false, errors.New("TETHER_API_TOKEN needs at least 32 characters")
	}

	var missing []string
	if !apiToken {
		missing = append(missing,
			"Move the token to TETHER_API_TOKEN with at least 32 characters, and use it on the agents and in Traefik")
	}
	if len(os.Getenv("TETHER_PASSWORD")) < 12 &&
		os.Getenv("TETHER_PASSWORD_FILE") == "" &&
		os.Getenv("TETHER_OIDC_ISSUER") == "" {
		missing = append(missing,
			"Set TETHER_PASSWORD with at least 12 characters or OIDC, without either v0.8.0 only answers on localhost")
	}
	switch strings.ToLower(strings.TrimSpace(trustedProxies)) {
	case "direct", "none", "cloudflare", "traefik", "nginx", "standard":
		missing = append(missing,
			"Set TETHER_TRUSTED_PROXIES to IPs or ranges such as 172.18.0.0/16, or remove it")
	}
	if len(missing) == 0 {
		return true, nil
	}
	slog.Warn("v0.8.0 on January 13 changes sign in and this version already reads the new settings, " +
		"see https://github.com/MizuchiLabs/tether#moving-to-v080")
	for _, m := range missing {
		slog.Warn(m)
	}
	return false, nil
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
