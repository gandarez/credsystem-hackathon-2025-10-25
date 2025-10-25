package openrouter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_Defaults(t *testing.T) {
	baseURL := "https://api.openrouter.ai/v1"
	c := NewClient(baseURL)

	if c == nil {
		t.Fatalf("NewClient returned nil")
	}

	if c.baseURL != baseURL {
		t.Fatalf("baseURL mismatch: got %q, want %q", c.baseURL, baseURL)
	}

	if c.client == nil {
		t.Fatalf("http client should not be nil")
	}

	// Transport should be *http.Transport configured by NewTransport
	tr, ok := c.client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("client transport should be *http.Transport, got %T", c.client.Transport)
	}

	if !tr.ForceAttemptHTTP2 {
		t.Errorf("ForceAttemptHTTP2 = false, want true")
	}

	if tr.MaxConnsPerHost != 10 {
		t.Errorf("MaxConnsPerHost = %d, want 10", tr.MaxConnsPerHost)
	}

	if tr.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns = %d, want 10", tr.MaxIdleConns)
	}

	if tr.MaxIdleConnsPerHost != 10 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 10", tr.MaxIdleConnsPerHost)
	}

	// Verify that Do() sets Accept header and performs the request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept header = %q, want %q", got, "application/json")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest error: %v", err)
	}

	if _, err := c.Do(context.Background(), req); err != nil {
		t.Fatalf("Do() error: %v", err)
	}
}

func TestNewClient_WithAuthOption(t *testing.T) {
	const token = "test-token"

	// Server validates Authorization and Accept headers
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+token {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer "+token)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept header = %q, want %q", got, "application/json")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient("irrelevant", WithAuth(token))

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest error: %v", err)
	}

	if _, err := c.Do(context.Background(), req); err != nil {
		t.Fatalf("Do() error: %v", err)
	}
}
