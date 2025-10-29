package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

var (
	openRouterClient *OpenRouterClient
	logger           *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	openRouterClient = NewOpenRouterClient()
}

func FindServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		logger.Warn("method not allowed", "method", r.Method, "endpoint", "/api/find-service")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FindServiceResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req FindServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid request body", "error", err)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FindServiceResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if req.Intent == "" {
		logger.Warn("intent is required")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FindServiceResponse{
			Success: false,
			Error:   "intent is required",
		})
		return
	}

	serviceData, err := openRouterClient.FindServiceByIntent(r.Context(), req.Intent)
	if err != nil {
		logger.Error("failed to find service", "intent", req.Intent, "error", err)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FindServiceResponse{
			Success: false,
			Error:   "failed to process intent: " + err.Error(),
		})
		return
	}

	logger.Info("service found", "intent", req.Intent, "service_id", serviceData.ServiceID, "service_name", serviceData.ServiceName)

	response := FindServiceResponse{
		Success: true,
		Data:    serviceData,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status: "ok",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
