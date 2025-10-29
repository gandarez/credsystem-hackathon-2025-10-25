package nlp

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// TFIDFVectorizer implements TF-IDF (Term Frequency-Inverse Document Frequency) vectorization.
type TFIDFVectorizer struct {
	mu sync.RWMutex

	// vocabulary maps terms to their index in the vector
	vocabulary map[string]int

	// idf stores the inverse document frequency for each term
	idf map[string]float64

	// documentCount is the total number of documents
	documentCount int

	// normalized indicates if vectors should be normalized (for cosine similarity)
	normalized bool
}

// NewTFIDFVectorizer creates a new TF-IDF vectorizer.
func NewTFIDFVectorizer(normalized bool) *TFIDFVectorizer {
	return &TFIDFVectorizer{
		vocabulary:    make(map[string]int),
		idf:           make(map[string]float64),
		normalized:    normalized,
		documentCount: 0,
	}
}

// Fit trains the vectorizer on a corpus of documents.
func (v *TFIDFVectorizer) Fit(documents []string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if len(documents) == 0 {
		return fmt.Errorf("cannot fit on empty document set")
	}

	v.documentCount = len(documents)

	// Build vocabulary and count document frequencies
	termDocFreq := make(map[string]int)

	for _, doc := range documents {
		terms := strings.Fields(doc)
		seenInDoc := make(map[string]bool)

		for _, term := range terms {
			if term == "" {
				continue
			}

			// Add to vocabulary if new
			if _, exists := v.vocabulary[term]; !exists {
				v.vocabulary[term] = len(v.vocabulary)
			}

			// Count document frequency (each document counted once)
			if !seenInDoc[term] {
				termDocFreq[term]++
				seenInDoc[term] = true
			}
		}
	}

	// Calculate IDF for each term
	// IDF(t) = log(N / df(t)) where N is total docs and df(t) is doc frequency of term t
	for term, df := range termDocFreq {
		v.idf[term] = math.Log(float64(v.documentCount) / float64(df))
	}

	return nil
}

// Transform converts a document into a TF-IDF vector.
func (v *TFIDFVectorizer) Transform(document string) ([]float64, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if len(v.vocabulary) == 0 {
		return nil, fmt.Errorf("vectorizer must be fitted before transform")
	}

	// Initialize vector with zeros
	vector := make([]float64, len(v.vocabulary))

	// Calculate term frequencies in this document
	terms := strings.Fields(document)
	termFreq := make(map[string]int)

	for _, term := range terms {
		if term != "" {
			termFreq[term]++
		}
	}

	// Calculate TF-IDF for each term
	docLength := float64(len(terms))
	if docLength == 0 {
		return vector, nil
	}

	var norm float64

	for term, freq := range termFreq {
		if idx, exists := v.vocabulary[term]; exists {
			// TF = frequency / doc_length
			tf := float64(freq) / docLength

			// TF-IDF = TF * IDF
			tfidf := tf * v.idf[term]
			vector[idx] = tfidf

			if v.normalized {
				norm += tfidf * tfidf
			}
		}
	}

	// Normalize vector (for cosine similarity)
	if v.normalized && norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vector {
			vector[i] /= norm
		}
	}

	return vector, nil
}

// FitTransform fits the vectorizer and transforms all documents in one step.
func (v *TFIDFVectorizer) FitTransform(documents []string) ([][]float64, error) {
	if err := v.Fit(documents); err != nil {
		return nil, err
	}

	vectors := make([][]float64, len(documents))
	for i, doc := range documents {
		vec, err := v.Transform(doc)
		if err != nil {
			return nil, fmt.Errorf("error transforming document %d: %w", i, err)
		}
		vectors[i] = vec
	}

	return vectors, nil
}

// VocabularySize returns the number of unique terms in the vocabulary.
func (v *TFIDFVectorizer) VocabularySize() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.vocabulary)
}

// GetVocabulary returns a copy of the vocabulary map.
func (v *TFIDFVectorizer) GetVocabulary() map[string]int {
	v.mu.RLock()
	defer v.mu.RUnlock()

	vocab := make(map[string]int, len(v.vocabulary))
	for k, v := range v.vocabulary {
		vocab[k] = v
	}
	return vocab
}
