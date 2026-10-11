package weaviate

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/weaviate"
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
)

var (
	TEST_VECTOR_DIMENSION int = 512
	TEST_CONTENT_LENGTH   int = 3000
)

func createTestCollectionConfig(collection string) weaviate.CollectionConfig {
	return weaviate.CollectionConfig{
		Name:        collection,
		Description: fmt.Sprintf("test collection for collection %v unit testing", collection),
		Dimension:   int64(TEST_VECTOR_DIMENSION),
		VectorIndex: weaviate.HNSWIndexType,
		Distance:    weaviate.DistanceCosine,
		HNSW: &weaviate.HNSWConfig{
			EF:             weaviate.EF_SMALL,
			EFConstruction: weaviate.EF_CONSTRUCTION_SMALL,
			MaxConnections: weaviate.MAX_CONNECTIONS_SMALL,
		},
		Fields: []weaviate.FieldConfig{
			{
				Name:        vectordb.DocumentIDColumnName,
				Description: fmt.Sprintf("description for %v", vectordb.DocumentIDColumnName),
				DataType:    []string{"text"},
			},
			{
				Name:        vectordb.ChunkIDColumnName,
				Description: fmt.Sprintf("description for %v", vectordb.ChunkIDColumnName),
				DataType:    []string{"text"},
			},
			{
				Name:        vectordb.ContentColumnName,
				Description: fmt.Sprintf("description for %v", vectordb.ContentColumnName),
				DataType:    []string{"text"},
			},
			{
				Name:        vectordb.MetadataColumnName,
				Description: fmt.Sprintf("description for %v", vectordb.MetadataColumnName),
				DataType:    []string{"object"},
				Fields: []weaviate.FieldConfig{
					{Name: "person", DataType: []string{"text[]"}},
					{Name: "place", DataType: []string{"text[]"}},
					{Name: "organization", DataType: []string{"text[]"}},
					{Name: "product", DataType: []string{"text[]"}},
					{Name: "event", DataType: []string{"text[]"}},
					{Name: "title", DataType: []string{"text"}},
					{Name: "tags", DataType: []string{"text[]"}},
					{Name: "description", DataType: []string{"text"}},
				},
			},
		},
		Vectorizer: "none",
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
