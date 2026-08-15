package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const promptPlaceholder = "{prompt}"

type Config struct {
	DefaultAgent  string           `toml:"default_agent"`
	DefaultPreset string           `toml:"default_preset"`
	DefaultRepos  string           `toml:"default_repos"`
	Agents        map[string]Agent `toml:"agents"`
}

type Agent struct {
	Command []string `toml:"command"`
}

func DefaultPath() string {
	if groveDir := os.Getenv("GROVE_DIR"); groveDir != "" {
		return filepath.Join(groveDir, "dispatch.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".grove", "dispatch.toml")
	}
	return filepath.Join(home, ".grove", "dispatch.toml")
}

func Load(path string) (Config, error) {
	cfg := Config{DefaultAgent: "pi", Agents: make(map[string]Agent)}
	if path == "" {
		path = DefaultPath()
	}

	metadata, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}
		return Config{}, fmt.Errorf("unknown config keys: %s", strings.Join(keys, ", "))
	}
	if cfg.DefaultAgent == "" {
		cfg.DefaultAgent = "pi"
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]Agent)
	}
	cfg.DefaultPreset = strings.TrimSpace(cfg.DefaultPreset)
	cfg.DefaultRepos = strings.TrimSpace(cfg.DefaultRepos)
	if cfg.DefaultPreset != "" && cfg.DefaultRepos != "" {
		return Config{}, fmt.Errorf("default_preset and default_repos are mutually exclusive")
	}
	for name, agent := range cfg.Agents {
		if strings.TrimSpace(name) == "" {
			return Config{}, fmt.Errorf("agent name cannot be empty")
		}
		if len(agent.Command) == 0 || strings.TrimSpace(agent.Command[0]) == "" {
			return Config{}, fmt.Errorf("agent %q command cannot be empty", name)
		}
		if strings.Contains(agent.Command[0], promptPlaceholder) {
			return Config{}, fmt.Errorf("agent %q executable cannot contain %s", name, promptPlaceholder)
		}
		hasPrompt := false
		for _, arg := range agent.Command[1:] {
			if strings.Contains(arg, promptPlaceholder) {
				hasPrompt = true
				break
			}
		}
		if !hasPrompt {
			return Config{}, fmt.Errorf("agent %q command must include %s", name, promptPlaceholder)
		}
	}
	return cfg, nil
}
