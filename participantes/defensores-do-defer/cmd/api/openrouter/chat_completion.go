package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		Model: "openai/gpt-4o-mini-2024-07-18",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role: "system",
				Content: `Você é um modelo de classificação de intenções especializado em atendimento financeiro em português do Brasil.
Receberá uma mensagem de cliente e deve classificá-la em UMA das intenções pré-definidas abaixo.

Responda **somente** com um JSON válido no formato:
{
  "service_id": número ou null,
  "service_name": "nome do serviço" ou "Unknown"
}

Lista de intenções possíveis:
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
12 - Consulta do Saldo
13 - Pagamento de contas
14 - Reclamações
15 - Atendimento humano
16 - Token de proposta

Se a frase não corresponder a nenhuma dessas categorias, retorne:
{
  "service_id": null,
  "service_name": "Unknown"
}

Exemplo de entrada:
"quero aumentar o limite do meu cartão"

Exemplo de resposta:
{"service_id": 6, "service_name": "Solicitação de aumento de limite"}`,
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
	fmt.Printf("Requisição para o OpenRouter %s", string(jsonBody))
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
	fmt.Printf("Resposta bruta do OpenRouter %s", string(body))
	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v. body: %s", err, string(body))
	}
	fmt.Printf("Resposta unmarshaled do OpenRouter %v", openRouterResp)
	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	var dataRes DataResponse
	if err := json.Unmarshal([]byte(openRouterResp.Choices[0].Message.Content), &dataRes); err != nil {
		return nil, fmt.Errorf("error unmarshaling data response: %v. content: %s", err, openRouterResp.Choices[0].Message.Content)
	}

	return &dataRes, nil
}
