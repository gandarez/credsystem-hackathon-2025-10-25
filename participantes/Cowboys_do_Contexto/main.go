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
	"time"
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
	12: "Consulta do Saldo Conta do Mais",
	13: "Pagamento de contas",
	14: "Reclamações",
	15: "Atendimento humano",
	16: "Token de proposta",
}

// ---------------------- HTTP client ----------------------

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxConnsPerHost:     0,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
	},
	Timeout: 8 * time.Second,
}

// ---------------------- LLM (OpenRouter) ----------------------

const llmModel = "openai/gpt-4o-mini"

// Saída do modelo: APENAS o número 1..16 (sem JSON/explicações)
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
12 Consulta do Saldo Conta do Mais
13 Pagamento de contas
14 Reclamações
15 Atendimento humano
16 Token de proposta
REGRAS:
- Retorne SOMENTE o número do ID (ex.: 7).
- Não retorne texto, explicações, JSON, aspas ou qualquer outro símbolo.
- Se houver dúvida ou ambiguidade, responda 15.`

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

	// healthz simples
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		returnJSON(w, map[string]string{"status": "ok"})
	})

	// Classificação de serviço
	mux.HandleFunc("/api/find-service", func(w http.ResponseWriter, r *http.Request) {
		var reqBody requestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || strings.TrimSpace(reqBody.Intent) == "" {
			returnJSON(w, apiResponse{
				Success: false,
				Data:    nil,
				Error:   "body inválido: {\"intent\": \"...\"}",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		id, err := classifyWithLLM(ctx, strings.TrimSpace(reqBody.Intent))
		if err != nil {
			returnJSON(w, apiResponse{
				Success: false,
				Data:    nil,
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "18020"
	}

	addr := ":" + port
	server := &http.Server{
		Addr:              addr,
		Handler:           makeServer(),
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       90 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}

	log.Printf("servidor ouvindo em %s", addr)
	log.Fatal(server.ListenAndServe())
}
