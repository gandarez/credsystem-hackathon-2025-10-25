package config

import (
	"os"
	"path/filepath"
)

type ClassifierType string

const (
	ClassifierOpenRouter ClassifierType = "openrouter"
	ClassifierTensorFlow ClassifierType = "tensorflow"
)

type Config struct {
	Port                string
	OpenRouterAPIKey    string
	TrainingDataPath    string
	ClassifierType      ClassifierType
	TensorFlowModelPath string
	TensorFlowServerURL string
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

	// Tipo de classificador (openrouter ou tensorflow)
	classifierType := ClassifierType(os.Getenv("CLASSIFIER_TYPE"))
	if classifierType == "" {
		classifierType = ClassifierOpenRouter // Default
	}

	// Caminho do modelo TensorFlow
	tfModelPath := os.Getenv("TENSORFLOW_MODEL_PATH")
	if tfModelPath == "" {
		tfModelPath = filepath.Join(".", "training", "service_intent_model.h5")
	}

	// URL do servidor Python TensorFlow
	tfServerURL := os.Getenv("TENSORFLOW_SERVER_URL")
	if tfServerURL == "" {
		tfServerURL = "http://localhost:5000"
	}

	return &Config{
		Port:                port,
		OpenRouterAPIKey:    os.Getenv("OPENROUTER_API_KEY"),
		TrainingDataPath:    trainingPath,
		ClassifierType:      classifierType,
		TensorFlowModelPath: tfModelPath,
		TensorFlowServerURL: tfServerURL,
	}
}
