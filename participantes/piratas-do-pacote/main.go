package main

import (
	"log"
	"os"

	"github.com/piratas-do-pacote/adapter/in"
	"github.com/piratas-do-pacote/adapter/out"
	"github.com/piratas-do-pacote/adapter/out/client"
	"github.com/piratas-do-pacote/core"
)

func main() {

	key := os.Getenv("OPENROUTER_API_KEY")
	if key == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable not set")
	}
	openRouterClient := client.NewOpenRouterClient(key)
	inferer := out.NewAiInferer(openRouterClient)
	useCase := core.NewInferenceUseCase(inferer)

	server := in.NewHttpHandler(useCase)
	server.ListenAndServe()
}
