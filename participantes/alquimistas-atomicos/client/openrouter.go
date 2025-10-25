package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type OpenRouterClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenRouterClient() *OpenRouterClient {
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "mistralai/mistral-7b-instruct" // Modelo padrão
	}

	return &OpenRouterClient{
		baseURL: "https://openrouter.ai/api/v1",
		apiKey:  os.Getenv("OPENROUTER_API_KEY"),
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type OpenRouterRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Service struct {
	ID   int    `json:"service_id"`
	Name string `json:"service_name"`
}

func (c *OpenRouterClient) ClassifyIntent(ctx context.Context, intent string) (*Service, error) {
	prompt := c.createClassificationPrompt(intent)

	requestBody := OpenRouterRequest{
		Model: c.model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "system",
				Content: prompt,
			},
			{
				Role:    "user",
				Content: intent,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("erro ao deserializar resposta: %v", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("nenhuma escolha na resposta")
	}

	content := openRouterResp.Choices[0].Message.Content

	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonStart >= jsonEnd {
		return nil, fmt.Errorf("resposta da IA não contém JSON válido")
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	var service Service
	if err := json.Unmarshal([]byte(jsonStr), &service); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse da resposta da IA: %v", err)
	}

	return &service, nil
}

func (c *OpenRouterClient) createClassificationPrompt(_ string) string {
	// Tenta ler o prompt do arquivo prompt.txt
	promptBytes, err := os.ReadFile("prompt.txt")
	if err != nil {
		// Se não conseguir ler o arquivo, usa o prompt padrão
		return `Você é um assistente especializado em classificar intenções de clientes bancários.

Sua tarefa é analisar a intenção do cliente e retornar o serviço mais adequado.

SERVIÇOS DISPONÍVEIS:
1 - Consulta Limite / Vencimento do cartão / Melhor dia de compra
2 - Segunda via de boleto de acordo
3 - Segunda via de Fatura
4 - Status de Entrega do Cartão
5 - Status de cartão
6 - Solicitação de aumento de limite
7 - Cancelamento de cartão
8 - Telefones de seguradoras
9 - Desbloqueio de Cartão
10 - Esqueceu senha / Troca de senha
11 - Perda e roubo
12 - Consulta do Saldo Conta do Mais
13 - Pagamento de contas
14 - Reclamações
15 - Atendimento humano
16 - Token de proposta

INSTRUÇÕES:
- Analise a intenção do cliente
- Escolha o serviço mais adequado (ID 1-16)
- Retorne APENAS um JSON no formato: {"service_id": X, "service_name": "Nome do Serviço"}
- Se não conseguir classificar, retorne erro
- Seja preciso e direto na classificação`
	}

	return string(promptBytes)
}
