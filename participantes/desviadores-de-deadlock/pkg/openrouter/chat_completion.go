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
		Model       string  `json:"model"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
		TopP        float64 `json:"top_p"`
		Messages    []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		ResponseFormat any      `json:"response_format,omitempty"` // <-- força JSON
		Stop           []string `json:"stop,omitempty"`
	}

	OpenRouterResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	DataResponse struct {
		ServiceID   *uint8 `json:"service_id"` // pode ser null
		ServiceName string `json:"service_name"`
		// Confidence  float64 `json:"confidence"`
		// Explanation string  `json:"explanation"`
	}
)

func (c *Client) ChatCompletion(ctx context.Context, intent string) (*DataResponse, error) {
	url := c.baseURL + "/chat/completions"

	requestBody := OpenRouterRequest{
		Model:          "anthropic/claude-haiku-4.5",
		Temperature:    0.2,
		MaxTokens:      512,
		TopP:           1.0,
		ResponseFormat: map[string]any{"type": "json_object"}, // <--- força JSON
		Stop:           []string{"\n\n"},                      // opcional: para o modelo não narrar
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role: "system",
				Content: `
Você é um classificador de intenções financeiras.
Leia uma frase curta e identifique a qual serviço FIXO ela pertence.
Use exatamente os IDs e nomes abaixo. Responda SOMENTE com um JSON válido (sem markdown, sem cercas de código, sem texto extra).

SERVIÇOS FIXOS:
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

REGRAS GERAIS
- Não invente serviços. Ignore caixa/acentos/pontuação simples.
- Use sinônimos/variações/gírias e erros leves de digitação.
- Compare a intenção (semântica), não apenas palavras exatas.

REGRAS DE DESAMBIGUAÇÃO (determinísticas)
• "saldo"/"conta"/"extrato" → 12
• "limite"/"melhor dia"/"disponível pra gastar"/"fatura do cartão" → 1
• "acordo"/"negociação"/"renegociar" → 2
• "fatura"/"boleto da fatura"/"quero meu boleto" → 3
• "pagar boleto/conta" sem “fatura/acordo” → 13
• "não passa"/"recusado"/"não funciona"/"problema com cartão" → 5
• "desbloquear"/"ativar"/"liberar" → 9
• "roubo"/"furto"/"extravio"/"perdi" → 11 (prioriza sobre 7)
• "cancelar"/"encerrar" → 7
• "reclamar"/"queixa"/"protocolo"/"registrar problema" → 14
• "falar com atendente"/"humano"/"pessoa" → 15
• "token"/"código da proposta"/"número de token" → 16
• "saldo do cartão" → 1

REGRAS ESPECÍFICAS PARA EVITAR ERROS OBSERVADOS
• **4/Status de Entrega**: frases de logística como "cartão em transporte", "rastreio", "previsão de entrega", "foi enviado?" → 4
• **8/Seguradoras**: "quero cancelar seguro", "cancelar assistência", "telefone do seguro", "contato da seguradora" → 8
• **9/Desbloqueio**: "cartão para uso imediato" e "desbloqueio para compras" → 9
• **3 vs 13 (caso especial)**:
   – "fatura para pagamento" → 3
   – "quero pagar fatura" → 13
• **16/Token**: "receber código do cartão", "número de token", "código de token da proposta" → 16

CONFIDÊNCIA E EMPATE
- Se confiança < 0.50 ou top-2 muito próximos (diferença < 0.05) → use Desconhecido.

VERIFICAÇÃO (interna, sem mostrar)
1) Confirme que termos sustentam o serviço escolhido e que nenhuma prioridade acima foi violada.
2) Ajuste a confiança (0–1 com duas casas).
3) Nunca altere nomes/IDs dos serviços.

SAÍDA OBRIGATÓRIA (todas as chaves, nessa ordem exata):
{
  "service_id": <número ou null>,
  "service_name": "<nome do serviço ou 'Desconhecido'>",
  "confidence": <número entre 0 e 1 com duas casas>,
  "explanation": "<frase curta explicando a decisão>"
}
`,
			},
			{Role: "user", Content: intent},
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
	// Dica: se usar OpenRouter, inclua também Authorization e cabeçalhos recomendados (HTTP-Referer/X-Title) no c.Client.Do(...)

	resp, err := c.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}
	fmt.Println("body:", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var orResp OpenRouterResponse
	if err := json.Unmarshal(body, &orResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v. body: %s", err, string(body))
	}
	if len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := orResp.Choices[0].Message.Content
	payload := extractJSON(content) // <--- remove cercas/pega primeiro objeto

	var dataRes DataResponse
	if err := json.Unmarshal([]byte(payload), &dataRes); err != nil {
		return nil, fmt.Errorf("error unmarshaling data response: %v. content: %s", err, payload)
	}

	return &dataRes, nil
}
