package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bandidos_do_byte/api/internal/adapters"
	"github.com/bandidos_do_byte/api/internal/config"
	"github.com/bandidos_do_byte/api/internal/domain"
	"github.com/bandidos_do_byte/api/internal/ports"
)

type TestResult struct {
	Intent        string
	ExpectedID    int
	ExpectedName  string
	PredictedID   int
	PredictedName string
	Confidence    float64
	Success       bool
	NotFound      bool
}

func main() {
	// Carrega variáveis de ambiente do .env se existir
	loadEnvFile(".env")

	fmt.Println("=== Teste de Acurácia do Classificador ===")
	fmt.Println()

	// Carrega configuração
	cfg := config.NewConfig()

	// Determina qual classificador usar
	classifierType := cfg.ClassifierType
	if classifierType == "" {
		classifierType = config.ClassifierOpenRouter
	}

	fmt.Printf("Classificador: %s\n", classifierType)
	fmt.Println()

	// Cria o classificador apropriado
	var classifier ports.IntentClassifier
	switch classifierType {
	case config.ClassifierTensorFlow:
		classifier = adapters.NewTensorFlowClassifier(cfg.TensorFlowModelPath, "")
	case config.ClassifierOpenRouter:
		if cfg.OpenRouterAPIKey == "" {
			fmt.Println("❌ ERRO: OPENROUTER_API_KEY não configurada")
			os.Exit(1)
		}
		classifier = adapters.NewOpenRouterClient(cfg.OpenRouterAPIKey)
	default:
		fmt.Printf("❌ ERRO: Tipo de classificador desconhecido: %s\n", classifierType)
		os.Exit(1)
	}

	// Carrega dados de teste
	testData, err := loadTestData(cfg.TrainingDataPath)
	if err != nil {
		fmt.Printf("❌ ERRO ao carregar dados: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total de exemplos: %d\n", len(testData))

	// Converte dados de teste em exemplos para o classificador
	examples := make([]domain.IntentExample, len(testData))
	for i, data := range testData {
		examples[i] = domain.IntentExample{
			ServiceID:   data.ServiceID,
			ServiceName: data.ServiceName,
			Intent:      data.Intent,
		}
	}

	// Testa todos os cenários
	fmt.Printf("Testando todos os %d cenários\n", len(testData))
	fmt.Println()

	// Executa testes
	results := runTests(classifier, testData, examples)

	// Mostra resultados
	printResults(results)
}

// IntentData representa uma linha do CSV
type IntentData struct {
	ServiceID   int
	ServiceName string
	Intent      string
}

func loadTestData(csvPath string) ([]IntentData, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = 3

	// Pula header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("erro ao ler header: %w", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler CSV: %w", err)
	}

	var data []IntentData
	for _, record := range records {
		if len(record) < 3 {
			continue
		}

		serviceID, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			continue
		}

		data = append(data, IntentData{
			ServiceID:   serviceID,
			ServiceName: strings.TrimSpace(record[1]),
			Intent:      strings.TrimSpace(record[2]),
		})
	}

	return data, nil
}

func runTests(classifier ports.IntentClassifier, samples []IntentData, examples []domain.IntentExample) []TestResult {
	results := make([]TestResult, 0, len(samples))

	for i, sample := range samples {
		fmt.Printf("\r[%d/%d] Testando...", i+1, len(samples))

		req := domain.IntentClassificationRequest{
			UserIntent: sample.Intent,
			Examples:   examples,
		}

		response, err := classifier.ClassifyIntent(req)

		result := TestResult{
			Intent:       sample.Intent,
			ExpectedID:   sample.ServiceID,
			ExpectedName: sample.ServiceName,
		}

		if err != nil {
			if err == domain.ErrNoServiceFound {
				result.NotFound = true
				result.Success = false
			} else {
				result.Success = false
				result.PredictedName = fmt.Sprintf("ERRO: %v", err)
			}
		} else {
			result.PredictedID = response.ServiceID
			result.PredictedName = response.ServiceName
			result.Confidence = response.Confidence
			result.Success = response.ServiceID == sample.ServiceID
		}

		results = append(results, result)

		// Delay para não sobrecarregar API (apenas para OpenRouter)
		if _, ok := classifier.(*adapters.OpenRouterClient); ok {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Println()
	return results
}

func printResults(results []TestResult) {
	fmt.Println()
	fmt.Println("=== Resultados ===")
	fmt.Println()

	correct := 0
	notFound := 0
	errors := 0
	failures := make([]TestResult, 0)

	for _, result := range results {
		if result.Success {
			correct++
		} else {
			if result.NotFound {
				notFound++
			} else if strings.Contains(result.PredictedName, "ERRO") {
				errors++
			}
			failures = append(failures, result)
		}
	}

	total := len(results)
	accuracy := float64(correct) / float64(total) * 100

	// Resumo
	fmt.Printf("Total de testes: %d\n", total)
	fmt.Printf("✅ Acertos: %d (%.2f%%)\n", correct, accuracy)
	fmt.Printf("❌ Erros: %d\n", len(failures))
	if notFound > 0 {
		fmt.Printf("   - Não encontrados: %d\n", notFound)
	}
	if errors > 0 {
		fmt.Printf("   - Erros de execução: %d\n", errors)
	}
	fmt.Println()

	// Mostra falhas se houver
	if len(failures) > 0 {
		fmt.Println("=== Detalhes das Falhas ===")
		fmt.Println()
		for i, failure := range failures {
			fmt.Printf("%d. Intent: \"%s\"\n", i+1, failure.Intent)
			fmt.Printf("   Esperado: [%d] %s\n", failure.ExpectedID, failure.ExpectedName)
			if failure.NotFound {
				fmt.Printf("   Obtido: ❌ Serviço não encontrado\n")
			} else if strings.Contains(failure.PredictedName, "ERRO") {
				fmt.Printf("   Obtido: %s\n", failure.PredictedName)
			} else {
				fmt.Printf("   Obtido: [%d] %s (Confiança: %.2f)\n",
					failure.PredictedID, failure.PredictedName, failure.Confidence)
			}
			fmt.Println()
		}
	}

	// Critério de sucesso
	fmt.Println("=== Avaliação ===")
	if accuracy >= 80.0 {
		fmt.Printf("✅ PASSOU - Acurácia de %.2f%% (meta: ≥80%%)\n", accuracy)
		os.Exit(0)
	} else if accuracy >= 70.0 {
		fmt.Printf("⚠️  ATENÇÃO - Acurácia de %.2f%% (meta: ≥80%%)\n", accuracy)
		os.Exit(0)
	} else {
		fmt.Printf("❌ FALHOU - Acurácia de %.2f%% (meta: ≥80%%)\n", accuracy)
		os.Exit(1)
	}
}

// loadEnvFile carrega variáveis de ambiente de um arquivo .env
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// Arquivo .env não é obrigatório
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignora linhas vazias e comentários
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Divide em chave=valor
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove aspas se houver
		value = strings.Trim(value, "\"'")

		// Define a variável de ambiente se ainda não estiver definida
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
