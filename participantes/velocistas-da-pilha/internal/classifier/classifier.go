package classifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"velocistas_da_pilha/internal/storage"
)

type IntentClassifier struct {
	knownIntents []storage.IntentEntry
	apiKey       string
	client       *http.Client
}

func NewIntentClassifier(intents []storage.IntentEntry, apiKey string) *IntentClassifier {
	return &IntentClassifier{
		knownIntents: intents,
		apiKey:       apiKey,
		client:       &http.Client{},
	}
}

// Classify usa uma abordagem híbrida: matching exato primeiro, depois LLM
func (ic *IntentClassifier) Classify(intent string) (int, string, error) {
	// 1. Tentar match exato ou muito similar (case-insensitive)
	normalized := strings.ToLower(strings.TrimSpace(intent))
	
	for _, known := range ic.knownIntents {
		if strings.ToLower(strings.TrimSpace(known.Intent)) == normalized {
			return known.ServiceID, known.ServiceName, nil
		}
	}

	// 2. Tentar match parcial forte (contém palavras-chave principais)
	bestMatch, confidence := ic.fuzzyMatch(normalized)
	if confidence > 0.8 {
		return bestMatch.ServiceID, bestMatch.ServiceName, nil
	}

	// 3. Usar LLM para classificar
	return ic.classifyWithLLM(intent)
}

// fuzzyMatch encontra o melhor match baseado em palavras-chave
func (ic *IntentClassifier) fuzzyMatch(intent string) (*storage.IntentEntry, float64) {
	words := strings.Fields(intent)
	var bestMatch *storage.IntentEntry
	bestScore := 0.0

	for i := range ic.knownIntents {
		knownWords := strings.Fields(strings.ToLower(ic.knownIntents[i].Intent))
		matchCount := 0
		
		for _, word := range words {
			if len(word) < 3 {
				continue // ignorar palavras muito curtas
			}
			for _, knownWord := range knownWords {
				if strings.Contains(knownWord, word) || strings.Contains(word, knownWord) {
					matchCount++
					break
				}
			}
		}

		if len(words) > 0 {
			score := float64(matchCount) / float64(len(words))
			if score > bestScore {
				bestScore = score
				bestMatch = &ic.knownIntents[i]
			}
		}
	}

	return bestMatch, bestScore
}

// classifyWithLLM usa o modelo da OpenRouter para classificar
func (ic *IntentClassifier) classifyWithLLM(intent string) (int, string, error) {
	// Criar prompt com exemplos dos serviços
	prompt := ic.buildPrompt(intent)

	// Chamar API
	response, err := ic.callOpenRouter(prompt)
	if err != nil {
		return 0, "", err
	}

	// Parsear resposta
	return ic.parseResponse(response)
}

func (ic *IntentClassifier) buildPrompt(intent string) string {
	// Agrupar intenções por serviço para criar exemplos
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
				// Pegar até 3 exemplos
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
	Model    string          `json:"model"`
	Messages []Message       `json:"messages"`
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

func (ic *IntentClassifier) parseResponse(response string) (int, string, error) {
	// Limpar resposta
	response = strings.TrimSpace(response)
	
	// Tentar extrair número
	var serviceID int
	_, err := fmt.Sscanf(response, "%d", &serviceID)
	if err != nil {
		// Tentar encontrar número na string
		for _, char := range response {
			if char >= '0' && char <= '9' {
				serviceID = int(char - '0')
				// Se for dois dígitos
				if serviceID == 1 && strings.Contains(response, "16") {
					serviceID = 16
				} else if serviceID == 1 && strings.Contains(response, "15") {
					serviceID = 15
				} else if serviceID == 1 && strings.Contains(response, "14") {
					serviceID = 14
				} else if serviceID == 1 && strings.Contains(response, "13") {
					serviceID = 13
				} else if serviceID == 1 && strings.Contains(response, "12") {
					serviceID = 12
				} else if serviceID == 1 && strings.Contains(response, "11") {
					serviceID = 11
				} else if serviceID == 1 && strings.Contains(response, "10") {
					serviceID = 10
				}
				break
			}
		}
	}

	// Validar ID
	if serviceID < 1 || serviceID > 16 {
		return 15, "Atendimento humano", nil // Fallback para atendimento humano
	}

	// Encontrar nome do serviço
	for _, entry := range ic.knownIntents {
		if entry.ServiceID == serviceID {
			return serviceID, entry.ServiceName, nil
		}
	}

	return 0, "", fmt.Errorf("serviço ID %d não encontrado", serviceID)
}