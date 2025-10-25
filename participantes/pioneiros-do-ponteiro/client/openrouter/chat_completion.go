package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	url := c.baseURL + "chat/completions"

	requestBody := OpenRouterRequest{
		Model: "openai/gpt-4o-mini",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "system",
				Content: `Você é um assistente que recebe uma entrada de texto do usuário descrevendo uma necessidade ou problema. Com base nessa entrada, você deve identificar o serviço mais adequado da lista abaixo e retornar apenas o ID do serviço correspondente. Use os exemplos fornecidos como referência para entender como mapear corretamente as intenções do usuário.\n\nLista de serviços:\n1 - Consulta Limite / Vencimento / Melhor dia de compra\n2 - Segunda via de boleto de acordo\n3 - Segunda via de fatura\n4 - Status de entrega do cartão\n5 - Status de cartão\n6 - Solicitação de aumento de limite\n7 - Cancelamento de cartão\n8 - Telefones de seguradoras\n9 - Desbloqueio de cartão\n10 - Esqueceu / Troca de senha\n11 - Perda e roubo\n12 - Consulta do saldo conta do Mais\n13 - Pagamento de contas\n14 - Reclamações\n15 - Atendimento humano\n16- Token de proposta\n\nExemplos:\nEntrada: Quanto tem disponível para usar → Resposta: 1\nEntrada: quando fecha minha fatura → Resposta: 1\nEntrada: Quando vence meu cartão → Resposta: 1\nEntrada: quando posso comprar → Resposta: 1\nEntrada: vencimento da fatura → Resposta: 1\nEntrada: valor para gastar → Resposta: 1\nEntrada: segunda via boleto de acordo → Resposta: 2\nEntrada: Boleto para pagar minha negociação → Resposta: 2\nEntrada: código de barras acordo → Resposta: 2\nEntrada: preciso pagar negociação → Resposta: 2\nEntrada: enviar boleto acordo → Resposta: 2\nEntrada: boleto da negociação → Resposta: 2\nEntrada: quero meu boleto → Resposta: 3\nEntrada: segunda via de fatura → Resposta: 3\nEntrada: código de barras fatura → Resposta: 3\nEntrada: quero a fatura do cartão → Resposta: 3\nEntrada: enviar boleto da fatura → Resposta: 3\nEntrada: fatura para pagamento → Resposta: 3\nEntrada: onde está meu cartão → Resposta: 4\nEntrada: meu cartão não chegou → Resposta: 4\nEntrada: status da entrega do cartão → Resposta: 4\nEntrada: cartão em transporte → Resposta: 4\nEntrada: previsão de entrega do cartão → Resposta: 4\nEntrada: cartão foi enviado? → Resposta: 4\nEntrada: não consigo passar meu cartão → Resposta: 5\nEntrada: meu cartão não funciona → Resposta: 5\nEntrada: cartão recusado → Resposta: 5\nEntrada: cartão não está passando → Resposta: 5\nEntrada: status do cartão ativo → Resposta: 5\nEntrada: problema com cartão → Resposta: 5\nEntrada: quero mais limite → Resposta: 6\nEntrada: aumentar limite do cartão → Resposta: 6\nEntrada: solicitar aumento de crédito → Resposta: 6\nEntrada: preciso de mais limite → Resposta: 6\nEntrada: pedido de aumento de limite → Resposta: 6\nEntrada: limite maior no cartão → Resposta: 6\nEntrada: cancelar cartão → Resposta: 7\nEntrada: quero encerrar meu cartão → Resposta: 7\nEntrada: bloquear cartão definitivamente → Resposta: 7\nEntrada: cancelamento de crédito → Resposta: 7\nEntrada: desistir do cartão → Resposta: 7\nEntrada: quero cancelar seguro → Resposta: 8\nEntrada: telefone do seguro → Resposta: 8\nEntrada: contato da seguradora → Resposta: 8\nEntrada: preciso falar com o seguro → Resposta: 8\nEntrada: seguro do cartão → Resposta: 8\nEntrada: cancelar assistência → Resposta: 8\nEntrada: desbloquear cartão → Resposta: 9\nEntrada: ativar cartão novo → Resposta: 9\nEntrada: como desbloquear meu cartão → Resposta: 9\nEntrada: quero desbloquear o cartão → Resposta: 9\nEntrada: cartão para uso imediato → Resposta: 9\nEntrada: desbloqueio para compras → Resposta: 9\nEntrada: não tenho mais a senha do cartão → Resposta: 10\nEntrada: esqueci minha senha → Resposta: 10\nEntrada: trocar senha do cartão → Resposta: 10\nEntrada: preciso de nova senha → Resposta: 10\nEntrada: recuperar senha → Resposta: 10\nEntrada: senha bloqueada → Resposta: 10\nEntrada: perdi meu cartão → Resposta: 11\nEntrada: roubaram meu cartão → Resposta: 11\nEntrada: cartão furtado → Resposta: 11\nEntrada: perda do cartão → Resposta: 11\nEntrada: bloquear cartão por roubo → Resposta: 11\nEntrada: extravio de cartão → Resposta: 11\nEntrada: saldo conta corrente → Resposta: 12\nEntrada: consultar saldo → Resposta: 12\nEntrada: quanto tenho na conta → Resposta: 12\nEntrada: extrato da conta → Resposta: 12\nEntrada: saldo disponível → Resposta: 12\nEntrada: meu saldo atual → Resposta: 12\nEntrada: quero pagar minha conta → Resposta: 13\nEntrada: pagar boleto → Resposta: 13\nEntrada: pagamento de conta → Resposta: 13\nEntrada: quero pagar fatura → Resposta: 13\nEntrada: efetuar pagamento → Resposta: 13\nEntrada: quero reclamar → Resposta: 14\nEntrada: abrir reclamação → Resposta: 14\nEntrada: fazer queixa → Resposta: 14\nEntrada: reclamar atendimento → Resposta: 14\nEntrada: registrar problema → Resposta: 14\nEntrada: protocolo de reclamação → Resposta: 14\nEntrada: falar com uma pessoa → Resposta: 15\nEntrada: preciso de humano → Resposta: 15\nEntrada: transferir para atendente → Resposta: 15\nEntrada: quero falar com atendente → Resposta: 15\nEntrada: atendimento pessoal → Resposta: 15\nEntrada: código para fazer meu cartão → Resposta: 16\nEntrada: token de proposta → Resposta: 16\nEntrada: receber código do cartão → Resposta: 16\nEntrada: proposta token → Resposta: 16\nEntrada: número de token → Resposta: 1\nEntrada: código de token da proposta → Resposta: 16\n\nSempre responda apenas com o número do ID do serviço mais indicado, sem explicações.`,
			},
			{
				Role:    "user",
				Content: intent,
			},
		},
	}

	fmt.Println("Url: ", url)

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

	fmt.Println("response: ", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("status Code openrouter: ", resp.StatusCode)

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v. body: %s", err, string(body))
	}

	fmt.Println("openRouterResp: ", openRouterResp)

	fmt.Println("len(openRouterResp.Choices)", len(openRouterResp.Choices))

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	num, _ := strconv.ParseUint(openRouterResp.Choices[0].Message.Content, 10, 8)
	u8 := uint8(num)

	return GetDataResponse(u8), nil
}

func GetDataResponse(serviceID uint8) *DataResponse {

	names := map[uint8]string{
		1:  "Consulta Limite / Vencimento / Melhor dia de compra",
		2:  "Segunda via de boleto de acordo",
		3:  "Segunda via de fatura",
		4:  "Status de entrega do cartão",
		5:  "Status de cartão",
		6:  "Solicitação de aumento de limite",
		7:  "Cancelamento de cartão",
		8:  "Telefones de seguradoras",
		9:  "Desbloqueio de cartão",
		10: "Esqueceu / Troca de senha",
		11: "Perda e roubo",
		12: "Consulta do saldo conta do Mais",
		13: "Pagamento de contas",
		14: "Reclamações",
		15: "Atendimento humano",
		16: "Token de proposta",
	}
	return &DataResponse{
		ServiceID:   serviceID,
		ServiceName: names[serviceID],
	}
}
