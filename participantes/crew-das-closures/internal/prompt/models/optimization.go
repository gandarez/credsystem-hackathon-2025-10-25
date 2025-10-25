package models

import (
	"strings"
	"time"
	"unicode"
)

// ModelSelector handles model selection and optimization logic
type ModelSelector struct {
	primaryModel   string
	fallbackModel  string
	costThreshold  float64
	performanceLog *PerformanceMonitor
}

// ModelConfig holds configuration for model selection
type ModelConfig struct {
	PrimaryModel     string  `json:"primary_model"`
	FallbackModel    string  `json:"fallback_model"`
	CostThreshold    float64 `json:"cost_threshold"`
	EnableMonitoring bool    `json:"enable_monitoring"`
}

// RequestComplexity represents the complexity analysis of a user request
type RequestComplexity struct {
	Score           float64 `json:"score"`
	WordCount       int     `json:"word_count"`
	HasNumbers      bool    `json:"has_numbers"`
	HasSpecialChars bool    `json:"has_special_chars"`
	Language        string  `json:"language"`
	Ambiguity       float64 `json:"ambiguity"`
}

// ModelRecommendation represents the recommended model for a request
type ModelRecommendation struct {
	ModelName     string            `json:"model_name"`
	Reason        string            `json:"reason"`
	Complexity    RequestComplexity `json:"complexity"`
	EstimatedCost float64           `json:"estimated_cost"`
	Priority      string            `json:"priority"` // "cost", "accuracy", "speed"
}

// PerformanceMetrics holds performance data for a model
type PerformanceMetrics struct {
	ModelName      string        `json:"model_name"`
	AverageLatency time.Duration `json:"average_latency"`
	SuccessRate    float64       `json:"success_rate"`
	TotalRequests  int64         `json:"total_requests"`
	FailedRequests int64         `json:"failed_requests"`
	LastUsed       time.Time     `json:"last_used"`
	CostPerRequest float64       `json:"cost_per_request"`
}

// PerformanceMonitor tracks performance metrics for different models
type PerformanceMonitor struct {
	metrics map[string]*PerformanceMetrics
	enabled bool
}

// Model constants
const (
	ModelMistral7B = "mistralai/mistral-7b-instruct"
	ModelGPT4OMini = "openai/gpt-4o-mini"

	// Cost estimates (tokens per dollar - approximate)
	MistralCostPerToken = 0.00001 // $0.01 per 1K tokens
	GPTCostPerToken     = 0.00015 // $0.15 per 1K tokens

	// Complexity thresholds
	LowComplexityThreshold    = 1.5
	MediumComplexityThreshold = 3.5
	HighComplexityThreshold   = 4.0

	// Performance thresholds
	MaxAcceptableLatency     = 5 * time.Second
	MinAcceptableSuccessRate = 0.85
)

// NewModelSelector creates a new model selector with default configuration
func NewModelSelector() *ModelSelector {
	return &ModelSelector{
		primaryModel:   ModelMistral7B,
		fallbackModel:  ModelGPT4OMini,
		costThreshold:  0.001, // $0.001 per request threshold
		performanceLog: NewPerformanceMonitor(),
	}
}

// NewModelSelectorWithConfig creates a model selector with custom configuration
func NewModelSelectorWithConfig(config ModelConfig) *ModelSelector {
	primaryModel := config.PrimaryModel
	if primaryModel == "" {
		primaryModel = ModelMistral7B
	}

	fallbackModel := config.FallbackModel
	if fallbackModel == "" {
		fallbackModel = ModelGPT4OMini
	}

	costThreshold := config.CostThreshold
	if costThreshold <= 0 {
		costThreshold = 0.001
	}

	var monitor *PerformanceMonitor
	if config.EnableMonitoring {
		monitor = NewPerformanceMonitor()
	}

	return &ModelSelector{
		primaryModel:   primaryModel,
		fallbackModel:  fallbackModel,
		costThreshold:  costThreshold,
		performanceLog: monitor,
	}
}

