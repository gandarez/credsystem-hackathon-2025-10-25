package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// ServiceExample represents a service with its intent and embedding
type ServiceExample struct {
	ServiceID   int       `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Intent      string    `json:"intent"`
	Embedding   []float64 `json:"embedding"`
}

// EmbeddingRequest represents the request to OpenRouter embeddings API
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// EmbeddingResponse represents the response from OpenRouter embeddings API
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// ServiceEmbeddings represents the complete embedding dataset
type ServiceEmbeddings struct {
	Services []ServiceExample `json:"services"`
	Metadata struct {
		GeneratedAt string `json:"generated_at"`
		TotalCount  int    `json:"total_count"`
		Model       string `json:"model"`
	} `json:"metadata"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <csv_file_path>")
	}

	csvPath := os.Args[1]
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	// Load services from CSV
	services, err := loadServicesFromCSV(csvPath)
	if err != nil {
		log.Fatal("Failed to load services:", err)
	}

	log.Printf("Loaded %d service examples from %s", len(services), csvPath)

	// Generate embeddings
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	embeddings := &ServiceEmbeddings{
		Services: make([]ServiceExample, 0, len(services)),
		Metadata: struct {
			GeneratedAt string `json:"generated_at"`
			TotalCount  int    `json:"total_count"`
			Model       string `json:"model"`
		}{
			GeneratedAt: time.Now().Format(time.RFC3339),
			TotalCount:  len(services),
			Model:       "text-embedding-ada-002",
		},
	}

	log.Printf("Generating embeddings for %d services...", len(services))

	for i, service := range services {
		embedding, err := getEmbedding(client, apiKey, service.Intent)
		if err != nil {
			log.Printf("Failed to get embedding for service %d: %v", service.ServiceID, err)
			continue
		}

		service.Embedding = embedding
		embeddings.Services = append(embeddings.Services, service)

		if (i+1)%10 == 0 {
			log.Printf("Generated embeddings for %d/%d services", i+1, len(services))
		}

		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	// Save embeddings to JSON file
	outputFile := "service_embeddings.json"
	if err := saveEmbeddings(embeddings, outputFile); err != nil {
		log.Fatal("Failed to save embeddings:", err)
	}

	log.Printf("Successfully generated and saved %d embeddings to %s", len(embeddings.Services), outputFile)
}

func loadServicesFromCSV(filename string) ([]ServiceExample, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';' // Use semicolon as delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	var services []ServiceExample

	// Skip header row
	for i, record := range records[1:] {
		if len(record) < 3 {
			continue
		}

		serviceID, err := strconv.Atoi(record[0])
		if err != nil {
			log.Printf("Warning: Invalid service ID at line %d: %s", i+2, record[0])
			continue
		}

		service := ServiceExample{
			ServiceID:   serviceID,
			ServiceName: record[1],
			Intent:      record[2],
		}

		services = append(services, service)
	}

	return services, nil
}

func getEmbedding(client *http.Client, apiKey, text string) ([]float64, error) {
	url := "https://openrouter.ai/api/v1/embeddings"

	requestBody := EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: text,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v. body: %s", err, string(body))
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embeddingResp.Data[0].Embedding, nil
}

func saveEmbeddings(embeddings *ServiceEmbeddings, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(embeddings); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	return nil
}
