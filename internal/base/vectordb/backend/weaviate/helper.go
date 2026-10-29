package weaviate

import (
	"aiml-infrastructure/internal/base/vectordb"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/weaviate/weaviate/entities/models"
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

func createWeaviateProperties(data vectordb.Data) (map[string]interface{}, error) {
	properties := map[string]interface{}{
		vectordb.DocumentIDColumnName: data.DocumentID,
		vectordb.ChunkIDColumnName:    data.ChunkID,
		vectordb.ContentColumnName:    data.Content,
		vectordb.MetadataColumnName:   data.Metadata,
	}

	if len(data.SparseEmbeddings) > 0 {
		for name, sparseEmbedding := range data.SparseEmbeddings {
			jsonData, err := json.Marshal(sparseEmbedding)
			if err != nil {
				return nil, err
			}
			properties[fmt.Sprintf("%s_%s",
				vectordb.SparseEmbeddingsColumnName,
				name)] = string(jsonData)
		}
	}

	return properties, nil
}

func createWeaviateVectors(data vectordb.Data) models.Vectors {
	vectors := make(models.Vectors, len(data.DenseEmbeddings))

	for name, denseEmbedding := range data.DenseEmbeddings {
		vectors[fmt.Sprintf("%s_%s",
			vectordb.DenseEmbeddingsColumnName,
			name)] = denseEmbedding
	}

	return vectors
}

func normalizeMetadata(metadata map[string]interface{}) {
	for key, value := range metadata {
		items, ok := value.([]interface{})
		if !ok {
			continue
		}

		if values, ok := toTypedSlice[string](items); ok {
			metadata[key] = values
		} else if values, ok := toTypedSlice[bool](items); ok {
			metadata[key] = values
		} else if values, ok := toTypedSlice[float64](items); ok {
			metadata[key] = values
		}
	}
}

func toTypedSlice[T any](items []interface{}) ([]T, bool) {
	result := make([]T, len(items))

	for index, item := range items {
		value, ok := item.(T)
		if !ok {
			return nil, false
		}
		result[index] = value
	}

	return result, true
}
