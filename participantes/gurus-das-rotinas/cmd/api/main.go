package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
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
	openrouterClient   *openrouter.Client
	enhancedClassifier *EnhancedKeywordClassifier
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

	// Create enhanced keyword classifier
	enhancedClassifier := NewEnhancedKeywordClassifier()

	// Create server instance
	server := &Server{
		openrouterClient:   client,
		enhancedClassifier: enhancedClassifier,
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

	// Call OpenRouter API with shorter timeout for faster responses
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	dataResp, err := s.openrouterClient.ChatCompletion(ctx, req.Intent)
	if err != nil {
		log.Printf("OpenRouter API error: %v", err)
		// Try enhanced keyword classification
		if enhancedResp := s.tryEnhancedClassification(req.Intent); enhancedResp != nil {
			log.Printf("Using enhanced classification for: %s", req.Intent)
			dataResp = enhancedResp
		} else {
			sendErrorResponse(w, "Failed to process intent with AI service")
			return
		}
	}

	// Check for out-of-context responses
	if dataResp.ServiceName == "out of context" {
		log.Printf("Out of context request detected: %s", req.Intent)
		sendErrorResponse(w, "Request is not related to banking services")
		return
	}

	// Validate service ID
	if dataResp.ServiceID < 1 || dataResp.ServiceID > 16 {
		log.Printf("Invalid service ID returned: %d", dataResp.ServiceID)
		// Try enhanced classification
		if enhancedResp := s.tryEnhancedClassification(req.Intent); enhancedResp != nil {
			log.Printf("Using enhanced classification for invalid ID: %s", req.Intent)
			dataResp = enhancedResp
		} else {
			sendErrorResponse(w, "AI returned invalid service ID")
			return
		}
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

// tryFallbackClassification attempts to classify common patterns without AI
func (s *Server) tryFallbackClassification(intent string) *openrouter.DataResponse {
	intent = strings.ToLower(intent)

	// Common patterns for each service
	patterns := map[uint8][]string{
		1:  {"limite", "vencimento", "melhor dia", "disponível", "quanto", "saldo"},
		2:  {"boleto", "segunda via", "boleto acordo"},
		3:  {"fatura", "segunda via fatura"},
		4:  {"entrega", "cartão entrega", "status entrega"},
		5:  {"status cartão", "cartão status"},
		6:  {"aumento", "limite", "solicitar"},
		7:  {"cancelar", "cancelamento", "bloquear"},
		8:  {"seguradora", "telefone", "seguro"},
		9:  {"desbloqueio", "desbloquear"},
		10: {"senha", "esqueci", "trocar", "alterar"},
		11: {"perdi", "roubo", "perda", "furtado"},
		12: {"saldo", "consulta saldo", "quanto tenho"},
		13: {"pagamento", "pagar", "conta"},
		14: {"reclamação", "reclamar", "problema"},
		15: {"humano", "atendente", "pessoa"},
		16: {"token", "proposta"},
	}

	for serviceID, keywords := range patterns {
		for _, keyword := range keywords {
			if strings.Contains(intent, keyword) {
				return &openrouter.DataResponse{
					ServiceID:   serviceID,
					ServiceName: ServiceMap[serviceID],
				}
			}
		}
	}

	return nil
}

// tryEnhancedClassification attempts to classify using enhanced keyword matching
func (s *Server) tryEnhancedClassification(intent string) *openrouter.DataResponse {
	if s.enhancedClassifier == nil {
		return nil
	}

	// Classify using enhanced keyword matching
	dataResp, confidence := s.enhancedClassifier.ClassifyWithScore(intent)
	if dataResp != nil {
		log.Printf("Enhanced classification confidence: %.3f", confidence)
		return dataResp
	}

	return nil
}
