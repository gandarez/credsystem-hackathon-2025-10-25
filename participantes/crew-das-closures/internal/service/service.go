package service

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dyammarcano/crew-das-closures/internal/client/openrouter"
	"github.com/dyammarcano/crew-das-closures/internal/core"
	"github.com/spf13/cobra"
)

func Service(_ *cobra.Command, _ []string) error {
	// Read environment variables
	openRouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterKey == "" {
		return fmt.Errorf("environment variable OPENROUTER_API_KEY is required")
	}

	router := http.NewServeMux()

	server := http.Server{
		Addr:    ":8080",
		Handler: forceStatusOK(router),
	}

	urlStr := "https://openrouter.ai/api/v1"
	opts := openrouter.WithAuth(openRouterKey)

	aks, err := core.NewCore(urlStr, opts)
	if err != nil {
		return fmt.Errorf("failed to initialize core: %w", err)
	}

	router.HandleFunc("GET /api/health", healthHandler)
	router.HandleFunc("POST /api/find-service", findServiceHandler(aks))

	return server.ListenAndServe()
}
