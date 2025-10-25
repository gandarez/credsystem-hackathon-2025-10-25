package main

import (
	"defensoresdefer/cmd/api/openrouter"
	"defensoresdefer/cmd/configs"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type IntentUser struct {
	Intent string `json:"intent"`
}

type Response struct {
	Success bool         `json:"success"`
	Error   string       `json:"error,omitempty"`
	Data    *DataService `json:"data,omitempty"`
}

type DataService struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/api/healthz", ConsultaHealthz)
	r.Post("/api/find-service", FindService)

	http.ListenAndServe(":8080", r)
}

func ConsultaHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func FindService(w http.ResponseWriter, r *http.Request) {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load configs: %v", err), http.StatusInternalServerError)
		return
	}

	var intent IntentUser
	if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	client := openrouter.NewClient(
		"https://openrouter.ai/api/v1",
		openrouter.WithAuth(cfg.OPENROUTER_API_KEY),
	)

	dataResp, err := client.ChatCompletion(r.Context(), intent.Intent)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
		return
	}

	response := Response{
		Success: true,
		Data: &DataService{
			ServiceID:   int(dataResp.ServiceID),
			ServiceName: dataResp.ServiceName,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
