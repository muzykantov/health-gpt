package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Bot represents the main configuration structure
type Bot struct {
	Telegram `yaml:"telegram"`
	Storage  `yaml:"storage"`
	LLM      `yaml:"llm"`
}

// Read parses configuration from reader in YAML format
func Read(r io.Reader) (*Bot, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	raw = []byte(os.ExpandEnv(string(raw)))

	var cfg Bot
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// FromFile reads configuration from file
func FromFile(path string) (*Bot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	return Read(f)
}
