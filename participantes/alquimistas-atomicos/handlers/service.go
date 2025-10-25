package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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
	var req models.FindServiceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Erro ao decodificar requisição", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		h.sendErrorResponse(w, "Campo 'intent' é obrigatório", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	aiService, err := h.openRouterClient.ClassifyIntent(ctx, req.Intent)
	if err != nil {
		log.Printf("Erro na classificação por IA: %v", err)
		h.sendErrorResponse(w, "Não foi possível classificar a intenção", http.StatusInternalServerError)
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
	response := models.HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ServiceHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := models.FindServiceResponse{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
