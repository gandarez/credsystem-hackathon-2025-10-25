package models

// Service representa um serviço disponível
type Service struct {
	ID   int    `json:"service_id"`
	Name string `json:"service_name"`
}

// FindServiceRequest representa a requisição para encontrar serviço
type FindServiceRequest struct {
	Intent string `json:"intent"`
}

// FindServiceResponse representa a resposta do endpoint find-service
type FindServiceResponse struct {
	Success bool     `json:"success"`
	Data    *Service `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// HealthResponse representa a resposta do endpoint healthz
type HealthResponse struct {
	Status string `json:"status"`
}
