package qdrant

import "aiml-infrastructure/internal/base/vectordb/backend/qdrant"

func createTestCollectionConfig() qdrant.CollectionConfig {
	return qdrant.CollectionConfig{
		Name:                   "test_collection",
		Dimension:              128,
		Distance:               qdrant.DistanceCosine,
		ShardNumber:            2,
		ReplicationFactor:      2,
		WriteConsistencyFactor: 2,
	}
}
