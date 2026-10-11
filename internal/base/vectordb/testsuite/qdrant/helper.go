package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/qdrant"
	"math/rand/v2"

	"github.com/google/uuid"
)

var (
	TEST_VECTOR_DIMENSION int = 128
	TEST_CONTENT_LENGTH   int = 1000
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

func generateBulkTestData(dataID []string) []vectordb.Data {

	bulkData := make([]vectordb.Data, len(dataID))

	for index, id := range dataID {
		bulkData[index] = generateTestData(id)
	}

	return bulkData
}

func generateBulkTestDataID(dataCount int) []string {
	bulkDataID := make([]string, dataCount)

	for index := 0; index < dataCount; index++ {
		bulkDataID[index] = uuid.New().String()
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
