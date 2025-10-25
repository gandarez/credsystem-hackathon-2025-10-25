package models

import (
	"testing"
)

func TestValidateServiceID(t *testing.T) {
	tests := []struct {
		name      string
		serviceID int
		wantError bool
	}{
		{"Valid ID 1", 1, false},
		{"Valid ID 8", 8, false},
		{"Valid ID 16", 16, false},
		{"Invalid ID 0", 0, true},
		{"Invalid ID 17", 17, true},
		{"Invalid ID -1", -1, true},
		{"Invalid ID 100", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServiceID(tt.serviceID)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateServiceID(%d) error = %v, wantError %v", tt.serviceID, err, tt.wantError)
			}
		})
	}
}

func TestIsValidServiceID(t *testing.T) {
	tests := []struct {
		name      string
		serviceID int
		want      bool
	}{
		{"Valid ID 1", 1, true},
		{"Valid ID 8", 8, true},
		{"Valid ID 16", 16, true},
		{"Invalid ID 0", 0, false},
		{"Invalid ID 17", 17, false},
		{"Invalid ID -1", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidServiceID(tt.serviceID); got != tt.want {
				t.Errorf("IsValidServiceID(%d) = %v, want %v", tt.serviceID, got, tt.want)
			}
		})
	}
}

func TestServiceDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		service ServiceDefinition
		wantErr bool
	}{
		{
			name: "Valid service definition",
			service: ServiceDefinition{
				ID:          1,
				Name:        "Test Service",
				Keywords:    []string{"test", "service"},
				Description: "A test service",
			},
			wantErr: false,
		},
		{
			name: "Invalid service ID",
			service: ServiceDefinition{
				ID:          0,
				Name:        "Test Service",
				Keywords:    []string{"test"},
				Description: "A test service",
			},
			wantErr: true,
		},
		{
			name: "Empty service name",
			service: ServiceDefinition{
				ID:          1,
				Name:        "",
				Keywords:    []string{"test"},
				Description: "A test service",
			},
			wantErr: true,
		},
		{
			name: "Empty description",
			service: ServiceDefinition{
				ID:          1,
				Name:        "Test Service",
				Keywords:    []string{"test"},
				Description: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.service.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ServiceDefinition.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceResponse_Validate(t *testing.T) {
	tests := []struct {
		name     string
		response ServiceResponse
		wantErr  bool
	}{
		{
			name: "Valid service response",
			response: ServiceResponse{
				ServiceID:   1,
				ServiceName: "Test Service",
			},
			wantErr: false,
		},
		{
			name: "Invalid service ID",
			response: ServiceResponse{
				ServiceID:   0,
				ServiceName: "Test Service",
			},
			wantErr: true,
		},
		{
			name: "Empty service name",
			response: ServiceResponse{
				ServiceID:   1,
				ServiceName: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.response.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ServiceResponse.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewServiceRegistry(t *testing.T) {
	registry := NewServiceRegistry()

	// Test that all 16 services are loaded
	if len(registry.services) != 16 {
		t.Errorf("Expected 16 services, got %d", len(registry.services))
	}

	// Test that fallback service is set correctly
	if registry.fallbackService.ID != 15 {
		t.Errorf("Expected fallback service ID 15, got %d", registry.fallbackService.ID)
	}

	// Test that all service IDs from 1-16 exist
	for i := 1; i <= 16; i++ {
		if _, exists := registry.services[i]; !exists {
			t.Errorf("Service ID %d not found in registry", i)
		}
	}
}

func TestServiceRegistry_GetServiceByID(t *testing.T) {
	registry := NewServiceRegistry()

	tests := []struct {
		name     string
		id       int
		wantName string
		wantOK   bool
	}{
		{"Valid ID 1", 1, "Consulta Limite / Vencimento do cartão / Melhor dia de compra", true},
		{"Valid ID 15", 15, "Atendimento humano", true},
		{"Valid ID 16", 16, "Token de proposta", true},
		{"Invalid ID 0", 0, "", false},
		{"Invalid ID 17", 17, "", false},
		{"Invalid ID -1", -1, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, ok := registry.GetServiceByID(tt.id)
			if ok != tt.wantOK {
				t.Errorf("GetServiceByID(%d) ok = %v, want %v", tt.id, ok, tt.wantOK)
			}
			if ok && service.Name != tt.wantName {
				t.Errorf("GetServiceByID(%d) name = %v, want %v", tt.id, service.Name, tt.wantName)
			}
		})
	}
}

func TestServiceRegistry_GetServiceNameByID(t *testing.T) {
	registry := NewServiceRegistry()

	tests := []struct {
		name     string
		id       int
		wantName string
	}{
		{"Valid ID 1", 1, "Consulta Limite / Vencimento do cartão / Melhor dia de compra"},
		{"Valid ID 15", 15, "Atendimento humano"},
		{"Invalid ID 0", 0, "Atendimento humano"},   // Should return fallback
		{"Invalid ID 17", 17, "Atendimento humano"}, // Should return fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := registry.GetServiceNameByID(tt.id)
			if name != tt.wantName {
				t.Errorf("GetServiceNameByID(%d) = %v, want %v", tt.id, name, tt.wantName)
			}
		})
	}
}

func TestServiceRegistry_ValidateAndMapService(t *testing.T) {
	registry := NewServiceRegistry()

	tests := []struct {
		name   string
		id     int
		wantID int
	}{
		{"Valid ID 1", 1, 1},
		{"Valid ID 8", 8, 8},
		{"Valid ID 16", 16, 16},
		{"Invalid ID 0", 0, 15},   // Should return fallback (ID 15)
		{"Invalid ID 17", 17, 15}, // Should return fallback (ID 15)
		{"Invalid ID -1", -1, 15}, // Should return fallback (ID 15)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := registry.ValidateAndMapService(tt.id)
			if service.ID != tt.wantID {
				t.Errorf("ValidateAndMapService(%d) ID = %v, want %v", tt.id, service.ID, tt.wantID)
			}
		})
	}
}

func TestServiceRegistry_CreateServiceResponse(t *testing.T) {
	registry := NewServiceRegistry()

	tests := []struct {
		name           string
		id             int
		wantServiceID  int
		wantValidation bool
	}{
		{"Valid ID 1", 1, 1, true},
		{"Valid ID 15", 15, 15, true},
		{"Invalid ID 0", 0, 15, true},   // Should return fallback but still be valid
		{"Invalid ID 17", 17, 15, true}, // Should return fallback but still be valid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := registry.CreateServiceResponse(tt.id)
			if response.ServiceID != tt.wantServiceID {
				t.Errorf("CreateServiceResponse(%d) ServiceID = %v, want %v", tt.id, response.ServiceID, tt.wantServiceID)
			}
			if response.ServiceName == "" {
				t.Errorf("CreateServiceResponse(%d) ServiceName is empty", tt.id)
			}
			// Test that the response validates correctly
			if err := response.Validate(); (err == nil) != tt.wantValidation {
				t.Errorf("CreateServiceResponse(%d).Validate() error = %v, wantValidation %v", tt.id, err, tt.wantValidation)
			}
		})
	}
}

func TestServiceRegistry_GetFallbackService(t *testing.T) {
	registry := NewServiceRegistry()
	fallback := registry.GetFallbackService()

	if fallback.ID != 15 {
		t.Errorf("GetFallbackService() ID = %v, want 15", fallback.ID)
	}
	if fallback.Name != "Atendimento humano" {
		t.Errorf("GetFallbackService() Name = %v, want 'Atendimento humano'", fallback.Name)
	}
}

func TestServiceRegistry_GetAllServices(t *testing.T) {
	registry := NewServiceRegistry()
	services := registry.GetAllServices()

	if len(services) != 16 {
		t.Errorf("GetAllServices() length = %v, want 16", len(services))
	}

	// Verify all services have valid data
	for id, service := range services {
		if service.ID != id {
			t.Errorf("Service ID mismatch: map key %d, service ID %d", id, service.ID)
		}
		if service.Name == "" {
			t.Errorf("Service ID %d has empty name", id)
		}
		if service.Description == "" {
			t.Errorf("Service ID %d has empty description", id)
		}
		if len(service.Keywords) == 0 {
			t.Errorf("Service ID %d has no keywords", id)
		}
	}
}
