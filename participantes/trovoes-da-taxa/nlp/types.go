package nlp

import "fmt"

// IntentVector represents a preprocessed intent with its TF-IDF vector.
type IntentVector struct {
	// Original is the original intent text before preprocessing
	Original string

	// Processed is the text after preprocessing (lowercase, no stopwords, stemmed)
	Processed string

	// Vector is the TF-IDF vector representation
	Vector []float64

	// Category is the service category for this intent
	Category string

	// Metadata stores any additional information
	Metadata map[string]interface{}
}

// Pipeline combines all NLP operations into a single workflow.
type Pipeline struct {
	Preprocessor  *Preprocessor
	Vectorizer    *TFIDFVectorizer
	IntentVectors []IntentVector
}

// NewPipeline creates a new NLP pipeline with the specified language.
func NewPipeline(language string, normalized bool) (*Pipeline, error) {
	preprocessor, err := NewPreprocessor(language)
	if err != nil {
		return nil, err
	}

	vectorizer := NewTFIDFVectorizer(normalized)

	return &Pipeline{
		Preprocessor:  preprocessor,
		Vectorizer:    vectorizer,
		IntentVectors: make([]IntentVector, 0),
	}, nil
}

// Train trains the pipeline on a set of intents.
func (p *Pipeline) Train(intents []string, categories []string) error {
	if len(intents) != len(categories) {
		return fmt.Errorf("intents and categories must have same length")
	}

	// Step 1: Preprocess all intents
	processed := p.Preprocessor.ProcessBatch(intents)

	// Step 2: Fit and transform with TF-IDF
	vectors, err := p.Vectorizer.FitTransform(processed)
	if err != nil {
		return fmt.Errorf("error during vectorization: %w", err)
	}

	// Step 3: Store intent vectors
	p.IntentVectors = make([]IntentVector, len(intents))
	for i := range intents {
		p.IntentVectors[i] = IntentVector{
			Original:  intents[i],
			Processed: processed[i],
			Vector:    vectors[i],
			Category:  categories[i],
			Metadata:  make(map[string]interface{}),
		}
	}

	return nil
}

// Predict finds the most similar intent for a given query.
func (p *Pipeline) Predict(query string) (*IntentVector, float64, error) {
	// Preprocess the query
	processed := p.Preprocessor.Process(query)

	// Transform to vector
	queryVector, err := p.Vectorizer.Transform(processed)
	if err != nil {
		return nil, 0, fmt.Errorf("error transforming query: %w", err)
	}

	// Extract just the vectors for comparison
	vectors := make([][]float64, len(p.IntentVectors))
	for i := range p.IntentVectors {
		vectors[i] = p.IntentVectors[i].Vector
	}

	// Find most similar
	idx, similarity, err := FindMostSimilar(queryVector, vectors)
	if err != nil {
		return nil, 0, fmt.Errorf("error finding similar intent: %w", err)
	}

	return &p.IntentVectors[idx], similarity, nil
}

// PredictTopK finds the top K most similar intents for a given query.
func (p *Pipeline) PredictTopK(query string, k int) ([]IntentVector, []float64, error) {
	// Preprocess the query
	processed := p.Preprocessor.Process(query)

	// Transform to vector
	queryVector, err := p.Vectorizer.Transform(processed)
	if err != nil {
		return nil, nil, fmt.Errorf("error transforming query: %w", err)
	}

	// Extract just the vectors for comparison
	vectors := make([][]float64, len(p.IntentVectors))
	for i := range p.IntentVectors {
		vectors[i] = p.IntentVectors[i].Vector
	}

	// Find top K similar
	results, err := FindTopKSimilar(queryVector, vectors, k)
	if err != nil {
		return nil, nil, fmt.Errorf("error finding similar intents: %w", err)
	}

	// Extract intent vectors and similarities
	topIntents := make([]IntentVector, len(results))
	similarities := make([]float64, len(results))

	for i, result := range results {
		topIntents[i] = p.IntentVectors[result.Index]
		similarities[i] = result.Similarity
	}

	return topIntents, similarities, nil
}
