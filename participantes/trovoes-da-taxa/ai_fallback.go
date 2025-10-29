package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ValidationError representa um erro de validação de entrada (não erro técnico)
// Este tipo de erro indica que o input é inválido e não deve fazer fallback
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError cria um novo erro de validação
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// IsValidationError verifica se um erro é de validação
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// AIClient representa um cliente para a API da OpenRouter
type AIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	model      string
	intents    []Intent // Cache dos intents para construir prompts melhores
}

// NewAIClient cria um novo cliente AI
func NewAIClient() *AIClient {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("WARNING: OPENROUTER_API_KEY not set, AI fallback will not work")
	}

	return &AIClient{
		baseURL: "https://openrouter.ai/api/v1",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 45 * time.Second, // Aumentado para modelos mais potentes
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		// Usando o modelo mais potente disponível para máxima precisão
		// Alternativas: "anthropic/claude-3.5-sonnet", "openai/gpt-4-turbo"
		model: "openai/gpt-4o", // Modelo mais avançado da OpenAI
	}
}

// SetIntents define os intents disponíveis para melhorar o prompt
func (c *AIClient) SetIntents(intents []Intent) {
	c.intents = intents
}

type openRouterRequest struct {
	Model     string    `json:"model"`
	Messages  []message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// aiResponse representa a resposta estruturada da IA
type aiResponse struct {
	Success     bool   `json:"success"`
	ServiceID   int    `json:"service_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	Error       string `json:"error,omitempty"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// buildPrompt constrói o prompt para a IA com as instruções precisas
// Usa os intents pré-carregados como exemplos para melhorar a classificação
// Aplica técnicas avançadas de prompt engineering:
// - Chain-of-Thought reasoning
// - Few-shot learning com exemplos positivos e negativos
// - Instruções estruturadas e claras
func (c *AIClient) buildPrompt(intentText string, services map[int]string) string {
	var sb strings.Builder

	// === ROLE & CONTEXT ===
	sb.WriteString("# FUNÇÃO\n")
	sb.WriteString("Você é um especialista em classificação de intenções para o sistema Credsystem.\n")
	sb.WriteString("Sua tarefa é identificar qual serviço bancário/financeiro o cliente deseja acessar com base na descrição de sua intenção.\n\n")

	// === INSTRUÇÕES ESTRUTURADAS ===
	sb.WriteString("# INSTRUÇÕES CRÍTICAS\n\n")
	sb.WriteString("## 1. VALIDAÇÃO DE ENTRADA\n")
	sb.WriteString("ANTES de classificar, verifique se a intenção:\n")
	sb.WriteString("- É uma frase coerente em português relacionada a serviços bancários/financeiros\n")
	sb.WriteString("- Contém palavras ou conceitos relacionados aos serviços disponíveis\n")
	sb.WriteString("- NÃO é apenas ruído, caracteres aleatórios, emojis, ou texto sem sentido\n")
	sb.WriteString("- NÃO é uma pergunta sem relação com serviços bancários\n\n")

	sb.WriteString("## 2. CASOS QUE DEVEM RETORNAR ERRO\n")
	sb.WriteString("Retorne {\"success\": false, \"error\": \"descrição\"} se a intenção:\n")
	sb.WriteString("- É texto aleatório/sem sentido: \"asdfjkl\", \"xpto123\", \"zzzzz\"\n")
	sb.WriteString("- São apenas emojis ou símbolos: \"😀😀😀\", \"!!!!\", \"###\"\n")
	sb.WriteString("- É muito curta e vaga: \"oi\", \"olá\", \"help\", \"?\"\n")
	sb.WriteString("- Não tem NENHUMA relação com os serviços listados\n")
	sb.WriteString("- É uma pergunta filosófica, piada, ou completamente fora do contexto bancário\n\n")

	sb.WriteString("## 3. PROCESSO DE CLASSIFICAÇÃO (Chain-of-Thought)\n")
	sb.WriteString("Para cada entrada VÁLIDA:\n")
	sb.WriteString("a) Identifique palavras-chave e conceitos principais\n")
	sb.WriteString("b) Compare com os exemplos de cada serviço\n")
	sb.WriteString("c) Considere sinônimos, variações de escrita e contexto\n")
	sb.WriteString("d) Escolha o serviço com maior correspondência semântica\n")
	sb.WriteString("e) Se não houver correspondência clara (< 60% de certeza), retorne erro\n\n")

	sb.WriteString("## 4. FORMATO DE RESPOSTA\n")
	sb.WriteString("Responda SOMENTE com JSON puro, sem markdown, crases ou explicações:\n")
	sb.WriteString("- Sucesso: {\"success\": true, \"service_id\": <int>, \"service_name\": \"<string>\"}\n")
	sb.WriteString("- Falha: {\"success\": false, \"error\": \"<razão específica>\"}\n\n")

	// === EXEMPLOS FEW-SHOT ===
	sb.WriteString("# EXEMPLOS DE CLASSIFICAÇÃO\n\n")

	// Exemplos NEGATIVOS (casos que devem retornar erro)
	sb.WriteString("## Exemplos de REJEIÇÃO (inputs inválidos):\n")
	sb.WriteString("Input: \"asdfghjkl\"\n")
	sb.WriteString("Output: {\"success\": false, \"error\": \"Entrada inválida: texto sem sentido ou caracteres aleatórios\"}\n\n")

	sb.WriteString("Input: \"😀😀😀😀\"\n")
	sb.WriteString("Output: {\"success\": false, \"error\": \"Entrada inválida: apenas emojis ou símbolos\"}\n\n")

	sb.WriteString("Input: \"o que é a vida?\"\n")
	sb.WriteString("Output: {\"success\": false, \"error\": \"Intenção não relacionada aos serviços disponíveis\"}\n\n")

	sb.WriteString("Input: \"oi\"\n")
	sb.WriteString("Output: {\"success\": false, \"error\": \"Entrada muito vaga ou genérica, sem contexto suficiente\"}\n\n")

	// Exemplos POSITIVOS com raciocínio
	sb.WriteString("## Exemplos de SUCESSO (inputs válidos):\n")
	sb.WriteString("Input: \"quero saber meu limite\"\n")
	sb.WriteString("Raciocínio: \"limite\" relaciona-se com crédito do cartão\n")
	sb.WriteString("Output: {\"success\": true, \"service_id\": 1, \"service_name\": \"Consulta Limite / Vencimento do cartão / Melhor dia de compra\"}\n\n")

	sb.WriteString("Input: \"boleto do acordo que fiz\"\n")
	sb.WriteString("Raciocínio: \"boleto\" + \"acordo\" indica renegociação de dívida\n")
	sb.WriteString("Output: {\"success\": true, \"service_id\": 2, \"service_name\": \"Segunda via de boleto de acordo\"}\n\n")

	// === DATASET COMPLETO ===
	if len(c.intents) > 0 {
		sb.WriteString("# BASE DE CONHECIMENTO - SERVIÇOS E EXEMPLOS\n\n")

		// Agrupar intents por serviço
		serviceExamples := make(map[int][]string)
		for _, intent := range c.intents {
			serviceExamples[intent.ServiceID] = append(
				serviceExamples[intent.ServiceID],
				intent.IntentText,
			)
		}

		// Mostrar TODOS os exemplos (não limitar a 3) para máxima precisão
		for id, name := range services {
			examples := serviceExamples[id]
			if len(examples) > 0 {
				sb.WriteString(fmt.Sprintf("## Serviço %d: %s\n", id, name))
				sb.WriteString("Exemplos de intenções válidas:\n")

				// Mostrar até 10 exemplos representativos para não sobrecarregar
				maxExamples := 10
				if len(examples) < maxExamples {
					maxExamples = len(examples)
				}

				for i := 0; i < maxExamples; i++ {
					sb.WriteString(fmt.Sprintf("- \"%s\"\n", examples[i]))
				}

				if len(examples) > maxExamples {
					sb.WriteString(fmt.Sprintf("... e mais %d variações similares\n", len(examples)-maxExamples))
				}
				sb.WriteString("\n")
			}
		}
	} else {
		// Fallback se não tivermos intents
		sb.WriteString("# SERVIÇOS DISPONÍVEIS\n\n")
		for id, name := range services {
			sb.WriteString(fmt.Sprintf("- %d: %s\n", id, name))
		}
	}

	// === TAREFA ATUAL ===
	sb.WriteString("\n# TAREFA\n")
	sb.WriteString("Analise a seguinte intenção do cliente e classifique:\n\n")
	sb.WriteString(fmt.Sprintf("Input: \"%s\"\n\n", intentText))

	sb.WriteString("LEMBRE-SE:\n")
	sb.WriteString("1. Valide PRIMEIRO se a entrada é coerente e relacionada a serviços bancários\n")
	sb.WriteString("2. Se for inválida ou sem sentido, retorne erro com descrição específica\n")
	sb.WriteString("3. Se for válida, identifique o serviço correspondente\n")
	sb.WriteString("4. Use os exemplos acima como referência\n")
	sb.WriteString("5. Responda APENAS com JSON, sem explicações adicionais\n\n")

	sb.WriteString("Output:")

	return sb.String()
}

// ClassifyWithAI usa a API da OpenRouter para classificar a intenção
func (c *AIClient) ClassifyWithAI(ctx context.Context, intentText string, services map[int]string) (*ClassificationResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY not configured")
	}

	prompt := c.buildPrompt(intentText, services)

	reqBody := openRouterRequest{
		Model: c.model,
		Messages: []message{
			{
				Role:    "system",
				Content: "Você é um especialista em classificação de intenções bancárias. Responda sempre em JSON puro, sem markdown ou explicações. Valide rigorosamente se a entrada é coerente antes de classificar.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 250, // Aumentado para suportar respostas com raciocínio
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openRouterResp openRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("unmarshal openrouter response: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in AI response")
	}

	content := strings.TrimSpace(openRouterResp.Choices[0].Message.Content)

	// Limpar possíveis wrappers de markdown
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Remover possível texto "Output:" ou similar no início
	if strings.HasPrefix(content, "Output:") {
		content = strings.TrimPrefix(content, "Output:")
		content = strings.TrimSpace(content)
	}

	// Tentar extrair apenas o JSON se houver texto extra
	if jsonStart := strings.Index(content, "{"); jsonStart >= 0 {
		if jsonEnd := strings.LastIndex(content, "}"); jsonEnd > jsonStart {
			content = content[jsonStart : jsonEnd+1]
		}
	}

	// Parse a resposta JSON estruturada
	var aiResp aiResponse
	if err := json.Unmarshal([]byte(content), &aiResp); err != nil {
		fmt.Printf("AI JSON Parse Error: %v\nRaw Content: %s\n", err, content)
		return nil, fmt.Errorf("failed to parse AI JSON response: %w", err)
	}

	// Verificar se a IA conseguiu classificar
	if !aiResp.Success {
		// Log detalhado para análise
		fmt.Printf("AI Rejected Intent: %q - Reason: %s\n", intentText, aiResp.Error)

		// Retornar ValidationError para indicar que é um problema com o input,
		// não um erro técnico da IA. Isso impede o fallback para NLP local.
		return nil, NewValidationError(aiResp.Error)
	}

	// Verificar se o ID é válido
	serviceName, exists := services[aiResp.ServiceID]
	if !exists {
		return nil, fmt.Errorf("AI returned invalid service ID: %d", aiResp.ServiceID)
	}

	return &ClassificationResult{
		ServiceID:   aiResp.ServiceID,
		ServiceName: serviceName,
		Confidence:  1.0, // AI não fornece confiança, usamos 1.0
	}, nil
}
