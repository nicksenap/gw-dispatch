package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/nicksenap/gw-dispatch/internal/config"
	"github.com/nicksenap/gw-dispatch/internal/dispatch"
)

func TestCommandPassesDispatchOptions(t *testing.T) {
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
		"--config", t.TempDir() + "/missing.toml",
	})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := dispatch.Options{Branch: "feat/one", Repos: "api,web", Prompt: "build it", Agent: "claude"}
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
