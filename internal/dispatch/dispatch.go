package dispatch

import (
	"crypto/sha256"
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
	NoHooks   bool
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

	command, builtin := builtinAgents[name]
	custom, configured := cfg.Agents[name]
	if configured {
		command = custom.Command
	}
	if !builtin && !configured {
		return nil, fmt.Errorf("unknown agent %q (built-ins: pi, claude, codex, opencode)", name)
	}

	resolved := make([]string, len(command))
	for i, arg := range command {
		promptArg := prompt
		if strings.HasPrefix(arg, promptPlaceholder) && strings.HasPrefix(prompt, "-") {
			promptArg = "\n" + prompt
		}
		resolved[i] = strings.ReplaceAll(arg, promptPlaceholder, promptArg)
	}
	return resolved, nil
}

func DeriveBranchFromPrompt(prompt string) string {
	normalized := strings.ToLower(strings.Join(strings.Fields(prompt), " "))
	sum := sha256.Sum256([]byte(normalized))

	var slug strings.Builder
	lastWasSeparator := false
	for _, char := range normalized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			slug.WriteRune(char)
			lastWasSeparator = false
		} else if slug.Len() > 0 && !lastWasSeparator {
			slug.WriteByte('-')
			lastWasSeparator = true
		}
		if slug.Len() >= 40 {
			break
		}
	}
	readable := strings.Trim(slug.String(), "-")
	if readable == "" {
		readable = "task"
	}
	return fmt.Sprintf("dispatch/%s-%x", readable, sum[:4])
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
	branch := strings.TrimSpace(opts.Branch)
	repos := strings.TrimSpace(opts.Repos)
	preset := strings.TrimSpace(opts.Preset)
	if strings.TrimSpace(opts.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	if branch == "" {
		branch = DeriveBranchFromPrompt(opts.Prompt)
	}
	if repos == "" && preset == "" {
		repos = strings.TrimSpace(cfg.DefaultRepos)
		preset = strings.TrimSpace(cfg.DefaultPreset)
	}
	if (repos == "") == (preset == "") {
		return fmt.Errorf("exactly one of repos or preset is required, either as a flag or dispatch config default")
	}

	agentCommand, err := ResolveAgent(cfg, opts.Agent, opts.Prompt)
	if err != nil {
		return err
	}
	gwPath, err := resolveExecutable(runner, "gw")
	if err != nil {
		return fmt.Errorf("gw executable not found: %w", err)
	}
	agentPath, err := resolveExecutable(runner, agentCommand[0])
	if err != nil {
		return fmt.Errorf("agent executable %q not found: %w", agentCommand[0], err)
	}

	createArgs := []string{"create", "--branch", branch}
	if repos != "" {
		createArgs = append(createArgs, "--repos", repos)
	} else {
		createArgs = append(createArgs, "--preset", preset)
	}
	if opts.NoHooks {
		createArgs = append(createArgs, "--no-hooks")
	}
	if err := runner.Run(gwPath, createArgs, ""); err != nil {
		return fmt.Errorf("gw create failed: %w", err)
	}

	statePath := opts.StatePath
	if statePath == "" {
		statePath = DefaultStatePath()
	}
	workspaceName := DeriveWorkspaceName(branch)
	workspacePath, err := FindWorkspacePath(statePath, workspaceName)
	if err != nil {
		return fmt.Errorf("workspace %q was created but its path could not be resolved: %w; inspect with gw list", workspaceName, err)
	}
	if err := runner.Run(agentPath, agentCommand[1:], workspacePath); err != nil {
		return fmt.Errorf("agent failed: %w; workspace retained at %s", err, workspacePath)
	}
	return nil
}

func resolveExecutable(runner Runner, name string) (string, error) {
	resolved, err := runner.LookPath(name)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(resolved) {
		return resolved, nil
	}
	return filepath.Abs(resolved)
}
