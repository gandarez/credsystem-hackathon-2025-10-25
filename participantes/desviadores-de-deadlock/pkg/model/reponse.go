package model

import "desviadores-de-deadlock/pkg/openrouter"

type Reponse struct {
	Success bool                     `json:"success"`
	Data    *openrouter.DataResponse `json:"data"`
	Error   string                   `json:"error"`
}
