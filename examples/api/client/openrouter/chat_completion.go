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
		Model: "openai/gpt-4o-mini",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role: "system",
				Content: `Você é um sistema de classificação de solicitações de clientes da Credsystem.
						Seu trabalho é identificar a intenção do cliente e retornar o service_id e service_name corretos.

						REGRAS OBRIGATÓRIAS:
						1. Retorne APENAS JSON válido, sem texto adicional
						2. Use EXATAMENTE um dos 16 serviços listados abaixo
						3. Não invente novos serviços. Caso o intento não corresponda a nenhum serviço da lista, retorne o service_id 0 e service_name "Serviço não identificado".
						4. Analise a intenção e encontre o serviço mais próximo

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
						12 - Consulta do Saldo
						13 - Pagamento de contas
						14 - Reclamações
						15 - Atendimento humano
						16 - Token de proposta

						EXEMPLOS DE INTENÇÕES CONHECIDAS:
						- "Quanto tem disponível para usar" → serviço 1
						- "quando fecha minha fatura" → serviço 1
						- "segunda via boleto de acordo" → serviço 2
						- "Boleto para pagar minha negociação" → serviço 2
						- "quero meu boleto" → serviço 3
						- "segunda via de fatura" → serviço 3
						- "onde está meu cartão" → serviço 4
						- "meu cartão não chegou" → serviço 4
						- "não consigo passar meu cartão" → serviço 5
						- "meu cartão não funciona" → serviço 5
						- "quero mais limite" → serviço 6
						- "aumentar limite do cartão" → serviço 6
						- "cancelar cartão" → serviço 7
						- "quero encerrar meu cartão" → serviço 7
						- "quero cancelar seguro" → serviço 8
						- "telefone do seguro" → serviço 8
						- "desbloquear cartão" → serviço 9
						- "ativar cartão novo" → serviço 9
						- "não tenho mais a senha do cartão" → serviço 10
						- "esqueci minha senha" → serviço 10
						- "perdi meu cartão" → serviço 11
						- "roubaram meu cartão" → serviço 11
						- "saldo conta corrente" → serviço 12
						- "consultar saldo" → serviço 12
						- "quero pagar minha conta" → serviço 13
						- "pagar boleto" → serviço 13
						- "quero reclamar" → serviço 14
						- "abrir reclamação" → serviço 14
						- "falar com uma pessoa" → serviço 15
						- "preciso de humano" → serviço 15
						- "código para fazer meu cartão" → serviço 16
						- "token de proposta" → serviço 16

						Formato de resposta obrigatório:
						{"service_id": <número>, "service_name": "<nome exato do serviço>"}`,
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

	var dataRes DataResponse
	if err := json.Unmarshal([]byte(openRouterResp.Choices[0].Message.Content), &dataRes); err != nil {
		return nil, fmt.Errorf("error unmarshaling data response: %v. content: %s", err, openRouterResp.Choices[0].Message.Content)
	}

	return &dataRes, nil
}
