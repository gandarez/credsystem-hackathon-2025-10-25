package client

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type (
	Intent struct {
		ServiceID   uint8
		ServiceName string
		Intent      string
	}

	ServiceInfo struct {
		ID       uint8
		Name     string
		Examples []string
	}
)

var (
	loadedIntents  []Intent
	loadedServices map[uint8]*ServiceInfo
)

// LoadIntentsFromCSV carrega os intents do arquivo CSV
func LoadIntentsFromCSV(filePath string) ([]Intent, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.TrimLeadingSpace = true

	// Ler header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	var intents []Intent
	services := make(map[uint8]*ServiceInfo)

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		if len(record) != 3 {
			continue
		}

		serviceID, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			continue
		}

		serviceName := strings.TrimSpace(record[1])
		intentText := strings.TrimSpace(record[2])

		intent := Intent{
			ServiceID:   uint8(serviceID),
			ServiceName: serviceName,
			Intent:      intentText,
		}

		intents = append(intents, intent)

		// Agregar serviços
		if _, exists := services[uint8(serviceID)]; !exists {
			services[uint8(serviceID)] = &ServiceInfo{
				ID:       uint8(serviceID),
				Name:     serviceName,
				Examples: []string{},
			}
		}
		services[uint8(serviceID)].Examples = append(services[uint8(serviceID)].Examples, intentText)
	}

	loadedIntents = intents
	loadedServices = services

	return intents, nil
}

// GetServices retorna o mapa de serviços com seus exemplos
func GetServices() map[uint8]*ServiceInfo {
	return loadedServices
}

// GetLoadedIntents retorna todos os intents carregados
func GetLoadedIntents() []Intent {
	return loadedIntents
}

// BuildPromptWithExamples constrói o prompt com exemplos dos intents carregados
func BuildPromptWithExamples() string {
	var sb strings.Builder
	
	sb.WriteString("Classifique a intenção e retorne JSON: {\"service_id\": N, \"service_name\": \"Nome\"}\n\n")
	sb.WriteString("SERVIÇOS E EXEMPLOS:\n")
	
	for i := uint8(1); i <= 16; i++ {
		if service, exists := loadedServices[i]; exists {
			sb.WriteString(fmt.Sprintf("%d. %s\n", service.ID, service.Name))
			// Adicionar até 3 exemplos por serviço
			exampleCount := len(service.Examples)
			if exampleCount > 3 {
				exampleCount = 3
			}
			for j := 0; j < exampleCount; j++ {
				sb.WriteString(fmt.Sprintf("   - %s\n", service.Examples[j]))
			}
		}
	}
	
	sb.WriteString("\nRetorne APENAS o JSON, sem texto adicional.")
	return sb.String()
}
