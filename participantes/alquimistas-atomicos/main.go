package main

import (
	"bufio"
	"ivr-service/client"
	"ivr-service/handlers"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}

	return scanner.Err()
}

func main() {
	// Carregar variáveis de ambiente do arquivo config.env
	if err := loadEnvFile("config.env"); err != nil {
		log.Printf("Aviso: não foi possível carregar config.env: %v", err)
	}

	openRouterClient := client.NewOpenRouterClient()

	serviceHandler := handlers.NewServiceHandler(openRouterClient)

	r := mux.NewRouter()
	r.HandleFunc("/api/find-service", serviceHandler.FindService).Methods("POST")
	r.HandleFunc("/api/healthz", serviceHandler.HealthCheck).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Servidor iniciado na porta %s", port)
	log.Fatal(server.ListenAndServe())
}
