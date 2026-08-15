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

func TestLoadDefaultSelector(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dispatch.toml")
	contents := []byte("default_repos = \"api,web\"\n")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DefaultRepos != "api,web" {
		t.Fatalf("DefaultRepos = %q, want api,web", cfg.DefaultRepos)
	}
}

func TestLoadRejectsMultipleDefaultSelectors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dispatch.toml")
	contents := []byte("default_repos = \"api\"\ndefault_preset = \"backend\"\n")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want mutually exclusive selector error")
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

func TestLoadRejectsUnknownKeys(t *testing.T) {
	tests := map[string]string{
		"top level":   "default_aget = \"claude\"\n",
		"agent field": "[agents.aider]\ncomand = [\"aider\", \"{prompt}\"]\n",
	}
	for name, contents := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "dispatch.toml")
			if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil {
				t.Fatal("Load() error = nil, want unknown key error")
			}
		})
	}
}

func TestDefaultPathUsesGroveDir(t *testing.T) {
	t.Setenv("GROVE_DIR", filepath.Join(t.TempDir(), "grove"))
	want := filepath.Join(os.Getenv("GROVE_DIR"), "dispatch.toml")
	if got := DefaultPath(); got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}
