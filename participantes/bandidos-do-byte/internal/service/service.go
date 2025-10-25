package service

import (
	"fmt"

	"github.com/bandidos_do_byte/api/internal/domain"
	"github.com/bandidos_do_byte/api/internal/ports"
)

type ServiceFinder interface {
	FindService(intent string) (*domain.ServiceData, error)
	HealthCheck() string
}

type serviceFinder struct {
	intentClassifier ports.IntentClassifier
	trainingData     ports.TrainingDataRepository
	cachedExamples   []domain.IntentExample
	examplesLoaded   bool
}

func NewServiceFinder(classifier ports.IntentClassifier, trainingRepo ports.TrainingDataRepository) ServiceFinder {
	return &serviceFinder{
		intentClassifier: classifier,
		trainingData:     trainingRepo,
		examplesLoaded:   false,
	}
}

// FindService implementa a lógica para encontrar o serviço apropriado usando IA
func (s *serviceFinder) FindService(intent string) (*domain.ServiceData, error) {
	// Lazy load training examples on first use
	if !s.examplesLoaded {
		examples, err := s.trainingData.LoadIntentExamples()
		if err != nil {
			return nil, fmt.Errorf("failed to load training examples: %w", err)
		}
		s.cachedExamples = examples
		s.examplesLoaded = true
	}

	// Build classification request
	classificationReq := domain.IntentClassificationRequest{
		UserIntent: intent,
		Examples:   s.cachedExamples,
	}

	// Use AI to classify the intent
	result, err := s.intentClassifier.ClassifyIntent(classificationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to classify intent: %w", err)
	}

	return &domain.ServiceData{
		ServiceID:   result.ServiceID,
		ServiceName: result.ServiceName,
	}, nil
}

// HealthCheck verifica a saúde do serviço
func (s *serviceFinder) HealthCheck() string {
	if err := s.intentClassifier.HealthCheck(); err != nil {
		return "unhealthy: " + err.Error()
	}
	return "ok"
}
