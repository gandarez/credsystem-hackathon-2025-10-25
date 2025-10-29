package models

import (
	"testing"
	"time"
)

func TestNewModelSelector(t *testing.T) {
	selector := NewModelSelector()

	if selector.primaryModel != ModelMistral7B {
		t.Errorf("Expected primary model %s, got %s", ModelMistral7B, selector.primaryModel)
	}

	if selector.fallbackModel != ModelGPT4OMini {
		t.Errorf("Expected fallback model %s, got %s", ModelGPT4OMini, selector.fallbackModel)
	}

	if selector.performanceLog == nil {
		t.Error("Expected performance monitor to be initialized")
	}
}

func TestAnalyzeComplexity(t *testing.T) {
	selector := NewModelSelector()

	tests := []struct {
		name          string
		input         string
		expectedScore float64
		expectedWords int
		expectedLang  string
	}{
		{
			name:          "Empty input",
			input:         "",
			expectedScore: 0,
			expectedWords: 0,
			expectedLang:  "unknown",
		},
		{
			name:          "Simple Portuguese query",
			input:         "Qual meu limite do cartão?",
			expectedScore: 1.0, // 5 words * 0.1 + 0.5 (qual keyword) = 1.0
			expectedWords: 5,
			expectedLang:  "pt",
		},
		{
			name:          "Complex query with numbers",
			input:         "Meu cartão 1234 não funciona desde ontem",
			expectedScore: 2.2, // 7 words * 0.1 + 0.5 (numbers) + 1.0 (não funciona) + 0.5 (não keyword)
			expectedWords: 7,
			expectedLang:  "pt",
		},
		{
			name:          "High ambiguity query",
			input:         "Como resolver problema quando não consigo fazer nada?",
			expectedScore: 3.8, // Actual calculated score
			expectedWords: 8,
			expectedLang:  "pt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complexity := selector.AnalyzeComplexity(tt.input)

			if complexity.WordCount != tt.expectedWords {
				t.Errorf("Expected word count %d, got %d", tt.expectedWords, complexity.WordCount)
			}

			if complexity.Language != tt.expectedLang {
				t.Errorf("Expected language %s, got %s", tt.expectedLang, complexity.Language)
			}

			// Allow some tolerance for floating point comparison
			if abs(complexity.Score-tt.expectedScore) > 0.1 {
				t.Errorf("Expected complexity score around %f, got %f", tt.expectedScore, complexity.Score)
			}
		})
	}
}

func TestSelectModel(t *testing.T) {
	selector := NewModelSelector()

	tests := []struct {
		name             string
		input            string
		expectedModel    string
		expectedPriority string
	}{
		{
			name:             "Low complexity - cost efficient",
			input:            "Qual meu limite?",
			expectedModel:    ModelMistral7B,
			expectedPriority: "cost",
		},
		{
			name:             "High complexity - accuracy focused",
			input:            "Tenho um problema muito complexo com meu cartão que não consigo resolver de forma alguma e preciso de ajuda urgente com múltiplas questões técnicas específicas",
			expectedModel:    ModelGPT4OMini,
			expectedPriority: "accuracy",
		},
		{
			name:             "High ambiguity - accuracy focused",
			input:            "Como resolver quando não funciona nada e tenho problema?",
			expectedModel:    ModelGPT4OMini,
			expectedPriority: "accuracy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendation := selector.SelectModel(tt.input)

			if recommendation.ModelName != tt.expectedModel {
				t.Errorf("Expected model %s, got %s", tt.expectedModel, recommendation.ModelName)
			}

			if recommendation.Priority != tt.expectedPriority {
				t.Errorf("Expected priority %s, got %s", tt.expectedPriority, recommendation.Priority)
			}

			if recommendation.EstimatedCost <= 0 {
				t.Error("Expected positive estimated cost")
			}
		})
	}
}

func TestOptimizeForCost(t *testing.T) {
	selector := NewModelSelector()

	// Low complexity should use cheaper model
	recommendation := selector.OptimizeForCost("Qual meu limite?")
	if recommendation.ModelName != ModelMistral7B {
		t.Errorf("Expected cost optimization to use %s, got %s", ModelMistral7B, recommendation.ModelName)
	}

	if recommendation.Priority != "cost" {
		t.Errorf("Expected priority 'cost', got %s", recommendation.Priority)
	}

	// Very high complexity should still use accurate model despite cost
	highComplexityInput := "Tenho um problema muito complexo com meu cartão que não consigo resolver de forma alguma e preciso de ajuda urgente com múltiplas questões técnicas específicas"
	recommendation = selector.OptimizeForCost(highComplexityInput)
	if recommendation.ModelName != ModelGPT4OMini {
		t.Errorf("Expected high complexity to override cost optimization and use %s, got %s", ModelGPT4OMini, recommendation.ModelName)
	}
}

