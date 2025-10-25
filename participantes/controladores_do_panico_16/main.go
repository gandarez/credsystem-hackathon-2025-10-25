package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type (
	Response struct {
		Success bool   `json:"success"`
		Data    *Data  `json:"data,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	Data struct {
		ServiceID   uint8  `json:"service_id"`
		ServiceName string `json:"service_name"`
	}

	IntentRequest struct {
		Intent string `json:"intent"`
	}
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/healthz", healthCheckHandler)
	http.HandleFunc("/api/find-service", findServiceHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func findServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		sendErrorResponse(w, "Intent is required", http.StatusBadRequest)
		return
	}

	// TODO: implementar OpenRouter para classificação de intenções
	sendErrorResponse(w, "Not implemented yet", http.StatusNotImplemented)
}

func sendErrorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	})
}
