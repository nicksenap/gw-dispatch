package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nicksenap/gw-dispatch/internal/config"
	"github.com/nicksenap/gw-dispatch/internal/dispatch"
)

func TestCommandPassesDispatchOptions(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "dispatch.toml")
	if err := os.WriteFile(configPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	var got dispatch.Options
	command := newRootCommand(func(opts dispatch.Options, _ config.Config) error {
		got = opts
		return nil
	})
	command.SetArgs([]string{
		"--branch", "feat/one",
		"--repos", "api,web",
		"--prompt", "build it",
		"--agent", "claude",
		"--config", configPath,
	})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := dispatch.Options{Branch: "feat/one", Repos: "api,web", Prompt: "build it", Agent: "claude"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
}

func TestCommandAcceptsShortFlags(t *testing.T) {
	var got dispatch.Options
	command := newRootCommand(func(opts dispatch.Options, _ config.Config) error {
		got = opts
		return nil
	})
	command.SetArgs([]string{
		"-b", "feat/one",
		"-p", "backend",
		"-P", "build it",
		"-n",
	})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := dispatch.Options{Branch: "feat/one", Preset: "backend", Prompt: "build it", NoHooks: true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
}

func TestCommandAllowsSelectorFromDispatchConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "dispatch.toml")
	if err := os.WriteFile(configPath, []byte("default_preset = \"backend\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var got config.Config
	command := newRootCommand(func(_ dispatch.Options, cfg config.Config) error {
		got = cfg
		return nil
	})
	command.SetArgs([]string{"--prompt", "build it", "--config", configPath})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.DefaultPreset != "backend" {
		t.Fatalf("DefaultPreset = %q, want backend", got.DefaultPreset)
	}
}

func TestCommandAllowsGeneratedBranch(t *testing.T) {
	var got dispatch.Options
	command := newRootCommand(func(opts dispatch.Options, _ config.Config) error {
		got = opts
		return nil
	})
	command.SetArgs([]string{"--repos", "api", "--prompt", "build it"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.Branch != "" {
		t.Fatalf("Branch = %q, want empty for dispatch-layer derivation", got.Branch)
	}
}

func TestCommandRejectsMissingRequiredFlags(t *testing.T) {
	tests := map[string][]string{
		"prompt": {"--branch", "feat/one", "--repos", "api"},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			command := newRootCommand(func(dispatch.Options, config.Config) error { return nil })
			command.SetArgs(args)
			if err := command.Execute(); err == nil {
				t.Fatal("Execute() error = nil, want required flag error")
			}
		})
	}
}

func TestCommandRejectsPositionalArguments(t *testing.T) {
	command := newRootCommand(func(dispatch.Options, config.Config) error { return nil })
	command.SetArgs([]string{
		"--branch", "feat/one",
		"--repos", "api",
		"--prompt", "build it",
		"unexpected-name",
	})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want positional argument error")
	}
}

func TestCommandRejectsExplicitMissingConfig(t *testing.T) {
	command := newRootCommand(func(dispatch.Options, config.Config) error { return nil })
	command.SetArgs([]string{
		"--branch", "feat/one",
		"--repos", "api",
		"--prompt", "build it",
		"--config", t.TempDir() + "/missing.toml",
	})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want missing config error")
	}
}

func TestCommandRejectsReposAndPresetTogether(t *testing.T) {
	command := newRootCommand(func(dispatch.Options, config.Config) error { return nil })
	command.SetArgs([]string{
		"--branch", "feat/one",
		"--repos", "api",
		"--preset", "backend",
		"--prompt", "build it",
	})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "if any flags in the group") {
		t.Fatalf("Execute() error = %v, want mutual exclusion error", err)
	}
}
