package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type (
	OpenRouterRequest struct {
		Model    string        `json:"model"`
		Messages []ChatMessage `json:"messages"`
		Stream   bool          `json:"stream,omitempty"`
	}

	ChatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	OpenRouterResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	ServiceData struct {
		ServiceID   int    `json:"service_id"`
		ServiceName string `json:"service_name"`
	}

	DataResponse struct {
		Success bool         `json:"success"`
		Data    *ServiceData `json:"data"`
		Error   string       `json:"error"`
	}
)

func (c *Client) ChatCompletion(ctx context.Context, intent string) (*DataResponse, error) {
	intent = strings.TrimSpace(intent)
	if intent == "" {
		return nil, errors.New("intent cannot be empty")
	}

	if strings.TrimSpace(c.model) == "" {
		return nil, errors.New("model must be configured")
	}

	if strings.TrimSpace(c.systemPrompt) == "" {
		return nil, errors.New("system prompt must be configured")
	}

	url := strings.TrimRight(c.baseURL, "/") + "/chat/completions"

	requestBody := OpenRouterRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: c.systemPrompt,
			},
			{
				Role:    "user",
				Content: intent,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v. body: %s", err, string(body))
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := sanitizeContent(openRouterResp.Choices[0].Message.Content)

	var dataRes DataResponse
	if err := json.Unmarshal([]byte(content), &dataRes); err != nil {
		return nil, fmt.Errorf("error unmarshaling data response: %v. content: %s", err, openRouterResp.Choices[0].Message.Content)
	}

	if dataRes.Success {
		if dataRes.Data == nil {
			return nil, fmt.Errorf("success response missing data: %+v", dataRes)
		}

		dataRes.Data.ServiceName = strings.TrimSpace(dataRes.Data.ServiceName)
		if dataRes.Data.ServiceID == 0 || dataRes.Data.ServiceName == "" {
			return nil, fmt.Errorf("incomplete data response: %+v", dataRes)
		}
	} else {
		dataRes.Error = strings.TrimSpace(dataRes.Error)
	}

	return &dataRes, nil
}

func sanitizeContent(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)

		// Remove optional language identifier (e.g., ```json)
		if idx := strings.Index(content, "\n"); idx != -1 {
			firstLine := content[:idx]
			if !strings.HasPrefix(firstLine, "{") && !strings.HasPrefix(firstLine, "[") {
				content = strings.TrimSpace(content[idx+1:])
			}
		}
	}

	return content
}
