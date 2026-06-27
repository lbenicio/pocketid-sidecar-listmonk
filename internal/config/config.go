package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the sync service.
type Config struct {
	PocketIDBaseURL  string
	PocketIDAPIKey   string
	ListmonkBaseURL  string
	ListmonkUsername string
	ListmonkPassword string
	ListmonkListID   int
	DryRun           bool
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	listID, err := strconv.Atoi(os.Getenv("LISTMONK_LIST_ID"))
	if err != nil {
		return nil, fmt.Errorf("LISTMONK_LIST_ID: must be a valid integer: %w", err)
	}

	dryRun := false
	if v := os.Getenv("SYNC_DRY_RUN"); v == "true" || v == "1" {
		dryRun = true
	}

	cfg := &Config{
		PocketIDBaseURL:  os.Getenv("POCKETID_BASE_URL"),
		PocketIDAPIKey:   os.Getenv("POCKETID_API_KEY"),
		ListmonkBaseURL:  os.Getenv("LISTMONK_BASE_URL"),
		ListmonkUsername: os.Getenv("LISTMONK_USERNAME"),
		ListmonkPassword: os.Getenv("LISTMONK_PASSWORD"),
		ListmonkListID:   listID,
		DryRun:           dryRun,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.PocketIDBaseURL == "" {
		return fmt.Errorf("POCKETID_BASE_URL is required")
	}
	if c.PocketIDAPIKey == "" {
		return fmt.Errorf("POCKETID_API_KEY is required")
	}
	if c.ListmonkBaseURL == "" {
		return fmt.Errorf("LISTMONK_BASE_URL is required")
	}
	if c.ListmonkUsername == "" {
		return fmt.Errorf("LISTMONK_USERNAME is required")
	}
	if c.ListmonkPassword == "" {
		return fmt.Errorf("LISTMONK_PASSWORD is required")
	}
	if c.ListmonkListID <= 0 {
		return fmt.Errorf("LISTMONK_LIST_ID must be a positive integer")
	}
	return nil
}
