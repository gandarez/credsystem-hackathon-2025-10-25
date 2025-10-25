package modelia

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type FindServiceRequest struct {
	Intent string `json:"intent"`
}

type Service struct {
	ID   int    `json:"service_id"`
	Name string `json:"service_name"`
}

type FindServiceResponse struct {
	Success bool     `json:"success"`
	Data    *Service `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type IntentCSV struct {
	Intent      string
	ServiceID   int
	ServiceName string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	intents, err := loadIntentsCSV("./assets/intents_pre_loaded.csv")
	if err != nil {
		log.Fatalf("Erro ao carregar CSV: %v", err)
	}

	http.HandleFunc("/api/find-service", findServiceHandler(intents))
	http.HandleFunc("/api/healthz", healthHandler)

	log.Printf("Servidor rodando na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loadIntentsCSV(path string) ([]IntentCSV, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var intents []IntentCSV
	for _, r := range records[1:] {
		var id int
		fmt.Sscanf(r[1], "%d", &id)
		intents = append(intents, IntentCSV{
			Intent:      r[0],
			ServiceID:   id,
			ServiceName: r[2],
		})
	}

	return intents, nil
}

func buildPrompt(intent string, intents []IntentCSV) string {
	var sb strings.Builder

	sb.WriteString("Você é um assistente que identifica qual serviço da Credsystem corresponde à intenção do cliente.\n")
	sb.WriteString("Regras importantes:\n")
	sb.WriteString("1. Use apenas os serviços listados abaixo.\n")
	sb.WriteString("2. Não invente serviços.\n")
	sb.WriteString("3. Compare a intenção do cliente com os exemplos fornecidos, usando palavras-chave, sinônimos e variações de escrita.\n")
	sb.WriteString("   - Se encontrar uma correspondência aproximada, escolha o serviço correspondente.\n")
	sb.WriteString("   - Se não houver correspondência clara, retorne o JSON de erro.\n")
	sb.WriteString("4. Responda **exatamente** com JSON puro, sem crases, markdown ou texto adicional.\n")
	sb.WriteString("5. Formato do JSON:\n")
	sb.WriteString("   - Sucesso: {\"success\": true, \"service_id\": int, \"service_name\": string}\n")
	sb.WriteString("   - Erro: {\"success\": false, \"error\": string}\n")
	sb.WriteString("6. Se a intenção do cliente não corresponder a nenhum serviço, retorne exatamente:\n")
	sb.WriteString("{\"success\": false, \"error\": \"Intenção desconhecida\"}\n")
	sb.WriteString("Serviços disponíveis:\n")

	for _, i := range intents {
		sb.WriteString(fmt.Sprintf("- %d: %s (exemplos: %s)\n", i.ServiceID, i.ServiceName, i.Intent))
	}

	sb.WriteString(fmt.Sprintf("\nIntenção do cliente: %s\n", intent))
	sb.WriteString("Escolha o serviço correto se existir, considerando variações de escrita, palavras-chave e sinônimos.\n")
	sb.WriteString("Responda **somente** com JSON válido, sem nenhum texto extra.\n")

	return sb.String()
}

func callOpenRouter(prompt string) (*Service, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY não definido")
	}

	reqBody := map[string]interface{}{
		"model": "openai/gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": prompt},
		},
		"max_tokens": 100,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(context.Background(),
		"POST",
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error status %d: %s", resp.StatusCode, string(data))
	}

	var respData struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}

	if len(respData.Choices) == 0 {
		return nil, fmt.Errorf("resposta vacía de OpenRouter")
	}

	fmt.Println(respData.Choices[0].Message.Content)

	var respJSON FindServiceResponse
	if err := json.Unmarshal([]byte(respData.Choices[0].Message.Content), &respJSON); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %v\nContenido: %s", err, respData.Choices[0].Message.Content)
	}

	if !respJSON.Success {
		return nil, fmt.Errorf(respJSON.Error)
	}

	return respJSON.Data, nil
}

func findServiceHandler(intents []IntentCSV) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req FindServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(FindServiceResponse{Success: false, Error: "invalid request"})
			return
		}

		prompt := buildPrompt(req.Intent, intents)
		service, err := callOpenRouter(prompt)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(FindServiceResponse{Success: false, Error: err.Error()})
			return
		}

		json.NewEncoder(w).Encode(FindServiceResponse{Success: true, Data: service})
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
