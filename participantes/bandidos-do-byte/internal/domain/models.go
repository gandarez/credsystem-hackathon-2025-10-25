package domain

// FindServiceRequest representa a requisição para encontrar um serviço
type FindServiceRequest struct {
	Intent string `json:"intent"`
}

// ServiceData representa os dados do serviço encontrado
type ServiceData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// FindServiceResponse representa a resposta do endpoint find-service
type FindServiceResponse struct {
	Success bool         `json:"success"`
	Data    *ServiceData `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// HealthResponse representa a resposta do endpoint healthz
type HealthResponse struct {
	Status string `json:"status"`
}
