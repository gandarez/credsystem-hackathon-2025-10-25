package modelia

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	openai "github.com/sashabaranov/go-openai"
)

// Estrutura que vamos gerar
type IntentEmbedding struct {
	Intent      string    `json:"intent"`
	ServiceID   int       `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Embedding   []float32 `json:"embedding"`
}

func Embed() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Defina OPENROUTER_API_KEY no .env")
	}

	client := openai.NewClient(apiKey)

	file, err := os.Open("assets/intents_pre_loaded.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var embeddings []IntentEmbedding

	for i, row := range rows {
		if i == 0 {
			continue
		}
		serviceID := row[0]
		serviceName := row[1]
		intent := row[2]

		resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
			Model: openai.AdaEmbeddingV2,
			Input: []string{intent},
		})
		if err != nil {
			log.Printf("Erro no embedding para '%s': %v\n", intent, err)
			continue
		}

		vec := make([]float32, len(resp.Data[0].Embedding))
		for j, v := range resp.Data[0].Embedding {
			vec[j] = float32(v)
		}

		embeddings = append(embeddings, IntentEmbedding{
			Intent:      intent,
			ServiceID:   atoi(serviceID),
			ServiceName: serviceName,
			Embedding:   vec,
		})
	}

	outFile, err := os.Create("embeddings.json")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	enc := json.NewEncoder(outFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(embeddings); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Embeddings gerados em embeddings.json")
}

func atoi(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}
