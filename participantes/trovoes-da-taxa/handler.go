package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// classificationResult representa o resultado de uma classificação (local ou IA)
type classificationResult struct {
	serviceID   int
	serviceName string
	confidence  float64
	usedAI      bool
	err         error
}

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

// classifyParallel executa NLP local e IA em paralelo usando goroutines
// Retorna o resultado do NLP local se a confiança for alta, caso contrário usa o resultado da IA
func (s *Server) classifyParallel(ctx context.Context, intentText string) classificationResult {
	// Canais para receber os resultados
	localChan := make(chan classificationResult, 1)
	aiChan := make(chan classificationResult, 1)

	// Goroutine 1: Classificação NLP local (geralmente mais rápida)
	go func() {
		classification := s.knnService.Classify(intentText)
		localChan <- classificationResult{
			serviceID:   classification.ServiceID,
			serviceName: classification.ServiceName,
			confidence:  classification.Confidence,
			usedAI:      false,
			err:         nil,
		}
	}()

	// Goroutine 2: Classificação com IA (pode demorar mais)
	go func() {
		aiResponse, err := s.aiClient.ClassifyWithAI(ctx, intentText, s.serviceMap)
		if err != nil {
			aiChan <- classificationResult{
				err: err,
			}
			return
		}
		aiChan <- classificationResult{
			serviceID:   aiResponse.ServiceID,
			serviceName: aiResponse.ServiceName,
			confidence:  aiResponse.Confidence,
			usedAI:      true,
			err:         nil,
		}
	}()

	// Esperar resultados com select unificado
	var localResult classificationResult
	var hasLocalResult bool

	for {
		select {
		case localResult = <-localChan:
			hasLocalResult = true
			// Verificar se a confiança local é suficiente
			if localResult.confidence >= s.confidenceThreshold {
				// Alta confiança no NLP local, usar esse resultado
				// A goroutine da IA vai completar em background e o resultado será descartado
				log.Printf("Using LOCAL - Intent: %q, Confidence: %.4f", intentText, localResult.confidence)
				return localResult
			}
			// Confiança baixa, continuar esperando a IA
			log.Printf("LOW confidence LOCAL (%.4f), waiting for AI...", localResult.confidence)

		case aiResult := <-aiChan:
			if aiResult.err != nil {
				// Verificar se é um erro de VALIDAÇÃO (input inválido)
				// ou um erro TÉCNICO (timeout, API error, etc)
				if IsValidationError(aiResult.err) {
					// Erro de validação: input é inválido
					// NÃO fazer fallback para NLP local, propagar o erro
					log.Printf("AI Validation Error: %v - Input rejected", aiResult.err)
					return aiResult // Retorna com erro para propagar ao cliente
				}

				// Erro técnico: tentar fallback para NLP local
				if hasLocalResult {
					log.Printf("AI technical error: %v, using LOCAL fallback", aiResult.err)
					return localResult
				}
				// IA falhou tecnicamente e ainda não temos resultado local, esperar local
				log.Printf("AI technical error: %v, waiting for LOCAL...", aiResult.err)
				localResult = <-localChan
				return localResult
			}
			// IA sucedeu
			log.Printf("Using AI - Intent: %q", intentText)
			if hasLocalResult {
				aiResult.confidence = localResult.confidence // Preservar confiança do NLP para estatísticas
			}
			return aiResult

		case <-ctx.Done():
			// Contexto cancelado/timeout
			if hasLocalResult {
				log.Printf("Context cancelled, using LOCAL fallback")
				return localResult
			}
			// Ainda não temos resultado local, esperar
			log.Printf("Context cancelled, waiting for LOCAL...")
			localResult = <-localChan
			return localResult

		case <-time.After(25 * time.Second):
			// Timeout geral
			if hasLocalResult {
				log.Printf("Timeout, using LOCAL fallback")
				return localResult
			}
			// Ainda não temos resultado local, esperar
			log.Printf("Timeout, waiting for LOCAL...")
			localResult = <-localChan
			return localResult
		}
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
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	// Decodificar request
	var req APIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	// Validar intent
	if req.Intent == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "intent cannot be empty",
		})
		return
	}

	// Classificar usando execução paralela (NLP local + IA em goroutines)
	result := s.classifyParallel(r.Context(), req.Intent)

	// Verificar se há erro de validação
	if result.err != nil {
		elapsed := time.Since(startTime)
		log.Printf("VALIDATION ERROR - Intent: %q, Error: %v, Time: %v", req.Intent, result.err, elapsed)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   result.err.Error(),
		})
		return
	}

	// Preparar resposta de sucesso
	response := APIResponse{
		Success: true,
		Data: &ServiceData{
			ServiceID:   result.serviceID,
			ServiceName: result.serviceName,
		},
	}

	elapsed := time.Since(startTime)
	method := "LOCAL"
	if result.usedAI {
		method = "AI"
	}

	log.Printf("%s - Intent: %q, ServiceID: %d, ServiceName: %q, Confidence: %.4f, Time: %v",
		method, req.Intent, response.Data.ServiceID, response.Data.ServiceName, result.confidence, elapsed)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// testBatchHandler responde ao endpoint /api/test-batch
