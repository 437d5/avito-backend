package config

import (
	"os"
)

type Config struct {
	SERVER_ADDRESS    string
	POSTGRES_CONN     string
	POSTGRES_JDBC_URL string
	POSTGRES_USERNAME string
	POSTGRES_PASSWORD string
	POSTGRES_HOST     string
	POSTGRES_PORT     string
	POSTGRES_DATABASE string
}

func NewConfig() (*Config, error) {
	config := Config{
		SERVER_ADDRESS:    os.Getenv("SERVER_ADDRESS"),
		POSTGRES_CONN:     os.Getenv("POSTGRES_CONN"),
		POSTGRES_JDBC_URL: os.Getenv("POSTGRES_JDBC_URL"),
		POSTGRES_USERNAME: os.Getenv("POSTGRES_USERNAME"),
		POSTGRES_PASSWORD: os.Getenv("POSTGRES_PASSWORD"),
		POSTGRES_HOST:     os.Getenv("POSTGRES_HOST"),
		POSTGRES_PORT:     os.Getenv("POSTGRES_PORT"),
		POSTGRES_DATABASE: os.Getenv("POSTGRES_DATABASE"),
	}

	// if config.SERVER_ADDRESS == "" {
	// 	return nil, errors.New("SERVER_ADDRESS variable is not provided")
	// }

	return &config, nil
}
