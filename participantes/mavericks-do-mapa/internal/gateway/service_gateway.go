package gateway

import (
	"context"
	"errors"
	"strings"

	"mavericksdomapa/internal/domain"
)

var ErrServiceNotFound = errors.New("service not found")

type ServiceGateway interface {
	FindService(ctx context.Context, intent string) (*domain.Service, error)
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
