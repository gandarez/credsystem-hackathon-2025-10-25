package nlp

import (
	"fmt"
	"math"
)

// CosineSimilarity calculates the cosine similarity between two vectors.
// Returns a value between 0 and 1, where:
// - 1 means vectors are identical
// - 0 means vectors are orthogonal (no similarity)
//
// For normalized vectors (as from TF-IDF with normalization enabled),
// this is simply the dot product.
func CosineSimilarity(vec1, vec2 []float64) (float64, error) {
	if len(vec1) != len(vec2) {
		return 0, fmt.Errorf("vectors must have same length: %d != %d", len(vec1), len(vec2))
	}

	if len(vec1) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var dotProduct, norm1, norm2 float64

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	// Handle zero vectors
	if norm1 == 0 || norm2 == 0 {
		return 0, nil
	}

	norm1 = math.Sqrt(norm1)
	norm2 = math.Sqrt(norm2)

	similarity := dotProduct / (norm1 * norm2)

	// Clamp to [0, 1] to handle floating-point errors
	if similarity > 1.0 {
		similarity = 1.0
	} else if similarity < 0.0 {
		similarity = 0.0
	}

	return similarity, nil
}

// FindMostSimilar finds the index of the most similar vector in a collection.
// Returns the index and the similarity score.
func FindMostSimilar(query []float64, vectors [][]float64) (int, float64, error) {
	if len(vectors) == 0 {
		return -1, 0, fmt.Errorf("vector collection is empty")
	}

	maxSimilarity := -1.0
	maxIndex := -1

	for i, vec := range vectors {
		similarity, err := CosineSimilarity(query, vec)
		if err != nil {
			return -1, 0, fmt.Errorf("error calculating similarity with vector %d: %w", i, err)
		}

		if similarity > maxSimilarity {
			maxSimilarity = similarity
			maxIndex = i
		}
	}

	if maxIndex == -1 {
		return -1, 0, fmt.Errorf("no valid similarity found")
	}

	return maxIndex, maxSimilarity, nil
}

// TopKSimilar finds the K most similar vectors and their indices.
// Returns a slice of results sorted by similarity (highest first).
type SimilarityResult struct {
	Index      int
	Similarity float64
}

// FindTopKSimilar finds the top K most similar vectors.
func FindTopKSimilar(query []float64, vectors [][]float64, k int) ([]SimilarityResult, error) {
	if len(vectors) == 0 {
		return nil, fmt.Errorf("vector collection is empty")
	}

	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	if k > len(vectors) {
		k = len(vectors)
	}

	// Calculate all similarities
	results := make([]SimilarityResult, len(vectors))
	for i, vec := range vectors {
		similarity, err := CosineSimilarity(query, vec)
		if err != nil {
			return nil, fmt.Errorf("error calculating similarity with vector %d: %w", i, err)
		}
		results[i] = SimilarityResult{
			Index:      i,
			Similarity: similarity,
		}
	}

	// Sort by similarity (descending)
	// Using a simple bubble sort for the top K elements (efficient for small K)
	for i := 0; i < k; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Similarity > results[i].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results[:k], nil
}
