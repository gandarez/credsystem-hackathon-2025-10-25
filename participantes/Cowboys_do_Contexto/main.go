package main

import (
	"bufio"
	"bytes"
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

// ---------- Chamada ao LLM (OpenRouter) ----------

func classifyWithLLM(intent string, timeout time.Duration) (*responseData, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENROUTER_API_KEY não definido")
	}

	model := "openai/gpt-4o-mini"

	// Prompt ÚNICO: a IA decide tudo
	systemPrompt := `Você é um CLASSIFICADOR DE INTENÇÕES para uma URA. Sua tarefa é mapear a frase do usuário
para EXATAMENTE UM serviço dentre a lista CANÔNICA abaixo.

REGRAS GERAIS (SIGA À RISCA):
1) Escolha apenas entre estes serviços (ID -> Nome canônico):
   1  -> Consulta Limite / Vencimento do cartão / Melhor dia de compra
   2  -> Segunda via de boleto de acordo
   3  -> Segunda via de Fatura
   4  -> Status de Entrega do Cartão
   5  -> Status de cartão
   6  -> Solicitação de aumento de limite
   7  -> Cancelamento de cartão
   8  -> Telefones de seguradoras
   9  -> Desbloqueio de Cartão
   10 -> Esqueceu senha / Troca de senha
   11 -> Perda e roubo
   12 -> Consulta do Saldo Conta do Mais
   13 -> Pagamento de contas
   14 -> Reclamações
   15 -> Atendimento humano
   16 -> Token de proposta
   17 -> Atualização de dados cadastrais

2) SAÍDA OBRIGATÓRIA: retorne SOMENTE um JSON compacto, sem texto extra:
   {"service_id": <int>, "service_name": "<string exata da lista acima>"}

3) Se houver AMBIGUIDADE OU FALTAR CONTEXTO, retorne:
   {"service_id": 15, "service_name": "Atendimento humano"}

4) NÃO invente novos serviços, NÃO traduza nomes, NÃO explique.

5) Critérios rápidos de mapeamento (PT-BR):
   - Fatura do cartão → 3
   - Boleto de acordo/renegociação → 2
   - Limite/melhor dia/vencimento → 1
   - Entrega/rastreamento do cartão → 4
   - Status genérico do cartão → 5
   - Aumento de limite → 6
   - Cancelamento do cartão → 7
   - Telefones de seguradoras → 8
   - Desbloqueio do cartão → 9
   - Esqueci/Trocar senha → 10
   - Perda/Roubo/Furto/Extravio → 11
   - Saldo Conta do Mais → 12
   - Pagamento de contas → 13
   - Reclamações → 14
   - Falar com atendente → 15
   - Token de proposta/código proposta → 16
   - Atualização de dados/telefone/endereço → 17

6) Desempate: se duas opções parecerem possíveis, prefira a MAIS ESPECÍFICA.
   Se ainda empatar → 15 (Atendimento humano).

7) Trate variações, erros de digitação e sinônimos do PT-BR como equivalentes.`

	// few-shots curtos só para “ancorar” o formato
	fewshots := []map[string]string{
		{"role": "user", "content": "perdi meu cartão no ônibus"},
		{"role": "assistant", "content": `{"service_id": 11, "service_name": "Perda e roubo"}`},

		{"role": "user", "content": "quero 2a via da fatura"},
		{"role": "assistant", "content": `{"service_id": 3, "service_name": "Segunda via de Fatura"}`},

		{"role": "user", "content": "atualizar meu telefone e endereço"},
		{"role": "assistant", "content": `{"service_id": 17, "service_name": "Atualização de dados cadastrais"}`},
	}

	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
	}
	messages = append(messages, fewshots...)
	messages = append(messages, map[string]string{"role": "user", "content": intent})

	payload := map[string]interface{}{
		"model":       model,
		"temperature": 0,
		"top_p":       1,
		"max_tokens":  50,
		"messages":    messages,
		"response_format": map[string]string{
			"type": "json_object",
		},
		// evita texto fora do JSON
		"stop": []string{"\n\n", "\nUsuário:", "Resposta (JSON apenas):"},
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "ura-classifier-proxy")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(bufio.NewReader(resp.Body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openrouter: http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// estrutura mínima da resposta do OpenRouter
	var or struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &or); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta do LLM: %w", err)
	}
	if len(or.Choices) == 0 {
		return nil, errors.New("openrouter: nenhuma choice retornada")
	}

	content := strings.TrimSpace(or.Choices[0].Message.Content)

	// Tenta decodificar como JSON direto
	var out responseData
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		// fallback: extrai primeiro bloco { ... }
		re := regexp.MustCompile(`\{[\s\S]*?\}`)
		m := re.FindString(content)
		if m == "" {
			return nil, fmt.Errorf("resposta do LLM sem JSON: %s", content)
		}
		if err2 := json.Unmarshal([]byte(m), &out); err2 != nil {
			return nil, fmt.Errorf("JSON inválido do LLM: %v | %s", err2, m)
		}
	}

	// Sem validação de catálogo aqui — a IA é a “fonte da verdade”
	// (opcional: validar service_id 1..17 e nomes, se quiser “fail-fast”)
	if out.ServiceID == 0 || strings.TrimSpace(out.ServiceName) == "" {
		return nil, fmt.Errorf("JSON do LLM sem campos esperados: %s", content)
	}
	return &out, nil
}

// ---------- HTTP ----------

func makeServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/find-service", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(apiResponse{Success: false, Error: "use POST"})
			return
		}

		var reqBody requestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil ||
			strings.TrimSpace(reqBody.Intent) == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(apiResponse{Success: false, Error: "body inválido: {\"intent\": \"...\"}"})
			return
		}

		out, err := classifyWithLLM(reqBody.Intent, 12*time.Second)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(apiResponse{Success: false, Error: err.Error()})
			return
		}

		_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: out})
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
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("servidor ouvindo em %s", addr)
	log.Fatal(server.ListenAndServe())
}
