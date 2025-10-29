package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
var openRouterRespPool = sync.Pool{New: func() any { return new(OpenRouterResponse) }}

type (
	OpenRouterRequest struct {
		Model    string          `json:"model"`
		Messages []PromptMessage `json:"messages"`
	}

	OpenRouterResponse struct {
		Choices []Choice `json:"choices"`
	}

	Choice struct {
		Message Message
	}

	Message struct {
		Content string `json:"content"`
	}

	DataResponse struct {
		ServiceID   uint8  `json:"service_id"`
		ServiceName string `json:"service_name"`
		Result      string `json:"result"`
	}

	ContextPrompt struct {
		Prompt   string          `json:"prompt"`
		Model    string          `json:"model"`
		Messages []PromptMessage `json:"messages"`
	}

	PromptMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
)

func (c *Client) ChatCompletion(ctx context.Context, request *OpenRouterRequest) (*DataResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	url := c.baseURL + "/chat/completions"

	// Encode request using a pooled buffer to reduce allocations
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	enc := json.NewEncoder(buf)
	if err := enc.Encode(request); err != nil {
		buf.Reset()
		bufPool.Put(buf)
		return nil, fmt.Errorf("error encoding request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		buf.Reset()
		bufPool.Put(buf)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(ctx, req)
	// We can safely return the buffer after request is created; to be conservative, return it after Do completes
	buf.Reset()
	bufPool.Put(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	openRouterResp := openRouterRespPool.Get().(*OpenRouterResponse)
	*openRouterResp = OpenRouterResponse{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(openRouterResp); err != nil {
		openRouterRespPool.Put(openRouterResp)
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if len(openRouterResp.Choices) == 0 {
		openRouterRespPool.Put(openRouterResp)
		return nil, fmt.Errorf("no choices in response")
	}

	reasoning, response, err := filterReasoning(openRouterResp.Choices)
	// Return the pooled object after extracting needed data
	openRouterRespPool.Put(openRouterResp)
	if err != nil {
		return nil, fmt.Errorf("error filtering reasoning: %v", err)
	}

	var dataRes DataResponse
	if err := json.NewDecoder(strings.NewReader(response)).Decode(&dataRes); err != nil {
		return nil, fmt.Errorf("error unmarshaling data response: %v", err)
	}

	// Optionally log reasoning for debugging
	if reasoning != "" {
		fmt.Printf("Reasoning: %s\n", reasoning)
	}

	return &dataRes, nil
}

func filterReasoning(choices []Choice) (reasoning string, response string, err error) {
	if len(choices) == 0 {
		return "", "", fmt.Errorf("no choices available")
	}

	content := choices[0].Message.Content

	// Find reasoning block
	reasoningStart := strings.Index(content, "<reasoning>")
	reasoningEnd := strings.Index(content, "</reasoning>")

	if reasoningStart != -1 && reasoningEnd != -1 {
		reasoning = strings.TrimSpace(content[reasoningStart+len("<reasoning>") : reasoningEnd])

		// Get content after reasoning block
		afterReasoning := content[reasoningEnd+len("</reasoning>"):]
		response = strings.TrimSpace(afterReasoning)
	} else {
		// No reasoning block found, treat everything as response
		response = strings.TrimSpace(content)
	}

	return reasoning, response, nil
}
