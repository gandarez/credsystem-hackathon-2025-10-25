package util

import "participantes/galaxia-dos-gophers/internal/dto"

func DetermineService(req *dto.OpenRouterRequest) (uint8, string) {
	// TODO make logical

	if len(req.Messages) > 0 {
		content := req.Messages[0].Content

		if len(content) > 100 {
			return 1, ""
		} else if len(content) > 50 {
			return 2, ""
		}
	}

	return 3, ""
}
