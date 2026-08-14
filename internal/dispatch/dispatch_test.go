package dispatch

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nicksenap/gw-dispatch/internal/config"
)

func TestResolveAgentBuiltins(t *testing.T) {
	tests := map[string][]string{
		"pi":       {"pi", "do the work"},
		"claude":   {"claude", "do the work"},
		"codex":    {"codex", "do the work"},
		"opencode": {"opencode", "--prompt", "do the work"},
	}
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ResolveAgent(config.Config{}, name, "do the work")
			if err != nil {
				t.Fatalf("ResolveAgent() error = %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("ResolveAgent() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestResolveAgentCustomOverridesBuiltinWithoutShellExpansion(t *testing.T) {
	cfg := config.Config{Agents: map[string]config.Agent{
		"pi": {Command: []string{"custom-pi", "--message={prompt}"}},
	}}
	prompt := `fix "quotes"; rm -rf /`

	got, err := ResolveAgent(cfg, "pi", prompt)
	if err != nil {
		t.Fatalf("ResolveAgent() error = %v", err)
	}
	want := []string{"custom-pi", "--message=" + prompt}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveAgent() = %#v, want %#v", got, want)
	}
}

func TestResolveAgentUnknown(t *testing.T) {
	_, err := ResolveAgent(config.Config{}, "missing", "prompt")
	if err == nil || !strings.Contains(err.Error(), "unknown agent") {
		t.Fatalf("ResolveAgent() error = %v, want unknown agent", err)
	}
}

func TestDeriveWorkspaceName(t *testing.T) {
	if got := DeriveWorkspaceName("feat/my task"); got != "feat-my-task" {
		t.Fatalf("DeriveWorkspaceName() = %q, want feat-my-task", got)
	}
}

func TestFindWorkspacePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	state := `[{"name":"feat-one","path":"/tmp/feat-one"},{"name":"other","path":"/tmp/other"}]`
	if err := os.WriteFile(path, []byte(state), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := FindWorkspacePath(path, "feat-one")
	if err != nil {
		t.Fatalf("FindWorkspacePath() error = %v", err)
	}
	if got != "/tmp/feat-one" {
		t.Fatalf("FindWorkspacePath() = %q, want /tmp/feat-one", got)
	}
}

type recordedCommand struct {
	name string
	args []string
	dir  string
}

type fakeRunner struct {
	commands  []recordedCommand
	statePath string
	workspace string
	createErr error
	agentErr  error
}

func (f *fakeRunner) LookPath(name string) (string, error) {
	return "/fake/" + name, nil
}

func (f *fakeRunner) Run(name string, args []string, dir string) error {
	f.commands = append(f.commands, recordedCommand{name: name, args: append([]string(nil), args...), dir: dir})
	if name == "gw" {
		if f.createErr != nil {
			return f.createErr
		}
		state := `[{"name":"feat-one","path":"` + f.workspace + `"}]`
		return os.WriteFile(f.statePath, []byte(state), 0o600)
	}
	return f.agentErr
}

func TestRunCreatesWorkspaceThenStartsAgentThere(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	runner := &fakeRunner{statePath: statePath, workspace: "/workspaces/feat-one"}
	opts := Options{Branch: "feat/one", Repos: "api,web", Prompt: "build it", Agent: "pi", StatePath: statePath}

	if err := Run(opts, config.Config{}, runner); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := []recordedCommand{
		{name: "gw", args: []string{"create", "--branch", "feat/one", "--repos", "api,web"}},
		{name: "pi", args: []string{"build it"}, dir: "/workspaces/feat-one"},
	}
	if !reflect.DeepEqual(runner.commands, want) {
		t.Fatalf("commands = %#v, want %#v", runner.commands, want)
	}
}

func TestRunUsesPresetAndConfiguredDefaultAgent(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	runner := &fakeRunner{statePath: statePath, workspace: "/workspaces/feat-one"}
	cfg := config.Config{DefaultAgent: "custom", Agents: map[string]config.Agent{
		"custom": {Command: []string{"custom-agent", "--prompt", "{prompt}"}},
	}}
	opts := Options{Branch: "feat/one", Preset: "backend", Prompt: "build it", StatePath: statePath}

	if err := Run(opts, cfg, runner); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := runner.commands[0].args; !reflect.DeepEqual(got, []string{"create", "--branch", "feat/one", "--preset", "backend"}) {
		t.Fatalf("create args = %#v", got)
	}
	if got := runner.commands[1].name; got != "custom-agent" {
		t.Fatalf("agent command = %q, want custom-agent", got)
	}
}

func TestRunDoesNotStartAgentWhenCreateFails(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	runner := &fakeRunner{statePath: statePath, createErr: errors.New("create failed")}
	opts := Options{Branch: "feat/one", Repos: "api", Prompt: "build it", Agent: "pi", StatePath: statePath}

	err := Run(opts, config.Config{}, runner)
	if err == nil || !strings.Contains(err.Error(), "create failed") {
		t.Fatalf("Run() error = %v", err)
	}
	if len(runner.commands) != 1 {
		t.Fatalf("commands = %d, want only gw create", len(runner.commands))
	}
}

func TestRunReportsRetainedWorkspaceWhenAgentFails(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	runner := &fakeRunner{statePath: statePath, workspace: "/workspaces/feat-one", agentErr: errors.New("agent failed")}
	opts := Options{Branch: "feat/one", Repos: "api", Prompt: "build it", Agent: "pi", StatePath: statePath}

	err := Run(opts, config.Config{}, runner)
	if err == nil || !strings.Contains(err.Error(), "workspace retained at /workspaces/feat-one") {
		t.Fatalf("Run() error = %v, want retained workspace path", err)
	}
}
