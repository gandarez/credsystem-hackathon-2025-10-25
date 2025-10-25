package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// loadIntentsFromCSV carrega as intenções do arquivo CSV
func loadIntentsFromCSV(filePath string) ([]Intent, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'             // CSV usa ponto e vírgula como delimitador
	reader.FieldsPerRecord = -1    // Permite número variável de campos
	reader.TrimLeadingSpace = true // Remove espaços em branco

	// Ler todas as linhas
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Pular o cabeçalho (primeira linha) e linhas vazias
	records = records[1:]

	// Filtrar linhas vazias
	var validRecords [][]string
	for _, record := range records {
		// Pular linhas vazias ou com apenas um campo vazio
		if len(record) == 0 || (len(record) == 1 && record[0] == "") {
			continue
		}
		validRecords = append(validRecords, record)
	}
	records = validRecords

	intents := make([]Intent, 0, len(records))
	for i, record := range records {
		if len(record) < 3 {
			continue // Pular linhas inválidas
		}

		// Parsear ServiceID
		serviceID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("invalid service_id at line %d: %w", i+2, err)
		}

		intent := Intent{
			ServiceID:   serviceID,
			ServiceName: record[1],
			IntentText:  record[2],
		}

		intents = append(intents, intent)
	}

	return intents, nil
}