// SelectModel chooses the optimal model based on request complexity and performance data
func (ms *ModelSelector) SelectModel(userIntent string) ModelRecommendation {
	complexity := ms.AnalyzeComplexity(userIntent)

	// Get performance data for both models
	primaryPerf := ms.performanceLog.GetMetrics(ms.primaryModel)
	fallbackPerf := ms.performanceLog.GetMetrics(ms.fallbackModel)

	// Decision logic based on complexity and performance
	recommendation := ms.makeModelDecision(complexity, primaryPerf, fallbackPerf)

	return recommendation
}

// AnalyzeComplexity analyzes the complexity of a user request
func (ms *ModelSelector) AnalyzeComplexity(userIntent string) RequestComplexity {
	if userIntent == "" {
		return RequestComplexity{
			Score:     0,
			WordCount: 0,
			Language:  "unknown",
		}
	}

	words := strings.Fields(userIntent)
	wordCount := len(words)

	// Base complexity from word count
	complexityScore := float64(wordCount) * 0.1

	// Check for numbers (might indicate specific account queries)
	hasNumbers := strings.ContainsAny(userIntent, "0123456789")
	if hasNumbers {
		complexityScore += 0.5
	}

	// Check for special characters (might indicate technical issues)
	hasSpecialChars := false
	for _, r := range userIntent {
		if unicode.IsPunct(r) && r != '.' && r != ',' && r != '?' && r != '!' {
			hasSpecialChars = true
			complexityScore += 0.3
			break
		}
	}

	// Language detection (simple heuristic for Portuguese)
	language := "unknown"
	portugueseWords := []string{"cartão", "limite", "fatura", "conta", "senha", "boleto", "qual", "meu", "não", "como", "quando", "onde"}
	portugueseCount := 0
	lowerIntent := strings.ToLower(userIntent)

	for _, word := range portugueseWords {
		if strings.Contains(lowerIntent, word) {
			portugueseCount++
		}
	}

	if portugueseCount > 0 {
		language = "pt"
	} else {
		complexityScore += 1.0 // Unknown language is more complex
	}

	// Ambiguity detection (multiple question words, vague terms)
	ambiguityWords := []string{"como", "quando", "onde", "por que", "qual", "quais"}
	ambiguityScore := 0.0
	for _, word := range ambiguityWords {
		if strings.Contains(lowerIntent, word) {
			ambiguityScore += 0.5
		}
	}

	// Vague terms that might require human interpretation
	vagueTerms := []string{"problema", "erro", "não funciona", "não consigo"}
	for _, term := range vagueTerms {
		if strings.Contains(lowerIntent, term) {
			ambiguityScore += 1.0
		}
	}

	complexityScore += ambiguityScore

	return RequestComplexity{
		Score:           complexityScore,
		WordCount:       wordCount,
		HasNumbers:      hasNumbers,
		HasSpecialChars: hasSpecialChars,
		Language:        language,
		Ambiguity:       ambiguityScore,
	}
}

