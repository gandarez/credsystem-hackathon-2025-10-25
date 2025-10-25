package main

import (
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

	"github.com/go-chi/chi/v5"
)

type FindServiceRequest struct {
	Intent string `json:"intent"`
}

type FindServiceResponse struct {
	Success bool             `json:"success"`
	Data    *FindServiceData `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type FindServiceData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type openRouterResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

var services = []struct {
	ID   int
	Name string
}{
	{1, "Consulta Limite / Vencimento do cartão / Melhor dia de compra"},
	{2, "Segunda via de boleto de acordo"},
	{3, "Segunda via de Fatura"},
	{4, "Status de Entrega do Cartão"},
	{5, "Status de cartão"},
	{6, "Solicitação de aumento de limite"},
	{7, "Cancelamento de cartão"},
	{8, "Telefones de seguradoras"},
	{9, "Desbloqueio de Cartão"},
	{10, "Esqueceu senha / Troca de senha"},
	{11, "Perda e roubo"},
	{12, "Consulta do Saldo"},
	{13, "Pagamento de contas"},
	{14, "Reclamações"},
	{15, "Atendimento humano"},
	{16, "Token de proposta"},
}

func getServiceByID(id int) (string, bool) {
	for _, s := range services {
		if s.ID == id {
			return s.Name, true
		}
	}
	return "", false
}

func resolveWithLLM(ctx context.Context, intent string) (FindServiceData, bool, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return FindServiceData{}, false, errors.New("OPENROUTER_API_KEY não configurada")
	}

	// Monta lista de serviços como referência para o modelo
	list := ""
	for _, s := range services {
		list += fmt.Sprintf("%d - %s\n", s.ID, s.Name)
	}

	// Prompt otimizado: 100% acurácia + velocidade
	prompt := fmt.Sprintf(`Classifique intenção de cliente brasileiro sobre CARTÃO DE CRÉDITO/BANCO. Aceite gírias, erros e variações.

IMPORTANTE: Se NÃO for sobre cartão/banco/fatura, retorne {"id":0,"name":""}.

Serviços:
%s
REGRAS OBRIGATÓRIAS (priorize estas regras):

1. LIMITE/VENCIMENTO (ID 1):
   • "disponível usar/gastar/comprar/tem/valor"→1
   • "quando fecha/vence/vencimento fatura"→1
   • "melhor dia compra"→1

2. BOLETO ACORDO (ID 2):
   • "pagar negociação/acordo"→2 (obter boleto)
   • "segunda via acordo"→2

3. FATURA (ID 3):
   • "segunda via fatura"→3
   • "fatura para pagamento/cartão"→3 (obter)
   • "meu boleto" (sem especificar)→3
   • "código barras fatura"→3

4. ENTREGA CARTÃO (ID 4):
   • "onde está/não chegou/enviado cartão"→4

5. STATUS CARTÃO (ID 5):
   • "não funciona/recusado/não passa"→5
   • "problema cartão"→5

6. AUMENTO LIMITE (ID 6):
   • "quero mais limite/aumentar/maior"→6

7. CANCELAMENTO CARTÃO (ID 7):
   • "cancelar/encerrar/desistir cartão"→7
   • "cancelamento crédito"→7
   • "bloquear cartão" (SEM mencionar perda/roubo)→7
   • "bloquear por suspeita/golpe/fraude" (sem perda física)→7
   • "bloquear definitivamente"→7

8. SEGURO (ID 8):
   • "cancelar/quero cancelar seguro/assistência"→8
   • "telefone/contato seguro/seguradora"→8
   • "falar/preciso falar com seguro"→8
   • "seguro do cartão"→8

9. DESBLOQUEIO (ID 9):
   • "desbloquear/ativar cartão"→9
   • "cartão para uso imediato"→9

10. SENHA (ID 10):
    • "esqueci/trocar/recuperar senha"→10
    • "senha bloqueada"→10
    • "não tenho mais senha"→10
    • "preciso nova senha"→10

11. PERDA/ROUBO (ID 11):
    • "perdi/roubaram/furtado cartão"→11
    • "extravio/perda do cartão"→11
    • "bloquear por roubo/perda" (menciona perda REAL)→11

12. SALDO (ID 12):
    • "saldo conta corrente/disponível"→12
    • "consultar saldo"→12
    • "extrato da conta"→12
    • "quanto tenho na conta/meu saldo"→12

13. PAGAMENTO (ID 13):
    • "quero pagar conta/boleto"→13 (efetuar)
    • "pagar boleto" (sem especificar)→13
    • "pagamento conta"→13
    • "efetuar pagamento"→13
    • "quero/vou pagar fatura"→13

14. RECLAMAÇÕES (ID 14):
    • "quero reclamar"→14
    • "fazer queixa"→14
    • "abrir/registrar reclamação"→14
    • "registrar problema"→14
    • "protocolo reclamação"→14

15. ATENDIMENTO HUMANO (ID 15):
    • "falar pessoa/humano/atendente"→15
    • "preciso humano"→15
    • "transferir atendente"→15
    • "atendimento pessoal"→15

16. TOKEN (ID 16):
    • "token/código proposta/fazer cartão"→16
    • "receber código cartão"→16
    • "número token"→16

DIFERENÇAS CRÍTICAS:
• "bloquear" sem roubo/perda→7 | "bloquear por roubo/perda"→11
• "perdi/roubaram" (explícito)→11 | "suspeita" (sem perda)→7
• "cartão não funciona"→5 | "bloquear cartão"→7
• "pagar acordo"→2 | "pagar fatura/boleto"→13
• "obter fatura"→3 | "efetuar pagamento"→13

Frase: "%s"
JSON: {"id":N,"name":"nome exato da lista"}`, list, intent)

	reqBody := map[string]any{
		//		"model": "mistralai/mistral-7b-instruct",
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": "Você é um classificador de intenções de cliente."},
			{"role": "user", "content": prompt},
		},
		"temperature":     0.0,
		"response_format": map[string]string{"type": "json_object"},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	start := time.Now()
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Falha HTTP: %v\n", err)
		return FindServiceData{}, false, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	elapsed := time.Since(start).Milliseconds()
	log.Printf("[DEBUG] OpenRouter status: %s (%dms)\nBody: %s\n", resp.Status, elapsed, string(raw))

	if resp.StatusCode >= 400 {
		return FindServiceData{}, false, fmt.Errorf("erro HTTP %s", resp.Status)
	}

	var or openRouterResp
	if err := json.Unmarshal(raw, &or); err != nil {
		return FindServiceData{}, false, fmt.Errorf("falha ao decodificar resposta: %v", err)
	}
	if len(or.Choices) == 0 {
		return FindServiceData{}, false, errors.New("resposta vazia da IA")
	}

	type LLMOutput struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var out LLMOutput
	if err := json.Unmarshal([]byte(or.Choices[0].Message.Content), &out); err != nil {
		log.Printf("[WARN] Resposta não JSON: %s\n", or.Choices[0].Message.Content)
		return FindServiceData{}, false, err
	}

	if out.ID == 0 {
		return FindServiceData{}, false, nil
	}

	name, ok := getServiceByID(out.ID)
	if !ok {
		return FindServiceData{}, false, nil
	}

	return FindServiceData{ServiceID: out.ID, ServiceName: name}, true, nil
}

/* ================================
   Handlers HTTP
================================ */

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func findService(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start).Milliseconds()
		log.Printf("[INFO] POST /api/find-service processed in %dms\n", elapsed)
	}()

	var req FindServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"error":"payload inválido"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	result, ok, err := resolveWithLLM(ctx, req.Intent)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		writeJSON(w, FindServiceResponse{Success: false, Error: "erro ao consultar IA"})
		return
	}

	if !ok {
		writeJSON(w, FindServiceResponse{Success: false, Error: "Serviço não encontrado"})
		return
	}

	writeJSON(w, FindServiceResponse{Success: true, Data: &result})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

/* ================================
   Main
================================ */

func main() {
	r := chi.NewRouter()
	r.Get("/api/healthz", healthz)
	r.Post("/api/find-service", findService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 API online em http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
