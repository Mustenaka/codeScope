package config

import (
	"os"
	"strconv"
)

type HTTPConfig struct {
	Host string
	Port int
}

type Config struct {
	AppName string
	HTTP    HTTPConfig
}

func Load() Config {
	return Config{
		AppName: envOrDefault("CODESCOPE_SERVER_APP_NAME", "codeScope Server"),
		HTTP: HTTPConfig{
			Host: envOrDefault("CODESCOPE_SERVER_HOST", "0.0.0.0"),
			Port: envIntOrDefault("CODESCOPE_SERVER_PORT", 8080),
		},
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
