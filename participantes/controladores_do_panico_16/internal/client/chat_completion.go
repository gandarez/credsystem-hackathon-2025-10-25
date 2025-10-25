package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type (
	OpenRouterRequest struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}

	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	OpenRouterResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage,omitempty"`
		Error *struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error,omitempty"`
	}

	DataResponse struct {
		ServiceID   uint8  `json:"service_id"`
		ServiceName string `json:"service_name"`
	}
)

func (c *Client) ChatCompletion(ctx context.Context, intent string) (*DataResponse, error) {
	// Usar prompt com exemplos dos intents carregados
	systemPrompt := BuildPromptWithExamples()
	url := c.baseURL + "/chat/completions"

	requestBody := OpenRouterRequest{
		Model: "openai/gpt-4o-mini", // Modelo para classificação
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: intent,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("DEBUG: Authorization header: %s\n", req.Header.Get("Authorization"))

	resp, err := c.Do(ctx, req)
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

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v. body: %s", err, string(body))
	}

	// Verificar se há erro na resposta da API
	if openRouterResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s (code: %s)", openRouterResp.Error.Message, openRouterResp.Error.Code)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Log de uso de tokens
	if openRouterResp.Usage != nil {
		fmt.Printf("\n[TOKEN USAGE]\n")
		fmt.Printf("  Prompt Tokens:     %d\n", openRouterResp.Usage.PromptTokens)
		fmt.Printf("  Completion Tokens: %d\n", openRouterResp.Usage.CompletionTokens)
		fmt.Printf("  Total Tokens:      %d\n\n", openRouterResp.Usage.TotalTokens)
	}

	content := strings.TrimSpace(openRouterResp.Choices[0].Message.Content)

	// Tentar extrair JSON se vier com texto extra
	if strings.Contains(content, "{") && strings.Contains(content, "}") {
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}") + 1
		content = content[start:end]
	}

	var dataRes DataResponse
	if err := json.Unmarshal([]byte(content), &dataRes); err != nil {
		return nil, fmt.Errorf("error unmarshaling data response: %v. content: %s", err, content)
	}

	// Validação rigorosa: service_id deve estar entre 1 e 16
	if dataRes.ServiceID < 1 || dataRes.ServiceID > 16 {
		return nil, fmt.Errorf("invalid service_id: %d (must be between 1 and 16)", dataRes.ServiceID)
	}

	// Validação: service_name não pode estar vazio
	if dataRes.ServiceName == "" {
		return nil, fmt.Errorf("service_name cannot be empty")
	}

	return &dataRes, nil
}
