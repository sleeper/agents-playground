package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config captures all server configuration options loaded from YAML.
type Config struct {
	Server struct {
		Addr           string   `yaml:"addr"`
		AllowedOrigins []string `yaml:"allowed_origins"`
	} `yaml:"server"`

	Confluence struct {
		BaseURL   string `yaml:"base_url"`
		AuthToken string `yaml:"auth_token"`
	} `yaml:"confluence"`

	AccessControl struct {
		SharedSecret  string   `yaml:"shared_secret"`
		DefaultSpaces []string `yaml:"default_spaces"`
		Policies      []Policy `yaml:"policies"`
	} `yaml:"access_control"`
}

// Policy ties a subject identifier to the set of Confluence spaces they may query.
type Policy struct {
	Subject       string   `yaml:"subject"`
	AllowedSpaces []string `yaml:"allowed_spaces"`
}

// Load reads and decodes YAML configuration from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}

	return &cfg, nil
}