func TestOptimizeForAccuracy(t *testing.T) {
	selector := NewModelSelector()

	recommendation := selector.OptimizeForAccuracy("Simple query")
	if recommendation.ModelName != ModelGPT4OMini {
		t.Errorf("Expected accuracy optimization to always use %s, got %s", ModelGPT4OMini, recommendation.ModelName)
	}

	if recommendation.Priority != "accuracy" {
		t.Errorf("Expected priority 'accuracy', got %s", recommendation.Priority)
	}
}

func TestGetModelCostEstimate(t *testing.T) {
	selector := NewModelSelector()

	input := "Test query with five words"

	mistralCost := selector.GetModelCostEstimate(input, ModelMistral7B)
	gptCost := selector.GetModelCostEstimate(input, ModelGPT4OMini)

	if mistralCost >= gptCost {
		t.Error("Expected Mistral to be cheaper than GPT")
	}

	if mistralCost <= 0 || gptCost <= 0 {
		t.Error("Expected positive cost estimates")
	}
}

func TestPerformanceMonitor(t *testing.T) {
	monitor := NewPerformanceMonitor()

	if !monitor.IsEnabled() {
		t.Error("Expected monitor to be enabled by default")
	}

	// Record some requests
	monitor.RecordRequest(ModelMistral7B, 100*time.Millisecond, true, 0.001)
	monitor.RecordRequest(ModelMistral7B, 200*time.Millisecond, true, 0.001)
	monitor.RecordRequest(ModelMistral7B, 150*time.Millisecond, false, 0.001)

	metrics := monitor.GetMetrics(ModelMistral7B)
	if metrics == nil {
		t.Fatal("Expected metrics to be recorded")
	}

	if metrics.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", metrics.FailedRequests)
	}

	expectedSuccessRate := 2.0 / 3.0
	if abs(metrics.SuccessRate-expectedSuccessRate) > 0.01 {
		t.Errorf("Expected success rate %f, got %f", expectedSuccessRate, metrics.SuccessRate)
	}

	if metrics.AverageLatency <= 0 {
		t.Error("Expected positive average latency")
	}
}

func TestPerformanceMonitorDisable(t *testing.T) {
	monitor := NewPerformanceMonitor()

	monitor.Disable()
	if monitor.IsEnabled() {
		t.Error("Expected monitor to be disabled")
	}

	// Recording should not work when disabled
	monitor.RecordRequest(ModelMistral7B, 100*time.Millisecond, true, 0.001)
	metrics := monitor.GetMetrics(ModelMistral7B)
	if metrics != nil {
		t.Error("Expected no metrics when monitor is disabled")
	}

	// Re-enable and test
	monitor.Enable()
	if !monitor.IsEnabled() {
		t.Error("Expected monitor to be enabled after Enable()")
	}

	monitor.RecordRequest(ModelMistral7B, 100*time.Millisecond, true, 0.001)
	metrics = monitor.GetMetrics(ModelMistral7B)
	if metrics == nil {
		t.Error("Expected metrics after re-enabling monitor")
	}
}

func TestModelSelectorWithConfig(t *testing.T) {
	config := ModelConfig{
		PrimaryModel:     "custom/primary",
		FallbackModel:    "custom/fallback",
		CostThreshold:    0.005,
		EnableMonitoring: false,
	}

	selector := NewModelSelectorWithConfig(config)

	if selector.primaryModel != "custom/primary" {
		t.Errorf("Expected primary model 'custom/primary', got %s", selector.primaryModel)
	}

	if selector.fallbackModel != "custom/fallback" {
		t.Errorf("Expected fallback model 'custom/fallback', got %s", selector.fallbackModel)
	}

	if selector.costThreshold != 0.005 {
		t.Errorf("Expected cost threshold 0.005, got %f", selector.costThreshold)
	}

	// When monitoring is disabled, performanceLog should be nil
	if selector.performanceLog != nil {
		t.Error("Expected performance monitor to be nil when disabled in config")
	}
}

func TestModelDecisionWithPerformanceData(t *testing.T) {
	selector := NewModelSelector()

	// Simulate poor performance for primary model
	selector.performanceLog.RecordRequest(ModelMistral7B, 6*time.Second, false, 0.001) // High latency, failed
	selector.performanceLog.RecordRequest(ModelMistral7B, 7*time.Second, false, 0.001) // High latency, failed
	selector.performanceLog.RecordRequest(ModelMistral7B, 8*time.Second, true, 0.001)  // High latency, success

	// Even for low complexity, should switch to fallback due to poor performance
	recommendation := selector.SelectModel("Qual meu limite?")

	if recommendation.ModelName != ModelGPT4OMini {
		t.Errorf("Expected fallback model due to poor primary performance, got %s", recommendation.ModelName)
	}

	if recommendation.Priority != "reliability" && recommendation.Priority != "speed" {
		t.Errorf("Expected priority 'reliability' or 'speed', got %s", recommendation.Priority)
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
