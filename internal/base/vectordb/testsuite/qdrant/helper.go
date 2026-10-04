package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/qdrant"
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
)

var (
	TEST_VECTOR_DIMENSION int = 512
	TEST_CONTENT_LENGTH   int = 3000
)

func createTestCollectionConfig(collection string) qdrant.CollectionConfig {
	return qdrant.CollectionConfig{
		Name:      collection,
		Dimension: uint64(TEST_VECTOR_DIMENSION),
		Distance:  qdrant.DistanceCosine,
		HNSW: &qdrant.HNSWConfig{
			M:                 qdrant.M_SMALL,
			EFConstruction:    qdrant.EF_CONSTRUCTION_SMALL,
			FullScanThreshold: qdrant.FULL_SCAN_THRESHOLD_SMALL,
		},
		ShardNumber:            2,
		ReplicationFactor:      2,
		WriteConsistencyFactor: 2,
	}
}

func generateTestData(id string) vectordb.Data {
	return vectordb.Data{
		ID:         id,
		DocumentID: uuid.New().String(),
		ChunkID:    uuid.New().String(),
		Content:    generateRandomString(TEST_CONTENT_LENGTH),
		Embedding:  generateRandomVector(TEST_VECTOR_DIMENSION),
		Metadata: map[string]interface{}{
			"person":       []string{"personA", "personB", "personC"},
			"place":        []string{"placeA", "placeB", "placeC"},
			"organization": []string{"orgA", "orgB", "orgC"},
			"product":      []string{"productA", "productB", "productC"},
			"event":        []string{"eventA", "eventB", "eventC"},
			"title":        "titleA",
			"tags":         []string{"tagsA", "tagsB", "tagsC"},
			"description":  "descriptionA",
		},
	}
}

func generateBulkTestData(dataCount int, prefixID string) []vectordb.Data {

	bulkData := make([]vectordb.Data, dataCount)

	for index := 0; index < dataCount; index++ {
		bulkData[index] = generateTestData(
			fmt.Sprintf("%s-%d", prefixID, index),
		)
	}

	return bulkData
}

func generateBulkTestDataID(dataCount int, prefixID string) []string {
	bulkDataID := make([]string, dataCount)

	for index := 0; index < dataCount; index++ {
		bulkDataID[index] = fmt.Sprintf("%s-%d", prefixID, index)
	}

	return bulkDataID
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
