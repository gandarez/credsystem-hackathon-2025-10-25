package main

import (
	"fmt"

	"github.com/credsystem/hackathon/knn/nlp"
)

// KNNService encapsula o pipeline NLP e fornece métodos de alto nível
type KNNService struct {
	pipeline   *nlp.Pipeline
	intents    []Intent
	serviceMap map[int]string
}

// NewKNNService cria um novo serviço KNN com o pipeline NLP
func NewKNNService() (*KNNService, error) {
	// Criar pipeline NLP otimizado para português
	pipeline, err := nlp.NewPipeline("portuguese", true)
	if err != nil {
		return nil, fmt.Errorf("failed to create NLP pipeline: %w", err)
	}

	return &KNNService{
		pipeline:   pipeline,
		intents:    make([]Intent, 0),
		serviceMap: make(map[int]string),
	}, nil
}

// LoadIntents carrega e treina o modelo com as intents do CSV
func (s *KNNService) LoadIntents(intents []Intent) error {
	if len(intents) == 0 {
		return fmt.Errorf("no intents provided")
	}

	// Criar mapa de serviços para lookup rápido
	for _, intent := range intents {
		s.serviceMap[intent.ServiceID] = intent.ServiceName
	}

	// Extrair textos e categorias para treinar o pipeline
	documents := make([]string, len(intents))
	categories := make([]string, len(intents))

	for i, intent := range intents {
		documents[i] = intent.IntentText
		// Usar o ServiceID como categoria (convertido para string)
		categories[i] = fmt.Sprintf("%d", intent.ServiceID)
	}

	// Treinar o pipeline NLP
	if err := s.pipeline.Train(documents, categories); err != nil {
		return fmt.Errorf("failed to train pipeline: %w", err)
	}

	// Após o treino, copiar os vetores do pipeline para os intents
	s.intents = make([]Intent, len(intents))
	for i := range intents {
		s.intents[i] = Intent{
			ServiceID:   intents[i].ServiceID,
			ServiceName: intents[i].ServiceName,
			IntentText:  intents[i].IntentText,
			Vector:      s.pipeline.IntentVectors[i].Vector,
		}
	}

	return nil
}

// Classify classifica uma intenção do usuário
func (s *KNNService) Classify(intentText string) ClassificationResult {
	// Predição usando o pipeline NLP
	match, confidence, err := s.pipeline.Predict(intentText)
	if err != nil {
		// Em caso de erro, retornar resultado vazio com confiança zero
		return ClassificationResult{
			ServiceID:   0,
			ServiceName: "",
			Confidence:  0.0,
		}
	}

	// Converter a categoria (ServiceID em string) de volta para int
	var serviceID int
	fmt.Sscanf(match.Category, "%d", &serviceID)

	// Buscar o nome do serviço
	serviceName := s.serviceMap[serviceID]

	return ClassificationResult{
		ServiceID:   serviceID,
		ServiceName: serviceName,
		Confidence:  confidence,
	}
}

// ClassifyTopK retorna os K melhores matches
func (s *KNNService) ClassifyTopK(intentText string, k int) []ClassificationResult {
	matches, confidences, err := s.pipeline.PredictTopK(intentText, k)
	if err != nil {
		return []ClassificationResult{}
	}

	results := make([]ClassificationResult, len(matches))
	for i, match := range matches {
		var serviceID int
		fmt.Sscanf(match.Category, "%d", &serviceID)

		results[i] = ClassificationResult{
			ServiceID:   serviceID,
			ServiceName: s.serviceMap[serviceID],
			Confidence:  confidences[i],
		}
	}

	return results
}

// VocabularySize retorna o tamanho do vocabulário treinado
func (s *KNNService) VocabularySize() int {
	return s.pipeline.Vectorizer.VocabularySize()
}

// ClassifyWithSafetyCheck classifica uma intenção e indica se o resultado é confiável.
// Esta é a interface pública que primeiro transforma o texto em vetor e então aplica a verificação de segurança.
func (s *KNNService) ClassifyWithSafetyCheck(intentText string) (predictedID int, predictedName string, confidence float64, isSafe bool, err error) {
	// Pré-processar a intenção
	processed := s.pipeline.Preprocessor.Process(intentText)

	// Transformar em vetor
	vector, err := s.pipeline.Vectorizer.Transform(processed)
	if err != nil {
		return 0, "", 0.0, false, fmt.Errorf("error transforming intent: %w", err)
	}

	// Aplicar a classificação com verificação de segurança
	predictedID, predictedName, confidence, isSafe = s.classifyLocallyWithSafetyCheck(vector)

	return predictedID, predictedName, confidence, isSafe, nil
}

// classifyLocallyWithSafetyCheck implementa uma regra de decisão avançada para prevenir ambiguidades.
// Retorna a melhor predição e um booleano `isSafe` que indica se o resultado passou pelos critérios de segurança:
// 1. CRITÉRIO DE CONFIANÇA MÍNIMA: A melhor correspondência deve ter uma pontuação acima do threshold mínimo
// 2. CRITÉRIO DE MARGEM DE AMBIGUIDADE: A diferença entre a melhor e a segunda melhor deve ser significativa
func (s *KNNService) classifyLocallyWithSafetyCheck(newIntentVector []float64) (predictedID int, predictedName string, confidence float64, isSafe bool) {
	// Constantes finais definidas com base na análise de todos os ciclos de teste.
	// Elas são otimizadas para maximizar a precisão e evitar os erros de -50 pontos.
	const confidenceThreshold = 0.55
	const ambiguityMargin = 0.25

	// Inicializar as duas melhores correspondências
	var bestMatch struct {
		serviceID   int
		serviceName string
		confidence  float64
	}

	var secondBestMatch struct {
		confidence float64
	}

	// Inicializar com valores que garantem que qualquer correspondência válida seja melhor
	bestMatch.confidence = -1.0
	secondBestMatch.confidence = -1.0

	// Iterar por todas as intenções pré-carregadas
	for _, intent := range s.intents {
		// Calcular a similaridade de cossenos entre o vetor da nova intenção e o vetor da intenção atual
		similarity, err := nlp.CosineSimilarity(newIntentVector, intent.Vector)
		if err != nil {
			// Em caso de erro no cálculo, continuar para a próxima intenção
			continue
		}

		// Atualizar as duas melhores correspondências
		if similarity > bestMatch.confidence {
			// A similaridade atual é melhor que a melhor correspondência
			// A antiga melhor correspondência se torna a segunda melhor
			secondBestMatch.confidence = bestMatch.confidence

			// Atualizar a melhor correspondência
			bestMatch.serviceID = intent.ServiceID
			bestMatch.serviceName = intent.ServiceName
			bestMatch.confidence = similarity
		} else if similarity > secondBestMatch.confidence {
			// A similaridade atual não é a melhor, mas é melhor que a segunda melhor
			secondBestMatch.confidence = similarity
		}
	}

	// Aplicar a regra de decisão final usando as constantes definidas
	// Um resultado é considerado "seguro" se:
	// 1. A confiança da melhor correspondência está acima do threshold mínimo
	// 2. A margem entre a melhor e a segunda melhor é suficientemente grande

	confidenceCheckPassed := bestMatch.confidence >= confidenceThreshold
	ambiguityCheckPassed := (bestMatch.confidence - secondBestMatch.confidence) >= ambiguityMargin

	isSafe = confidenceCheckPassed && ambiguityCheckPassed

	// Retornar a melhor predição encontrada e o indicador de segurança
	return bestMatch.serviceID, bestMatch.serviceName, bestMatch.confidence, isSafe
}
