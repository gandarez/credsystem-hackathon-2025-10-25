package service

import (
	"github.com/bandidos_do_byte/api/internal/domain"
)

type ServiceFinder interface {
	FindService(intent string) (*domain.ServiceData, error)
	HealthCheck() string
}

type serviceFinder struct {
	// Aqui você pode adicionar dependências como clients HTTP, repositórios, etc.
}

func NewServiceFinder() ServiceFinder {
	return &serviceFinder{}
}

// FindService implementa a lógica para encontrar o serviço apropriado
func (s *serviceFinder) FindService(intent string) (*domain.ServiceData, error) {
	// TODO: Implementar lógica com IA para determinar o serviço
	// Por enquanto retorna um serviço dummy
	return &domain.ServiceData{
		ServiceID:   1,
		ServiceName: "Consulta Limite / Vencimento do cartão / Melhor dia de compra",
	}, nil
}

// HealthCheck verifica a saúde do serviço
func (s *serviceFinder) HealthCheck() string {
	return "ok"
}
