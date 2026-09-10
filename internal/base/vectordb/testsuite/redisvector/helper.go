package redisvector

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/redisvector"
)

var (
	TEST_HASH_COLLECTION string = "hash-collection"
	TEST_JSON_COLLECTION string = "json-collection"
)

func createTestCollectionConfig(name string, on redisvector.IndexType) redisvector.CollectionConfig {
	collectionConfig := redisvector.CollectionConfig{
		Name: name,
		On:   on,
		Fields: []redisvector.FieldConfig{
			{
				Type: redisvector.FieldTypeVector,
				VectorField: &redisvector.VectorFieldConfig{
					Name:       "vector",
					Dimension:  4,
					Algorithm:  redisvector.HNSW,
					Metric:     redisvector.COSINE,
					Type:       redisvector.VECTOR_FLOAT32,
					InitialCap: 1000,
					HNSW: &redisvector.HNSWConfig{
						M:              redisvector.M_SMALL,
						EFConstruction: redisvector.EF_CONSTRUCTION_SMALL,
						EFRuntime:      redisvector.EF_RUNTIME_SMALL,
					},
				},
			},
		},
	}
	return collectionConfig
}

func generateTestData(id string) vectordb.Data {
	return vectordb.Data{
		ID:     id,
		Vector: []float32{0.1, 0.2, 0.3, 0.4},
		Fields: vectordb.Metadata{
			"name":   "test-data",
			"source": "unit-test",
			"active": true,
			"score":  67,
		},
	}
}
