package handler

type FindServiceRequest struct {
	Intent string `json:"intent"`
}

type FindServiceResponse struct {
	Success bool             `json:"success"`
	Data    *FindServiceData `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type FindServiceData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
