package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
)

// ServiceExample represents a service with its intent and embedding
type ServiceExample struct {
	ServiceID   int       `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Intent      string    `json:"intent"`
	Embedding   []float64 `json:"embedding"`
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

// EmbeddingClassifier handles classification using pre-computed embeddings
type EmbeddingClassifier struct {
	services []ServiceExample
}

// NewEmbeddingClassifier creates a new embedding classifier
func NewEmbeddingClassifier(embeddingsFile string) (*EmbeddingClassifier, error) {
	file, err := os.Open(embeddingsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open embeddings file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read embeddings file: %v", err)
	}

	var embeddings ServiceEmbeddings
	if err := json.Unmarshal(data, &embeddings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embeddings: %v", err)
	}

	return &EmbeddingClassifier{
		services: embeddings.Services,
	}, nil
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ClassifyWithEmbedding classifies intent using embedding similarity
func (ec *EmbeddingClassifier) ClassifyWithEmbedding(inputEmbedding []float64, threshold float64) (*DataResponse, float64) {
	var bestMatch *ServiceExample
	var bestScore float64

	// Calculate cosine similarity with all services
	for i := range ec.services {
		if len(ec.services[i].Embedding) == 0 {
			continue // Skip services without embeddings
		}

		score := CosineSimilarity(inputEmbedding, ec.services[i].Embedding)
		if score > bestScore {
			bestScore = score
			bestMatch = &ec.services[i]
		}
	}

	if bestMatch == nil || bestScore < threshold {
		return nil, bestScore
	}

	return &DataResponse{
		ServiceID:   uint8(bestMatch.ServiceID),
		ServiceName: bestMatch.ServiceName,
	}, bestScore
}

// GetEmbedding gets embedding for text using OpenRouter API
func (c *Client) GetEmbedding(text string) ([]float64, error) {
	url := c.baseURL + "/embeddings"

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

	resp, err := c.Do(context.Background(), req)
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
