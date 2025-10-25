package openrouter

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultTimeoutSecs = 30
	DefaultModel       = "openrouter/auto"
)

type Client struct {
	baseURL      string
	client       *http.Client
	doFunc       func(c *Client, req *http.Request) (*http.Response, error)
	model        string
	systemPrompt string
	referer      string
	title        string
}

func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		client: &http.Client{
			Transport: NewTransport(),
		},
		doFunc: func(c *Client, req *http.Request) (*http.Response, error) {
			req.Header.Set("Accept", "application/json")
			if strings.TrimSpace(c.referer) != "" {
				req.Header.Set("HTTP-Referer", c.referer)
			}
			if strings.TrimSpace(c.title) != "" {
				req.Header.Set("X-Title", c.title)
			}
			return c.client.Do(req)
		},
		model: DefaultModel,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	resp, err := c.doFunc(c, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// NewTransport initializes a new http.Transport.
func NewTransport() *http.Transport {
	return &http.Transport{
		ForceAttemptHTTP2:   true,
		MaxConnsPerHost:     10,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		Proxy:               nil,
		TLSHandshakeTimeout: DefaultTimeoutSecs * time.Second,
	}
}
