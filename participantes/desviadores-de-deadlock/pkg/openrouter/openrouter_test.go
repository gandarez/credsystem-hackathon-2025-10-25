package openrouter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		opts     []Option
		validate func(t *testing.T, c *Client)
	}{
		{
			name:    "client without options",
			baseURL: "https://api.openrouter.ai",
			opts:    nil,
			validate: func(t *testing.T, c *Client) {
				if c.baseURL != "https://api.openrouter.ai" {
					t.Errorf("expected baseURL %s, got %s", "https://api.openrouter.ai", c.baseURL)
				}
				if c.client == nil {
					t.Error("expected client to be initialized")
				}
				if c.doFunc == nil {
					t.Error("expected doFunc to be initialized")
				}
			},
		},
		{
			name:    "client with auth option",
			baseURL: "https://api.openrouter.ai",
			opts:    []Option{WithAuth("test-token")},
			validate: func(t *testing.T, c *Client) {
				if c.baseURL != "https://api.openrouter.ai" {
					t.Errorf("expected baseURL %s, got %s", "https://api.openrouter.ai", c.baseURL)
				}
				if c.client == nil {
					t.Error("expected client to be initialized")
				}
				if c.doFunc == nil {
					t.Error("expected doFunc to be initialized")
				}
			},
		},
		{
			name:    "client with empty baseURL",
			baseURL: "",
			opts:    nil,
			validate: func(t *testing.T, c *Client) {
				if c.baseURL != "" {
					t.Errorf("expected baseURL to be empty, got %s", c.baseURL)
				}
				if c.client == nil {
					t.Error("expected client to be initialized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.opts...)
			tt.validate(t, client)
		})
	}
}

func TestNewClient_HTTPClient(t *testing.T) {
	client := NewClient("https://api.openrouter.ai")

	// Verificar se o client HTTP tem as configurações corretas
	if client.client.Timeout != 0 {
		t.Errorf("expected timeout to be 0, got %v", client.client.Timeout)
	}

	// Verificar se o transport foi configurado
	transport, ok := client.client.Transport.(*http.Transport)
	if !ok {
		t.Error("expected transport to be *http.Transport")
	}

	// Verificar configurações do transport
	if !transport.ForceAttemptHTTP2 {
		t.Error("expected ForceAttemptHTTP2 to be true")
	}
	if transport.MaxConnsPerHost != 10 {
		t.Errorf("expected MaxConnsPerHost to be 10, got %d", transport.MaxConnsPerHost)
	}
	if transport.MaxIdleConns != 10 {
		t.Errorf("expected MaxIdleConns to be 10, got %d", transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("expected MaxIdleConnsPerHost to be 10, got %d", transport.MaxIdleConnsPerHost)
	}
	if transport.TLSHandshakeTimeout != DefaultTimeoutSecs*time.Second {
		t.Errorf("expected TLSHandshakeTimeout to be %v, got %v", DefaultTimeoutSecs*time.Second, transport.TLSHandshakeTimeout)
	}
}

func TestNewClient_DoFunc(t *testing.T) {
	// Criar um servidor de teste
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar se o header Accept foi definido
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept header to be 'application/json', got %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	// Criar uma requisição de teste
	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Executar a requisição
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestNewClient_WithAuth(t *testing.T) {
	// Criar um servidor de teste
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar se o header Authorization foi definido
		expectedAuth := "Bearer test-token"
		if r.Header.Get("Authorization") != expectedAuth {
			t.Errorf("expected Authorization header to be '%s', got %s", expectedAuth, r.Header.Get("Authorization"))
		}
		// Verificar se o header Accept ainda está presente
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept header to be 'application/json', got %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := NewClient(server.URL, WithAuth("test-token"))

	// Criar uma requisição de teste
	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Executar a requisição
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestNewClient_MultipleOptions(t *testing.T) {
	// Criar um servidor de teste
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar se ambos os headers foram definidos
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header to be 'Bearer test-token', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept header to be 'application/json', got %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	// Aplicar a mesma opção duas vezes para testar múltiplas opções
	client := NewClient(server.URL, WithAuth("test-token"), WithAuth("another-token"))

	// Criar uma requisição de teste
	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Executar a requisição
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
