package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/victorbrittoferreira/credsystem-hackathon-2025-10-25/participantes/controladores_do_panico_16/internal/client"
)

type (
	Response struct {
		Success bool   `json:"success"`
		Data    *Data  `json:"data,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	Data struct {
		ServiceID   uint8  `json:"service_id"`
		ServiceName string `json:"service_name"`
	}

	IntentRequest struct {
		Intent string `json:"intent"`
	}
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1"
)

var openRouterClient *client.Client

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Ler variáveis de ambiente
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	// Debug: mostrar tamanho da chave
	log.Printf("DEBUG: API Key length: %d characters", len(apiKey))
	log.Printf("DEBUG: API Key starts with: %s", apiKey[:min(20, len(apiKey))])

	// Carregar intents do CSV
	csvPath := "/assets/intents_pre_loaded.csv"
	log.Printf("Loading intents from CSV: %s", csvPath)
	intents, err := client.LoadIntentsFromCSV(csvPath)
	if err != nil {
		log.Fatalf("Failed to load intents from CSV: %v", err)
	}
	log.Printf("Loaded %d intents from CSV", len(intents))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Inicializar cliente OpenRouter
	openRouterClient = client.NewClient(
		openRouterBaseURL,
		client.WithAuth(apiKey),
	)

	http.HandleFunc("/api/healthz", healthCheckHandler)
	http.HandleFunc("/api/find-service", findServiceHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if r.Method != http.MethodGet {
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func findServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Sempre retornar 200 OK, usar success: true/false
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Validar Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "content-type must be application/json",
		})
		return
	}

	var req IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.Intent == "" {
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "Intent is required",
		})
		return
	}

	// Contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Chamar OpenRouter para classificar a intenção
	data, err := openRouterClient.ChatCompletion(ctx, req.Intent)
	if err != nil {
		log.Printf("Error calling ChatCompletion: %v", err)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   fmt.Sprintf("failed to process intent: %v", err),
		})
		return
	}

	// Resposta de sucesso
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data: &Data{
			ServiceID:   data.ServiceID,
			ServiceName: data.ServiceName,
		},
	})
}
