package weaviate

import "aiml-infrastructure/internal/base/vectordb/backend/weaviate"

func createTestCollectionConfig() weaviate.CollectionConfig {
	return weaviate.CollectionConfig{
		Name:        "test_collection",
		Description: "Test collection for unit testing",
		Properties: []weaviate.PropertyConfig{
			{
				Name:         "id",
				Description:  "Primary identifier",
				DataType:     []string{"text"},
				Tokenization: "field",
			},
			{
				Name:        "metadata",
				Description: "Additional metadata",
				DataType:    []string{"object"},
			},
		},
		Vectorizer: "none",
	}
}
