package main

// Intent representa uma intenção pré-carregada do CSV
type Intent struct {
	ServiceID   int
	ServiceName string
	IntentText  string
	Vector      []float64
}

// ClassificationResult representa o resultado da classificação
type ClassificationResult struct {
	ServiceID   int
	ServiceName string
	Confidence  float64
}

// APIRequest representa a requisição recebida pela API
type APIRequest struct {
	Intent string `json:"intent"`
}

// APIResponse representa a resposta da API
type APIResponse struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error string `json:"error"`
}

// TestCase representa um caso de teste individual
type TestCase struct {
	Intent            string `json:"intent"`
	ExpectedServiceID int    `json:"expected_service_id,omitempty"`
}

// TestBatchRequest representa uma requisição de lote de testes
type TestBatchRequest struct {
	TestCases []TestCase `json:"test_cases"`
}

// APITestResult representa o resultado de um teste individual via API
type APITestResult struct {
	Intent            string  `json:"intent"`
	ExpectedServiceID int     `json:"expected_service_id,omitempty"`
	PredictedID       int     `json:"predicted_service_id"`
	PredictedName     string  `json:"predicted_service_name"`
	Confidence        float64 `json:"confidence"`
	IsCorrect         bool    `json:"is_correct,omitempty"`
}

// TestBatchStats representa as estatísticas do lote de testes
type TestBatchStats struct {
	TotalTests           int                       `json:"total_tests"`
	CorrectPredictions   int                       `json:"correct_predictions,omitempty"`
	IncorrectPredictions int                       `json:"incorrect_predictions,omitempty"`
	AccuracyRate         float64                   `json:"accuracy_rate,omitempty"`
	AverageConfidence    float64                   `json:"average_confidence"`
	HighConfidence       int                       `json:"high_confidence_count"`   // >= 80%
	MediumConfidence     int                       `json:"medium_confidence_count"` // 50-80%
	LowConfidence        int                       `json:"low_confidence_count"`    // < 50%
	ByService            map[int]*ServiceTestStats `json:"by_service,omitempty"`
}

// ServiceTestStats representa estatísticas por serviço
type ServiceTestStats struct {
	ServiceID          int     `json:"service_id"`
	ServiceName        string  `json:"service_name"`
	TotalTests         int     `json:"total_tests"`
	CorrectPredictions int     `json:"correct_predictions"`
	AccuracyRate       float64 `json:"accuracy_rate"`
	AverageConfidence  float64 `json:"average_confidence"`
}

// TestBatchResponse representa a resposta completa do lote de testes
type TestBatchResponse struct {
	Results    []APITestResult `json:"results"`
	Statistics TestBatchStats  `json:"statistics"`
}
