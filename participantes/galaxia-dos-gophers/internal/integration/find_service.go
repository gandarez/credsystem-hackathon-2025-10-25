package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"participantes/galaxia-dos-gophers/internal/dto"
	"participantes/galaxia-dos-gophers/internal/util"
	"participantes/galaxia-dos-gophers/openrouter"
	"time"
)

type FindServiceRequest struct {
	Intent string `json:"intent"`
}

type ApiResponse struct {
	Success bool              `json:"success"`
	Data    *dto.DataResponse `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

func FindServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FindServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ApiResponse{Success: false, Error: "invalid JSON"})
		return
	}
	defer r.Body.Close()

	baseURL := "https://openrouter.ai/api/v1"
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	client := openrouter.NewClient(baseURL, apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	intent, err := client.ChatCompletion(ctx, req.Intent)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ApiResponse{Success: false, Error: "failed to call OpenRouter"})
		return
	}

	serviceID, serviceName := util.DetermineService(intent)

	resp := ApiResponse{
		Success: true,
		Data: &dto.DataResponse{
			ServiceID:   serviceID,
			ServiceName: serviceName,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
