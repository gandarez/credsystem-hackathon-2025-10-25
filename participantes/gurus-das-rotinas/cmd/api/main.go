package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"participantes/gurus-das-rotinas/client/openrouter"
)

// ServiceMap contains all available services
var ServiceMap = map[uint8]string{
	1:  "Consulta Limite / Vencimento do cartão / Melhor dia de compra",
	2:  "Segunda via de boleto de acordo",
	3:  "Segunda via de Fatura",
	4:  "Status de Entrega do Cartão",
	5:  "Status de cartão",
	6:  "Solicitação de aumento de limite",
	7:  "Cancelamento de cartão",
	8:  "Telefones de seguradoras",
	9:  "Desbloqueio de Cartão",
	10: "Esqueceu senha / Troca de senha",
	11: "Perda e roubo",
	12: "Consulta do Saldo",
	13: "Pagamento de contas",
	14: "Reclamações",
	15: "Atendimento humano",
	16: "Token de proposta",
}

// Request structures
type FindServiceRequest struct {
	Intent string `json:"intent"`
}

type FindServiceResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ServiceID   uint8  `json:"service_id"`
		ServiceName string `json:"service_name"`
	} `json:"data"`
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type Server struct {
	openrouterClient *openrouter.Client
}

func main() {
	// Get environment variables
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create OpenRouter client
	client := openrouter.NewClient("https://openrouter.ai/api/v1", openrouter.WithAuth(apiKey))

	// Create server instance
	server := &Server{
		openrouterClient: client,
	}

	// Setup routes
	http.HandleFunc("/api/find-service", server.handleFindService)
	http.HandleFunc("/api/healthz", server.handleHealth)

	// Start server
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (s *Server) handleFindService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FindServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid JSON request body")
		return
	}

	if req.Intent == "" {
		sendErrorResponse(w, "Intent field is required")
		return
	}

	// Call OpenRouter API
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dataResp, err := s.openrouterClient.ChatCompletion(ctx, req.Intent)
	if err != nil {
		log.Printf("OpenRouter API error: %v", err)
		sendErrorResponse(w, "Failed to process intent with AI service")
		return
	}

	// Validate service ID
	if dataResp.ServiceID < 1 || dataResp.ServiceID > 16 {
		log.Printf("Invalid service ID returned: %d", dataResp.ServiceID)
		sendErrorResponse(w, "AI returned invalid service ID")
		return
	}

	// Validate service name matches our map
	expectedName, exists := ServiceMap[dataResp.ServiceID]
	if !exists || dataResp.ServiceName != expectedName {
		log.Printf("Service name mismatch for ID %d: got '%s', expected '%s'",
			dataResp.ServiceID, dataResp.ServiceName, expectedName)
		// Use the correct name from our map
		dataResp.ServiceName = expectedName
	}

	// Send success response
	response := FindServiceResponse{
		Success: true,
		Data: struct {
			ServiceID   uint8  `json:"service_id"`
			ServiceName string `json:"service_name"`
		}{
			ServiceID:   dataResp.ServiceID,
			ServiceName: dataResp.ServiceName,
		},
		Error: "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, errorMsg string) {
	response := FindServiceResponse{
		Success: false,
		Data: struct {
			ServiceID   uint8  `json:"service_id"`
			ServiceName string `json:"service_name"`
		}{
			ServiceID:   0,
			ServiceName: "",
		},
		Error: errorMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Per requirements, always return 200 with success: false
	json.NewEncoder(w).Encode(response)
}
