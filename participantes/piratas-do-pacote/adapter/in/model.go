package in

import "github.com/piratas-do-pacote/core"

type HttpHandler struct {
	inferenceUseCase *core.InferenceUseCase
}

func NewHttpHandler(inferenceUseCase *core.InferenceUseCase) *HttpHandler {
	return &HttpHandler{
		inferenceUseCase: inferenceUseCase,
	}
}
