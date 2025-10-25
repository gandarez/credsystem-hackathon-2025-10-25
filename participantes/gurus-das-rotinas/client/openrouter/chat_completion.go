package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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
		Model: "mistralai/mistral-7b-instruct",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role: "system",
				Content: `Você é um classificador especializado em intenções bancárias. Analise a solicitação e retorne APENAS JSON válido.

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

REGRAS CRÍTICAS:
1. Retorne SEMPRE: {"service_id": N, "service_name": "nome exato"}
2. Use EXATAMENTE os nomes dos serviços acima
3. Para não-bancário: {"service_id": null, "service_name": "out of context"}

CLASSIFICAÇÃO EXATA POR PALAVRAS-CHAVE:

SERVIÇO 1 - CONSULTAS LIMITE/VENCIMENTO (PERGUNTAS SOBRE LIMITE/VENCIMENTO):
- "quanto tem disponível" → ID=1
- "quando fecha minha fatura" → ID=1 (PERGUNTA sobre vencimento)
- "quando vence meu cartão" → ID=1
- "quando posso comprar" → ID=1
- "vencimento da fatura" → ID=1
- "valor para gastar" → ID=1 (PERGUNTA sobre limite disponível)

SERVIÇO 2 - BOLETO DE ACORDO (NEGOCIAÇÃO/ACORDO):
- "segunda via boleto de acordo" → ID=2
- "boleto para pagar minha negociação" → ID=2 (CONTÉM "negociação")
- "código de barras acordo" → ID=2
- "preciso pagar negociação" → ID=2 (CONTÉM "negociação")
- "enviar boleto acordo" → ID=2
- "boleto da negociação" → ID=2 (CONTÉM "negociação")

SERVIÇO 3 - SEGUNDA VIA FATURA (FATURA SEM NEGOCIAÇÃO):
- "quero meu boleto" → ID=3 (SEM "acordo" ou "negociação")
- "segunda via de fatura" → ID=3 (CONTÉM "fatura")
- "código de barras fatura" → ID=3 (CONTÉM "fatura")
- "quero a fatura do cartão" → ID=3 (CONTÉM "fatura")
- "enviar boleto da fatura" → ID=3 (CONTÉM "fatura")
- "fatura para pagamento" → ID=3 (CONTÉM "fatura")

SERVIÇO 4 - ENTREGA CARTÃO (ONDE ESTÁ/TRANSPORTE):
- "onde está meu cartão" → ID=4
- "meu cartão não chegou" → ID=4
- "status da entrega do cartão" → ID=4
- "cartão em transporte" → ID=4 (TRANSPORTE = ENTREGA)
- "previsão de entrega do cartão" → ID=4
- "cartão foi enviado?" → ID=4

SERVIÇO 5 - STATUS CARTÃO (FUNCIONAMENTO/PROBLEMAS):
- "não consigo passar meu cartão" → ID=5 (PROBLEMA DE FUNCIONAMENTO)
- "meu cartão não funciona" → ID=5 (PROBLEMA DE FUNCIONAMENTO)
- "cartão recusado" → ID=5 (PROBLEMA DE FUNCIONAMENTO)
- "cartão não está passando" → ID=5 (PROBLEMA DE FUNCIONAMENTO)
- "status do cartão ativo" → ID=5
- "problema com cartão" → ID=5 (PROBLEMA DE FUNCIONAMENTO)

SERVIÇO 6 - AUMENTO LIMITE (SOLICITAR MAIS LIMITE):
- "quero mais limite" → ID=6
- "aumentar limite do cartão" → ID=6
- "solicitar aumento de crédito" → ID=6 (SOLICITAÇÃO DE AUMENTO)
- "preciso de mais limite" → ID=6
- "pedido de aumento de limite" → ID=6
- "limite maior no cartão" → ID=6

SERVIÇO 7 - CANCELAMENTO (CANCELAR/ENCERRAR):
- "cancelar cartão" → ID=7
- "quero encerrar meu cartão" → ID=7
- "bloquear cartão definitivamente" → ID=7 (BLOQUEIO DEFINITIVO = CANCELAMENTO)
- "cancelamento de crédito" → ID=7
- "desistir do cartão" → ID=7

SERVIÇO 8 - SEGURADORAS (SEGURO/ASSISTÊNCIA):
- "quero cancelar seguro" → ID=8
- "telefone do seguro" → ID=8
- "contato da seguradora" → ID=8
- "preciso falar com o seguro" → ID=8
- "seguro do cartão" → ID=8
- "cancelar assistência" → ID=8 (ASSISTÊNCIA = SEGURO)

SERVIÇO 9 - DESBLOQUEIO (ATIVAR/LIBERAR):
- "desbloquear cartão" → ID=9
- "ativar cartão novo" → ID=9
- "liberar cartão" → ID=9
- "habilitar cartão" → ID=9
- "cartão para uso imediato" → ID=9 (USO IMEDIATO = ATIVAR)

SERVIÇO 10 - SENHA:
- "esqueci minha senha" → ID=10
- "trocar senha" → ID=10
- "alterar senha" → ID=10
- "nova senha" → ID=10
- "resetar senha" → ID=10

SERVIÇO 11 - PERDA/ROUBO:
- "perdi meu cartão" → ID=11
- "roubaram meu cartão" → ID=11
- "furtado cartão" → ID=11
- "sumiu meu cartão" → ID=11
- "desapareceu cartão" → ID=11

SERVIÇO 12 - SALDO:
- "consulta saldo" → ID=12
- "quanto tenho na conta" → ID=12
- "valor disponível na conta" → ID=12
- "dinheiro na conta" → ID=12

SERVIÇO 13 - PAGAMENTO (PAGAR CONTAS/BOLETOS):
- "pagar conta" → ID=13
- "transferência" → ID=13
- "débito automático" → ID=13
- "pix" → ID=13
- "pagar boleto" → ID=13 (PAGAR = PAGAMENTO, não fatura)

SERVIÇO 14 - RECLAMAÇÕES:
- "reclamação" → ID=14
- "problema" → ID=14
- "erro" → ID=14
- "não funciona" → ID=14
- "complicado" → ID=14

SERVIÇO 15 - ATENDIMENTO HUMANO:
- "falar com atendente" → ID=15
- "atendimento pessoal" → ID=15
- "operador" → ID=15
- "pessoa" → ID=15
- "transferir para atendente" → ID=15

SERVIÇO 16 - TOKEN PROPOSTA (CÓDIGO/APROVAÇÃO):
- "token de proposta" → ID=16
- "aprovação de proposta" → ID=16
- "solicitação proposta" → ID=16
- "código para fazer meu cartão" → ID=16 (CÓDIGO = TOKEN)
- "receber código do cartão" → ID=16 (CÓDIGO = TOKEN)

INSTRUÇÕES FINAIS:
1. Analise palavra por palavra
2. Identifique a categoria exata
3. Retorne JSON válido
4. Use IDs e nomes exatos

CASOS CRÍTICOS - ATENÇÃO ESPECIAL:
- "quero meu boleto" (SEM "acordo" ou "negociação") = ID=3
- "quando fecha minha fatura" (PERGUNTA sobre vencimento) = ID=1
- "segunda via de fatura" (CONTÉM "fatura") = ID=3
- "boleto para pagar minha negociação" (CONTÉM "negociação") = ID=2
- "cartão em transporte" (TRANSPORTE = ENTREGA) = ID=4
- "não consigo passar meu cartão" (PROBLEMA DE FUNCIONAMENTO) = ID=5
- "cartão não está passando" (PROBLEMA DE FUNCIONAMENTO) = ID=5
- "problema com cartão" (PROBLEMA DE FUNCIONAMENTO) = ID=5
- "bloquear cartão definitivamente" (BLOQUEIO DEFINITIVO = CANCELAMENTO) = ID=7
- "cancelar assistência" (ASSISTÊNCIA = SEGURO) = ID=8
- "valor para gastar" (PERGUNTA sobre limite) = ID=1
- "solicitar aumento de crédito" (SOLICITAÇÃO DE AUMENTO) = ID=6
- "cartão para uso imediato" (USO IMEDIATO = ATIVAR) = ID=9
- "pagar boleto" (PAGAR = PAGAMENTO) = ID=13
- "receber código do cartão" (CÓDIGO = TOKEN) = ID=16`,
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

	// Retry mechanism with exponential backoff
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.Do(ctx, req)
		if err != nil {
			lastErr = fmt.Errorf("failed to execute request (attempt %d): %v", attempt+1, err)
			if attempt < maxRetries-1 {
				// Exponential backoff: 100ms, 200ms, 400ms
				time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("error reading response (attempt %d): %v", attempt+1, err)
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API request failed with status %d (attempt %d): %s", resp.StatusCode, attempt+1, string(body))
			if attempt < maxRetries-1 && resp.StatusCode >= 500 {
				// Retry on server errors
				time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		var openRouterResp OpenRouterResponse
		if err := json.Unmarshal(body, &openRouterResp); err != nil {
			lastErr = fmt.Errorf("error unmarshaling response (attempt %d): %v. body: %s", attempt+1, err, string(body))
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		if len(openRouterResp.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response (attempt %d)", attempt+1)
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		// Clean the response content to extract JSON
		content := openRouterResp.Choices[0].Message.Content

		// Find JSON object in the response (handle cases where AI adds extra text)
		startIdx := strings.Index(content, "{")
		endIdx := strings.LastIndex(content, "}")
		if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
			lastErr = fmt.Errorf("no valid JSON found in response (attempt %d): %s", attempt+1, content)
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		jsonStr := content[startIdx : endIdx+1]

		var dataRes DataResponse
		if err := json.Unmarshal([]byte(jsonStr), &dataRes); err != nil {
			lastErr = fmt.Errorf("error unmarshaling data response (attempt %d): %v. content: %s", attempt+1, err, jsonStr)
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		// Success! Return the result
		return &dataRes, nil
	}

	return nil, lastErr
}
