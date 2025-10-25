package handler

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//go:embed prompts/system_prompt.txt
var systemPrompt string

const defaultModel = "meta-llama/llama-4-scout"

var serviceMap = map[int]string{
	1:  "Consulta Limite / Vencimento do cartão / Melhor dia de compra",
	2:  "Segunda via de boleto de acordo",
	3:  "Segunda via de Fatura",
	4:  "Status de Entrega do Cartão",
	5:  "Status de cartão",
	6:  "Solicitação de aumento de limite",
	7:  "Cancelamento de cartão",
	8:  "Telefones de seguradoras",
	9:  "Desbloqueio de Cartão",
	10: "Esqueceu senha / Troca de senha",
	11: "Perda e roubo",
	12: "Consulta do Saldo Conta do Mais",
	13: "Pagamento de contas",
	14: "Reclamações",
	15: "Atendimento humano",
	16: "Token de proposta",
}

type OpenRouterClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
	Provider *openRouterProvider `json:"provider,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterProvider struct {
	Order             []string `json:"order,omitempty"`
	AllowFallbacks    *bool    `json:"allow_fallbacks,omitempty"`
	RequireParameters *bool    `json:"require_parameters,omitempty"`
	DataCollection    string   `json:"data_collection,omitempty"`
	Sort              string   `json:"sort,omitempty"`
}

type openRouterResponse struct {
	ID      string             `json:"id"`
	Choices []openRouterChoice `json:"choices"`
	Error   *openRouterError   `json:"error,omitempty"`
}

type openRouterChoice struct {
	Message openRouterMessage `json:"message"`
	Index   int               `json:"index"`
}

type openRouterError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func NewOpenRouterClient() *OpenRouterClient {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		panic("OPENROUTER_API_KEY environment variable is required")
	}

	logger.Info("openrouter client initialized", "model", defaultModel, "provider", "Groq")

	return &OpenRouterClient{
		apiKey: apiKey,
		model:  defaultModel,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OpenRouterClient) FindServiceByIntent(ctx context.Context, intent string) (*FindServiceData, error) {
	reqBody := openRouterRequest{
		Model: c.model,
		Messages: []openRouterMessage{
			{
				Role:    "system",
				Content: strings.TrimSpace(systemPrompt),
			},
			{
				Role:    "user",
				Content: intent,
			},
		},
		Provider: &openRouterProvider{
			Sort: "latency",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("failed to marshal request", "error", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "Esquadrao do Scheduler")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("failed to send request", "error", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("openrouter returned non-200 status", "status", resp.StatusCode, "body", string(bodyBytes))
		return nil, fmt.Errorf("openrouter returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	logger.Info("openrouter api response", "body", string(bodyBytes))

	var orResp openRouterResponse
	if err := json.Unmarshal(bodyBytes, &orResp); err != nil {
		logger.Error("failed to unmarshal openrouter response", "error", err, "body", string(bodyBytes))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if orResp.Error != nil {
		logger.Error("openrouter returned error", "error", orResp.Error.Message, "type", orResp.Error.Type)
		return nil, fmt.Errorf("openrouter error: %s", orResp.Error.Message)
	}

	if len(orResp.Choices) == 0 {
		logger.Error("no response received from model")
		return nil, fmt.Errorf("no response received from model")
	}

	content := orResp.Choices[0].Message.Content
	id, err := strconv.Atoi(content)
	if err != nil {
		logger.Error("service not found in map", "service_id", content)
		return nil, fmt.Errorf("service with ID %s not found", content)

	}

	serviceName, exists := serviceMap[id]
	if !exists {
		logger.Error("service not found in map", "service_id", content)
		return nil, fmt.Errorf("service with ID %s not found", content)
	}

	return &FindServiceData{
		ServiceID:   id,
		ServiceName: serviceName,
	}, nil
}
