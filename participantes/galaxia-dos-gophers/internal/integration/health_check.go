package integration

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Success bool `json:"success"`
}

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{Success: true}
	json.NewEncoder(w).Encode(response)
}
