package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port             string
	OpenRouterAPIKey string
	TrainingDataPath string
}

func NewConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
	}

	trainingPath := os.Getenv("TRAINING_DATA_PATH")
	if trainingPath == "" {
		// Default to relative path from project root
		trainingPath = filepath.Join(".", "training", "intents_pre_loaded.csv")
	}

	return &Config{
		Port:             port,
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
		TrainingDataPath: trainingPath,
	}
}
