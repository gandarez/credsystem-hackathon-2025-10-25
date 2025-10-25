package main

import (
	"defensoresdefer/cmd/api/openrouter"
	"defensoresdefer/cmd/configs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type IntentUser struct {
	Intent string
}

type Response struct {
	success     bool
	error       string
	dataService DataService
}

type DataService struct {
	service_id   int
	service_name string
}

func main() {
	r := chi.NewRouter()
	//r.Use(LogRequest)
	r.Use(middleware.Logger)

	//r.Post("/users", )
	//r.Post("/users/generate_token", userHandler.GetJwt)
	r.Get("/api/healthz/*", ConsultaHealthz)
	r.Post("/api/find-service", FindService)

	http.ListenAndServe(":8000", r)

}

func ConsultaHealthz(w http.ResponseWriter, r *http.Request) {
	// Consulta serviço
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func FindService(w http.ResponseWriter, r *http.Request) {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("failed to load configs: %v", err)
	}
	var intent IntentUser
	fmt.Printf("Recebido request para FindService %s", intent.Intent)
	err = json.NewDecoder(r.Body).Decode(&intent)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Lógica simples para determinar o serviço com base na intenção
	var response Response

	var client openrouter.Client
	client = *openrouter.NewClient("https://openrouter.ai/api/v1",
		openrouter.WithAuth(configs.OPENROUTER_API_KEY),
	)
	fmt.Printf("Client OpenRouter criado %v", client)

	dataResp, err := client.ChatCompletion(r.Context(), intent.Intent)
	if err != nil {
		http.Error(w, "Error processing request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Resposta recebida do OpenRouter %v", dataResp)

	response = Response{
		success: true,
		dataService: DataService{
			service_id:   int(dataResp.ServiceID),
			service_name: dataResp.ServiceName,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
