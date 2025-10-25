package main

import (
	"ivr-service/client"
	"ivr-service/handlers"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
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
