package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Route struct {
	Path     string   `yaml:"path"`
	Provider string   `yaml:"provider"`
	Secret   string   `yaml:"secret"`
	Targets  []string `yaml:"targets"`
}

type Config struct {
	ListenAddr string  `yaml:"listen_addr"`
	Routes     []Route `yaml:"routes"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}

	return &cfg, nil
}
