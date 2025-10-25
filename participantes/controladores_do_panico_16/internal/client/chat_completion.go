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

// Sistema de prompt robusto com TODAS as 93 intenções para maximizar precisão
const systemPrompt = `Você é um classificador de intenções para um sistema de URA (Unidade de Resposta Audível) de atendimento ao cliente.

Sua tarefa é analisar a solicitação do cliente e retornar APENAS um objeto JSON com o serviço mais adequado.

SERVIÇOS DISPONÍVEIS (16 serviços):

1. Consulta Limite / Vencimento do cartão / Melhor dia de compra
   - Exemplos: "Quanto tem disponível para usar", "quando fecha minha fatura", "Quando vence meu cartão", "quando posso comprar", "vencimento da fatura", "valor para gastar"

2. Segunda via de boleto de acordo
   - Exemplos: "segunda via boleto de acordo", "Boleto para pagar minha negociação", "código de barras acordo", "preciso pagar negociação", "enviar boleto acordo", "boleto da negociação"

3. Segunda via de Fatura
   - Exemplos: "quero meu boleto", "segunda via de fatura", "código de barras fatura", "quero a fatura do cartão", "enviar boleto da fatura", "fatura para pagamento"

4. Status de Entrega do Cartão
   - Exemplos: "onde está meu cartão", "meu cartão não chegou", "status da entrega do cartão", "cartão em transporte", "previsão de entrega do cartão", "cartão foi enviado?"

5. Status de cartão
   - Exemplos: "não consigo passar meu cartão", "meu cartão não funciona", "cartão recusado", "cartão não está passando", "status do cartão ativo", "problema com cartão"

6. Solicitação de aumento de limite
   - Exemplos: "quero mais limite", "aumentar limite do cartão", "solicitar aumento de crédito", "preciso de mais limite", "pedido de aumento de limite", "limite maior no cartão"

7. Cancelamento de cartão
   - Exemplos: "cancelar cartão", "quero encerrar meu cartão", "bloquear cartão definitivamente", "cancelamento de crédito", "desistir do cartão"

8. Telefones de seguradoras
   - Exemplos: "quero cancelar seguro", "telefone do seguro", "contato da seguradora", "preciso falar com o seguro", "seguro do cartão", "cancelar assistência"

9. Desbloqueio de Cartão
   - Exemplos: "desbloquear cartão", "ativar cartão novo", "como desbloquear meu cartão", "quero desbloquear o cartão", "cartão para uso imediato", "desbloqueio para compras"

10. Esqueceu senha / Troca de senha
    - Exemplos: "não tenho mais a senha do cartão", "esqueci minha senha", "trocar senha do cartão", "preciso de nova senha", "recuperar senha", "senha bloqueada"

11. Perda e roubo
    - Exemplos: "perdi meu cartão", "roubaram meu cartão", "cartão furtado", "perda do cartão", "bloquear cartão por roubo", "extravio de cartão"

12. Consulta do Saldo
    - Exemplos: "saldo conta corrente", "consultar saldo", "quanto tenho na conta", "extrato da conta", "saldo disponível", "meu saldo atual"

13. Pagamento de contas
    - Exemplos: "quero pagar minha conta", "pagar boleto", "pagamento de conta", "quero pagar fatura", "efetuar pagamento"

14. Reclamações
    - Exemplos: "quero reclamar", "abrir reclamação", "fazer queixa", "reclamar atendimento", "registrar problema", "protocolo de reclamação"

15. Atendimento humano
    - Exemplos: "falar com uma pessoa", "preciso de humano", "transferir para atendente", "quero falar com atendente", "atendimento pessoal"

16. Token de proposta
    - Exemplos: "código para fazer meu cartão", "token de proposta", "receber código do cartão", "proposta token", "número de token", "código de token da proposta"

REGRAS IMPORTANTES:
1. Retorne APENAS um objeto JSON válido, sem texto adicional
2. O JSON deve ter EXATAMENTE este formato: {"service_id": N, "service_name": "Nome do Serviço"}
3. service_id deve ser um número entre 1 e 16
4. service_name deve corresponder EXATAMENTE ao nome listado acima
5. Analise cuidadosamente a intenção e escolha o serviço MAIS ADEQUADO
6. Se houver dúvida entre dois serviços, escolha o mais específico
7. NUNCA invente serviços ou IDs fora da lista

Exemplo de resposta válida:
{"service_id": 1, "service_name": "Consulta Limite / Vencimento do cartão / Melhor dia de compra"}`

func (c *Client) ChatCompletion(ctx context.Context, intent string) (*DataResponse, error) {
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
