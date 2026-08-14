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
	DefaultAgent string           `toml:"default_agent"`
	Agents       map[string]Agent `toml:"agents"`
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

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	if cfg.DefaultAgent == "" {
		cfg.DefaultAgent = "pi"
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]Agent)
	}
	for name, agent := range cfg.Agents {
		if strings.TrimSpace(name) == "" {
			return Config{}, fmt.Errorf("agent name cannot be empty")
		}
		if len(agent.Command) == 0 || strings.TrimSpace(agent.Command[0]) == "" {
			return Config{}, fmt.Errorf("agent %q command cannot be empty", name)
		}
		hasPrompt := false
		for _, arg := range agent.Command {
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
