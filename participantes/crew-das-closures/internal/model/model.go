package model

// Intent representa una intención del CSV
type Intent struct {
	Text        string
	ServiceID   int
	ServiceName string
}

// FindServiceRequest representa el request del endpoint
type FindServiceRequest struct {
	Intent string `json:"intent"`
}

// FindServiceResponse representa el response del endpoint
type FindServiceResponse struct {
	Success bool         `json:"success"`
	Data    *ServiceData `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// ServiceData contiene los datos del servicio
type ServiceData struct {
	ServiceID   uint8  `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// HealthResponse representa el response del healthcheck
type HealthResponse struct {
	Status string `json:"status"`
}
