package config

import (
	"os"
)

type Config struct {
	Port             string
	OpenRouterAPIKey string
}

func NewConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
	}

	return &Config{
		Port:             port,
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
	}
}
