package main

import (
	"log/slog"
	"net/http"
	"os"
	handler "ura-ai/handlers"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	http.HandleFunc("/api/find-service", handler.FindServiceHandler)
	http.HandleFunc("/api/healthz", handler.HealthHandler)

	port := ":8080"

	logger.Info("starting server", "port", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
