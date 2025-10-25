package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/TaysonMartinss/cacadores-de-corrida/participantes/cacadores-de-corrida/agent"
	"github.com/TaysonMartinss/cacadores-de-corrida/participantes/cacadores-de-corrida/validator"
	"github.com/joho/godotenv"
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

var classifier *agent.ServiceClassifier

func main() {
	// Carregar variáveis de ambiente
	godotenv.Load()

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY não configurada")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
	}

	// Inicializar o classificador
	var err error
	classifier, err = agent.NewServiceClassifier(apiKey)
	if err != nil {
		log.Fatalf("Erro ao inicializar classificador: %v", err)
	}

	// Configurar rotas
	http.HandleFunc("/api/find-service", findServiceHandler)
	http.HandleFunc("/api/healthz", healthzHandler)

	log.Printf("Servidor rodando na porta %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func findServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req FindServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Erro ao decodificar requisição", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		sendErrorResponse(w, "Intent não pode ser vazio", http.StatusBadRequest)
		return
	}

	// Classificar a intenção usando o agente
	serviceID, serviceName, err := classifier.Classify(req.Intent)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Erro ao classificar intenção: %v", err), http.StatusInternalServerError)
		return
	}

	// Validar a resposta do agente
	if !validator.IsValidService(serviceID) {
		sendErrorResponse(w, "Serviço inválido retornado pelo classificador", http.StatusInternalServerError)
		return
	}

	// Retornar resposta de sucesso
	response := FindServiceResponse{
		Success: true,
		Data: &ServiceData{
			ServiceID:   serviceID,
			ServiceName: serviceName,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := FindServiceResponse{
		Success: false,
		Error:   errorMsg,
	}
	
	json.NewEncoder(w).Encode(response)
}
