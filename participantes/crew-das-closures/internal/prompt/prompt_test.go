package prompt

import (
	"strings"
	"testing"
)

func TestGenerateClassificationPrompt_Success(t *testing.T) {
	pm := NewPromptManager()

	intent := "Quero saber meu limite do cartão"
	out, err := pm.GenerateClassificationPrompt(intent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == "" {
		t.Fatalf("expected non-empty prompt output")
	}

	// Contains a well-known snippet from the default system prompt
	if !strings.Contains(out, "expert AI assistant specialized") {
		t.Errorf("prompt should contain system prompt content")
	}

	// Contains the user intent
	if !strings.Contains(out, "User Intent: "+intent) {
		t.Errorf("prompt should contain the user intent; got: %s", out)
	}

	// Contains the classification instruction
	if !strings.Contains(out, "Please classify this intent") {
		t.Errorf("prompt should contain classification instruction")
	}
}

func TestGenerateClassificationPrompt_EmptyIntent(t *testing.T) {
	pm := NewPromptManager()
	if _, err := pm.GenerateClassificationPrompt(""); err == nil {
		t.Fatalf("expected error for empty intent, got nil")
	}
}

func TestGenerateModelSpecificPrompt_MistralFormat(t *testing.T) {
	pm := NewPromptManager()
	intent := "Preciso da segunda via da fatura"

	out, err := pm.GenerateModelSpecificPrompt(intent, "mistral-7b-instruct")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.HasPrefix(out, "<s>[INST]") || !strings.Contains(out, "[/INST]") {
		t.Errorf("mistral prompt should be wrapped in [INST] tags; got: %s", out)
	}

	if !strings.Contains(out, intent) {
		t.Errorf("mistral prompt should contain the user intent")
	}
}

func TestGenerateModelSpecificPrompt_DefaultFormat(t *testing.T) {
	pm := NewPromptManager()
	intent := "Quero parcelar minha fatura"

	out, err := pm.GenerateModelSpecificPrompt(intent, "gpt-4o-mini")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if strings.HasPrefix(out, "<s>[INST]") {
		t.Errorf("default prompt should not use mistral [INST] format")
	}

	if !strings.Contains(out, intent) {
		t.Errorf("default prompt should contain the user intent")
	}

	if !strings.Contains(out, "Classify this intent and provide your reasoning") {
		t.Errorf("default prompt should contain the classification instruction text")
	}
}

func TestNewPromptManagerWithConfig_CustomSystemPrompt(t *testing.T) {
	cfg := PromptConfig{SystemPromptTemplate: "CUSTOM SYSTEM PROMPT"}
	pm := NewPromptManagerWithConfig(cfg)

	if got := pm.GetSystemPrompt(); got != "CUSTOM SYSTEM PROMPT" {
		t.Fatalf("expected custom system prompt to be applied, got: %s", got)
	}

	out, err := pm.GenerateClassificationPrompt("Teste")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(out, "CUSTOM SYSTEM PROMPT") {
		t.Errorf("generated prompt should include the custom system prompt")
	}
}
