package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("config not found")

type Config struct {
	LeaderboardURL string `json:"leaderboard_url"`
}

func Default() Config {
	return Config{}
}

// Path returns the config file path, e.g. ~/.config/aoccli/config.json.
func Path() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "aoccli", "config.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "aoccli", "config.json"), nil
}

func Load() (Config, error) {
	p, err := Path()
	if err != nil {
		return Default(), err
	}

	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), ErrNotFound
	}
	if err != nil {
		return Default(), err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}
	return cfg, nil
}

func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, p)
}
