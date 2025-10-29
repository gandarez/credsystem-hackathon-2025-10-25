package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bandidos_do_byte/api/internal/domain"
	"github.com/bandidos_do_byte/api/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	serviceFinder service.ServiceFinder
}

func NewHandler(serviceFinder service.ServiceFinder) *Handler {
	return &Handler{
		serviceFinder: serviceFinder,
	}
}

// RegisterRoutes registra todas as rotas da aplicação
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/api/find-service", h.FindService)
	r.Get("/api/healthz", h.HealthCheck)
}

// FindService handler para o endpoint POST /api/find-service
func (h *Handler) FindService(w http.ResponseWriter, r *http.Request) {
	var req domain.FindServiceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	serviceData, err := h.serviceFinder.FindService(req.Intent)
	if err != nil {
		// Se for erro de "serviço não encontrado", retorna 200 com success: false
		if errors.Is(err, domain.ErrNoServiceFound) {
			response := domain.FindServiceResponse{
				Success: false,
				Error:   "No suitable service found for your request",
			}
			h.sendJSONResponse(w, response, http.StatusOK)
			return
		}

		// Outros erros retornam 500
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := domain.FindServiceResponse{
		Success: true,
		Data:    serviceData,
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// HealthCheck handler para o endpoint GET /api/healthz
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := domain.HealthResponse{
		Status: h.serviceFinder.HealthCheck(),
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// sendJSONResponse envia uma resposta JSON
func (h *Handler) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendErrorResponse envia uma resposta de erro
func (h *Handler) sendErrorResponse(w http.ResponseWriter, errorMsg string, statusCode int) {
	response := domain.FindServiceResponse{
		Success: false,
		Error:   errorMsg,
	}
	h.sendJSONResponse(w, response, statusCode)
}
