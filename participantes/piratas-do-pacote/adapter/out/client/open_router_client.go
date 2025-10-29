package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"
)

type OpenRouterClient struct {
	hc     *http.Client
	apiKey string
	base   string
}

func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   defaultDialTimeout,
			KeepAlive: defaultKeepAlive,
		}).DialContext,
		TLSHandshakeTimeout:   1 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          defaultMaxIdleConns,
		MaxIdleConnsPerHost:   defaultMaxIdleConnsPerHost,
		IdleConnTimeout:       defaultIdleConnTimeout,
		ExpectContinueTimeout: 0,
		DisableKeepAlives:     false,
	}
	return &OpenRouterClient{
		hc: &http.Client{
			Transport: tr,
			Timeout:   defaultTimeout,
		},
		apiKey: apiKey,
		base:   openRouterUrl,
	}
}

var ErrNon2xx = errors.New("non-2xx response")

func (c *OpenRouterClient) PostJSON(ctx context.Context, path string, in any, out any) (int, error) {
	u := c.base + path

	b, err := json.Marshal(in)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(http.MaxBytesReader(nil, resp.Body, 1024))
		return resp.StatusCode, errors.Join(ErrNon2xx, errors.New(buf.String()))
	}

	if out != nil {
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(out); err != nil {
			return resp.StatusCode, err
		}
	}
	return resp.StatusCode, nil
}
