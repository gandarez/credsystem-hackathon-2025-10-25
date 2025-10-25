package main

import (
	"strings"

	"participantes/gurus-das-rotinas/client/openrouter"
)

// EnhancedKeywordClassifier provides sophisticated keyword-based classification
type EnhancedKeywordClassifier struct {
	servicePatterns map[uint8][]string
	serviceWeights  map[uint8]map[string]int
}

// NewEnhancedKeywordClassifier creates a new enhanced classifier
func NewEnhancedKeywordClassifier() *EnhancedKeywordClassifier {
	classifier := &EnhancedKeywordClassifier{
		servicePatterns: make(map[uint8][]string),
		serviceWeights:  make(map[uint8]map[string]int),
	}

	// Load embedded patterns
	classifier.loadDefaultPatterns()

	return classifier
}

// loadDefaultPatterns loads comprehensive default patterns
func (ekc *EnhancedKeywordClassifier) loadDefaultPatterns() {
	// Initialize weight maps
	for i := uint8(1); i <= 16; i++ {
		ekc.serviceWeights[i] = make(map[string]int)
	}

	// Comprehensive patterns based on the expanded dataset
	patterns := map[uint8][]string{
		1: {
			"limite", "vencimento", "melhor dia", "disponível", "quanto", "saldo",
			"fatura", "ciclo", "atualizou", "consulta limite", "valor gastar",
			"quanto tem", "quando fecha", "quando vence", "quando posso",
		},
		2: {
			"boleto", "segunda via", "boleto acordo", "código barras", "negociação",
			"pagar acordo", "segunda via boleto",
		},
		3: {
			"fatura", "segunda via fatura", "fatura cartão", "extrato",
		},
		4: {
			"entrega", "cartão entrega", "status entrega", "quando chega",
			"onde está", "rastreamento",
		},
		5: {
			"status cartão", "cartão status", "situação cartão", "ativo",
			"bloqueado", "funcionando",
		},
		6: {
			"aumento", "limite", "solicitar", "mais limite", "ampliar",
			"elevar limite", "incrementar",
		},
		7: {
			"cancelar", "cancelamento", "bloquear", "parar", "encerrar",
			"desativar cartão",
		},
		8: {
			"seguradora", "telefone", "seguro", "contato", "número",
			"ligar seguradora",
		},
		9: {
			"desbloqueio", "desbloquear", "liberar", "ativar", "habilitar",
		},
		10: {
			"senha", "esqueci", "trocar", "alterar", "mudar", "nova senha",
			"resetar senha", "recuperar senha",
		},
		11: {
			"perdi", "roubo", "perda", "furtado", "roubaram", "sumiu",
			"desapareceu", "não encontro",
		},
		12: {
			"saldo", "consulta saldo", "quanto tenho", "valor disponível",
			"dinheiro", "conta corrente",
		},
		13: {
			"pagamento", "pagar", "conta", "boleto", "transferência",
			"débito", "pix",
		},
		14: {
			"reclamação", "reclamar", "problema", "erro", "defeito",
			"não funciona", "complicado",
		},
		15: {
			"humano", "atendente", "pessoa", "operador", "falar com",
			"atendimento pessoal", "suporte humano",
		},
		16: {
			"token", "proposta", "aprovação", "solicitação", "pedido",
			"processo", "análise",
		},
	}

	for serviceID, keywords := range patterns {
		ekc.servicePatterns[serviceID] = keywords
		for _, keyword := range keywords {
			ekc.serviceWeights[serviceID][keyword] = 3 // Default weight
		}
	}
}

