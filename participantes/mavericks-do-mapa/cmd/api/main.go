package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"mavericksdomapa/internal/controller"
	"mavericksdomapa/internal/gateway"
	"mavericksdomapa/internal/handler"
)

func main() {
	app := fiber.New()

	serviceGateway := gateway.NewStaticServiceGateway()
	serviceController := controller.NewServiceController(serviceGateway)
	serviceHandler := handler.NewServiceHandler(serviceController)

	handler.RegisterRoutes(app, serviceHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
