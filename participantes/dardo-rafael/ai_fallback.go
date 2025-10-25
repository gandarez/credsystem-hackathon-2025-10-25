package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIClient representa um cliente para a API da OpenRouter
type AIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	model      string
}

// NewAIClient cria um novo cliente AI
func NewAIClient() *AIClient {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("WARNING: OPENROUTER_API_KEY not set, AI fallback will not work")
	}

	return &AIClient{
		baseURL: "https://openrouter.ai/api/v1",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		model: "mistralai/mistral-7b-instruct:free", // Modelo gratuito
	}
}

type openRouterRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// buildPrompt constrói o prompt para a IA com as instruções precisas
func (c *AIClient) buildPrompt(intentText string, services map[int]string) string {
	var sb strings.Builder

	sb.WriteString("Você é um assistente de classificação de intenções. ")
	sb.WriteString("Dado a solicitação do usuário, retorne APENAS o ID numérico do serviço correspondente. ")
	sb.WriteString("Não adicione nenhuma outra palavra, apenas o número.\n\n")
	sb.WriteString("Serviços disponíveis:\n")

	for id, name := range services {
		sb.WriteString(fmt.Sprintf("%d - %s\n", id, name))
	}

	sb.WriteString(fmt.Sprintf("\nSolicitação do usuário: \"%s\"\n\n", intentText))
	sb.WriteString("ID do Serviço:")

	return sb.String()
}

// ClassifyWithAI usa a API da OpenRouter para classificar a intenção
func (c *AIClient) ClassifyWithAI(ctx context.Context, intentText string, services map[int]string) (*APIResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY not configured")
	}

	prompt := c.buildPrompt(intentText, services)

	reqBody := openRouterRequest{
		Model: c.model,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var aiResp openRouterResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(aiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in AI response")
	}

	content := strings.TrimSpace(aiResp.Choices[0].Message.Content)

	// Tentar extrair o ID do serviço
	var serviceID int
	if _, err := fmt.Sscanf(content, "%d", &serviceID); err != nil {
		return nil, fmt.Errorf("failed to parse service ID from AI response: %s", content)
	}

	// Verificar se o ID é válido
	serviceName, exists := services[serviceID]
	if !exists {
		return nil, fmt.Errorf("AI returned invalid service ID: %d", serviceID)
	}

	return &APIResponse{
		ServiceID:   serviceID,
		ServiceName: serviceName,
	}, nil
}
