package openrouter

import (
	"context"
	"net/http"
	"time"
)

const DefaultTimeoutSecs = 30

type Client struct {
	baseURL string
	client  *http.Client
	doFunc  func(c *Client, req *http.Request) (*http.Response, error)
}

func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		client: &http.Client{
			Transport: NewTransport(),
		},
		doFunc: func(c *Client, req *http.Request) (*http.Response, error) {
			req.Header.Set("Accept", "application/json")
			return c.client.Do(req)
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
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
