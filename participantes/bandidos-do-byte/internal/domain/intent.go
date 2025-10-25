package domain

// IntentExample representa um exemplo de intent do arquivo de treinamento
type IntentExample struct {
	ServiceID   int
	ServiceName string
	Intent      string
}

// IntentClassificationRequest representa a requisição para classificação de intent
type IntentClassificationRequest struct {
	UserIntent string
	Examples   []IntentExample
}

// IntentClassificationResponse representa a resposta da classificação
type IntentClassificationResponse struct {
	ServiceID   int
	ServiceName string
	Confidence  float64
}
