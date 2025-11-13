package config

import (
	"os"
)

// Config holds runtime configuration values.
type Config struct {
	HTTPAddress string
	DatabaseDSN string
}

// Load reads configuration from environment variables with defaults.
func Load() Config {
	cfg := Config{
		HTTPAddress: ":8080",
		DatabaseDSN: "file:data/app.db?_fk=1",
	}
	if v := os.Getenv("HTTP_ADDRESS"); v != "" {
		cfg.HTTPAddress = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.DatabaseDSN = v
	}
	return cfg
}
