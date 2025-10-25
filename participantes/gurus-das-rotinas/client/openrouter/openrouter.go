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
		ForceAttemptHTTP2:     true,
		MaxConnsPerHost:       50,               // Increased for better concurrency
		MaxIdleConns:          50,               // Increased for better connection reuse
		MaxIdleConnsPerHost:   50,               // Increased for better connection reuse
		IdleConnTimeout:       90 * time.Second, // Keep connections alive longer
		DisableKeepAlives:     false,            // Enable keep-alive
		DisableCompression:    false,            // Enable compression
		Proxy:                 nil,
		TLSHandshakeTimeout:   10 * time.Second, // Reduced from 30s
		ResponseHeaderTimeout: 15 * time.Second, // Add response timeout
		ExpectContinueTimeout: 1 * time.Second,  // Add expect continue timeout
	}
}
