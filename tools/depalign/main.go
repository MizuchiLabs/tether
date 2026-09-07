package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const traefikModule = "github.com/traefik/traefik/v3"

// Placeholder version traefik's go.mod uses for modules it replaces locally,
// like dynamic/ext. Those are handled by the replace directive in our go.mod.
const replacedPlaceholder = "v0.0.0-00010101000000-000000000000"

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "depalign:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	pins, err := traefikPins(ctx)
	if err != nil {
		return fmt.Errorf("collecting traefik pins: %w", err)
	}

	current, err := requires("go.mod", true)
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}
	direct, err := requires("go.mod", false)
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}

	var args []string
	for mod, ver := range pins {
		if ver == replacedPlaceholder || direct[mod] != "" || current[mod] == "" {
			continue
		}
		args = append(args, mod+"@"+ver)
	}
	if len(args) == 0 {
		return nil
	}
	sort.Strings(args)

	if err := goCmd(ctx, append([]string{"get"}, args...)...); err != nil {
		return fmt.Errorf("go get: %w", err)
	}
	return goCmd(ctx, "mod", "tidy")
}

func goCmd(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func traefikPins(ctx context.Context) (map[string]string, error) {
	out, err := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Version}}", traefikModule).Output()
	if err != nil {
		return nil, fmt.Errorf("finding traefik version: %w", err)
	}
	version := strings.TrimSpace(string(out))

	dlArgs := []string{"mod", "download", "-json", traefikModule + "@" + version}
	out, err = exec.CommandContext(ctx, "go", dlArgs...).Output()
	if err != nil {
		return nil, fmt.Errorf("downloading traefik go.mod: %w", err)
	}
	var download struct {
		GoMod string `json:"GoMod"`
	}
	if err := json.Unmarshal(out, &download); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(download.GoMod)
	if err != nil {
		return nil, err
	}
	return parseRequires(data, true), nil
}

func requires(path string, keepIndirect bool) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseRequires(data, keepIndirect), nil
}

func parseRequires(data []byte, keepIndirect bool) map[string]string {
	pins := map[string]string{}
	inBlock := false
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "require (":
			inBlock = true
			continue
		case inBlock && line == ")":
			inBlock = false
			continue
		case !inBlock && !strings.HasPrefix(line, "require "):
			continue
		}
		line = strings.TrimPrefix(line, "require ")

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if !keepIndirect && strings.Contains(line, "// indirect") {
			continue
		}
		pins[fields[0]] = fields[1]
	}
	return pins
}
