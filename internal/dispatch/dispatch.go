package dispatch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicksenap/gw-dispatch/internal/config"
)

const promptPlaceholder = "{prompt}"

var builtinAgents = map[string][]string{
	"pi":       {"pi", promptPlaceholder},
	"claude":   {"claude", promptPlaceholder},
	"codex":    {"codex", promptPlaceholder},
	"opencode": {"opencode", "--prompt", promptPlaceholder},
}

type Options struct {
	Branch    string
	Repos     string
	Preset    string
	Prompt    string
	Agent     string
	StatePath string
}

type Runner interface {
	LookPath(name string) (string, error)
	Run(name string, args []string, dir string) error
}

func ResolveAgent(cfg config.Config, name, prompt string) ([]string, error) {
	if name == "" {
		name = cfg.DefaultAgent
		if name == "" {
			name = "pi"
		}
	}

	command, ok := builtinAgents[name]
	if custom, customOK := cfg.Agents[name]; customOK {
		command = custom.Command
		ok = true
	}
	if !ok {
		return nil, fmt.Errorf("unknown agent %q (built-ins: pi, claude, codex, opencode)", name)
	}

	resolved := make([]string, len(command))
	for i, arg := range command {
		resolved[i] = strings.ReplaceAll(arg, promptPlaceholder, prompt)
	}
	return resolved, nil
}

func DeriveWorkspaceName(branch string) string {
	name := strings.ReplaceAll(branch, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")
	return strings.Trim(name, "-")
}

func DefaultStatePath() string {
	if statePath := os.Getenv("GROVE_STATE"); statePath != "" {
		return statePath
	}
	if groveDir := os.Getenv("GROVE_DIR"); groveDir != "" {
		return filepath.Join(groveDir, "state.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".grove", "state.json")
	}
	return filepath.Join(home, ".grove", "state.json")
}

func FindWorkspacePath(statePath, name string) (string, error) {
	data, err := os.ReadFile(statePath)
	if err != nil {
		return "", fmt.Errorf("read Grove state %s: %w", statePath, err)
	}
	var workspaces []struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(data, &workspaces); err != nil {
		return "", fmt.Errorf("parse Grove state %s: %w", statePath, err)
	}
	for _, workspace := range workspaces {
		if workspace.Name == name {
			if workspace.Path == "" {
				return "", fmt.Errorf("workspace %q has no path in Grove state", name)
			}
			return workspace.Path, nil
		}
	}
	return "", fmt.Errorf("workspace %q not found in Grove state", name)
}

func Run(opts Options, cfg config.Config, runner Runner) error {
	if strings.TrimSpace(opts.Branch) == "" {
		return fmt.Errorf("branch is required")
	}
	if strings.TrimSpace(opts.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	if (opts.Repos == "") == (opts.Preset == "") {
		return fmt.Errorf("exactly one of repos or preset is required")
	}

	agentCommand, err := ResolveAgent(cfg, opts.Agent, opts.Prompt)
	if err != nil {
		return err
	}
	if _, err := runner.LookPath("gw"); err != nil {
		return fmt.Errorf("gw executable not found: %w", err)
	}
	if _, err := runner.LookPath(agentCommand[0]); err != nil {
		return fmt.Errorf("agent executable %q not found: %w", agentCommand[0], err)
	}

	createArgs := []string{"create", "--branch", opts.Branch}
	if opts.Repos != "" {
		createArgs = append(createArgs, "--repos", opts.Repos)
	} else {
		createArgs = append(createArgs, "--preset", opts.Preset)
	}
	if err := runner.Run("gw", createArgs, ""); err != nil {
		return fmt.Errorf("gw create failed: %w", err)
	}

	statePath := opts.StatePath
	if statePath == "" {
		statePath = DefaultStatePath()
	}
	workspacePath, err := FindWorkspacePath(statePath, DeriveWorkspaceName(opts.Branch))
	if err != nil {
		return err
	}
	if err := runner.Run(agentCommand[0], agentCommand[1:], workspacePath); err != nil {
		return fmt.Errorf("agent failed: %w; workspace retained at %s", err, workspacePath)
	}
	return nil
}
