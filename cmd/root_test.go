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

func TestCommandAcceptsShortCreateFlags(t *testing.T) {
	var got dispatch.Options
	command := newRootCommand(func(opts dispatch.Options, _ config.Config) error {
		got = opts
		return nil
	})
	command.SetArgs([]string{
		"-b", "feat/one",
		"-p", "backend",
		"--prompt", "build it",
	})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := dispatch.Options{Branch: "feat/one", Preset: "backend", Prompt: "build it"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
}

func TestCommandRequiresRepoOrPreset(t *testing.T) {
	command := newRootCommand(func(dispatch.Options, config.Config) error { return nil })
	command.SetArgs([]string{"--branch", "feat/one", "--prompt", "build it"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "at least one of the flags") {
		t.Fatalf("Execute() error = %v, want repo selector error", err)
	}
}

func TestCommandRejectsMissingRequiredFlags(t *testing.T) {
	tests := map[string][]string{
		"branch": {"--repos", "api", "--prompt", "build it"},
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
