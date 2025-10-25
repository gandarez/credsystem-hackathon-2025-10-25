package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

var Intent struct {
	Intent string `json:"intent"`
}

func main() {
	/*configs, err := LoadConfigs(".")
	if err != nil {
		log.Fatalf("failed to load configs: %v", err)
	}*/

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
	/*var intent Intent
	err := json.NewDecoder(r.Body).Decode(&intent)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	fmt.Println(intent.Intent)
	// Lógica simples para determinar o serviço com base na intenção
	*/
}
