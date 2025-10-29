package gateway

import (
	"context"
	"errors"
	"strings"

	"mavericksdomapa/client/openrouter"
	"mavericksdomapa/internal/domain"
	prompttemplate "mavericksdomapa/prompttemplate"
)

var ErrServiceNotFound = errors.New("service not found")

type ServiceGateway interface {
	FindService(ctx context.Context, intent string) (*domain.Service, error)
}

type openRouterClient interface {
	ChatCompletion(ctx context.Context, intent string) (*openrouter.DataResponse, error)
}

type StaticServiceGateway struct{}

func NewStaticServiceGateway() *StaticServiceGateway {
	return &StaticServiceGateway{}
}

func (g *StaticServiceGateway) FindService(_ context.Context, intent string) (*domain.Service, error) {
	intent = strings.ToLower(strings.TrimSpace(intent))

	switch {
	case strings.Contains(intent, "cart") || strings.Contains(intent, "credit"):
		return &domain.Service{ID: 1, Name: "Cartoes de Credito"}, nil
	case strings.Contains(intent, "emprest") || strings.Contains(intent, "loan"):
		return &domain.Service{ID: 2, Name: "Emprestimos Pessoais"}, nil
	case strings.Contains(intent, "invest") || strings.Contains(intent, "application"):
		return &domain.Service{ID: 3, Name: "Investimentos"}, nil
	case strings.Contains(intent, "seguro") || strings.Contains(intent, "insurance"):
		return &domain.Service{ID: 4, Name: "Seguros"}, nil
	default:
		return nil, ErrServiceNotFound
	}
}

type OpenRouterServiceGateway struct {
	client openRouterClient
}

func NewOpenRouterServiceGateway(client openRouterClient) *OpenRouterServiceGateway {
	return &OpenRouterServiceGateway{
		client: client,
	}
}

func (g *OpenRouterServiceGateway) FindService(ctx context.Context, intent string) (*domain.Service, error) {
	intent = strings.TrimSpace(intent)
	if intent == "" {
		return nil, ErrServiceNotFound
	}

	prompt, err := prompttemplate.BuildPrompt(intent)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.ChatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}

	if resp == nil || !resp.Success {
		return nil, ErrServiceNotFound
	}

	if resp.Data == nil || resp.Data.ServiceID == 0 || strings.TrimSpace(resp.Data.ServiceName) == "" {
		return nil, ErrServiceNotFound
	}

	return &domain.Service{
		ID:   resp.Data.ServiceID,
		Name: resp.Data.ServiceName,
	}, nil
}
