package main

import (
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Determinar o caminho do arquivo CSV
	csvPath := os.Getenv("INTENTS_CSV_PATH")
	if csvPath == "" {
		// Caminho padrão relativo ao projeto
		csvPath = filepath.Join("..", "..", "assets", "intents_pre_loaded.csv")
	}

	log.Printf("Loading intents from: %s", csvPath)

	// Carregar intenções do CSV
	intents, err := loadIntentsFromCSV(csvPath)
	if err != nil {
		log.Fatalf("Failed to load intents: %v", err)
	}

	log.Printf("Loaded %d intents from CSV", len(intents))

	// Criar serviço KNN com pipeline NLP
	log.Println("Initializing NLP-based KNN service...")
	knnService, err := NewKNNService()
	if err != nil {
		log.Fatalf("Failed to create KNN service: %v", err)
	}

	// Carregar e treinar com as intents
	log.Println("Training NLP pipeline with intents...")
	if err := knnService.LoadIntents(intents); err != nil {
		log.Fatalf("Failed to load intents into service: %v", err)
	}
	log.Printf("NLP pipeline ready - Vocabulary size: %d", knnService.VocabularySize())

	// Criar cliente AI para fallback
	aiClient := NewAIClient()
	// Configurar intents no cliente AI para melhorar os prompts
	aiClient.SetIntents(intents)
	log.Println("AI client configured with intents")

	// Criar servidor
	server := NewServer(knnService, aiClient, intents)

	// Obter porta do ambiente ou usar padrão
	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
	}

	// Iniciar servidor
	log.Printf("Starting server on port %s", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
