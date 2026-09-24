package weaviate

import "aiml-infrastructure/internal/base/vectordb/backend/weaviate"

func createTestCollectionConfig() weaviate.CollectionConfig {
	return weaviate.CollectionConfig{
		Name:        "TestCollection",
		Description: "Test collection for unit testing",
		Properties: []weaviate.PropertyConfig{
			{
				Name:         "uuid",
				Description:  "Primary identifier",
				DataType:     []string{"text"},
				Tokenization: "field",
			},
			{
				Name:        "metadata",
				Description: "additional metadata",
				DataType:    []string{"object"},
				Properties: []weaviate.PropertyConfig{
					{
						Name:         "source",
						Description:  "data source",
						DataType:     []string{"text"},
						Tokenization: "field",
					},
					{
						Name:         "category",
						Description:  "document category",
						DataType:     []string{"text"},
						Tokenization: "field",
					},
				},
			},
		},
		Vectorizer: "none",
	}
}
