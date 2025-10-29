package models

import "time"

// Config represents the overall configuration for the OpenRouter integration
type Config struct {
	OpenRouter OpenRouterConfig `json:"openrouter"`
	Services   ServicesConfig   `json:"services"`
	Prompts    PromptsConfig    `json:"prompts"`
}

// OpenRouterConfig holds OpenRouter API specific configuration
type OpenRouterConfig struct {
	APIKey     string        `json:"api_key"`
	BaseURL    string        `json:"base_url"`
	Model      string        `json:"model"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
}

// ServicesConfig holds service classification configuration
type ServicesConfig struct {
	FallbackServiceID int     `json:"fallback_service_id"`
	ValidServiceIDs   []int   `json:"valid_service_ids"`
	MinConfidence     float64 `json:"min_confidence"`
}

// PromptsConfig holds prompt management configuration
type PromptsConfig struct {
	SystemPromptPath     string `json:"system_prompt_path"`
	UseHardcodedFallback bool   `json:"use_hardcoded_fallback"`
}
