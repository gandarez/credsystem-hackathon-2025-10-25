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
	0:  "Serviço não identificado",
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
}

// ---------------------- LLM (OpenRouter) ----------------------

const llmModel = "openai/gpt-4o-mini"

const systemPrompt = `{"t":"Classificar intenção (0–16) e retornar só o número. Se dúvida, 15 ou 0, avalie o contexto.","r":[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"s":{"t":"string","id":"int(0–16)"},"e":[{"t":"Serviço não identificado","i":0},{"t":"Melhor dia de compra","i":1},{"t":"2ª via boleto","i":2},{"t":"2ª via fatura","i":3},{"t":"Entrega do cartão","i":4},{"t":"Cartão ativo?","i":5},{"t":"Aumento limite","i":6},{"t":"Cancelar cartão","i":7},{"t":"Telefone seguradora","i":8},{"t":"Desbloquear cartão","i":9},{"t":"Esqueci senha","i":10},{"t":"Perdi cartão","i":11},{"t":"Consultar saldo","i":12},{"t":"Pagar conta","i":13},{"t":"Reclamação","i":14},{"t":"Atendente","i":15},{"t":"Token proposta","i":16},{"t":"Ajuda","i":15}],"examples":[{"service_id":0,"intent":"oi tudo bem"},{"service_id":0,"intent":"preciso de ajuda mas não sei com o que"},{"service_id":0,"intent":"como funciona o aplicativo"},{"service_id":0,"intent":"quero saber mais sobre vocês"},{"service_id":0,"intent":"não entendi o que fazer"},{"service_id":0,"intent":"ajuda geral"},{"service_id":1,"intent":"Quanto tem disponível para usar"},{"service_id":1,"intent":"quando fecha minha fatura"},{"service_id":1,"intent":"Quando vence meu cartão"},{"service_id":1,"intent":"quando posso comprar"},{"service_id":1,"intent":"vencimento da fatura"},{"service_id":1,"intent":"valor para gastar"},{"service_id":2,"intent":"segunda via boleto de acordo"},{"service_id":2,"intent":"Boleto para pagar minha negociação"},{"service_id":2,"intent":"código de barras acordo"},{"service_id":2,"intent":"preciso pagar negociação"},{"service_id":2,"intent":"enviar boleto acordo"},{"service_id":2,"intent":"boleto da negociação"},{"service_id":3,"intent":"quero meu boleto"},{"service_id":3,"intent":"segunda via de fatura"},{"service_id":3,"intent":"código de barras fatura"},{"service_id":3,"intent":"quero a fatura do cartão"},{"service_id":3,"intent":"enviar boleto da fatura"},{"service_id":3,"intent":"fatura para pagamento"},{"service_id":4,"intent":"onde está meu cartão"},{"service_id":4,"intent":"meu cartão não chegou"},{"service_id":4,"intent":"status da entrega do cartão"},{"service_id":4,"intent":"cartão em transporte"},{"service_id":4,"intent":"previsão de entrega do cartão"},{"service_id":4,"intent":"cartão foi enviado?"},{"service_id":5,"intent":"não consigo passar meu cartão"},{"service_id":5,"intent":"meu cartão não funciona"},{"service_id":5,"intent":"cartão recusado"},{"service_id":5,"intent":"cartão não está passando"},{"service_id":5,"intent":"status do cartão ativo"},{"service_id":5,"intent":"problema com cartão"},{"service_id":6,"intent":"quero mais limite"},{"service_id":6,"intent":"aumentar limite do cartão"},{"service_id":6,"intent":"solicitar aumento de crédito"},{"service_id":6,"intent":"preciso de mais limite"},{"service_id":6,"intent":"pedido de aumento de limite"},{"service_id":6,"intent":"limite maior no cartão"},{"service_id":7,"intent":"cancelar cartão"},{"service_id":7,"intent":"quero encerrar meu cartão"},{"service_id":7,"intent":"bloquear cartão definitivamente"},{"service_id":7,"intent":"cancelamento de crédito"},{"service_id":7,"intent":"desistir do cartão"},{"service_id":8,"intent":"quero cancelar seguro"},{"service_id":8,"intent":"telefone do seguro"},{"service_id":8,"intent":"contato da seguradora"},{"service_id":8,"intent":"preciso falar com o seguro"},{"service_id":8,"intent":"seguro do cartão"},{"service_id":8,"intent":"cancelar assistência"},{"service_id":9,"intent":"desbloquear cartão"},{"service_id":9,"intent":"ativar cartão novo"},{"service_id":9,"intent":"como desbloquear meu cartão"},{"service_id":9,"intent":"quero desbloquear o cartão"},{"service_id":9,"intent":"cartão para uso imediato"},{"service_id":9,"intent":"desbloqueio para compras"},{"service_id":10,"intent":"não tenho mais a senha do cartão"},{"service_id":10,"intent":"esqueci minha senha"},{"service_id":10,"intent":"trocar senha do cartão"},{"service_id":10,"intent":"preciso de nova senha"},{"service_id":10,"intent":"recuperar senha"},{"service_id":10,"intent":"senha bloqueada"},{"service_id":11,"intent":"perdi meu cartão"},{"service_id":11,"intent":"roubaram meu cartão"},{"service_id":11,"intent":"cartão furtado"},{"service_id":11,"intent":"perda do cartão"},{"service_id":11,"intent":"bloquear cartão por roubo"},{"service_id":11,"intent":"extravio de cartão"},{"service_id":12,"intent":"saldo conta corrente"},{"service_id":12,"intent":"consultar saldo"},{"service_id":12,"intent":"quanto tenho na conta"},{"service_id":12,"intent":"extrato da conta"},{"service_id":12,"intent":"saldo disponível"},{"service_id":12,"intent":"meu saldo atual"},{"service_id":13,"intent":"quero pagar minha conta"},{"service_id":13,"intent":"pagar boleto"},{"service_id":13,"intent":"pagamento de conta"},{"service_id":13,"intent":"quero pagar fatura"},{"service_id":13,"intent":"efetuar pagamento"},{"service_id":14,"intent":"quero reclamar"},{"service_id":14,"intent":"abrir reclamação"},{"service_id":14,"intent":"fazer queixa"},{"service_id":14,"intent":"reclamar atendimento"},{"service_id":14,"intent":"registrar problema"},{"service_id":14,"intent":"protocolo de reclamação"},{"service_id":15,"intent":"falar com uma pessoa"},{"service_id":15,"intent":"preciso de humano"},{"service_id":15,"intent":"transferir para atendente"},{"service_id":15,"intent":"quero falar com atendente"},{"service_id":15,"intent":"atendimento pessoal"},{"service_id":16,"intent":"código para fazer meu cartão"},{"service_id":16,"intent":"token de proposta"},{"service_id":16,"intent":"receber código do cartão"},{"service_id":16,"intent":"proposta token"},{"service_id":16,"intent":"número de token"},{"service_id":16,"intent":"código de token da proposta"}]}`

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
		return 0, nil // serviço não identificado
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

	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		returnJSON(w, map[string]string{"status": "ok"})
	})

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
				Data: &responseData{
					ServiceID:   0,
					ServiceName: catalog[0],
				},
			})
			return
		}

		name := catalog[id]
		if id == 0 || name == "" {
			returnJSON(w, apiResponse{
				Success: false,
				Error:   "serviço não identificado",
				Data: &responseData{
					ServiceID:   0,
					ServiceName: catalog[0],
				},
			})
			return
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
	}

	log.Printf("servidor ouvindo em %s", addr)
	log.Fatal(server.ListenAndServe())
}
