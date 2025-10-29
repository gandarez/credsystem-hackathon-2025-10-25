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

// ServiceData representa os dados do serviço encontrado
type ServiceData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// APIResponse representa a resposta da API
type APIResponse struct {
	Success bool         `json:"success"`
	Data    *ServiceData `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
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
	UsedAI            bool    `json:"used_ai"` // Indica se foi usado AI para classificar
}

// TestBatchStats representa as estatísticas do lote de testes
type TestBatchStats struct {
	TotalTests           int     `json:"total_tests"`
	CorrectPredictions   int     `json:"correct_predictions,omitempty"`
	IncorrectPredictions int     `json:"incorrect_predictions,omitempty"`
	AccuracyRate         float64 `json:"accuracy_rate,omitempty"`
	AverageConfidence    float64 `json:"average_confidence"`
	HighConfidence       int     `json:"high_confidence_count"`   // >= 80%
	MediumConfidence     int     `json:"medium_confidence_count"` // 50-80%
	LowConfidence        int     `json:"low_confidence_count"`    // < 50%

	// Métricas de uso da IA
	AIUsageCount         int     `json:"ai_usage_count"`         // Quantos casos usaram IA
	AIUsagePercentage    float64 `json:"ai_usage_percentage"`    // % de casos que usaram IA
	LocalUsageCount      int     `json:"local_usage_count"`      // Quantos casos usaram NLP local
	LocalUsagePercentage float64 `json:"local_usage_percentage"` // % de casos que usaram NLP local

	// Métricas de acerto por método
	AICorrectPredictions    int     `json:"ai_correct_predictions,omitempty"`    // Acertos quando usou IA
	AIAccuracyRate          float64 `json:"ai_accuracy_rate,omitempty"`          // Taxa de acerto da IA
	LocalCorrectPredictions int     `json:"local_correct_predictions,omitempty"` // Acertos quando usou NLP local
	LocalAccuracyRate       float64 `json:"local_accuracy_rate,omitempty"`       // Taxa de acerto do NLP local

	ByService map[int]*ServiceTestStats `json:"by_service,omitempty"`
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
