package cmd

import (
	"log"
	"net/http"
	"os"
	"participantes/galaxia-dos-gophers/internal/integration"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/find-service", integration.FindServiceHandler)
	http.HandleFunc("/api/healthz", integration.HealthzHandler)

	log.Printf("Servidor rodando na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
