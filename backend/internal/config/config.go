package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	StoragePath string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		StoragePath: getEnv("STORAGE_PATH", "./storage"),
	}

	if err := validatePort(cfg.Port); err != nil {
		return Config{}, err
	}

	if cfg.StoragePath == "" {
		return Config{}, fmt.Errorf("STORAGE_PATH cannot be empty")
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return net.JoinHostPort("", c.Port)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func validatePort(port string) error {
	number, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("PORT must be a number: %w", err)
	}

	if number < 1 || number > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535")
	}

	return nil
}
