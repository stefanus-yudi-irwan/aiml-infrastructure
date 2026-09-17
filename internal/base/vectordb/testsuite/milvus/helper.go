package milvus

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/milvus"
	"math/rand/v2"

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
		ID:     id,
		Vector: generateRandomVector(128),
		Fields: map[string]interface{}{
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