func (s *Server) testBatchHandler(w http.ResponseWriter, r *http.Request) {
	// Validar método HTTP
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	// Decodificar request
	var req TestBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// Validar que há casos de teste
	if len(req.TestCases) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"error": "test_cases cannot be empty"})
		return
	}

	// Processar cada caso de teste
	var results []APITestResult
	stats := TestBatchStats{
		ByService: make(map[int]*ServiceTestStats),
	}
	totalConfidence := 0.0

	for _, testCase := range req.TestCases {
		// Classificar usando execução paralela (NLP local + IA em goroutines)
		classification := s.classifyParallel(r.Context(), testCase.Intent)

		// Criar resultado
		result := APITestResult{
			Intent:        testCase.Intent,
			PredictedID:   classification.serviceID,
			PredictedName: classification.serviceName,
			Confidence:    classification.confidence,
			UsedAI:        classification.usedAI,
		}

		// Verificar se houve erro de validação
		if classification.err != nil {
			// Input foi rejeitado pela validação
			result.PredictedID = 0
			result.PredictedName = "REJECTED: " + classification.err.Error()
			result.Confidence = 0.0

			// Se esperávamos rejeição (expected_service_id == 0), considerar correto
			if testCase.ExpectedServiceID == 0 {
				result.IsCorrect = true
				stats.CorrectPredictions++
				if classification.usedAI {
					stats.AICorrectPredictions++
				}
			} else {
				// Esperávamos classificação mas foi rejeitado
				result.ExpectedServiceID = testCase.ExpectedServiceID
				result.IsCorrect = false
			}
		} else {
			// Classificação normal (sem erro)

			// Atualizar contadores de uso
			if classification.usedAI {
				stats.AIUsageCount++
			} else {
				stats.LocalUsageCount++
			}

			// Se o esperado foi fornecido, calcular se está correto
			hasExpected := testCase.ExpectedServiceID > 0
			if hasExpected {
				result.ExpectedServiceID = testCase.ExpectedServiceID
				result.IsCorrect = classification.serviceID == testCase.ExpectedServiceID

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
				serviceStats.AverageConfidence += classification.confidence

				if result.IsCorrect {
					serviceStats.CorrectPredictions++
					stats.CorrectPredictions++

					// Atualizar métricas por método
					if classification.usedAI {
						stats.AICorrectPredictions++
					} else {
						stats.LocalCorrectPredictions++
					}
				}
			}
		} // fim do else (classificação normal)

		results = append(results, result)

		// Atualizar estatísticas gerais
		stats.TotalTests++
		totalConfidence += classification.confidence

		// Categorizar por confiança
		if classification.confidence >= 0.8 {
			stats.HighConfidence++
		} else if classification.confidence >= 0.5 {
			stats.MediumConfidence++
		} else {
			stats.LowConfidence++
		}
	}

	// Calcular taxas finais
	if stats.TotalTests > 0 {
		stats.AverageConfidence = totalConfidence / float64(stats.TotalTests)

		// Calcular percentuais de uso
		stats.AIUsagePercentage = float64(stats.AIUsageCount) / float64(stats.TotalTests) * 100
		stats.LocalUsagePercentage = float64(stats.LocalUsageCount) / float64(stats.TotalTests) * 100

		// Se há valores esperados, calcular taxa de acerto
		if stats.CorrectPredictions > 0 || stats.TotalTests > 0 {
			stats.IncorrectPredictions = stats.TotalTests - stats.CorrectPredictions
			stats.AccuracyRate = float64(stats.CorrectPredictions) / float64(stats.TotalTests) * 100
		}

		// Calcular taxas de acerto por método
		if stats.AIUsageCount > 0 {
			stats.AIAccuracyRate = float64(stats.AICorrectPredictions) / float64(stats.AIUsageCount) * 100
		}
		if stats.LocalUsageCount > 0 {
			stats.LocalAccuracyRate = float64(stats.LocalCorrectPredictions) / float64(stats.LocalUsageCount) * 100
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

	log.Printf("Test batch completed - Total: %d, Accuracy: %.2f%%, AI Usage: %.1f%% (%.1f%% accuracy), Local: %.1f%% (%.1f%% accuracy)",
		stats.TotalTests, stats.AccuracyRate,
		stats.AIUsagePercentage, stats.AIAccuracyRate,
		stats.LocalUsagePercentage, stats.LocalAccuracyRate)

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
