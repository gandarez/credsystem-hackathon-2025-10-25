package integration

import (
	"encoding/json"
	"net/http"
	"participantes/galaxia-dos-gophers/internal/dto"
	"participantes/galaxia-dos-gophers/internal/util"
)

func FindServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req *dto.OpenRouterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	serviceID, serviceName := util.DetermineService(req)

	response := dto.DataResponse{
		ServiceID:   serviceID,
		ServiceName: serviceName,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
