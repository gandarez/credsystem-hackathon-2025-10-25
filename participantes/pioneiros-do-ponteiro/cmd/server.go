package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Intent struct {
	Text        string
	ServiceID   int
	ServiceName string
	Embedding   []float64
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

var (
	intents          []Intent
	openrouterAPIKey = os.Getenv("OPENAI_API_KEY")
	embeddingModel   = "text-embedding-3-small"
)

// --- Utils ---

func cosineSimilarity(a, b []float64) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func getEmbedding(text string) ([]float64, error) {
	body, _ := json.Marshal(map[string]any{
		"model": embeddingModel,
		"input": text,
	})
	fmt.Println(text)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openrouterAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return result.Data[0].Embedding, nil
}

func loadIntents() {
	file, err := os.Open("./assets/intents_pre_loaded.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	_, _ = reader.Read() // skip header

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		serviceID, _ := strconv.Atoi(record[0])
		intents = append(intents, Intent{
			ServiceName: strings.TrimSpace(record[1]),
			Text:        strings.TrimSpace(record[2]),
			ServiceID:   serviceID,
		})
	}

	// Generate embeddings for each intent
	for i := range intents {
		emb, err := getEmbedding(intents[i].Text)
		if err != nil {
			log.Printf("embedding error: %v", err)
			continue
		}
		intents[i].Embedding = emb
	}
}

// --- API Handler ---

func findServiceHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userEmb, err := getEmbedding(body.Message)
	if err != nil {
		http.Error(w, "error generating embedding", http.StatusInternalServerError)
		return
	}

	fmt.Printf("user embedding generated: (%v)", len(userEmb))

	// Find best match
	var best Intent
	bestScore := -1.0
	for _, intent := range intents {
		sim := cosineSimilarity(userEmb, intent.Embedding)
		if sim > bestScore {
			bestScore = sim
			best = intent
		}
	}

	//TODO: Criar struct com modelo do desafio e tratar similaridade minima 10%
	// {
	// "success": "bool",
	// "data": {
	// 	"service_id": "int",
	// 	"service_name": "string",
	// },
	// "error": "string"
	// }

	resp := map[string]any{
		"service_id":   best.ServiceID,
		"similarity":   bestScore,
		"service_name": serviceName(best.ServiceID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"status": "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func serviceName(id int) string {
	names := map[int]string{
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
	return names[id]
}

func main() {

	//TODO: criar um config global e ler .env para configurar porta, chave api, modelo de embedding, etc

	// TODO: rodar com goroutines
	loadIntents()

	http.HandleFunc("/api/find-service", findServiceHandler)
	http.HandleFunc("/api/healthz", healthCheckHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
