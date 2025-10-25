package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// ---------------------- Tipos ----------------------

type requestBody struct {
	Intent string `json:"intent"`
}

type responseData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type apiResponse struct {
	Success bool          `json:"success"`
	Data    *responseData `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// ---------------------- Catálogo ----------------------

var catalog = map[int]string{
	1:  "Consulta Limite / Vencimento do cartão / Melhor dia de compra",
	2:  "Segunda via de boleto de acordo",
	3:  "Segunda via de Fatura",
	4:  "Status de Entrega do Cartão",
	5:  "Status de cartão",
	6:  "Solicitação de aumento de limite",
	7:  "Cancelamento de cartão",
	8:  "Telefones de seguradoras",
	9:  "Desbloqueio de Cartão",
	10: "Esqueceu senha / Troca de senha",
	11: "Perda e roubo",
	12: "Consulta do Saldo",
	13: "Pagamento de contas",
	14: "Reclamações",
	15: "Atendimento humano",
	16: "Token de proposta",
}

// ---------------------- HTTP client (sem timeout) ----------------------

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		ForceAttemptHTTP2:   true,
	},
	// Timeout removido completamente
}

// ---------------------- LLM (OpenRouter) ----------------------

const llmModel = "openai/gpt-4o-mini"

const systemPrompt = `Você deve classificar a intenção do usuário como UM ÚNICO ID entre 1 e 16 da lista:
1 Consulta Limite / Vencimento do cartão / Melhor dia de compra
2 Segunda via de boleto de acordo
3 Segunda via de Fatura
4 Status de Entrega do Cartão
5 Status de cartão
6 Solicitação de aumento de limite
7 Cancelamento de cartão
8 Telefones de seguradoras
9 Desbloqueio de Cartão
10 Esqueceu senha / Troca de senha
11 Perda e roubo
12 Consulta do Saldo
13 Pagamento de contas
14 Reclamações
15 Atendimento humano
16 Token de proposta
REGRAS:
- Retorne SOMENTE o número do ID (ex.: 7).
- Não retorne texto, explicações, JSON, aspas ou qualquer outro símbolo.
- Se houver dúvida ou ambiguidade, responda 15, mas antes, pense

Exemplos para você treinar:
[{"service_id":1,"service_name":"Consulta Limite \/ Vencimento do cartão \/ Melhor dia de compra","intent":"Quanto tem disponível para usar"},{"service_id":1,"service_name":"Consulta Limite \/ Vencimento do cartão \/ Melhor dia de compra","intent":"quando fecha minha fatura"},{"service_id":1,"service_name":"Consulta Limite \/ Vencimento do cartão \/ Melhor dia de compra","intent":"Quando vence meu cartão"},{"service_id":1,"service_name":"Consulta Limite \/ Vencimento do cartão \/ Melhor dia de compra","intent":"quando posso comprar"},{"service_id":1,"service_name":"Consulta Limite \/ Vencimento do cartão \/ Melhor dia de compra","intent":"vencimento da fatura"},{"service_id":1,"service_name":"Consulta Limite \/ Vencimento do cartão \/ Melhor dia de compra","intent":"valor para gastar"},{"service_id":2,"service_name":"Segunda via de boleto de acordo","intent":"segunda via boleto de acordo"},{"service_id":2,"service_name":"Segunda via de boleto de acordo","intent":"Boleto para pagar minha negociação"},{"service_id":2,"service_name":"Segunda via de boleto de acordo","intent":"código de barras acordo"},{"service_id":2,"service_name":"Segunda via de boleto de acordo","intent":"preciso pagar negociação"},{"service_id":2,"service_name":"Segunda via de boleto de acordo","intent":"enviar boleto acordo"},{"service_id":2,"service_name":"Segunda via de boleto de acordo","intent":"boleto da negociação"},{"service_id":3,"service_name":"Segunda via de Fatura","intent":"quero meu boleto"},{"service_id":3,"service_name":"Segunda via de Fatura","intent":"segunda via de fatura"},{"service_id":3,"service_name":"Segunda via de Fatura","intent":"código de barras fatura"},{"service_id":3,"service_name":"Segunda via de Fatura","intent":"quero a fatura do cartão"},{"service_id":3,"service_name":"Segunda via de Fatura","intent":"enviar boleto da fatura"},{"service_id":3,"service_name":"Segunda via de Fatura","intent":"fatura para pagamento"},{"service_id":4,"service_name":"Status de Entrega do Cartão","intent":"onde está meu cartão"},{"service_id":4,"service_name":"Status de Entrega do Cartão","intent":"meu cartão não chegou"},{"service_id":4,"service_name":"Status de Entrega do Cartão","intent":"status da entrega do cartão"},{"service_id":4,"service_name":"Status de Entrega do Cartão","intent":"cartão em transporte"},{"service_id":4,"service_name":"Status de Entrega do Cartão","intent":"previsão de entrega do cartão"},{"service_id":4,"service_name":"Status de Entrega do Cartão","intent":"cartão foi enviado?"},{"service_id":5,"service_name":"Status de cartão","intent":"não consigo passar meu cartão"},{"service_id":5,"service_name":"Status de cartão","intent":"meu cartão não funciona"},{"service_id":5,"service_name":"Status de cartão","intent":"cartão recusado"},{"service_id":5,"service_name":"Status de cartão","intent":"cartão não está passando"},{"service_id":5,"service_name":"Status de cartão","intent":"status do cartão ativo"},{"service_id":5,"service_name":"Status de cartão","intent":"problema com cartão"},{"service_id":6,"service_name":"Solicitação de aumento de limite","intent":"quero mais limite"},{"service_id":6,"service_name":"Solicitação de aumento de limite","intent":"aumentar limite do cartão"},{"service_id":6,"service_name":"Solicitação de aumento de limite","intent":"solicitar aumento de crédito"},{"service_id":6,"service_name":"Solicitação de aumento de limite","intent":"preciso de mais limite"},{"service_id":6,"service_name":"Solicitação de aumento de limite","intent":"pedido de aumento de limite"},{"service_id":6,"service_name":"Solicitação de aumento de limite","intent":"limite maior no cartão"},{"service_id":7,"service_name":"Cancelamento de cartão","intent":"cancelar cartão"},{"service_id":7,"service_name":"Cancelamento de cartão","intent":"quero encerrar meu cartão"},{"service_id":7,"service_name":"Cancelamento de cartão","intent":"bloquear cartão definitivamente"},{"service_id":7,"service_name":"Cancelamento de cartão","intent":"cancelamento de crédito"},{"service_id":7,"service_name":"Cancelamento de cartão","intent":"desistir do cartão"},{"service_id":8,"service_name":"Telefones de seguradoras","intent":"quero cancelar seguro"},{"service_id":8,"service_name":"Telefones de seguradoras","intent":"telefone do seguro"},{"service_id":8,"service_name":"Telefones de seguradoras","intent":"contato da seguradora"},{"service_id":8,"service_name":"Telefones de seguradoras","intent":"preciso falar com o seguro"},{"service_id":8,"service_name":"Telefones de seguradoras","intent":"seguro do cartão"},{"service_id":8,"service_name":"Telefones de seguradoras","intent":"cancelar assistência"},{"service_id":9,"service_name":"Desbloqueio de Cartão","intent":"desbloquear cartão"},{"service_id":9,"service_name":"Desbloqueio de Cartão","intent":"ativar cartão novo"},{"service_id":9,"service_name":"Desbloqueio de Cartão","intent":"como desbloquear meu cartão"},{"service_id":9,"service_name":"Desbloqueio de Cartão","intent":"quero desbloquear o cartão"},{"service_id":9,"service_name":"Desbloqueio de Cartão","intent":"cartão para uso imediato"},{"service_id":9,"service_name":"Desbloqueio de Cartão","intent":"desbloqueio para compras"},{"service_id":10,"service_name":"Esqueceu senha \/ Troca de senha","intent":"não tenho mais a senha do cartão"},{"service_id":10,"service_name":"Esqueceu senha \/ Troca de senha","intent":"esqueci minha senha"},{"service_id":10,"service_name":"Esqueceu senha \/ Troca de senha","intent":"trocar senha do cartão"},{"service_id":10,"service_name":"Esqueceu senha \/ Troca de senha","intent":"preciso de nova senha"},{"service_id":10,"service_name":"Esqueceu senha \/ Troca de senha","intent":"recuperar senha"},{"service_id":10,"service_name":"Esqueceu senha \/ Troca de senha","intent":"senha bloqueada"},{"service_id":11,"service_name":"Perda e roubo","intent":"perdi meu cartão"},{"service_id":11,"service_name":"Perda e roubo","intent":"roubaram meu cartão"},{"service_id":11,"service_name":"Perda e roubo","intent":"cartão furtado"},{"service_id":11,"service_name":"Perda e roubo","intent":"perda do cartão"},{"service_id":11,"service_name":"Perda e roubo","intent":"bloquear cartão por roubo"},{"service_id":11,"service_name":"Perda e roubo","intent":"extravio de cartão"},{"service_id":12,"service_name":"Consulta do Saldo","intent":"saldo conta corrente"},{"service_id":12,"service_name":"Consulta do Saldo","intent":"consultar saldo"},{"service_id":12,"service_name":"Consulta do Saldo","intent":"quanto tenho na conta"},{"service_id":12,"service_name":"Consulta do Saldo","intent":"extrato da conta"},{"service_id":12,"service_name":"Consulta do Saldo","intent":"saldo disponível"},{"service_id":12,"service_name":"Consulta do Saldo","intent":"meu saldo atual"},{"service_id":13,"service_name":"Pagamento de contas","intent":"quero pagar minha conta"},{"service_id":13,"service_name":"Pagamento de contas","intent":"pagar boleto"},{"service_id":13,"service_name":"Pagamento de contas","intent":"pagamento de conta"},{"service_id":13,"service_name":"Pagamento de contas","intent":"quero pagar fatura"},{"service_id":13,"service_name":"Pagamento de contas","intent":"efetuar pagamento"},{"service_id":14,"service_name":"Reclamações","intent":"quero reclamar"},{"service_id":14,"service_name":"Reclamações","intent":"abrir reclamação"},{"service_id":14,"service_name":"Reclamações","intent":"fazer queixa"},{"service_id":14,"service_name":"Reclamações","intent":"reclamar atendimento"},{"service_id":14,"service_name":"Reclamações","intent":"registrar problema"},{"service_id":14,"service_name":"Reclamações","intent":"protocolo de reclamação"},{"service_id":15,"service_name":"Atendimento humano","intent":"falar com uma pessoa"},{"service_id":15,"service_name":"Atendimento humano","intent":"preciso de humano"},{"service_id":15,"service_name":"Atendimento humano","intent":"transferir para atendente"},{"service_id":15,"service_name":"Atendimento humano","intent":"quero falar com atendente"},{"service_id":15,"service_name":"Atendimento humano","intent":"atendimento pessoal"},{"service_id":16,"service_name":"Token de proposta","intent":"código para fazer meu cartão"},{"service_id":16,"service_name":"Token de proposta","intent":"token de proposta"},{"service_id":16,"service_name":"Token de proposta","intent":"receber código do cartão"},{"service_id":16,"service_name":"Token de proposta","intent":"proposta token"},{"service_id":16,"service_name":"Token de proposta","intent":"número de token"},{"service_id":16,"service_name":"Token de proposta","intent":"código de token da proposta"}]`

func classifyWithLLM(ctx context.Context, intent string) (int, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return 0, errors.New("OPENROUTER_API_KEY não definido")
	}

	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": intent},
	}

	payload := map[string]interface{}{
		"model":       llmModel,
		"temperature": 0,
		"max_tokens":  4,
		"messages":    messages,
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "ura-classifier-proxy")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(bufio.NewReader(resp.Body))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("openrouter: http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var or struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &or); err != nil {
		return 0, fmt.Errorf("falha ao decodificar resposta do LLM: %w", err)
	}
	if len(or.Choices) == 0 {
		return 0, errors.New("openrouter: nenhuma choice retornada")
	}

	content := strings.TrimSpace(or.Choices[0].Message.Content)
	content = strings.Trim(content, " \t\r\n\"'`.,;")

	id, err := strconv.Atoi(content)
	if err != nil || id < 1 || id > 16 {
		return 15, nil // fallback seguro
	}
	return id, nil
}

// ---------------------- HTTP ----------------------

func returnJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func makeServer() http.Handler {
	mux := http.NewServeMux()

	// Healthz
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		returnJSON(w, map[string]string{"status": "ok"})
	})

	// Classificador
	mux.HandleFunc("/api/find-service", func(w http.ResponseWriter, r *http.Request) {
		var reqBody requestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || strings.TrimSpace(reqBody.Intent) == "" {
			returnJSON(w, apiResponse{
				Success: false,
				Error:   "body inválido: {\"intent\": \"...\"}",
			})
			return
		}

		intent := strings.TrimSpace(reqBody.Intent)
		id, err := classifyWithLLM(context.Background(), intent)
		if err != nil {
			returnJSON(w, apiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		name := catalog[id]
		if name == "" {
			id = 15
			name = catalog[id]
		}

		returnJSON(w, apiResponse{
			Success: true,
			Data:    &responseData{ServiceID: id, ServiceName: name},
			Error:   "",
		})
	})

	return mux
}

func main() {
	port := "8080"

	addr := ":" + port
	server := &http.Server{
		Addr:    addr,
		Handler: makeServer(),
		// sem timeouts
	}

	log.Printf("servidor ouvindo em %s", addr)
	log.Fatal(server.ListenAndServe())
}
