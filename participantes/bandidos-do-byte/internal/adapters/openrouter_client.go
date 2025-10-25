package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bandidos_do_byte/api/internal/domain"
)

const (
	OpenRouterAPIURL = "https://openrouter.ai/api/v1/chat/completions"
	MistralModel     = "mistralai/mistral-7b-instruct"
)

type OpenRouterClient struct {
	apiKey     string
	httpClient *http.Client
}

type openRouterRequest struct {
	Model    string          `json:"model"`
	Messages []openRouterMsg `json:"messages"`
}

type openRouterMsg struct {
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

type classificationResult struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *OpenRouterClient) ClassifyIntent(request domain.IntentClassificationRequest) (*domain.IntentClassificationResponse, error) {
	prompt := c.buildPrompt(request)

	reqBody := openRouterRequest{
		Model: MistralModel,
		Messages: []openRouterMsg{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", OpenRouterAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/bandidos_do_byte")
	req.Header.Set("X-Title", "Bandidos do Byte - Intent Classifier")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenRouter API returned status %d: %s", resp.StatusCode, string(body))
	}

	var openRouterResp openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenRouter")
	}

	result, err := c.parseResponse(openRouterResp.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	}

	return &domain.IntentClassificationResponse{
		ServiceID:   result.ServiceID,
		ServiceName: result.ServiceName,
		Confidence:  0.95, // Mistral doesn't return confidence, using default
	}, nil
}

func (c *OpenRouterClient) buildPrompt(request domain.IntentClassificationRequest) string {
	var sb strings.Builder

	sb.WriteString("Você é um assistente especializado em classificar intenções de clientes de um banco/financeira.\n\n")
	sb.WriteString("Abaixo está uma lista de serviços disponíveis com exemplos de intenções de clientes:\n\n")

	// Group examples by service
	serviceMap := make(map[int][]string)
	serviceNames := make(map[int]string)

	for _, example := range request.Examples {
		serviceMap[example.ServiceID] = append(serviceMap[example.ServiceID], example.Intent)
		serviceNames[example.ServiceID] = example.ServiceName
	}

	// Build context with examples
	for serviceID := 1; serviceID <= 16; serviceID++ {
		if name, exists := serviceNames[serviceID]; exists {
			sb.WriteString(fmt.Sprintf("Serviço ID %d - %s:\n", serviceID, name))
			examples := serviceMap[serviceID]
			for i, intent := range examples {
				if i < 3 { // Limit to 3 examples per service to keep prompt concise
					sb.WriteString(fmt.Sprintf("  - %s\n", intent))
				}
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(fmt.Sprintf("A intenção do cliente é: \"%s\"\n\n", request.UserIntent))
	sb.WriteString("Analise a intenção do cliente e retorne APENAS um JSON no seguinte formato (sem nenhum texto adicional):\n")
	sb.WriteString(`{"service_id": <número>, "service_name": "<nome do serviço>"}`)
	sb.WriteString("Caso intenção não seja compatível com um um serviço mapeado retorne {\"service_id\": 0, \"service_name\": \"Contate a URA\"}")

	return sb.String()
}

func (c *OpenRouterClient) parseResponse(content string) (*classificationResult, error) {
	// Remove markdown code blocks if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result classificationResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse classification result: %w", err)
	}

	return &result, nil
}

func (c *OpenRouterClient) HealthCheck() error {
	// Simple check to verify API key is set
	if c.apiKey == "" {
		return fmt.Errorf("OpenRouter API key not configured")
	}
	return nil
}
