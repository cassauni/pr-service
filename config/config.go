package config

import (
	"fmt"
	"log/slog"
	"os"
)

func NewConfig() (*ConfigModel, error) {
	cfg := &ConfigModel{
		HTTP: HTTPConfig{
			Host: env("HTTP_HOST", "0.0.0.0"),
			Port: env("HTTP_PORT", "8080"),
		},
		Postgres: PostgresConfig{
			Host:     env("POSTGRES_HOST", "pr-db"),
			Port:     env("POSTGRES_PORT", "5432"),
			User:     env("POSTGRES_USER", "postgres"),
			Password: env("POSTGRES_PASSWORD", "password"),
			DBName:   env("POSTGRES_DB", "postgres"),
			SSLMode:  env("POSTGRES_SSLMODE", "disable"),
			PgDriver: env("POSTGRES_DRIVER", "pgx"),
		},
	}

	if cfg.HTTP.Host == "" || cfg.HTTP.Port == "" {
		return nil, fmt.Errorf("HTTP_HOST and HTTP_PORT must be set")
	}
	if cfg.Postgres.Host == "" || cfg.Postgres.User == "" || cfg.Postgres.DBName == "" {
		return nil, fmt.Errorf("POSTGRES_HOST, POSTGRES_USER and POSTGRES_DB must be set")
	}

	slog.Info("config loaded from environment")
	return cfg, nil
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
