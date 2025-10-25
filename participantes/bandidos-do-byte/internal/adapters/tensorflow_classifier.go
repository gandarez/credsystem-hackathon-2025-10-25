package adapters

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/bandidos_do_byte/api/internal/domain"
)

// TensorFlowClassifier usa similaridade de texto com dados de treinamento para classificar intents
// Nota: Esta é uma implementação simplificada que não requer TensorFlow
// Para usar o modelo .h5 real, seria necessário um servidor Python ou biblioteca C
type TensorFlowClassifier struct {
	modelPath      string
	serviceMapping map[int]string
	tokenizer      *SimpleTokenizer
	trainingData   []TrainingExample
	idf            map[string]float64 // Inverse Document Frequency
	stopWords      map[string]bool
}

// TrainingExample representa um exemplo de treinamento
type TrainingExample struct {
	ServiceID   int
	ServiceName string
	Intent      string
	Tokens      map[string]float64
}

// SimpleTokenizer é um tokenizer básico para processar texto
type SimpleTokenizer struct {
	wordIndex   map[string]int
	maxWords    int
	maxSequence int
}

// NewTensorFlowClassifier cria uma nova instância do classificador TensorFlow
func NewTensorFlowClassifier(modelPath, _ string) *TensorFlowClassifier {
	classifier := &TensorFlowClassifier{
		modelPath:      modelPath,
		serviceMapping: initializeServiceMapping(),
		tokenizer:      NewSimpleTokenizer(10000, 50),
		trainingData:   []TrainingExample{},
		idf:            make(map[string]float64),
		stopWords:      initializeStopWords(),
	}

	// Carrega dados de treinamento
	if err := classifier.loadTrainingData(); err != nil {
		fmt.Printf("Warning: Failed to load training data: %v\n", err)
	}

	// Calcula IDF após carregar dados
	classifier.calculateIDF()

	return classifier
}

// initializeStopWords retorna palavras comuns que devem ser ignoradas
func initializeStopWords() map[string]bool {
	words := []string{
		"de", "a", "o", "que", "e", "do", "da", "em", "um", "para",
		"é", "com", "não", "uma", "os", "no", "se", "na", "por", "mais",
		"as", "dos", "como", "mas", "ao", "ele", "das", "à", "seu", "sua",
		"ou", "quando", "muito", "nos", "já", "eu", "também", "só", "pelo",
		"pela", "até", "isso", "ela", "entre", "depois", "sem", "mesmo",
		"aos", "seus", "quem", "nas", "me", "esse", "eles", "você", "essa",
		"num", "nem", "suas", "meu", "às", "minha", "numa", "pelos", "elas",
	}

	stopWordsMap := make(map[string]bool)
	for _, word := range words {
		stopWordsMap[word] = true
	}
	return stopWordsMap
}

// loadTrainingData carrega os dados de treinamento do CSV
func (c *TensorFlowClassifier) loadTrainingData() error {
	csvPath := c.getTrainingDataPath()

	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open training data: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = 3

	// Skip header
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Lê todos os exemplos
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV records: %w", err)
	}

	for _, record := range records {
		if len(record) < 3 {
			continue
		}

		var serviceID int
		fmt.Sscanf(strings.TrimSpace(record[0]), "%d", &serviceID)

		intent := strings.TrimSpace(record[2])
		tokens := c.tokenize(intent)

		example := TrainingExample{
			ServiceID:   serviceID,
			ServiceName: strings.TrimSpace(record[1]),
			Intent:      intent,
			Tokens:      tokens,
		}

		c.trainingData = append(c.trainingData, example)
	}

	return nil
}

// tokenize converte texto em bag of words com TF-IDF
func (c *TensorFlowClassifier) tokenize(text string) map[string]float64 {
	tokens := make(map[string]float64)
	words := strings.Fields(strings.ToLower(text))

	for _, word := range words {
		// Remove pontuação
		word = strings.Trim(word, ".,!?;:\"'()[]{}*/-+=_")

		// Remove stopwords
		if len(word) < 2 || c.stopWords[word] {
			continue
		}

		tokens[word]++
	}

	// Aplica TF-IDF
	for word := range tokens {
		tf := tokens[word]
		idf := c.idf[word]
		if idf == 0 {
			idf = 1.0 // Palavra não vista no treinamento
		}
		tokens[word] = tf * idf
	}

	// Normaliza
	total := 0.0
	for _, val := range tokens {
		total += val * val
	}
	if total > 0 {
		total = math.Sqrt(total)
		for word := range tokens {
			tokens[word] /= total
		}
	}

	return tokens
}

// calculateIDF calcula o Inverse Document Frequency para cada palavra
func (c *TensorFlowClassifier) calculateIDF() {
	if len(c.trainingData) == 0 {
		return
	}

	// Conta em quantos documentos cada palavra aparece
	docFreq := make(map[string]int)
	totalDocs := len(c.trainingData)

	for _, example := range c.trainingData {
		seenWords := make(map[string]bool)
		words := strings.Fields(strings.ToLower(example.Intent))

		for _, word := range words {
			word = strings.Trim(word, ".,!?;:\"'()[]{}*/-+=_")
			if len(word) < 2 || c.stopWords[word] {
				continue
			}
			if !seenWords[word] {
				docFreq[word]++
				seenWords[word] = true
			}
		}
	}

	// Calcula IDF
	for word, freq := range docFreq {
		c.idf[word] = math.Log(float64(totalDocs) / float64(freq))
	}
}

