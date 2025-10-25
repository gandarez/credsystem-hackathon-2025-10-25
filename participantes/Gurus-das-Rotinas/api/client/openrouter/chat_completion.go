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
				Content: `Para cada service_name há um service_id equivalente, identifique o intent recebido que seja direcionado a um service_name. 
				Baseie-se sua resposta a partir do intent recebido. RESPONDA APENAS O SERVICE_NAME e nada mais. 
				O contexto que estamos falando aqui é de um banco, logo, pergutnas do tipo segunda via do ceular, levando a crer que seja a parcela do celular, nao se aplica aqui.
				 ESTAMOS FALANDO DE UM BANCO Como exemplo, use essa lista em formato csv abaixo: 
service_id;service_name;intent
1;Consulta Limite / Vencimento do cartão / Melhor dia de compra;Quanto tem disponível para usar
1;Consulta Limite / Vencimento do cartão / Melhor dia de compra;quando fecha minha fatura
1;Consulta Limite / Vencimento do cartão / Melhor dia de compra;Quando vence meu cartão
1;Consulta Limite / Vencimento do cartão / Melhor dia de compra;quando posso comprar
1;Consulta Limite / Vencimento do cartão / Melhor dia de compra;vencimento da fatura
1;Consulta Limite / Vencimento do cartão / Melhor dia de compra;valor para gastar
2;Segunda via de boleto de acordo;segunda via boleto de acordo
2;Segunda via de boleto de acordo;Boleto para pagar minha negociação
2;Segunda via de boleto de acordo;código de barras acordo
2;Segunda via de boleto de acordo;preciso pagar negociação
2;Segunda via de boleto de acordo;enviar boleto acordo
2;Segunda via de boleto de acordo;boleto da negociação
3;Segunda via de Fatura;quero meu boleto
3;Segunda via de Fatura;segunda via de fatura
3;Segunda via de Fatura;código de barras fatura
3;Segunda via de Fatura;quero a fatura do cartão
3;Segunda via de Fatura;enviar boleto da fatura
3;Segunda via de Fatura;fatura para pagamento
4;Status de Entrega do Cartão;onde está meu cartão
4;Status de Entrega do Cartão;meu cartão não chegou
4;Status de Entrega do Cartão;status da entrega do cartão
4;Status de Entrega do Cartão;cartão em transporte
4;Status de Entrega do Cartão;previsão de entrega do cartão
4;Status de Entrega do Cartão;cartão foi enviado?
5;Status de cartão;não consigo passar meu cartão
5;Status de cartão;meu cartão não funciona
5;Status de cartão;cartão recusado
5;Status de cartão;cartão não está passando
5;Status de cartão;status do cartão ativo
5;Status de cartão;problema com cartão
6;Solicitação de aumento de limite;quero mais limite
6;Solicitação de aumento de limite;aumentar limite do cartão
6;Solicitação de aumento de limite;solicitar aumento de crédito
6;Solicitação de aumento de limite;preciso de mais limite
6;Solicitação de aumento de limite;pedido de aumento de limite
6;Solicitação de aumento de limite;limite maior no cartão
7;Cancelamento de cartão;cancelar cartão
7;Cancelamento de cartão;quero encerrar meu cartão
7;Cancelamento de cartão;bloquear cartão definitivamente
7;Cancelamento de cartão;cancelamento de crédito
7;Cancelamento de cartão;desistir do cartão
8;Telefones de seguradoras;quero cancelar seguro
8;Telefones de seguradoras;telefone do seguro
8;Telefones de seguradoras;contato da seguradora
8;Telefones de seguradoras;preciso falar com o seguro
8;Telefones de seguradoras;seguro do cartão
8;Telefones de seguradoras;cancelar assistência
9;Desbloqueio de Cartão;desbloquear cartão
9;Desbloqueio de Cartão;ativar cartão novo
9;Desbloqueio de Cartão;como desbloquear meu cartão
9;Desbloqueio de Cartão;quero desbloquear o cartão
9;Desbloqueio de Cartão;cartão para uso imediato
9;Desbloqueio de Cartão;desbloqueio para compras
10;Esqueceu senha / Troca de senha;não tenho mais a senha do cartão
10;Esqueceu senha / Troca de senha;esqueci minha senha
10;Esqueceu senha / Troca de senha;trocar senha do cartão
10;Esqueceu senha / Troca de senha;preciso de nova senha
10;Esqueceu senha / Troca de senha;recuperar senha
10;Esqueceu senha / Troca de senha;senha bloqueada
11;Perda e roubo;perdi meu cartão
11;Perda e roubo;roubaram meu cartão
11;Perda e roubo;cartão furtado
11;Perda e roubo;perda do cartão
11;Perda e roubo;bloquear cartão por roubo
11;Perda e roubo;extravio de cartão
12;Consulta do Saldo;saldo conta corrente
12;Consulta do Saldo;consultar saldo
12;Consulta do Saldo;quanto tenho na conta
12;Consulta do Saldo;extrato da conta
12;Consulta do Saldo;saldo disponível
12;Consulta do Saldo;meu saldo atual
13;Pagamento de contas;quero pagar minha conta
13;Pagamento de contas;pagar boleto
13;Pagamento de contas;pagamento de conta
13;Pagamento de contas;quero pagar fatura
13;Pagamento de contas;efetuar pagamento
14;Reclamações;quero reclamar
14;Reclamações;abrir reclamação
14;Reclamações;fazer queixa
14;Reclamações;reclamar atendimento
14;Reclamações;registrar problema
14;Reclamações;protocolo de reclamação
15;Atendimento humano;falar com uma pessoa
15;Atendimento humano;preciso de humano
15;Atendimento humano;transferir para atendente
15;Atendimento humano;quero falar com atendente
15;Atendimento humano;atendimento pessoal
16;Token de proposta;código para fazer meu cartão
16;Token de proposta;token de proposta
16;Token de proposta;receber código do cartão
16;Token de proposta;proposta token
16;Token de proposta;número de token
16;Token de proposta;código de token da proposta`,
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

	// Try to unmarshal as JSON first
	var dataRes DataResponse
	if err := json.Unmarshal([]byte(openRouterResp.Choices[0].Message.Content), &dataRes); err != nil {
		// If JSON unmarshaling fails, treat as plain text response
		content := openRouterResp.Choices[0].Message.Content

		// Map service names to IDs based on your list
		serviceID := mapServiceNameToID(content)

		dataRes = DataResponse{
			ServiceID:   serviceID,
			ServiceName: content,
		}
	}

	return &dataRes, nil
}

// mapServiceNameToID maps service names to their corresponding IDs
func mapServiceNameToID(serviceName string) uint8 {
	serviceMap := map[string]uint8{
		"Consulta Limite / Vencimento do cartão / Melhor dia de compra": 1,
		"Segunda via de boleto de acordo":                               2,
		"Segunda via de Fatura":                                         3,
		"Status de Entrega do Cartão":                                   4,
		"Status de cartão":                                              5,
		"Solicitação de aumento de limite":                              6,
		"Cancelamento de cartão":                                        7,
		"Telefones de seguradoras":                                      8,
		"Desbloqueio de Cartão":                                         9,
		"Esqueceu senha / Troca de senha":                               10,
		"Perda e roubo":                                                 11,
		"Consulta do Saldo":                                             12,
		"Pagamento de contas":                                           13,
		"Reclamações":                                                   14,
		"Atendimento humano":                                            15,
		"Token de proposta":                                             16,
	}

	if id, exists := serviceMap[serviceName]; exists {
		return id
	}
	return 0 // Default ID if not found
}
