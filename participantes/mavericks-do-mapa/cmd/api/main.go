package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"mavericksdomapa/client/openrouter"
	"mavericksdomapa/internal/controller"
	"mavericksdomapa/internal/gateway"
	"mavericksdomapa/internal/handler"
)

func main() {
	app := fiber.New()

	serviceGateway := initServiceGateway()
	serviceController := controller.NewServiceController(serviceGateway)
	serviceHandler := handler.NewServiceHandler(serviceController)

	handler.RegisterRoutes(app, serviceHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func initServiceGateway() gateway.ServiceGateway {
	apiKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if apiKey == "" {
		log.Println("OPENROUTER_API_KEY not set; using static service mappings")
		return gateway.NewStaticServiceGateway()
	}

	openRouterBaseURL := envOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1")
	openRouterModel := envOrDefault("OPENROUTER_MODEL", "google/gemini-2.5-flash-preview-09-2025")
	systemPrompt := os.Getenv("OPENROUTER_SYSTEM_PROMPT")
	if systemPrompt == "" {
		systemPrompt = gateway.OpenRouterSystemPrompt
	}

	opts := []openrouter.Option{
		openrouter.WithModel(openRouterModel),
		openrouter.WithSystemPrompt(systemPrompt),
		openrouter.WithAuth(apiKey),
	}

	referer := os.Getenv("OPENROUTER_HTTP_REFERER")
	title := os.Getenv("OPENROUTER_TITLE")
	if referer != "" || title != "" {
		opts = append(opts, openrouter.WithAttribution(referer, title))
	}

	if timeoutEnv := os.Getenv("OPENROUTER_TIMEOUT_SECS"); timeoutEnv != "" {
		if secs, err := strconv.Atoi(timeoutEnv); err == nil && secs > 0 {
			opts = append(opts, openrouter.WithTimeout(time.Duration(secs)*time.Second))
		} else if err != nil {
			log.Printf("invalid OPENROUTER_TIMEOUT_SECS %q: %v", timeoutEnv, err)
		}
	}

	openRouterClient := openrouter.NewClient(openRouterBaseURL, opts...)
	return gateway.NewOpenRouterServiceGateway(openRouterClient)
}

func envOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
