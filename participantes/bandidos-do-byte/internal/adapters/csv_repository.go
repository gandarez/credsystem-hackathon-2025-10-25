package adapters

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bandidos_do_byte/api/internal/domain"
)

type CSVTrainingRepository struct {
	filePath string
}

func NewCSVTrainingRepository(filePath string) *CSVTrainingRepository {
	return &CSVTrainingRepository{
		filePath: filePath,
	}
}

// findCSVFile tenta encontrar o arquivo CSV em vários locais possíveis
func (r *CSVTrainingRepository) findCSVFile() (string, error) {
	// Lista de caminhos possíveis para tentar
	possiblePaths := []string{
		r.filePath,                              // Caminho fornecido
		"training/intents_pre_loaded.csv",       // Relativo ao diretório atual
		"./training/intents_pre_loaded.csv",     // Relativo explícito
		"../training/intents_pre_loaded.csv",    // Um nível acima
		"../../training/intents_pre_loaded.csv", // Dois níveis acima
	}

	// Adiciona caminho relativo ao executável
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		possiblePaths = append(possiblePaths,
			filepath.Join(exeDir, "training", "intents_pre_loaded.csv"),
			filepath.Join(exeDir, "..", "training", "intents_pre_loaded.csv"),
			filepath.Join(exeDir, "..", "..", "training", "intents_pre_loaded.csv"),
		)
	}

	// Tenta cada caminho
	for _, path := range possiblePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("CSV file not found in any of the expected locations: %v", possiblePaths)
}

func (r *CSVTrainingRepository) LoadIntentExamples() ([]domain.IntentExample, error) {
	csvPath, err := r.findCSVFile()
	if err != nil {
		// Log para debug
		wd, _ := os.Getwd()
		return nil, fmt.Errorf("%w (current working directory: %s)", err, wd)
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file at %s: %w", csvPath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = 3

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var examples []domain.IntentExample
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		serviceID, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			continue
		}

		examples = append(examples, domain.IntentExample{
			ServiceID:   serviceID,
			ServiceName: strings.TrimSpace(record[1]),
			Intent:      strings.TrimSpace(record[2]),
		})
	}

	if len(examples) == 0 {
		return nil, fmt.Errorf("no intent examples loaded from CSV")
	}

	return examples, nil
}
