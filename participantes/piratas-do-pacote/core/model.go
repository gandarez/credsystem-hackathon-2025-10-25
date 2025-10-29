package core

import "github.com/piratas-do-pacote/adapter/out"

type InferenceUseCase struct {
	inferer *out.AiInferer
}

func NewInferenceUseCase(inferer *out.AiInferer) *InferenceUseCase {
	return &InferenceUseCase{
		inferer: inferer,
	}
}

type InferenceInput struct {
	Intent string `json:"intent"`
}
type InferenceData struct {
	ServiceId   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type InferenceResult struct {
	Success bool          `json:"success"`
	Data    InferenceData `json:"data"`
	Error   string        `json:"error"`
}
