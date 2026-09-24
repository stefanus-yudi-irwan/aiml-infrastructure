package milvus

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/milvus"
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/milvus-io/milvus/client/v2/entity"
)

func createTestCollectionConfig() milvus.CollectionConfig {
	return milvus.CollectionConfig{
		Name:        "test_collection",
		Description: "Test collection for unit testing",
		Fields: []milvus.FieldConfig{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				MaxLength:  64,
			},
			{
				Name:       "document_id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				MaxLength:  64,
			},
			{
				Name:       "chunk_id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				MaxLength:  64,
			},
			{
				Name:       "content",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				MaxLength:  65535,
			},
			{
				Name:      "vector",
				DataType:  entity.FieldTypeFloatVector,
				Dimension: 128,
				VectorConfig: &milvus.VectorIndexConfig{
					Algorithm: milvus.HNSW,
					Metric:    milvus.COSINE,
					HNSW: &milvus.HNSWConfig{
						M:              milvus.M_SMALL,
						EFConstruction: milvus.EF_CONSTRUCTION_SMALL,
					},
				},
			},
			{
				Name:     "metadata",
				DataType: entity.FieldTypeJSON,
			},
		},
	}
}

func createTestVector(id string) vectordb.Data {
	return vectordb.Data{
		ID:         id,
		DocumentID: uuid.New().String(),
		ChunkID:    uuid.New().String(),
		Content:    generateRandomString(65535),
		Embedding:  generateRandomVector(128),
		Metadata: map[string]interface{}{
			"metadata": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
}

func generateRandomVector(dim int) []float32 {
	vector := make([]float32, dim)

	for i := range vector {
		vector[i] = rand.Float32()
	}

	return vector
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}

	return string(result)
}