// extractKeywords extracts meaningful keywords from text
func (ekc *EnhancedKeywordClassifier) extractKeywords(text string) []string {
	// Remove common stop words
	stopWords := map[string]bool{
		"o": true, "a": true, "os": true, "as": true, "um": true, "uma": true,
		"de": true, "da": true, "do": true, "das": true, "dos": true,
		"em": true, "na": true, "no": true, "nas": true, "nos": true,
		"para": true, "com": true, "por": true, "sobre": true,
		"meu": true, "minha": true, "meus": true, "minhas": true,
		"que": true, "qual": true, "quando": true, "onde": true,
		"como": true, "porque": true, "por que": true,
	}

	words := strings.Fields(text)
	var keywords []string

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// isOutOfContext checks if the intent is completely unrelated to banking/financial services
func (ekc *EnhancedKeywordClassifier) isOutOfContext(intent string, keywords []string) bool {
	// Define out-of-context keywords that indicate non-banking requests
	outOfContextKeywords := map[string]bool{
		// Food and drinks
		"water": true, "food": true, "eat": true, "drink": true, "hungry": true, "thirsty": true,
		"coffee": true, "tea": true, "pizza": true, "burger": true, "restaurant": true,

		// Weather
		"weather": true, "rain": true, "sunny": true, "cold": true, "hot": true, "temperature": true,

		// Transportation
		"taxi": true, "uber": true, "bus": true, "train": true, "flight": true, "airplane": true,
		"car": true, "drive": true, "gas": true, "fuel": true,

		// Entertainment
		"movie": true, "cinema": true, "theater": true, "music": true, "song": true, "game": true,
		"play": true, "fun": true, "party": true,

		// Health and medical
		"doctor": true, "hospital": true, "medicine": true, "sick": true, "pain": true, "health": true,
		"medical": true, "pharmacy": true,

		// Shopping (non-financial)
		"shopping": true, "buy": true, "store": true, "mall": true, "clothes": true, "shoes": true,

		// Technology (non-financial)
		"computer": true, "phone": true, "internet": true, "wifi": true, "password": true, "email": true,

		// General out-of-context phrases
		"hello": true, "hi": true, "goodbye": true, "bye": true, "thanks": true, "thank you": true,
		"help": true, "information": true, "question": true, "problem": true,
	}

	// Check if any keyword is clearly out of context
	for _, keyword := range keywords {
		if outOfContextKeywords[keyword] {
			return true
		}
	}

	// Check for banking-related keywords to ensure it's in context
	bankingKeywords := map[string]bool{
		"cartão": true, "card": true, "conta": true, "account": true, "banco": true, "bank": true,
		"dinheiro": true, "money": true, "saldo": true, "balance": true, "limite": true, "limit": true,
		"fatura": true, "bill": true, "boleto": true, "payment": true, "pagamento": true, "pay": true,
		"senha": true, "password": true, "pin": true, "código": true, "code": true,
		"bloqueio": true, "block": true, "desbloqueio": true, "unblock": true,
		"cancelamento": true, "cancel": true, "cancelar": true,
		"aumento": true, "increase": true,
		"entrega": true, "delivery": true, "status": true, "situação": true,
		"seguradora": true, "insurance": true, "seguro": true,
		"reclamação": true, "complaint": true, "problema": true, "issue": true,
		"atendimento": true, "service": true, "suporte": true, "support": true,
		"token": true, "proposta": true, "proposal": true,
		"perda": true, "loss": true, "roubo": true, "theft": true, "furtado": true,
		"negociação": true, "negotiation": true, "acordo": true, "agreement": true,
	}

	// If no banking keywords are found, it's likely out of context
	hasBankingKeyword := false
	for _, keyword := range keywords {
		if bankingKeywords[keyword] {
			hasBankingKeyword = true
			break
		}
	}

	// Also check the full intent for banking-related terms
	for bankingTerm := range bankingKeywords {
		if strings.Contains(intent, bankingTerm) {
			hasBankingKeyword = true
			break
		}
	}

	// If it's a very short intent with no banking keywords, it's likely out of context
	if len(keywords) <= 2 && !hasBankingKeyword {
		return true
	}

	// If no banking keywords are found at all, it's out of context
	if !hasBankingKeyword {
		return true
	}

	return false
}

// ClassifyWithScore classifies intent and returns confidence score
func (ekc *EnhancedKeywordClassifier) ClassifyWithScore(intent string) (*openrouter.DataResponse, float64) {
	intent = strings.ToLower(intent)
	inputKeywords := ekc.extractKeywords(intent)

	// Check if intent is completely out of context
	if ekc.isOutOfContext(intent, inputKeywords) {
		return nil, 0.0
	}

	var bestServiceID uint8
	var bestScore float64

	// Calculate scores for each service
	for serviceID := uint8(1); serviceID <= 16; serviceID++ {
		score := ekc.calculateScore(inputKeywords, serviceID)
		if score > bestScore {
			bestScore = score
			bestServiceID = serviceID
		}
	}

	// Only return if confidence is above threshold
	if bestScore > 0.3 {
		return &openrouter.DataResponse{
			ServiceID:   bestServiceID,
			ServiceName: ServiceMap[bestServiceID],
		}, bestScore
	}

	return nil, bestScore
}

// calculateScore calculates the confidence score for a service
func (ekc *EnhancedKeywordClassifier) calculateScore(keywords []string, serviceID uint8) float64 {
	totalWeight := 0
	matchedWeight := 0

	weights := ekc.serviceWeights[serviceID]
	if weights == nil {
		return 0
	}

	// Calculate total possible weight
	for _, weight := range weights {
		totalWeight += weight
	}

	// Calculate matched weight
	for _, keyword := range keywords {
		if weight, exists := weights[keyword]; exists {
			matchedWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return float64(matchedWeight) / float64(totalWeight)
}
