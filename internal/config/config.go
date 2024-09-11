package config

import (
	"os"
)

type Config struct {
	SERVER_ADDRESS    string
	POSTGRES_CONN     string
	POSTGRES_JDBC_URL string
	POSTGRES_USERNAME string
	POSTGRES_HOST     string
	POSTGRES_PASSWORD string
	POSTGRES_PORT     string
	POSTGRES_DATABASE string
}

// NewConfig returns pointer to new Config instance and error if occurs
func NewConfig() (*Config, error) {
	config := Config{
		SERVER_ADDRESS:    envLoad("SERVER_ADDRESS", ":8080"),
		POSTGRES_CONN:     os.Getenv("POSTGRES_CONN"),
		POSTGRES_JDBC_URL: os.Getenv("POSTGRES_JDBC_URL"),
		POSTGRES_USERNAME: os.Getenv("POSTGRES_USERNAME"),
		POSTGRES_PASSWORD: os.Getenv("POSTGRES_PASSWORD"),
		POSTGRES_HOST:     os.Getenv("POSTGRES_HOST"),
		POSTGRES_PORT:     os.Getenv("POSTGRES_PORT"),
		POSTGRES_DATABASE: os.Getenv("POSTGRES_DATABASE"),
	}

	return &config, nil
}

// envLoad checks if env variable with name exists
// if not it will return fallback value
func envLoad(name, fallback string) string {
	var value string
	if value = os.Getenv(name); value == "" {
		value = fallback
	}

	return value
} 