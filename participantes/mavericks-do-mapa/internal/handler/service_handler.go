package handler

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"mavericksdomapa/internal/controller"
)

type ServiceHandler struct {
	controller *controller.ServiceController
}

func NewServiceHandler(controller *controller.ServiceController) *ServiceHandler {
	return &ServiceHandler{controller: controller}
}

type findServiceRequest struct {
	Intent string `json:"intent"`
}

type serviceData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type findServiceResponse struct {
	Success bool         `json:"success"`
	Data    *serviceData `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
}

func (h *ServiceHandler) FindService(c *fiber.Ctx) error {
	var req findServiceRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(findServiceResponse{
			Success: false,
			Error:   "invalid request payload",
		})
	}

	intent := strings.TrimSpace(req.Intent)
	if intent == "" {
		return c.Status(fiber.StatusBadRequest).JSON(findServiceResponse{
			Success: false,
			Error:   "intent is required",
		})
	}

	service, err := h.controller.FindService(c.UserContext(), intent)
	if err != nil {
		if errors.Is(err, controller.ErrServiceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(findServiceResponse{
				Success: false,
				Error:   "no service matches the provided intent",
			})
		}

		log.Printf("failed to find service: %v", err)

		return c.Status(fiber.StatusInternalServerError).JSON(findServiceResponse{
			Success: false,
			Error:   "internal error",
		})
	}

	return c.JSON(findServiceResponse{
		Success: true,
		Data: &serviceData{
			ServiceID:   service.ID,
			ServiceName: service.Name,
		},
	})
}

func RegisterRoutes(app *fiber.App, handler *ServiceHandler) {
	app.Get("/api/healthz", Health)
	app.Post("/api/find-service", handler.FindService)
}

func Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
