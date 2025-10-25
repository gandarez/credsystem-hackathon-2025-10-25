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
		Model            string  `json:"model"`
		Temperature      float64 `json:"temperature,omitempty"`
		TopP             float64 `json:"top_p,omitempty"`
		MaxTokens        int     `json:"max_tokens,omitempty"`
		PresencePenalty  float64 `json:"presence_penalty,omitempty"`
		FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
		Stream           bool    `json:"stream,omitempty"`
		Messages         []struct {
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
		// Model: "google/gemini-2.5-pro",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role: "system",
				Content: `Você é um especialista em classificação de intenções de clientes no contexto financeiro brasileiro.
						Seu objetivo é identificar qual serviço o cliente está solicitando, analisando o SIGNIFICADO COMPLETO da frase, não apenas palavras-chave isoladas.

						PROCESSO DE CLASSIFICAÇÃO:
						1. Identifique o VERBO PRINCIPAL da solicitação (consultar, pagar, cancelar, solicitar, etc.)
						2. Identifique o OBJETO da solicitação (cartão, boleto, fatura, limite, saldo, senha, etc.)
						3. Identifique o CONTEXTO financeiro (crédito, conta corrente, acordo, seguro, entrega, etc.)
						4. Combine essas informações para determinar a INTENÇÃO REAL do cliente

						REGRAS CRÍTICAS DE DESAMBIGUAÇÃO:
						AA. PRIORIDADE DE CONTEXTO SEGURO vs CARTÃO:
						- Sempre que a solicitação mencionar "seguro", "seguradora", "assistência" ou "proteção",
						priorize o Serviço 8 (Telefones de seguradoras), independentemente do verbo presente.
						Exemplo: "quero cancelar seguro", "cancelar assistência", "preciso falar com a seguradora".
						Esses casos NÃO devem ser classificados como cancelamento de cartão.

						REGRAS DE DESAMBIGUAÇÃO:
						1. LIMITE (cartão) vs SALDO (conta): "disponível" + cartão = Serviço 1; "disponível" + conta = Serviço 12
						- "quando posso comprar" = Serviço 1 (contexto de melhor dia de compra no cartão)
						2. BOLETO sem qualificador = Serviço 3 (Fatura); Com "acordo"/"negociação" = Serviço 2
						3. PAGAR (verbo ação) = Serviço 13; "fatura para pagamento" (obter documento) = Serviço 3
						4. CARTÃO "uso imediato"/"liberar" = Serviço 9 (Desbloqueio)
						5. "registrar problema" = Serviço 14 (Reclamações)
						6. "receber código/token" = Serviço 16

						GLOSSÁRIO FINANCEIRO CONTEXTUALIZADO:

						1 - Consulta Limite / Vencimento do cartão / Melhor dia de compra
							Contexto: Informações sobre CRÉDITO disponível no cartão
							Termos: limite, vencimento, fechamento, melhor dia de compra, quanto posso gastar no cartão
							Exemplos: "quanto tenho de limite", "quando fecha minha fatura", "qual dia melhor para comprar"

						2 - Segunda via de boleto de acordo
							Contexto: Renegociação de dívidas, acordos de pagamento
							Termos: acordo, negociação, parcelamento, renegociação, código de barras do acordo
							Exemplos: "boleto do acordo", "pagar minha negociação", "código de barras do parcelamento"

						3 - Segunda via de Fatura
							Contexto: Documento da fatura mensal do cartão
							Termos: fatura, conta do cartão, boleto da fatura, código de barras da fatura
							Exemplos: "meu boleto" (sem contexto de acordo), "segunda via da fatura", "ver minha fatura"

						4 - Status de Entrega do Cartão
							Contexto: Rastreamento físico do cartão
							Termos: entrega, transporte, chegou, enviado, rastreio, previsão
							Exemplos: "onde está meu cartão", "cartão foi enviado", "cartão em transporte"

						5 - Status de cartão
							Contexto: Funcionamento e situação do cartão (ativo/bloqueado)
							Termos: não funciona, recusado, bloqueado, inativo, problema com cartão, não passa
							Exemplos: "cartão recusado na maquininha", "meu cartão está funcionando?"

						6 - Solicitação de aumento de limite
							Contexto: Pedido de mais crédito
							Termos: aumentar, mais limite, crédito maior, solicitar aumento
							Exemplos: "quero mais limite", "aumentar meu crédito"

						7 - Cancelamento de cartão
							Contexto: Encerramento definitivo do cartão
							Termos: cancelar, encerrar, desistir, bloqueio permanente
							Exemplos: "cancelar meu cartão", "não quero mais o cartão"

						8 - Telefones de seguradoras
							Contexto: Assuntos relacionados a seguros e assistências vinculadas ao cartão.
							Termos: seguro, seguradora, assistência, proteção, apólice, sinistro
							Exemplos: "cancelar seguro", "cancelar assistência", "telefone do seguro", "falar com a seguradora", "contato da seguradora"

						9 - Desbloqueio de Cartão
							Contexto: Ativar ou liberar cartão para uso
							Termos: desbloquear, ativar, liberar, habilitar, cartão novo
							Exemplos: "ativar meu cartão", "liberar para uso"

						10 - Esqueceu senha / Troca de senha
							Contexto: Problemas com senha do cartão
							Termos: senha, trocar senha, esqueci, recuperar senha
							Exemplos: "não lembro minha senha", "mudar senha do cartão"

						11 - Perda e roubo
							Contexto: Cartão perdido, roubado ou furtado
							Termos: perdi, roubaram, furtado, extraviado, sumiu
							Exemplos: "perdi meu cartão", "roubaram meu cartão"

						12 - Consulta do Saldo
							Contexto: Saldo em CONTA CORRENTE (não crédito)
							Termos: saldo, extrato, quanto tenho na conta, conta corrente
							Exemplos: "saldo da minha conta", "extrato bancário"

						13 - Pagamento de contas
							Contexto: AÇÃO de efetuar pagamento
							Termos: pagar (verbo de ação), efetuar pagamento, quitar
							Exemplos: "pagar minha conta", "quero pagar o boleto", "pagar fatura"

						14 - Reclamações
							Contexto: Insatisfação ou problemas
							Termos: reclamar, queixa, problema, insatisfeito
							Exemplos: "quero reclamar", "abrir uma reclamação"

						15 - Atendimento humano
							Contexto: Falar com pessoa física
							Termos: atendente, humano, pessoa, operador
							Exemplos: "falar com atendente", "preciso de uma pessoa"

						16 - Token de proposta
							Contexto: Código para aprovação de novo cartão
							Termos: token, código, proposta
							Exemplos: "token do cartão", "código da proposta"

						FORMATO DE RESPOSTA:
						Retorne OBRIGATORIAMENTE APENAS um objeto JSON válido, sem texto adicional, sem delimitadores de código (como tripla crase json) e sem nenhum caractere adicional:
						{"service_id": <número de 1 a 16>, "service_name": "<nome exato do serviço conforme listado acima>"}

						Se a solicitação for completamente fora do escopo financeiro, retorne:
						{"service_id": 0, "service_name": "Serviço não identificado"}`,
			},
			{
				Role:    "user",
				Content: intent,
			},
		},
		Temperature:      0.1,
		TopP:             0.3,
		MaxTokens:        70,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		Stream:           false,
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
