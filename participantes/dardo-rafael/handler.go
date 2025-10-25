package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Server representa o servidor HTTP
type Server struct {
	knnService          *KNNService
	aiClient            *AIClient
	serviceMap          map[int]string
	confidenceThreshold float64
}

// NewServer cria um novo servidor
func NewServer(knnService *KNNService, aiClient *AIClient, intents []Intent) *Server {
	// Criar mapa de service_id -> service_name
	serviceMap := make(map[int]string)
	for _, intent := range intents {
		serviceMap[intent.ServiceID] = intent.ServiceName
	}

	return &Server{
		knnService:          knnService,
		aiClient:            aiClient,
		serviceMap:          serviceMap,
		confidenceThreshold: 0.75, // Threshold padrão
	}
}

// healthzHandler responde ao endpoint /api/healthz
func (s *Server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// findServiceHandler responde ao endpoint /api/find-service
func (s *Server) findServiceHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Validar método HTTP
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	// Decodificar request
	var req APIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
		return
	}

	// Validar intent
	if req.Intent == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "intent cannot be empty"})
		return
	}

	// Classificar usando o serviço NLP
	result := s.knnService.Classify(req.Intent)

	log.Printf("NLP classification - Intent: %q, ServiceID: %d, Confidence: %.4f",
		req.Intent, result.ServiceID, result.Confidence)

	// Se a confiança for alta, retornar resultado local
	if result.Confidence >= s.confidenceThreshold {
		response := APIResponse{
			ServiceID:   result.ServiceID,
			ServiceName: result.ServiceName,
		}

		elapsed := time.Since(startTime)
		log.Printf("LOCAL - ServiceID: %d, ServiceName: %q, Confidence: %.4f, Time: %v",
			response.ServiceID, response.ServiceName, result.Confidence, elapsed)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fallback para IA
	log.Printf("Confidence too low (%.4f < %.2f), falling back to AI",
		result.Confidence, s.confidenceThreshold)

	aiResponse, err := s.aiClient.ClassifyWithAI(r.Context(), req.Intent, s.serviceMap)
	if err != nil {
		log.Printf("AI classification failed: %v", err)

		// Se a IA falhar, retornar o melhor palpite do KNN mesmo com baixa confiança
		response := APIResponse{
			ServiceID:   result.ServiceID,
			ServiceName: result.ServiceName,
		}

		elapsed := time.Since(startTime)
		log.Printf("FALLBACK_TO_KNN (AI failed) - ServiceID: %d, ServiceName: %q, Time: %v",
			response.ServiceID, response.ServiceName, elapsed)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	elapsed := time.Since(startTime)
	log.Printf("AI - ServiceID: %d, ServiceName: %q, Time: %v",
		aiResponse.ServiceID, aiResponse.ServiceName, elapsed)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(aiResponse)
}

// testBatchHandler responde ao endpoint /api/test-batch
func (s *Server) testBatchHandler(w http.ResponseWriter, r *http.Request) {
	// Validar método HTTP
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	// Decodificar request
	var req TestBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
		return
	}

	// Validar que há casos de teste
	if len(req.TestCases) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "test_cases cannot be empty"})
		return
	}

	// Processar cada caso de teste
	var results []APITestResult
	stats := TestBatchStats{
		ByService: make(map[int]*ServiceTestStats),
	}
	totalConfidence := 0.0

	for _, testCase := range req.TestCases {
		// Classificar usando o serviço NLP
		classification := s.knnService.Classify(testCase.Intent)

		// Criar resultado
		result := APITestResult{
			Intent:        testCase.Intent,
			PredictedID:   classification.ServiceID,
			PredictedName: classification.ServiceName,
			Confidence:    classification.Confidence,
		}

		// Se o esperado foi fornecido, calcular se está correto
		hasExpected := testCase.ExpectedServiceID > 0
		if hasExpected {
			result.ExpectedServiceID = testCase.ExpectedServiceID
			result.IsCorrect = classification.ServiceID == testCase.ExpectedServiceID

			// Atualizar estatísticas por serviço
			if _, exists := stats.ByService[testCase.ExpectedServiceID]; !exists {
				serviceName := s.serviceMap[testCase.ExpectedServiceID]
				stats.ByService[testCase.ExpectedServiceID] = &ServiceTestStats{
					ServiceID:   testCase.ExpectedServiceID,
					ServiceName: serviceName,
				}
			}

			serviceStats := stats.ByService[testCase.ExpectedServiceID]
			serviceStats.TotalTests++
			serviceStats.AverageConfidence += classification.Confidence

			if result.IsCorrect {
				serviceStats.CorrectPredictions++
				stats.CorrectPredictions++
			}
		}

		results = append(results, result)

		// Atualizar estatísticas gerais
		stats.TotalTests++
		totalConfidence += classification.Confidence

		// Categorizar por confiança
		if classification.Confidence >= 0.8 {
			stats.HighConfidence++
		} else if classification.Confidence >= 0.5 {
			stats.MediumConfidence++
		} else {
			stats.LowConfidence++
		}
	}

	// Calcular taxas finais
	if stats.TotalTests > 0 {
		stats.AverageConfidence = totalConfidence / float64(stats.TotalTests)

		// Se há valores esperados, calcular taxa de acerto
		if stats.CorrectPredictions > 0 || stats.TotalTests > 0 {
			stats.IncorrectPredictions = stats.TotalTests - stats.CorrectPredictions
			stats.AccuracyRate = float64(stats.CorrectPredictions) / float64(stats.TotalTests) * 100
		}
	}

	// Calcular taxas por serviço
	for _, serviceStats := range stats.ByService {
		if serviceStats.TotalTests > 0 {
			serviceStats.AccuracyRate = float64(serviceStats.CorrectPredictions) / float64(serviceStats.TotalTests) * 100
			serviceStats.AverageConfidence = serviceStats.AverageConfidence / float64(serviceStats.TotalTests)
		}
	}

	// Montar resposta
	response := TestBatchResponse{
		Results:    results,
		Statistics: stats,
	}

	log.Printf("Test batch completed - Total: %d, Accuracy: %.2f%%, Avg Confidence: %.4f",
		stats.TotalTests, stats.AccuracyRate, stats.AverageConfidence)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// loggingMiddleware registra todas as requisições
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	}
}

// StartServer inicia o servidor HTTP
func (s *Server) Start(port string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/healthz", loggingMiddleware(s.healthzHandler))
	mux.HandleFunc("/api/find-service", loggingMiddleware(s.findServiceHandler))
	mux.HandleFunc("/api/test-batch", loggingMiddleware(s.testBatchHandler))

	addr := ":" + port
	log.Printf("Server starting on %s", addr)

	return http.ListenAndServe(addr, mux)
}