// makeModelDecision decides which model to use based on complexity and performance
func (ms *ModelSelector) makeModelDecision(complexity RequestComplexity, primaryPerf, fallbackPerf *PerformanceMetrics) ModelRecommendation {
	// Estimate token count (rough approximation: 1 word ≈ 1.3 tokens)
	estimatedTokens := float64(complexity.WordCount) * 1.3

	// Calculate estimated costs
	primaryCost := estimatedTokens * MistralCostPerToken
	fallbackCost := estimatedTokens * GPTCostPerToken

	// Decision matrix based on complexity
	switch {
	case complexity.Score <= LowComplexityThreshold:
		// Low complexity: prioritize cost efficiency
		if primaryPerf == nil || primaryPerf.SuccessRate >= MinAcceptableSuccessRate {
			return ModelRecommendation{
				ModelName:     ms.primaryModel,
				Reason:        "Low complexity request, using cost-efficient primary model",
				Complexity:    complexity,
				EstimatedCost: primaryCost,
				Priority:      "cost",
			}
		}

	case complexity.Score <= MediumComplexityThreshold:
		// Medium complexity: balance cost and accuracy
		if primaryPerf == nil || (primaryPerf.SuccessRate >= 0.90 && primaryPerf.AverageLatency <= MaxAcceptableLatency) {
			return ModelRecommendation{
				ModelName:     ms.primaryModel,
				Reason:        "Medium complexity request, primary model performing well",
				Complexity:    complexity,
				EstimatedCost: primaryCost,
				Priority:      "balanced",
			}
		}

	case complexity.Score > HighComplexityThreshold:
		// High complexity: prioritize accuracy
		return ModelRecommendation{
			ModelName:     ms.fallbackModel,
			Reason:        "High complexity request, using high-accuracy fallback model",
			Complexity:    complexity,
			EstimatedCost: fallbackCost,
			Priority:      "accuracy",
		}
	}

	// Handle ambiguous cases
	if complexity.Ambiguity > 2.0 {
		return ModelRecommendation{
			ModelName:     ms.fallbackModel,
			Reason:        "High ambiguity detected, using more capable model",
			Complexity:    complexity,
			EstimatedCost: fallbackCost,
			Priority:      "accuracy",
		}
	}

	// Performance-based fallback decisions
	if primaryPerf != nil {
		// If primary model is failing too often, use fallback
		if primaryPerf.SuccessRate < MinAcceptableSuccessRate {
			return ModelRecommendation{
				ModelName:     ms.fallbackModel,
				Reason:        "Primary model success rate below threshold, using fallback",
				Complexity:    complexity,
				EstimatedCost: fallbackCost,
				Priority:      "reliability",
			}
		}

		// If primary model is too slow, consider fallback for time-sensitive requests
		if primaryPerf.AverageLatency > MaxAcceptableLatency {
			return ModelRecommendation{
				ModelName:     ms.fallbackModel,
				Reason:        "Primary model latency too high, using faster fallback",
				Complexity:    complexity,
				EstimatedCost: fallbackCost,
				Priority:      "speed",
			}
		}
	}

	// Default to primary model
	return ModelRecommendation{
		ModelName:     ms.primaryModel,
		Reason:        "Default selection: primary model",
		Complexity:    complexity,
		EstimatedCost: primaryCost,
		Priority:      "cost",
	}
}

// OptimizeForCost returns the most cost-effective model for the given request
func (ms *ModelSelector) OptimizeForCost(userIntent string) ModelRecommendation {
	complexity := ms.AnalyzeComplexity(userIntent)

	// For cost optimization, always prefer the cheaper model unless complexity is very high
	if complexity.Score > HighComplexityThreshold || complexity.Ambiguity > 3.0 {
		estimatedTokens := float64(complexity.WordCount) * 1.3
		return ModelRecommendation{
			ModelName:     ms.fallbackModel,
			Reason:        "High complexity requires accurate model despite higher cost",
			Complexity:    complexity,
			EstimatedCost: estimatedTokens * GPTCostPerToken,
			Priority:      "accuracy",
		}
	}

	estimatedTokens := float64(complexity.WordCount) * 1.3
	return ModelRecommendation{
		ModelName:     ms.primaryModel,
		Reason:        "Cost optimization: using cheaper primary model",
		Complexity:    complexity,
		EstimatedCost: estimatedTokens * MistralCostPerToken,
		Priority:      "cost",
	}
}

