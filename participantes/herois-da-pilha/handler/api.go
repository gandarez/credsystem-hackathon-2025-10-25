package handler

import (
	"encoding/json"
	"fmt"
	"herois-da-pilha/service"
	"herois-da-pilha/util"
	"net/http"
)

// APIHandler contém as referências necessárias para os handlers.
type APIHandler struct {
	FinderService *service.FinderService
}

// NewAPIHandler cria uma nova instância do handler.
func NewAPIHandler() *APIHandler {
	return &APIHandler{
		FinderService: service.NewFinderService(),
	}
}

// writeJSON é um utilitário para escrever a resposta JSON.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Trata o erro de encoding, embora seja raro em structs simples
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Erro ao escrever resposta JSON: %v\n", err)
	}
}

// HealthCheckHandler verifica a saúde do serviço.
// GET /api/healthz
func (h *APIHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// O roteador ServeMux do Go atende a todos os métodos,
	// mas para um healthcheck simples, podemos apenas verificar o path.
	if r.URL.Path != "/api/healthz" {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, util.HealthzResponse{
		Status: "ok",
	})
}

// FindServiceHandler processa a solicitação e chama a IA para roteamento.
// POST /api/find-service
func (h *APIHandler) FindServiceHandler(w http.ResponseWriter, r *http.Request) {
	// A rota só é acessada via POST, mas é bom garantir.
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusOK, util.FindServiceResponse{
			Success: false,
			Error:   "Método não permitido. Use POST.",
		})
		return
	}

	var req util.FindServiceRequest

	// 1. Binding do JSON de entrada (usando a biblioteca padrão)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Intent == "" {
		writeJSON(w, http.StatusOK, util.FindServiceResponse{
			Success: false,
			Error:   "Corpo da requisição inválido. Esperado {\"intent\": \"string\"}",
		})
		return
	}

	// 2. Chama o serviço de IA para encontrar o serviço mais adequado
	response := h.FinderService.FindService(req.Intent)

	// 3. Resposta
	writeJSON(w, http.StatusOK, response)
}