// getTrainingDataPath retorna o caminho do arquivo de treinamento
func (c *TensorFlowClassifier) getTrainingDataPath() string {
	return filepath.Join(filepath.Dir(c.modelPath), "intents_pre_loaded.csv")
}

// initializeServiceMapping mapeia IDs de serviço para nomes
func initializeServiceMapping() map[int]string {
	return map[int]string{
		0:  "Contate a URA",
		1:  "Abertura de Conta",
		2:  "Empréstimo Pessoal",
		3:  "Cartão de Crédito",
		4:  "Investimentos",
		5:  "Seguros",
		6:  "Consórcio",
		7:  "Financiamento Imobiliário",
		8:  "Financiamento de Veículos",
		9:  "Previdência Privada",
		10: "Conta Digital",
		11: "Portabilidade de Salário",
		12: "Renegociação de Dívidas",
		13: "Antecipação de FGTS",
		14: "Crédito Consignado",
		15: "Conta para Empresas",
		16: "Suporte Técnico",
	}
}

// ClassifyIntent classifica a intenção do usuário usando similaridade de texto com TF-IDF
func (c *TensorFlowClassifier) ClassifyIntent(request domain.IntentClassificationRequest) (*domain.IntentClassificationResponse, error) {
	if len(c.trainingData) == 0 {
		return nil, fmt.Errorf("training data not loaded")
	}

	// Tokeniza o texto de entrada
	inputTokens := c.tokenize(request.UserIntent)

	if len(inputTokens) == 0 {
		return nil, fmt.Errorf("could not tokenize input text")
	}

	// Calcula similaridade com cada exemplo e agrupa por serviço
	type ServiceScore struct {
		ServiceID   int
		ServiceName string
		Scores      []float64
		BestScore   float64
		AvgScore    float64
	}

	serviceScores := make(map[int]*ServiceScore)

	for _, example := range c.trainingData {
		similarity := c.cosineSimilarity(inputTokens, example.Tokens)

		// Só considera se similaridade for significativa
		if similarity < 0.1 {
			continue
		}

		if score, ok := serviceScores[example.ServiceID]; ok {
			score.Scores = append(score.Scores, similarity)
			if similarity > score.BestScore {
				score.BestScore = similarity
			}
		} else {
			serviceScores[example.ServiceID] = &ServiceScore{
				ServiceID:   example.ServiceID,
				ServiceName: example.ServiceName,
				Scores:      []float64{similarity},
				BestScore:   similarity,
			}
		}
	}

	if len(serviceScores) == 0 {
		return nil, domain.ErrNoServiceFound
	}

	// Calcula média ponderada (dá mais peso aos melhores scores)
	for _, score := range serviceScores {
		if len(score.Scores) == 0 {
			continue
		}

		// Ordena scores em ordem decrescente
		scores := score.Scores
		for i := 0; i < len(scores); i++ {
			for j := i + 1; j < len(scores); j++ {
				if scores[j] > scores[i] {
					scores[i], scores[j] = scores[j], scores[i]
				}
			}
		}

		// Média ponderada: top 3 scores com pesos 50%, 30%, 20%
		sum := 0.0
		if len(scores) >= 1 {
			sum += scores[0] * 0.5
		}
		if len(scores) >= 2 {
			sum += scores[1] * 0.3
		}
		if len(scores) >= 3 {
			sum += scores[2] * 0.2
		} else if len(scores) == 2 {
			// Se só tem 2, ajusta os pesos
			sum = scores[0]*0.6 + scores[1]*0.4
		}

		score.AvgScore = sum
	}

	// Encontra o serviço com maior pontuação média
	var bestService *ServiceScore
	for _, score := range serviceScores {
		if bestService == nil || score.AvgScore > bestService.AvgScore {
			bestService = score
		}
	}

	if bestService == nil {
		return nil, domain.ErrNoServiceFound
	}

	// Threshold mínimo de confiança - retorna erro se muito baixo
	if bestService.AvgScore < 0.20 {
		return nil, domain.ErrNoServiceFound
	}

	return &domain.IntentClassificationResponse{
		ServiceID:   bestService.ServiceID,
		ServiceName: bestService.ServiceName,
		Confidence:  bestService.AvgScore,
	}, nil
}

// cosineSimilarity calcula a similaridade de cosseno entre dois vetores
func (c *TensorFlowClassifier) cosineSimilarity(a, b map[string]float64) float64 {
	dotProduct := 0.0
	magA := 0.0
	magB := 0.0

	// Calcula produto escalar e magnitudes
	for word, valA := range a {
		magA += valA * valA
		if valB, ok := b[word]; ok {
			dotProduct += valA * valB
		}
	}

	for _, valB := range b {
		magB += valB * valB
	}

	magA = math.Sqrt(magA)
	magB = math.Sqrt(magB)

	if magA == 0 || magB == 0 {
		return 0
	}

	return dotProduct / (magA * magB)
}

// HealthCheck verifica se os dados de treinamento estão carregados
func (c *TensorFlowClassifier) HealthCheck() error {
	if len(c.trainingData) == 0 {
		return fmt.Errorf("training data not loaded")
	}

	return nil
}

// NewSimpleTokenizer cria um novo tokenizer (não usado nesta implementação)
func NewSimpleTokenizer(maxWords, maxSequence int) *SimpleTokenizer {
	return &SimpleTokenizer{
		wordIndex:   make(map[string]int),
		maxWords:    maxWords,
		maxSequence: maxSequence,
	}
}
