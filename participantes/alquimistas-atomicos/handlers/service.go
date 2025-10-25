package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"ivr-service/client"
	"ivr-service/models"
)

type ServiceHandler struct {
	openRouterClient *client.OpenRouterClient
}

func NewServiceHandler(openRouterClient *client.OpenRouterClient) *ServiceHandler {
	return &ServiceHandler{
		openRouterClient: openRouterClient,
	}
}

func (h *ServiceHandler) FindService(w http.ResponseWriter, r *http.Request) {
	// Middleware de recuperação de panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recuperado: %v\n%s", r, debug.Stack())
			h.sendErrorResponse(w, "Erro interno do servidor")
		}
	}()

	var req models.FindServiceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao decodificar requisição: %v", err)
		h.sendErrorResponse(w, "Erro ao decodificar requisição")
		return
	}

	if req.Intent == "" {
		h.sendErrorResponse(w, "Campo 'intent' é obrigatório")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	aiService, err := h.openRouterClient.ClassifyIntent(ctx, req.Intent)
	if err != nil {
		log.Printf("Erro na classificação por IA: %v", err)
		h.sendErrorResponse(w, "Não foi possível classificar a intenção")
		return
	}

	service := &models.Service{
		ID:   aiService.ID,
		Name: aiService.Name,
	}

	response := models.FindServiceResponse{
		Success: true,
		Data:    service,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ServiceHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Middleware de recuperação de panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recuperado no HealthCheck: %v\n%s", r, debug.Stack())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "error"})
		}
	}()

	response := models.HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ServiceHandler) sendErrorResponse(w http.ResponseWriter, message string) {
	// Garantir que sempre retornamos status 200
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Erro ao enviar resposta de erro: %v", r)
		}
	}()

	response := models.FindServiceResponse{
		Success: false,
		Error:   message,
	}

	// Limpar qualquer header que possa ter sido definido anteriormente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Erro ao codificar resposta de erro: %v", err)
		// Fallback: tentar escrever uma resposta simples
		w.Write([]byte(`{"success":false,"error":"Erro interno"}`))
	}
}
