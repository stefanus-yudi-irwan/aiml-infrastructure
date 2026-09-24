package pgvector

import "aiml-infrastructure/internal/base/vectordb/backend/pgvector"

func createTestCollectionConfig() pgvector.CollectionConfig {
	return pgvector.CollectionConfig{
		Name:               "test_collection",
		EmbeddingDimension: 128,
	}
}
