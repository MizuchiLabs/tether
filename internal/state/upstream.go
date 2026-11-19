package state

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	upstreamInterval = 5 * time.Second
	upstreamMaxBody  = 10 << 20
)

// WatchUpstream polls another traefik http provider endpoint until ctx is done.
// It shows up as an agent called name, so it merges and expires like one.
func (s *State) WatchUpstream(ctx context.Context, name, url string) {
	s.watchUpstream(ctx, name, url, upstreamInterval)
}

func (s *State) watchUpstream(ctx context.Context, name, url string, interval time.Duration) {
	client := &http.Client{Timeout: upstreamInterval}
	go func() {
		var last []byte
		connected, failing := false, false
		for {
			body, err := fetchUpstream(ctx, client, url)
			if ctx.Err() != nil {
				return
			}
			switch {
			case err != nil:
				if !failing {
					slog.Warn("Upstream unreachable", "upstream", name, "error", err)
					failing = true
				}
				if connected {
					s.AgentDisconnected(defaultEnv, name)
					connected = false
				}
				last = nil
			case !bytes.Equal(body, last):
				failing = false
				last = body
				var cfg Config
				if err := json.Unmarshal(body, &cfg); err != nil {
					slog.Error("Invalid upstream config", "upstream", name, "error", err)
					break
				}
				if !connected {
					s.AgentConnected(defaultEnv, name, url)
					connected = true
				}
				s.UpdateAgent(defaultEnv, name, cfg)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
		}
	}()
}

func fetchUpstream(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, upstreamMaxBody))
}
