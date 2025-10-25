package out

import (
	"flag"
	"fmt"

	"github.com/piratas-do-pacote/adapter/out/client"
)

type AiInferer struct {
	client *client.OpenRouterClient
	kb     *KB
}

func NewAiInferer(client *client.OpenRouterClient) *AiInferer {
	var kbPath string
	flag.StringVar(&kbPath, "kb", "kb.json", "caminho para o arquivo JSON de base de conhecimento")
	kb, err := loadKB(kbPath)
	if err != nil {
		fmt.Println(err)
	}
	return &AiInferer{
		client: client,
		kb:     kb,
	}
}

// ====== Tipos que espelham a resposta do OpenRouter ======
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	FinishReason       string      `json:"finish_reason"`
	NativeFinishReason string      `json:"native_finish_reason"`
	Message            ChatMessage `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Model   string   `json:"model"`
}

// ====== Tipos do request ======
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	// Controles de geração (todos opcionais)
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	// (opcionais adicionais, caso queira)
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`
	// Seed é suportado em alguns provedores/modelos; use só se seu gateway/documentação confirmar
	Seed *int `json:"seed,omitempty"`
}
