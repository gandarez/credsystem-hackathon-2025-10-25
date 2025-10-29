package main

import (
	"desviadores-de-deadlock/pkg/middleware"
	"desviadores-de-deadlock/pkg/service/health"
	"desviadores-de-deadlock/pkg/service/intent"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	port := "18020"

	server := createServer(port)

	printStartupMessages(port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

func createServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", middleware.LoggingMiddleware(health.HealthHandler))
	mux.HandleFunc("/api/intent", middleware.LoggingMiddleware(intent.IntentHandler))

	return &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func printStartupMessages(port string) {
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Health check available at http://localhost:%s/healthz\n", port)
	fmt.Printf("Intent endpoint available at http://localhost:%s/intent\n", port)
}
