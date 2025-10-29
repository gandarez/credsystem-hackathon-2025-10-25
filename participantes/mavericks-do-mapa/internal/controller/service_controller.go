package controller

import (
	"context"
	"errors"

	"mavericksdomapa/internal/domain"
	"mavericksdomapa/internal/gateway"
)

var ErrServiceNotFound = errors.New("service not found")

type ServiceController struct {
	gateway gateway.ServiceGateway
}

func NewServiceController(gateway gateway.ServiceGateway) *ServiceController {
	return &ServiceController{gateway: gateway}
}

func (c *ServiceController) FindService(ctx context.Context, intent string) (*domain.Service, error) {
	service, err := c.gateway.FindService(ctx, intent)
	if err != nil {
		if errors.Is(err, gateway.ErrServiceNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}

	return service, nil
}
