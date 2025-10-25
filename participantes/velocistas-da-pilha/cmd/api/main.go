package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"velocistas_da_pilha/internal/classifier"
	"velocistas_da_pilha/internal/storage"
)

type FindServiceRequest struct {
	Intent string `json:"intent"`
}

type FindServiceResponse struct {
	Success bool                   `json:"success"`
	Data    *ServiceData           `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type ServiceData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

var intentClassifier *classifier.IntentClassifier

func main() {
	// Carregar variáveis de ambiente
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY não definida")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
	}

	// Carregar intenções do CSV
	intents, err := storage.LoadIntentsCSV("assets/intents_pre_loaded.csv")
	if err != nil {
		log.Fatalf("Erro carregando intents: %v", err)
	}

	log.Printf("✅ Carregadas %d intenções do CSV", len(intents))

	// Inicializar classificador
	intentClassifier = classifier.NewIntentClassifier(intents, apiKey)

	// Rotas
	http.HandleFunc("/api/find-service", handleFindService)
	http.HandleFunc("/api/healthz", handleHealth)

	log.Printf("🚀 Servidor rodando na porta %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleFindService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req FindServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		respondError(w, "Intent vazio", http.StatusBadRequest)
		return
	}

	// Classificar intenção
	serviceID, serviceName, err := intentClassifier.Classify(req.Intent)
	if err != nil {
		log.Printf("❌ Erro classificando '%s': %v", req.Intent, err)
		respondError(w, fmt.Sprintf("Erro na classificação: %v", err), http.StatusInternalServerError)
		return
	}

	// log.Printf("✅ Intent: '%s' → Service: %d (%s)", req.Intent, serviceID, serviceName)

	respondSuccess(w, serviceID, serviceName)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}

func respondSuccess(w http.ResponseWriter, serviceID int, serviceName string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FindServiceResponse{
		Success: true,
		Data: &ServiceData{
			ServiceID:   serviceID,
			ServiceName: serviceName,
		},
	})
}

func respondError(w http.ResponseWriter, errMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(FindServiceResponse{
		Success: false,
		Error:   errMsg,
	})
}