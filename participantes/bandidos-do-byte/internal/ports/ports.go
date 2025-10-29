package ports

import "github.com/bandidos_do_byte/api/internal/domain"

// IntentClassifier é a porta de saída para classificação de intents usando IA
type IntentClassifier interface {
	ClassifyIntent(request domain.IntentClassificationRequest) (*domain.IntentClassificationResponse, error)
	HealthCheck() error
}

// TrainingDataRepository é a porta de saída para carregar dados de treinamento
type TrainingDataRepository interface {
	LoadIntentExamples() ([]domain.IntentExample, error)
}
