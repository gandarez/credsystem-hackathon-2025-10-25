package openrouter

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
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	OpenRouterResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	DataResponse struct {
		ServiceID   uint8  `json:"service_id"`
		ServiceName string `json:"service_name"`
	}
)

func (c *Client) ChatCompletion(ctx context.Context, intent string) (*DataResponse, error) {
	url := c.baseURL + "/chat/completions"

	requestBody := OpenRouterRequest{
		Model: "openai/gpt-4o-mini",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role: "system",
				Content: `Classifique a solicitação do cliente e retorne APENAS um JSON válido.

SERVIÇOS:
1: "Consulta Limite / Vencimento do cartão / Melhor dia de compra"
2: "Segunda via de boleto de acordo"
3: "Segunda via de Fatura"
4: "Status de Entrega do Cartão"
5: "Status de cartão"
6: "Solicitação de aumento de limite"
7: "Cancelamento de cartão"
8: "Telefones de seguradoras"
9: "Desbloqueio de Cartão"
10: "Esqueceu senha / Troca de senha"
11: "Perda e roubo"
12: "Consulta do Saldo"
13: "Pagamento de contas"
14: "Reclamações"
15: "Atendimento humano"
16: "Token de proposta"

EXEMPLOS:
"quanto tem disponível para usar" → {"service_id": 1, "service_name": "Consulta Limite / Vencimento do cartão / Melhor dia de compra"}
"quero cancelar meu cartão" → {"service_id": 7, "service_name": "Cancelamento de cartão"}
"perdi meu cartão" → {"service_id": 11, "service_name": "Perda e roubo"}

Retorne apenas: {"service_id": N, "service_name": "nome exato"}`,
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

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Clean the response content to extract JSON
	content := openRouterResp.Choices[0].Message.Content

	// Find JSON object in the response (handle cases where AI adds extra text)
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return nil, fmt.Errorf("no valid JSON found in response: %s", content)
	}

	jsonStr := content[startIdx : endIdx+1]

	var dataRes DataResponse
	if err := json.Unmarshal([]byte(jsonStr), &dataRes); err != nil {
		return nil, fmt.Errorf("error unmarshaling data response: %v. content: %s", err, jsonStr)
	}

	return &dataRes, nil
}
