package main

import (
	"log"

	"github.com/bandidos_do_byte/api/internal/adapters"
	"github.com/bandidos_do_byte/api/internal/config"
	"github.com/bandidos_do_byte/api/internal/handler"
	"github.com/bandidos_do_byte/api/internal/ports"
	"github.com/bandidos_do_byte/api/internal/server"
	"github.com/bandidos_do_byte/api/internal/service"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		// Providers
		fx.Provide(
			config.NewConfig,
			provideIntentClassifier,
			provideCSVRepository,
			service.NewServiceFinder,
			handler.NewHandler,
			server.NewServer,
		),
		// Invokers
		fx.Invoke(func(s *server.Server, lc fx.Lifecycle) {
			s.Start(lc)
		}),
	).Run()
}

// provideIntentClassifier cria o classificador de intents baseado na configuração
func provideIntentClassifier(cfg *config.Config) ports.IntentClassifier {
	switch cfg.ClassifierType {
	case config.ClassifierTensorFlow:
		log.Println("Using TensorFlow classifier")
		return adapters.NewTensorFlowClassifier(cfg.TensorFlowModelPath, cfg.TensorFlowServerURL)
	case config.ClassifierOpenRouter:
		log.Println("Using OpenRouter classifier")
		return adapters.NewOpenRouterClient(cfg.OpenRouterAPIKey)
	default:
		log.Printf("Unknown classifier type '%s', falling back to OpenRouter", cfg.ClassifierType)
		return adapters.NewOpenRouterClient(cfg.OpenRouterAPIKey)
	}
}

// provideCSVRepository cria o repositório CSV (adapter)
func provideCSVRepository(cfg *config.Config) ports.TrainingDataRepository {
	return adapters.NewCSVTrainingRepository(cfg.TrainingDataPath)
}
