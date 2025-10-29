package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type ServiceClassifier struct {
	apiKey     string
	httpClient *http.Client
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
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewServiceClassifier(apiKey string) (*ServiceClassifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key não pode ser vazia")
	}

	return &ServiceClassifier{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}, nil
}

func (sc *ServiceClassifier) Classify(intent string) (int, string, error) {
	systemPrompt := GetSystemPrompt()
	userPrompt := fmt.Sprintf("Intent do cliente: %s", intent)

	reqBody := OpenRouterRequest{
		Model: "openai/gpt-4o-mini", // Modelo eficiente e econômico
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/TaysonMartinss/cacadores-de-corrida")
	req.Header.Set("X-Title", "Cacadores de Corrida - Hackathon")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("erro na API: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp OpenRouterResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return 0, "", fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if apiResp.Error != nil {
		return 0, "", fmt.Errorf("erro da API: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return 0, "", fmt.Errorf("nenhuma resposta retornada pela API")
	}

	// Extrair service_id e service_name da resposta
	content := apiResp.Choices[0].Message.Content
	serviceID, serviceName, err := parseResponse(content)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao parsear resposta: %w", err)
	}

	return serviceID, serviceName, nil
}

type ServiceResponse struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

func parseResponse(content string) (int, string, error) {
	// Limpar conteúdo (remover possíveis markdown ou espaços extras)
	content = strings.TrimSpace(content)
	content = strings.Trim(content, "`")
	if strings.HasPrefix(content, "json") {
		content = strings.TrimPrefix(content, "json")
		content = strings.TrimSpace(content)
	}

	// Tentar parsear JSON direto (formato principal esperado)
	var svcResp ServiceResponse
	if err := json.Unmarshal([]byte(content), &svcResp); err == nil {
		if svcResp.ServiceID >= 1 && svcResp.ServiceID <= 16 {
			return svcResp.ServiceID, svcResp.ServiceName, nil
		}
	}

	// Fallback: Tentar extrair com regex (formato antigo)
	// Padrão 1: ID: 1, Nome: Consulta Limite / Vencimento do cartão / Melhor dia de compra
	re := regexp.MustCompile(`ID:\s*(\d+),?\s*Nome:\s*(.+?)(?:\n|$)`)
	matches := re.FindStringSubmatch(content)

	if len(matches) >= 3 {
		serviceID, err := strconv.Atoi(strings.TrimSpace(matches[1]))
		if err == nil {
			serviceName := strings.TrimSpace(matches[2])
			return serviceID, serviceName, nil
		}
	}

	// Fallback 2: service_id: 1
	re2 := regexp.MustCompile(`service_id[\"']?\s*:\s*(\d+)`)
	matches2 := re2.FindStringSubmatch(content)
	if len(matches2) >= 2 {
		serviceID, _ := strconv.Atoi(matches2[1])
		if serviceID >= 1 && serviceID <= 16 {
			return serviceID, getServiceNameByID(serviceID), nil
		}
	}

	return 0, "", fmt.Errorf("formato de resposta inválido: %s", content)
}

func getServiceNameByID(id int) string {
	services := map[int]string{
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
	
	if name, ok := services[id]; ok {
		return name
	}
	return ""
}