// OptimizeForAccuracy returns the most accurate model for the given request
func (ms *ModelSelector) OptimizeForAccuracy(userIntent string) ModelRecommendation {
	complexity := ms.AnalyzeComplexity(userIntent)
	estimatedTokens := float64(complexity.WordCount) * 1.3

	return ModelRecommendation{
		ModelName:     ms.fallbackModel,
		Reason:        "Accuracy optimization: using high-performance model",
		Complexity:    complexity,
		EstimatedCost: estimatedTokens * GPTCostPerToken,
		Priority:      "accuracy",
	}
}

// GetModelCostEstimate returns the estimated cost for a request with a specific model
func (ms *ModelSelector) GetModelCostEstimate(userIntent, modelName string) float64 {
	complexity := ms.AnalyzeComplexity(userIntent)
	estimatedTokens := float64(complexity.WordCount) * 1.3

	switch modelName {
	case ModelMistral7B:
		return estimatedTokens * MistralCostPerToken
	case ModelGPT4OMini:
		return estimatedTokens * GPTCostPerToken
	default:
		return estimatedTokens * MistralCostPerToken // Default to primary model cost
	}
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics: make(map[string]*PerformanceMetrics),
		enabled: true,
	}
}

// RecordRequest records a request and its performance metrics
func (pm *PerformanceMonitor) RecordRequest(modelName string, latency time.Duration, success bool, cost float64) {
	if !pm.enabled {
		return
	}

	if pm.metrics[modelName] == nil {
		pm.metrics[modelName] = &PerformanceMetrics{
			ModelName:      modelName,
			TotalRequests:  0,
			FailedRequests: 0,
			SuccessRate:    1.0,
			CostPerRequest: cost,
		}
	}

	metrics := pm.metrics[modelName]
	metrics.TotalRequests++
	metrics.LastUsed = time.Now()

	if !success {
		metrics.FailedRequests++
	}

	// Update success rate
	metrics.SuccessRate = float64(metrics.TotalRequests-metrics.FailedRequests) / float64(metrics.TotalRequests)

	// Update average latency (exponential moving average)
	if metrics.AverageLatency == 0 {
		metrics.AverageLatency = latency
	} else {
		// Use exponential moving average with alpha = 0.1
		alpha := 0.1
		metrics.AverageLatency = time.Duration(float64(metrics.AverageLatency)*(1-alpha) + float64(latency)*alpha)
	}

	// Update average cost per request
	if metrics.CostPerRequest == 0 {
		metrics.CostPerRequest = cost
	} else {
		alpha := 0.1
		metrics.CostPerRequest = metrics.CostPerRequest*(1-alpha) + cost*alpha
	}
}

// GetMetrics returns performance metrics for a specific model
func (pm *PerformanceMonitor) GetMetrics(modelName string) *PerformanceMetrics {
	if !pm.enabled {
		return nil
	}

	return pm.metrics[modelName]
}

// GetAllMetrics returns performance metrics for all models
func (pm *PerformanceMonitor) GetAllMetrics() map[string]*PerformanceMetrics {
	if !pm.enabled {
		return nil
	}

	// Return a copy to prevent external modification
	result := make(map[string]*PerformanceMetrics)
	for k, v := range pm.metrics {
		result[k] = v
	}
	return result
}

// ResetMetrics clears all performance metrics
func (pm *PerformanceMonitor) ResetMetrics() {
	if pm.enabled {
		pm.metrics = make(map[string]*PerformanceMetrics)
	}
}

// Enable enables performance monitoring
func (pm *PerformanceMonitor) Enable() {
	pm.enabled = true
	if pm.metrics == nil {
		pm.metrics = make(map[string]*PerformanceMetrics)
	}
}

// Disable disables performance monitoring
func (pm *PerformanceMonitor) Disable() {
	pm.enabled = false
}

// IsEnabled returns whether performance monitoring is enabled
func (pm *PerformanceMonitor) IsEnabled() bool {
	return pm.enabled
}

// GetPerformanceMonitor returns the performance monitor for the model selector
func (ms *ModelSelector) GetPerformanceMonitor() *PerformanceMonitor {
	return ms.performanceLog
}
