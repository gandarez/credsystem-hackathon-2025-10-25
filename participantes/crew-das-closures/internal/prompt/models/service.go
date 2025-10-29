package models

import (
	"fmt"
)

// ServiceDefinition represents a service with its metadata
type ServiceDefinition struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
}

// ServiceResponse represents the response structure for service classification
type ServiceResponse struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// ClassificationRequest represents a request for intent classification
type ClassificationRequest struct {
	UserIntent string `json:"user_intent"`
	Context    string `json:"context,omitempty"`
}

// ClassificationResponse represents the complete response from classification
type ClassificationResponse struct {
	Service    ServiceResponse `json:"service"`
	Confidence float64         `json:"confidence,omitempty"`
	Error      string          `json:"error,omitempty"`
}

// ValidationResult represents the result of service validation
type ValidationResult struct {
	IsValid     bool   `json:"is_valid"`
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
	Reason      string `json:"reason,omitempty"`
}

// ValidateServiceID validates if a service ID is within the valid range (1-16)
func ValidateServiceID(serviceID int) error {
	if serviceID < 1 || serviceID > 16 {
		return fmt.Errorf("invalid service ID %d: must be between 1 and 16", serviceID)
	}
	return nil
}

// IsValidServiceID checks if a service ID is valid without returning an error
func IsValidServiceID(serviceID int) bool {
	return serviceID >= 1 && serviceID <= 16
}

// ValidateServiceDefinition validates a ServiceDefinition struct
func (sd *ServiceDefinition) Validate() error {
	if err := ValidateServiceID(sd.ID); err != nil {
		return err
	}
	if sd.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if sd.Description == "" {
		return fmt.Errorf("service description cannot be empty")
	}
	return nil
}

// ValidateServiceResponse validates a ServiceResponse struct
func (sr *ServiceResponse) Validate() error {
	if err := ValidateServiceID(sr.ServiceID); err != nil {
		return err
	}
	if sr.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	return nil
}

// ServiceRegistry holds all predefined services and provides mapping functionality
type ServiceRegistry struct {
	services        map[int]ServiceDefinition
	fallbackService ServiceDefinition
}

