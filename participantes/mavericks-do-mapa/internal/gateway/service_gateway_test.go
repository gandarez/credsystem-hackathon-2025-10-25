package gateway

import (
	"context"
	"errors"
	"testing"

	"mavericksdomapa/client/openrouter"
)

type mockOpenRouterClient struct {
	response *openrouter.DataResponse
	err      error
}

func (m *mockOpenRouterClient) ChatCompletion(_ context.Context, _ string) (*openrouter.DataResponse, error) {
	return m.response, m.err
}

func TestOpenRouterServiceGateway_FindService_Success(t *testing.T) {
	mockClient := &mockOpenRouterClient{
		response: &openrouter.DataResponse{
			Success: true,
			Data: &openrouter.ServiceData{
				ServiceID:   3,
				ServiceName: "Segunda via de Fatura",
			},
		},
	}

	gw := NewOpenRouterServiceGateway(mockClient)

	service, err := gw.FindService(context.Background(), "quero meu boleto")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if service.ID != 3 || service.Name != "Segunda via de Fatura" {
		t.Fatalf("unexpected service response: %+v", service)
	}
}

func TestOpenRouterServiceGateway_FindService_NotFound(t *testing.T) {
	mockClient := &mockOpenRouterClient{
		response: &openrouter.DataResponse{
			Success: false,
			Error:   "nenhum serviço encontrado",
		},
	}

	gw := NewOpenRouterServiceGateway(mockClient)

	_, err := gw.FindService(context.Background(), "assunto desconhecido")
	if !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("expected ErrServiceNotFound, got %v", err)
	}
}

func TestOpenRouterServiceGateway_FindService_EmptyIntent(t *testing.T) {
	mockClient := &mockOpenRouterClient{}
	gw := NewOpenRouterServiceGateway(mockClient)

	_, err := gw.FindService(context.Background(), "   ")
	if !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("expected ErrServiceNotFound for empty intent, got %v", err)
	}
}

func TestOpenRouterServiceGateway_FindService_ClientError(t *testing.T) {
	mockClient := &mockOpenRouterClient{
		err: errors.New("api failure"),
	}

	gw := NewOpenRouterServiceGateway(mockClient)

	_, err := gw.FindService(context.Background(), "token de proposta")
	if err == nil || err.Error() != "api failure" {
		t.Fatalf("expected api failure error, got %v", err)
	}
}
