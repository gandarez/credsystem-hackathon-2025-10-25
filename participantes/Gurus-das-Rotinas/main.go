package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"gurus-das-rotinas/api/client/openrouter"

	"github.com/gorilla/mux"
)

type FindServiceRequest struct {
	Intent string `json:"intent"`
}

type FindServiceResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ServiceID   int    `json:"service_id"`
		ServiceName string `json:"service_name"`
	} `json:"data"`
	Error string `json:"error"`
}

func main() {
	// Get OpenRouter API key from environment variable
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	// Initialize OpenRouter client
	openRouterClient := openrouter.NewClient(
		"https://openrouter.ai/api/v1",
		openrouter.WithAuth(apiKey),
	)

	// Create router
	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	// Health check endpoint (alternative)
	r.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	}).Methods("GET")

	// Find service endpoint
	r.HandleFunc("/api/find-service", func(w http.ResponseWriter, r *http.Request) {
		handleFindService(w, r, openRouterClient)
	}).Methods("POST")

	// Add CORS middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler(r)))
}

func handleFindService(w http.ResponseWriter, r *http.Request, client *openrouter.Client) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req FindServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := FindServiceResponse{
			Success: false,
			Error:   "Invalid JSON",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate intent
	if req.Intent == "" {
		response := FindServiceResponse{
			Success: false,
			Error:   "intent field is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Call OpenRouter API
	ctx := context.Background()
	result, err := client.ChatCompletion(ctx, req.Intent)
	if err != nil {
		log.Printf("Error calling OpenRouter: %v", err)
		response := FindServiceResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to process intent: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if we got valid service data
	if result.ServiceID == 0 || result.ServiceName == "" {
		response := FindServiceResponse{
			Success: false,
			Error:   "Could not determine service from intent",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return successful response
	response := FindServiceResponse{
		Success: true,
		Data: struct {
			ServiceID   int    `json:"service_id"`
			ServiceName string `json:"service_name"`
		}{
			ServiceID:   int(result.ServiceID),
			ServiceName: result.ServiceName,
		},
		Error: "",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
