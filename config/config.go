package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func Load[T any](c *T) error {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	return loadFromEnv(c)
}

func loadFromEnv[T any](c *T) error {
	if err := envconfig.Process("", c); err != nil {
		return fmt.Errorf("failed to load envconfig: %w", err)
	}

	return nil
}
