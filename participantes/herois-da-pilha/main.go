package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"herois-da-pilha/handler"
)

func main() {
	// 1. Inicializar o Handler (que inicializa o serviço de IA e o cache)
	apiHandler := handler.NewAPIHandler()

	// 2. Configurar o Roteador (usando o ServeMux da biblioteca padrão)
	mux := http.NewServeMux()

	// O http.ServeMux usa HandleFunc
	mux.HandleFunc("/api/find-service", apiHandler.FindServiceHandler)
	mux.HandleFunc("/api/healthz", apiHandler.HealthCheckHandler)

	// 3. Ler a porta da variável de ambiente
	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
		fmt.Printf("AVISO: Variável de ambiente 'PORT' não definida, usando default: %s\n", port)
	}
	addr := fmt.Sprintf(":%s", port)

	// 4. Configurar e Iniciar o Servidor HTTP (Otimizado)
	// Usar http.Server para definir timeouts, o que é uma boa prática
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
		// Timeouts razoáveis para um serviço de baixa latência
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Serviço Credsystem/Golang SP (net/http) rodando na porta %s...\n", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}
