package config

import (
	"os"

	"mms_api/pkg/db/postgres"
	"mms_api/pkg/monitoring"
)

type Config struct {
	// Database configuration
	Database postgres.Config

	// MercadoBitcoin configuration
	MercadoBitcoinBaseURL string

	// Alert configuration
	AlertConfig monitoring.AlertConfig
}

// Load carrega as configurações do ambiente
func Load() (*Config, error) {
	return &Config{
		Database: postgres.Config{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   os.Getenv("DB_NAME"),
		},
		MercadoBitcoinBaseURL: os.Getenv("MB_API_URL"),
		AlertConfig: monitoring.AlertConfig{
			Enabled: os.Getenv("ALERTS_ENABLED") == "true",
		},
	}, nil
}
