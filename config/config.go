package config

import (
	"os"
	"strconv"
	"strings"

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
			Enabled: os.Getenv("ALERT_ENABLED") == "true",
			Email: monitoring.EmailConfig{
				Enabled:      os.Getenv("ALERT_EMAIL_ENABLED") == "true",
				SMTPHost:     os.Getenv("SMTP_HOST"),
				SMTPPort:     getEnvAsInt("SMTP_PORT", 587), // Porta padrão para SMTP com TLS
				SMTPUsername: os.Getenv("SMTP_USERNAME"),
				SMTPPassword: os.Getenv("SMTP_PASSWORD"),
				FromEmail:    os.Getenv("ALERT_FROM_EMAIL"),
				ToEmails:     getEnvAsSlice("ALERT_TO_EMAILS", ","),
			},
		},
	}, nil
}

// getEnvAsInt retorna uma variável de ambiente como inteiro
func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvAsSlice retorna uma variável de ambiente como slice usando o separador fornecido
func getEnvAsSlice(key string, sep string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, sep)
	}
	return []string{}
}
