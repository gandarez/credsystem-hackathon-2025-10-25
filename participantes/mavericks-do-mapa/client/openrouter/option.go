package openrouter

import (
	"net/http"
	"strings"
	"time"
)

type Option func(*Client)

func WithAuth(token string) Option {
	return func(c *Client) {
		next := c.doFunc
		c.doFunc = func(c *Client, req *http.Request) (*http.Response, error) {
			req.Header.Set("Authorization", "Bearer "+token)
			return next(c, req)
		}
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.client = httpClient
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.client.Timeout = timeout
	}
}

func WithModel(model string) Option {
	return func(c *Client) {
		if strings.TrimSpace(model) != "" {
			c.model = model
		}
	}
}

func WithSystemPrompt(prompt string) Option {
	return func(c *Client) {
		if strings.TrimSpace(prompt) != "" {
			c.systemPrompt = prompt
		}
	}
}

func WithAttribution(referer, title string) Option {
	return func(c *Client) {
		if trimmed := strings.TrimSpace(referer); trimmed != "" {
			c.referer = trimmed
		}
		if trimmed := strings.TrimSpace(title); trimmed != "" {
			c.title = trimmed
		}
	}
}