// NewServiceRegistry creates a new service registry with all predefined services
func NewServiceRegistry() *ServiceRegistry {
	services := map[int]ServiceDefinition{
		1: {
			ID:          1,
			Name:        "Consulta Limite / Vencimento do cartão / Melhor dia de compra",
			Keywords:    []string{"limite", "vencimento", "cartão", "melhor dia", "compra", "consulta"},
			Description: "Consultas sobre limite do cartão, data de vencimento e melhor dia para compras",
		},
		2: {
			ID:          2,
			Name:        "Segunda via de boleto de acordo",
			Keywords:    []string{"segunda via", "boleto", "acordo", "pagamento"},
			Description: "Solicitação de segunda via de boleto para acordos de pagamento",
		},
		3: {
			ID:          3,
			Name:        "Segunda via de Fatura",
			Keywords:    []string{"segunda via", "fatura", "conta", "cobrança"},
			Description: "Solicitação de segunda via da fatura do cartão",
		},
		4: {
			ID:          4,
			Name:        "Status de Entrega do Cartão",
			Keywords:    []string{"status", "entrega", "cartão", "envio", "correios"},
			Description: "Consulta sobre o status de entrega do cartão de crédito",
		},
		5: {
			ID:          5,
			Name:        "Status de cartão",
			Keywords:    []string{"status", "cartão", "situação", "ativo", "bloqueado"},
			Description: "Consulta sobre o status atual do cartão de crédito",
		},
		6: {
			ID:          6,
			Name:        "Solicitação de aumento de limite",
			Keywords:    []string{"aumento", "limite", "solicitação", "crédito"},
			Description: "Solicitação para aumento do limite do cartão de crédito",
		},
		7: {
			ID:          7,
			Name:        "Cancelamento de cartão",
			Keywords:    []string{"cancelamento", "cartão", "encerrar", "fechar"},
			Description: "Solicitação de cancelamento do cartão de crédito",
		},
		8: {
			ID:          8,
			Name:        "Telefones de seguradoras",
			Keywords:    []string{"telefone", "seguradora", "seguro", "contato"},
			Description: "Informações de contato das seguradoras parceiras",
		},
		9: {
			ID:          9,
			Name:        "Desbloqueio de Cartão",
			Keywords:    []string{"desbloqueio", "cartão", "desbloquear", "liberar"},
			Description: "Solicitação de desbloqueio do cartão de crédito",
		},
		10: {
			ID:          10,
			Name:        "Esqueceu senha / Troca de senha",
			Keywords:    []string{"senha", "esqueceu", "troca", "alterar", "redefinir"},
			Description: "Recuperação ou alteração de senha do cartão",
		},
		11: {
			ID:          11,
			Name:        "Perda e roubo",
			Keywords:    []string{"perda", "roubo", "perdeu", "roubaram", "furto"},
			Description: "Comunicação de perda ou roubo do cartão",
		},
		12: {
			ID:          12,
			Name:        "Consulta do Saldo Conta do Mais",
			Keywords:    []string{"saldo", "conta", "mais", "consulta", "extrato"},
			Description: "Consulta de saldo da Conta do Mais",
		},
		13: {
			ID:          13,
			Name:        "Pagamento de contas",
			Keywords:    []string{"pagamento", "contas", "pagar", "débito"},
			Description: "Serviços relacionados ao pagamento de contas",
		},
		14: {
			ID:          14,
			Name:        "Reclamações",
			Keywords:    []string{"reclamação", "problema", "insatisfação", "queixa"},
			Description: "Registro de reclamações e problemas",
		},
		15: {
			ID:          15,
			Name:        "Atendimento humano",
			Keywords:    []string{"atendimento", "humano", "pessoa", "operador"},
			Description: "Transferência para atendimento humano",
		},
		16: {
			ID:          16,
			Name:        "Token de proposta",
			Keywords:    []string{"token", "proposta", "código", "validação"},
			Description: "Solicitação ou validação de token de proposta",
		},
	}

	fallbackService := ServiceDefinition{
		ID:          15,
		Name:        "Atendimento humano",
		Keywords:    []string{"atendimento", "humano", "pessoa", "operador"},
		Description: "Transferência para atendimento humano",
	}

	return &ServiceRegistry{
		services:        services,
		fallbackService: fallbackService,
	}
}

// GetServiceByID returns a service definition by its ID
func (sr *ServiceRegistry) GetServiceByID(id int) (ServiceDefinition, bool) {
	service, exists := sr.services[id]
	return service, exists
}

// GetServiceNameByID returns the service name for a given ID
func (sr *ServiceRegistry) GetServiceNameByID(id int) string {
	if service, exists := sr.services[id]; exists {
		return service.Name
	}
	return sr.fallbackService.Name
}

// ValidateAndMapService validates a service ID and returns the corresponding service
func (sr *ServiceRegistry) ValidateAndMapService(id int) ServiceDefinition {
	if service, exists := sr.services[id]; exists {
		return service
	}
	return sr.fallbackService
}

// GetFallbackService returns the fallback service (Atendimento humano)
func (sr *ServiceRegistry) GetFallbackService() ServiceDefinition {
	return sr.fallbackService
}

// GetAllServices returns all available services
func (sr *ServiceRegistry) GetAllServices() map[int]ServiceDefinition {
	return sr.services
}

// CreateServiceResponse creates a ServiceResponse from a service ID with validation
func (sr *ServiceRegistry) CreateServiceResponse(id int) ServiceResponse {
	service := sr.ValidateAndMapService(id)
	return ServiceResponse{
		ServiceID:   service.ID,
		ServiceName: service.Name,
	}
}
