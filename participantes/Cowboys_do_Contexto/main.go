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
	17: "Atualização de dados cadastrais",
}

// ---------------------- HTTP client global (keep-alive/HTTP2) ----------------------

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxConnsPerHost:     0,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
	},
	Timeout: 8 * time.Second, // fail-fast
}

// ---------------------- LLM (OpenRouter) ----------------------

const llmModel = "openai/gpt-4o-mini"

// Prompt mínimo: saída só com {"service_id": N}
const systemPrompt = `Você classifica a intenção do usuário como UM ÚNICO ID entre 1..17 da lista:
1 Limite/Vencimento/Melhor dia; 2 2ª via boleto de acordo; 3 2ª via de Fatura; 4 Status de Entrega do Cartão;
5 Status de cartão; 6 Aumento de limite; 7 Cancelamento de cartão; 8 Telefones de seguradoras; 9 Desbloqueio de Cartão;
10 Esqueceu/Troca de senha; 11 Perda e roubo; 12 Saldo Conta do Mais; 13 Pagamento de contas; 14 Reclamações;
15 Atendimento humano; 16 Token de proposta; 17 Atualização cadastral.
Regras: Retorne SOMENTE JSON {"service_id":<int>}. Em caso de ambiguidade ou falta de contexto, retorne {"service_id":15}.`

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
		"max_tokens":  8, // só precisa do {"service_id":N}
		"messages":    messages,
		"response_format": map[string]string{
			"type": "json_object",
		},
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

	// Esperamos JSON estrito {"service_id":N}
	var out struct {
		ServiceID int `json:"service_id"`
	}
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return 0, fmt.Errorf("resposta não-JSON do LLM: %s", content)
	}
	if out.ServiceID < 1 || out.ServiceID > 17 {
		return 15, nil // fallback seguro
	}
	return out.ServiceID, nil
}

// ---------------------- HTTP ----------------------

func returnJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func makeServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		returnJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/find-service", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			returnJSON(w, http.StatusMethodNotAllowed, apiResponse{Success: false, Error: "use POST"})
			return
		}

		var reqBody requestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || strings.TrimSpace(reqBody.Intent) == "" {
			returnJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: "body inválido: {\"intent\": \"...\"}"})
			return
		}

		intent := strings.TrimSpace(reqBody.Intent)

		// chama o LLM sempre (sem fast-path)
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		id, err := classifyWithLLM(ctx, intent)
		if err != nil {
			returnJSON(w, http.StatusBadGateway, apiResponse{Success: false, Error: err.Error()})
			return
		}

		name := catalog[id]
		if name == "" {
			id = 15
			name = catalog[id]
		}

		returnJSON(w, http.StatusOK, apiResponse{
			Success: true,
			Data:    &responseData{ServiceID: id, ServiceName: name},
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
