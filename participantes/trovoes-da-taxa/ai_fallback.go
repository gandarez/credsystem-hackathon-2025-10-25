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
	intents    []Intent // Cache dos intents para construir prompts melhores
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
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		model: "openai/gpt-4o-mini", // Modelo mais avançado e ainda econômico
	}
}

// SetIntents define os intents disponíveis para melhorar o prompt
func (c *AIClient) SetIntents(intents []Intent) {
	c.intents = intents
}

type openRouterRequest struct {
	Model     string    `json:"model"`
	Messages  []message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// aiResponse representa a resposta estruturada da IA
type aiResponse struct {
	Success     bool   `json:"success"`
	ServiceID   int    `json:"service_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	Error       string `json:"error,omitempty"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// buildPrompt constrói o prompt para a IA com as instruções precisas
// Usa os intents pré-carregados como exemplos para melhorar a classificação
func (c *AIClient) buildPrompt(intentText string, services map[int]string) string {
	var sb strings.Builder

	sb.WriteString("Você é um assistente que identifica qual serviço corresponde à intenção do cliente.\n")
	sb.WriteString("Regras importantes:\n")
	sb.WriteString("1. Use apenas os serviços listados abaixo.\n")
	sb.WriteString("2. Não invente serviços.\n")
	sb.WriteString("3. Compare a intenção do cliente com os exemplos fornecidos, usando palavras-chave, sinônimos e variações de escrita.\n")
	sb.WriteString("   - Se encontrar uma correspondência aproximada, escolha o serviço correspondente.\n")
	sb.WriteString("   - Se não houver correspondência clara, retorne o JSON de erro.\n")
	sb.WriteString("4. Responda **exatamente** com JSON puro, sem crases, markdown ou texto adicional.\n")
	sb.WriteString("5. Formato do JSON:\n")
	sb.WriteString("   - Sucesso: {\"success\": true, \"service_id\": <número>, \"service_name\": \"<nome>\"}\n")
	sb.WriteString("   - Erro: {\"success\": false, \"error\": \"Intenção desconhecida\"}\n\n")

	// Se temos intents carregados, usá-los como exemplos
	if len(c.intents) > 0 {
		sb.WriteString("Serviços disponíveis com exemplos:\n")

		// Agrupar intents por serviço para mostrar exemplos
		serviceExamples := make(map[int][]string)
		for _, intent := range c.intents {
			serviceExamples[intent.ServiceID] = append(
				serviceExamples[intent.ServiceID],
				intent.IntentText,
			)
		}

		for id, name := range services {
			examples := serviceExamples[id]
			if len(examples) > 0 {
				// Limitar a 3 exemplos por serviço para não sobrecarregar o prompt
				if len(examples) > 3 {
					examples = examples[:3]
				}
				sb.WriteString(fmt.Sprintf("- %d: %s\n  Exemplos: %s\n",
					id, name, strings.Join(examples, " | ")))
			} else {
				sb.WriteString(fmt.Sprintf("- %d: %s\n", id, name))
			}
		}
	} else {
		// Fallback se não tivermos intents
		sb.WriteString("Serviços disponíveis:\n")
		for id, name := range services {
			sb.WriteString(fmt.Sprintf("- %d: %s\n", id, name))
		}
	}

	sb.WriteString(fmt.Sprintf("\nIntenção do cliente: \"%s\"\n\n", intentText))
	sb.WriteString("Escolha o serviço correto se existir, considerando variações de escrita, palavras-chave e sinônimos.\n")
	sb.WriteString("Responda **somente** com JSON válido, sem nenhum texto extra.\n")

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
		MaxTokens: 150, // Suficiente para resposta JSON
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

	var openRouterResp openRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("unmarshal openrouter response: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in AI response")
	}

	content := strings.TrimSpace(openRouterResp.Choices[0].Message.Content)

	// Limpar markdown code blocks se existirem (```json ... ```)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Parse a resposta JSON estruturada
	var aiResp aiResponse
	if err := json.Unmarshal([]byte(content), &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %w\nContent: %s", err, content)
	}

	// Verificar se a IA conseguiu classificar
	if !aiResp.Success {
		return nil, fmt.Errorf("AI could not classify: %s", aiResp.Error)
	}

	// Verificar se o ID é válido
	serviceName, exists := services[aiResp.ServiceID]
	if !exists {
		return nil, fmt.Errorf("AI returned invalid service ID: %d", aiResp.ServiceID)
	}

	return &APIResponse{
		ServiceID:   aiResp.ServiceID,
		ServiceName: serviceName,
	}, nil
}
