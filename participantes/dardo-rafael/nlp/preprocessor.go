package nlp

import (
	"strings"
	"unicode"

	"github.com/bbalet/stopwords"
)

// Preprocessor handles text preprocessing operations like stemming and stopword removal.
type Preprocessor struct {
	lang string
}

// NewPreprocessor creates a new preprocessor for the specified language.
// Supported languages: "portuguese", "english", etc.
func NewPreprocessor(lang string) (*Preprocessor, error) {
	return &Preprocessor{
		lang: mapLanguage(lang),
	}, nil
}

// mapLanguage maps full language names to stopwords package codes.
func mapLanguage(lang string) string {
	mapping := map[string]string{
		"portuguese": "pt",
		"english":    "en",
		"spanish":    "es",
	}

	if code, ok := mapping[lang]; ok {
		return code
	}
	return "pt" // default to Portuguese
}

// RemoveStopwords removes common words that don't add semantic value.
// It also removes punctuation if removePunctuation is true.
func (p *Preprocessor) RemoveStopwords(text string, removePunctuation bool) string {
	return stopwords.CleanString(text, p.lang, removePunctuation)
}

// stemPortuguese applies a simplified Portuguese stemming algorithm.
// It removes common suffixes to reduce words to their base form.
func stemPortuguese(word string) string {
	// List of common Portuguese suffixes ordered by length (longest first)
	suffixes := []string{
		"amentos", "imentos", "amento", "imento", "adora", "ância",
		"ância", "ências", "amento", "imento", "adora", "antes",
		"ância", "ência", "adora", "mente", "idade", "eiras",
		"ador", "ante", "ível", "eira", "osos", "osas", "ação",
		"ções", "ente", "ista", "ezas", "eza", "ica", "ico",
		"ada", "ado", "ida", "ido", "ura", "ura", "ara",
		"ira", "ava", "iam", "ado", "ido", "oso", "osa",
		"osa", "oso", "ção", "são", "vel", "eis", "ais",
		"amos", "emos", "imos", "emos", "emos", "ia",
		"as", "es", "is", "os", "us", "a", "e", "i", "o", "u",
	}

	// Minimum stem length to avoid over-stemming
	minStemLen := 3

	word = strings.ToLower(word)

	// Try to remove suffixes
	for _, suffix := range suffixes {
		if strings.HasSuffix(word, suffix) {
			stem := word[:len(word)-len(suffix)]
			// Only apply if the resulting stem is meaningful
			if len(stem) >= minStemLen {
				return stem
			}
		}
	}

	return word
}

// Stem reduces words to their root form using a simplified stemming algorithm.
// Example: "pagamento" -> "pag", "correção" -> "corre"
func (p *Preprocessor) Stem(text string) string {
	words := strings.Fields(text)
	stemmedWords := make([]string, 0, len(words))

	for _, word := range words {
		if word == "" {
			continue
		}

		// Remove any non-letter characters
		cleaned := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) {
				return unicode.ToLower(r)
			}
			return -1
		}, word)

		if cleaned == "" {
			continue
		}

		stemmed := stemPortuguese(cleaned)
		stemmedWords = append(stemmedWords, stemmed)
	}

	return strings.Join(stemmedWords, " ")
}

// Process applies the complete preprocessing pipeline:
// 1. Convert to lowercase
// 2. Remove stopwords
// 3. Apply stemming
func (p *Preprocessor) Process(text string) string {
	// 1. Normalize to lowercase
	text = strings.ToLower(text)

	// 2. Remove stopwords and punctuation
	text = p.RemoveStopwords(text, true)

	// 3. Apply stemming
	text = p.Stem(text)

	return text
}

// ProcessBatch processes multiple texts efficiently.
func (p *Preprocessor) ProcessBatch(texts []string) []string {
	processed := make([]string, len(texts))
	for i, text := range texts {
		processed[i] = p.Process(text)
	}
	return processed
}
