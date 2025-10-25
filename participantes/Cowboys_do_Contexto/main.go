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
	"regexp"
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

// ---------------------- Fast-path (regex) ----------------------
// Se bater aqui, não chamamos o LLM (reduz muito a latência média)

type rule struct {
	re *regexp.Regexp
	id int
}

var fastRules = []rule{
	// 1: Limite / Vencimento / Melhor dia
	{regexp.MustCompile(`(?i)\b(limite|melhor\s*dia|vencimento|data\s*de\s*vencimento)\b`), 1},
	// 2: 2ª via boleto de acordo
	{regexp.MustCompile(`(?i)\b(segunda?\s*via|2.?a\s*via)\b.*\b(boleto)\b.*\b(acordo|renegocia|negocia)\b`), 2},
	{regexp.MustCompile(`(?i)\b(boleto)\b.*\b(acordo|renegocia|negocia)\b`), 2},
	// 3: 2ª via de Fatura
	{regexp.MustCompile(`(?i)\b(segunda?\s*via|2.?a\s*via)\b.*\b(fatura)\b`), 3},
	{regexp.MustCompile(`(?i)\b(fatura|boleto\s*da\s*fatura)\b`), 3},
	// 4: Status de entrega do cartão
	{regexp.MustCompile(`(?i)\b(entrega|rastrea|tracking|rastreamento)\b.*\b(cart[aã]o)\b`), 4},
	// 5: Status genérico do cartão
	{regexp.MustCompile(`(?i)\b(status)\b.*\b(cart[aã]o)\b`), 5},
	// 6: Aumento de limite
	{regexp.MustCompile(`(?i)\b(aumento\s+de\s+limite|aumentar\s+limite|elevar\s+limite)\b`), 6},
	// 7: Cancelamento de cartão
	{regexp.MustCompile(`(?i)\b(cancelar|cancelamento)\b.*\b(cart[aã]o)\b`), 7},
	// 8: Telefones de seguradoras
	{regexp.MustCompile(`(?i)\b(telefone|contato)\b.*\b(seguradora|seguros)\b`), 8},
	// 9: Desbloqueio de cartão
	{regexp.MustCompile(`(?i)\b(desbloqueio|desbloquear)\b.*\b(cart[aã]o)\b`), 9},
	// 10: Esqueceu/Troca de senha
	{regexp.MustCompile(`(?i)\b(esqueci|esqueceu|trocar|redefinir|resetar)\b.*\b(senha|pin)\b`), 10},
	// 11: Perda e roubo
	{regexp.MustCompile(`(?i)\b(perdi|perda|roubo|furt[oa]|extravio|sumiu)\b.*\b(cart[aã]o)\b`), 11},
	// 12: Saldo Conta do Mais
	{regexp.MustCompile(`(?i)\b(saldo)\b.*\b(conta\s*do\s*mais|conta\s+mais)\b`), 12},
	// 13: Pagamento de contas
	{regexp.MustCompile(`(?i)\b(pagamento|pagar)\b.*\b(contas?|boleto[s]?)\b`), 13},
	// 14: Reclamações
	{regexp.MustCompile(`(?i)\b(reclama[cç][aã]o|reclamar|queixa)\b`), 14},
	// 15: Atendimento humano
	{regexp.MustCompile(`(?i)\b(falar\s+com\s+(atendente|humano)|atendimento\s+humano|transferir)\b`), 15},
	// 16: Token de proposta
	{regexp.MustCompile(`(?i)\b(token|c[oó]digo)\b.*\b(proposta)\b`), 16},
	// 17: Atualização cadastral
	{regexp.MustCompile(`(?i)\b(atualizar|atualiza[cç][aã]o|corrigir|alterar)\b.*\b(dados|cadastro|telefone|endereco|endereço|email|e-mail)\b`), 17},
}

func fastPath(intent string) (int, bool) {
	s := strings.ToLower(intent)
	for _, r := range fastRules {
		if r.re.MatchString(s) {
			return r.id, true
		}
	}
	return 0, false
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
	Timeout: 8 * time.Second, // fail-fast geral; o handler ainda impõe um deadline próprio
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

// classifyWithLLM chama o provedor com payload mínimo e extrai apenas o service_id
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
		"max_tokens":  8, // suficiente para {"service_id":N}
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
		// regra de segurança: se vier inválido, force atendimento humano
		return 15, nil
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

		// 1) Fast-path determinístico (latência ~0)
		if id, ok := fastPath(intent); ok {
			name := catalog[id]
			returnJSON(w, http.StatusOK, apiResponse{
				Success: true,
				Data:    &responseData{ServiceID: id, ServiceName: name},
			})
			return
		}

		// 2) Chamada ao LLM com deadline curto (SLA)
		ctx, cancel := context.WithTimeout(r.Context(), 800*time.Millisecond)
		defer cancel()

		id, err := classifyWithLLM(ctx, intent)
		if err != nil {
			// Em erro/transiente, degrade para humano (15) rapidamente
			returnJSON(w, http.StatusBadGateway, apiResponse{Success: false, Error: err.Error()})
			return
		}

		name := catalog[id]
		if name == "" {
			// sanity-check adicional
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
