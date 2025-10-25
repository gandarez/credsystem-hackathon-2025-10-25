package main

import (
	"fmt"
	"os"
	"pioneiros-do-ponteiro/client/openrouter"

	"github.com/gofiber/fiber/v2"
)

type FindServiceRequest struct {
	Intent string `json:"intent"`
}

func main() {

	var openRouterKey = os.Getenv("OPENROUTER_API_KEY")

	app := fiber.New()

	// Grupo de rotas /api
	api := app.Group("/api")

	// Endpoint de verificação de saúde
	api.Get("/healthz", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Endpoint de busca de serviço
	api.Get("/find-service", func(c *fiber.Ctx) error {

		var req FindServiceRequest

		// Faz o bind do JSON recebido
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		fmt.Println("openRouterKey: ", openRouterKey)

		client := openrouter.NewClient("https://openrouter.ai/api/v1/", openrouter.WithAuth(openRouterKey))
		response, err := client.ChatCompletion(c.Context(), req.Intent)

		resultado := true
		errorMessage := ""

		if err != nil {
			resultado = false
			errorMessage = err.Error()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get service from OpenRouter",
			})
		}

		fmt.Println("Response from OpenRouter:", response)

		// Simulação de busca de serviço
		service := fiber.Map{
			"success": resultado,
			"data": fiber.Map{
				"service_id":   response.ServiceID,
				"service_name": response.ServiceName,
			},
			"error": errorMessage,
		}
		return c.Status(fiber.StatusOK).JSON(service)
	})

	// Inicia o servidor na porta 18020
	app.Listen(":18020")
}
