package intent

import (
	"context"
	"desviadores-de-deadlock/pkg/model"
	"desviadores-de-deadlock/pkg/openrouter"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func IntentHandler(w http.ResponseWriter, r *http.Request) {
	// Fazer parse da requisição
	intentRequest, err := parseIntentRequest(r)
	if err != nil {
		response := model.Reponse{
			Success: false,
			Data:    nil,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fazer a chamada para o OpenRouter
	openrouterClient := openrouter.NewClient("https://openrouter.ai/api/v1", openrouter.WithAuth(os.Getenv("OPENROUTER_API_KEY")))
	data, err := openrouterClient.ChatCompletion(context.Background(), intentRequest.Intent)

	response := model.Reponse{
		Success: true,
		Data:    nil,
		Error:   "",
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else if data != nil {

		response.Success = true
		response.Data = data

		if data.ServiceID == nil {
			response.Success = false
			response.Data = nil
			response.Error = "Serviço não identificado"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// parseIntentRequest faz o parse do body da requisição e valida os dados
func parseIntentRequest(r *http.Request) (*model.IntentRequest, error) {
	// Ler o corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Fazer parse do JSON
	var intentRequest model.IntentRequest
	if err := json.Unmarshal(body, &intentRequest); err != nil {
		return nil, err
	}

	// Validar se o intent não está vazio
	if intentRequest.Intent == "" {
		return nil, fmt.Errorf("campo 'intent' é obrigatório")
	}

	return &intentRequest, nil
}
