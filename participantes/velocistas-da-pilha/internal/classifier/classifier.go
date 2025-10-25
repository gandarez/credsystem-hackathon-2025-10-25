package classifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"
	"velocistas_da_pilha/internal/storage"
)

// stopwords comuns que não ajudam na classificação
var stopwords = map[string]bool{
	"o": true, "a": true, "os": true, "as": true,
	"meu": true, "minha": true, "de": true, "do": true, "da": true,
	"para": true, "por": true, "com": true, "e": true,
}

// IntentClassifier mantém intenções conhecidas e cliente HTTP
type IntentClassifier struct {
	knownIntents []storage.IntentEntry
	apiKey       string
	client       *http.Client
}

// NewIntentClassifier cria um classificador
func NewIntentClassifier(intents []storage.IntentEntry, apiKey string) *IntentClassifier {
	return &IntentClassifier{
		knownIntents: intents,
		apiKey:       apiKey,
		client:       &http.Client{},
	}
}

// normalizeString remove acentos, pontuação e espaços extras
func normalizeString(s string) string {
	var b strings.Builder
	s = strings.ToLower(strings.TrimSpace(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Classify usa abordagem híbrida: match exato → fuzzy → LLM
func (ic *IntentClassifier) Classify(intent string) (int, string, error) {
	norm := normalizeString(intent)

	// 1️⃣ Match exato (após normalização)
	for _, known := range ic.knownIntents {
		if normalizeString(known.Intent) == norm {
			return known.ServiceID, known.ServiceName, nil
		}
	}

	// 2️⃣ Fuzzy match
	bestMatch, confidence := ic.fuzzyMatch(norm)
	if confidence > 0.5 { // threshold permissivo
		return bestMatch.ServiceID, bestMatch.ServiceName, nil
	}

	// 3️⃣ LLM fallback
	return ic.classifyWithLLM(intent)
}

// fuzzyMatch aprimorado
func (ic *IntentClassifier) fuzzyMatch(intent string) (*storage.IntentEntry, float64) {
	words := strings.Fields(intent)
	filtered := []string{}
	for _, w := range words {
		if len(w) >= 3 && !stopwords[w] {
			filtered = append(filtered, w)
		}
	}

	var bestMatch *storage.IntentEntry
	bestScore := 0.0

	for i := range ic.knownIntents {
		knownWords := strings.Fields(normalizeString(ic.knownIntents[i].Intent))
		knownFiltered := []string{}
		for _, w := range knownWords {
			if len(w) >= 3 && !stopwords[w] {
				knownFiltered = append(knownFiltered, w)
			}
		}

		matchCount := 0
		for _, w := range filtered {
			for _, kw := range knownFiltered {
				if strings.Contains(kw, w) || strings.Contains(w, kw) {
					matchCount++
					break
				}
			}
		}

		if len(filtered) > 0 {
			score := float64(matchCount) / float64(len(filtered))
			// ponderação baseada no tamanho da frase conhecida
			score *= 0.8 + 0.2*float64(len(knownFiltered))/10.0
			if score > bestScore {
				bestScore = score
				bestMatch = &ic.knownIntents[i]
			}
		}
	}

	return bestMatch, bestScore
}

// ---------------- LLM ----------------

func (ic *IntentClassifier) classifyWithLLM(intent string) (int, string, error) {
	prompt := ic.buildPrompt(intent)
	response, err := ic.callOpenRouter(prompt)
	if err != nil {
		return 0, "", err
	}
	return ic.parseResponse(response)
}

func (ic *IntentClassifier) buildPrompt(intent string) string {
	serviceExamples := make(map[int][]string)
	serviceNames := make(map[int]string)

	for _, entry := range ic.knownIntents {
		serviceExamples[entry.ServiceID] = append(serviceExamples[entry.ServiceID], entry.Intent)
		serviceNames[entry.ServiceID] = entry.ServiceName
	}

	var prompt strings.Builder
	prompt.WriteString("Você é um classificador de intenções para um sistema de URA. Analise a solicitação do cliente e retorne APENAS o ID do serviço correto.\n\n")
	prompt.WriteString("Serviços disponíveis:\n")

	for id := 1; id <= 16; id++ {
		if name, ok := serviceNames[id]; ok {
			prompt.WriteString(fmt.Sprintf("ID %d: %s\n", id, name))
			if examples, ok := serviceExamples[id]; ok && len(examples) > 0 {
				maxEx := 3
				if len(examples) < maxEx {
					maxEx = len(examples)
				}
				prompt.WriteString("  Exemplos: ")
				for i := 0; i < maxEx; i++ {
					if i > 0 {
						prompt.WriteString(", ")
					}
					prompt.WriteString(fmt.Sprintf("\"%s\"", examples[i]))
				}
				prompt.WriteString("\n")
			}
		}
	}

	prompt.WriteString(fmt.Sprintf("\nSolicitação do cliente: \"%s\"\n", intent))
	prompt.WriteString("\nResponda APENAS com o número do ID do serviço mais adequado (1-16). Sem texto adicional.")

	return prompt.String()
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func (ic *IntentClassifier) callOpenRouter(prompt string) (string, error) {
	reqBody := OpenRouterRequest{
		Model: "mistralai/mistral-7b-instruct",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ic.apiKey)

	resp, err := ic.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenRouter API error: %s - %s", resp.Status, string(body))
	}

	var apiResp OpenRouterResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", err
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta da API")
	}

	return apiResp.Choices[0].Message.Content, nil
}

// parseResponse mais confiável
func (ic *IntentClassifier) parseResponse(response string) (int, string, error) {
	response = strings.TrimSpace(response)
	var serviceID int

	// Tentar ler número direto
	_, err := fmt.Sscanf(response, "%d", &serviceID)
	if err != nil {
		// Procurar número de 1 a 16 na string
		for _, num := range []int{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1} {
			if strings.Contains(response, fmt.Sprintf("%d", num)) {
				serviceID = num
				break
			}
		}
	}

	if serviceID < 1 || serviceID > 16 {
		return 15, "Atendimento humano", nil
	}

	for _, entry := range ic.knownIntents {
		if entry.ServiceID == serviceID {
			return serviceID, entry.ServiceName, nil
		}
	}

	return 0, "", fmt.Errorf("serviço ID %d não encontrado", serviceID)
}
