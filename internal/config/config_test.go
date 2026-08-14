package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadMissingUsesPiDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DefaultAgent != "pi" {
		t.Fatalf("DefaultAgent = %q, want pi", cfg.DefaultAgent)
	}
}

func TestLoadCustomAgent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dispatch.toml")
	contents := []byte(`default_agent = "aider"

[agents.aider]
command = ["aider", "--message", "{prompt}"]
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := []string{"aider", "--message", "{prompt}"}
	if !reflect.DeepEqual(cfg.Agents["aider"].Command, want) {
		t.Fatalf("command = %#v, want %#v", cfg.Agents["aider"].Command, want)
	}
}

func TestLoadRejectsCustomCommandWithoutPrompt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dispatch.toml")
	if err := os.WriteFile(path, []byte("[agents.bad]\ncommand = [\"bad\"]\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want placeholder validation error")
	}
}

func TestLoadRejectsPromptPlaceholderInExecutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dispatch.toml")
	if err := os.WriteFile(path, []byte("[agents.bad]\ncommand = [\"{prompt}\", \"fixed\"]\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want executable validation error")
	}
}

func TestDefaultPathUsesGroveDir(t *testing.T) {
	t.Setenv("GROVE_DIR", filepath.Join(t.TempDir(), "grove"))
	want := filepath.Join(os.Getenv("GROVE_DIR"), "dispatch.toml")
	if got := DefaultPath(); got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}
