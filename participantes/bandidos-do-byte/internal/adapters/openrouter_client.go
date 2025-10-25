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
	GPT4oMiniModel   = "openai/gpt-4o-mini"
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
		Model: GPT4oMiniModel,
		Messages: []openRouterMsg{
			{Role: "user", Content: prompt},
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter API returned status %d: %s", resp.StatusCode, string(body))
	}

	bodyStr := strings.TrimSpace(string(body))
	if !strings.HasPrefix(bodyStr, "{") {
		preview := bodyStr
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf("invalid response format (not JSON): %s", preview)
	}

	var openRouterResp openRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		preview := bodyStr
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf("failed to decode response: %w (body: %s)", err, preview)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenRouter")
	}

	result, err := c.parseResponse(openRouterResp.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("%w (content: %s)", err, openRouterResp.Choices[0].Message.Content)
	}

	if result.ServiceID == 0 {
		return nil, domain.ErrNoServiceFound
	}

	return &domain.IntentClassificationResponse{
		ServiceID:   result.ServiceID,
		ServiceName: result.ServiceName,
		Confidence:  0.95,
	}, nil
}

func (c *OpenRouterClient) buildPrompt(request domain.IntentClassificationRequest) string {
	var sb strings.Builder

	sb.WriteString("Você é um assistente especializado em classificar intenções de clientes de um banco/financeira.\n\n")
	sb.WriteString("Serviços disponíveis com exemplos:\n\n")

	serviceMap := make(map[int][]string)
	serviceNames := make(map[int]string)

	for _, example := range request.Examples {
		serviceMap[example.ServiceID] = append(serviceMap[example.ServiceID], example.Intent)
		serviceNames[example.ServiceID] = example.ServiceName
	}

	for serviceID := 1; serviceID <= 16; serviceID++ {
		if name, exists := serviceNames[serviceID]; exists {
			sb.WriteString(fmt.Sprintf("Serviço ID %d - %s:\n", serviceID, name))
			for _, intent := range serviceMap[serviceID] {
				sb.WriteString(fmt.Sprintf("  - %s\n", intent))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("REGRAS (prioridade):\n")
	sb.WriteString("1. 'quando fecha', 'vencimento', 'melhor dia' + 'fatura' → 1 (Vencimento)\n")
	sb.WriteString("2. 'pagar fatura', 'quitar fatura' → 13 (Pagamento)\n")
	sb.WriteString("3. 'segunda via', 'ver fatura', 'enviar fatura' → 3 (Segunda via)\n")
	sb.WriteString("4. Boleto negociação/acordo → 2\n")
	sb.WriteString("5. Status geral cartão → 5\n")
	sb.WriteString("6. Status entrega → 4\n")
	sb.WriteString("7. Saldo → 12\n")
	sb.WriteString("8. Limite → 1\n")
	sb.WriteString("9. Cancelar assistência/seguro → 8\n")
	sb.WriteString("10. Cancelar cartão → 7\n\n")

	sb.WriteString(fmt.Sprintf("Intenção: \"%s\"\n\n", request.UserIntent))
	sb.WriteString("Retorne APENAS JSON: {\"service_id\": número, \"service_name\": \"nome\"}\n")
	sb.WriteString("Resposta:")

	return sb.String()
}

func (c *OpenRouterClient) parseResponse(content string) (*classificationResult, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "{") {
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")

		if start == -1 || end == -1 || start >= end {
			return nil, fmt.Errorf("no valid JSON found in response")
		}

		content = content[start : end+1]
	}

	var result classificationResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse classification result: %w", err)
	}

	return &result, nil
}

func (c *OpenRouterClient) HealthCheck() error {
	if c.apiKey == "" {
		return fmt.Errorf("OpenRouter API key not configured")
	}
	return nil
}
