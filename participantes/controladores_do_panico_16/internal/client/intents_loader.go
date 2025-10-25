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
	cachedIntents  []Intent
	cachedServices map[uint8]*ServiceInfo
)

// LoadIntentsFromCSV carrega os intents do arquivo CSV
func LoadIntentsFromCSV(filePath string) ([]Intent, error) {
	if cachedIntents != nil {
		return cachedIntents, nil
	}

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

	cachedIntents = intents
	cachedServices = services

	return intents, nil
}

// GetServices retorna o mapa de serviços com seus exemplos
func GetServices() map[uint8]*ServiceInfo {
	return cachedServices
}

// GetServicesList retorna uma lista formatada dos serviços (sem exemplos para prompt)
func GetServicesList() string {
	var sb strings.Builder

	for i := uint8(1); i <= 16; i++ {
		if service, exists := cachedServices[i]; exists {
			sb.WriteString(fmt.Sprintf("%d. %s\n", service.ID, service.Name))
		}
	}

	return sb.String()
}

// FindBestMatchInCSV busca o melhor match no CSV (busca simples por palavras-chave)
func FindBestMatchInCSV(userIntent string) *Intent {
	userLower := strings.ToLower(userIntent)
	
	var bestMatch *Intent
	maxScore := 0

	for i := range cachedIntents {
		intent := &cachedIntents[i]
		intentLower := strings.ToLower(intent.Intent)
		
		// Contar palavras em comum
		score := 0
		words := strings.Fields(userLower)
		for _, word := range words {
			if len(word) > 2 && strings.Contains(intentLower, word) {
				score++
			}
		}

		if score > maxScore {
			maxScore = score
			bestMatch = intent
		}
	}

	// Se encontrou um match razoável, retorna
	if maxScore >= 2 {
		return bestMatch
	}

	return nil
}
