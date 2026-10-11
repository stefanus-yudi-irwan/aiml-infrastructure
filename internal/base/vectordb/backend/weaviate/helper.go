package weaviate

import (
	"errors"
	"math"
)

func NormalizeVector(vector []float32) []float32 {
	var normValue float32
	for _, value := range vector {
		normValue += value * value
	}
	normValue = float32(math.Sqrt(float64(normValue)))

	normVector := make([]float32, len(vector))
	for index, value := range vector {
		normVector[index] = value / normValue
	}

	return normVector
}

func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("vectors must have the same dimension")
	}

	if len(a) == 0 {
		return 0, errors.New("vectors must not be empty")
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, errors.New("cosine similarity is undefined for zero vectors")
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB)))), nil
}
