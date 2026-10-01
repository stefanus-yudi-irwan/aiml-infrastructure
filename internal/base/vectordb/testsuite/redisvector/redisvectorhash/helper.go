package redisvector

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/redisvector/redisvectorhash"
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
)

var (
	TEST_VECTOR_DIMENSION int = 512
	TEST_CONTENT_LENGTH   int = 3000
)

func createTestCollectionConfig(collection string) redisvectorhash.CollectionConfig {
	collectionConfig := redisvectorhash.CollectionConfig{
		Name: collection,
		Fields: []redisvectorhash.FieldConfig{
			{
				Type: redisvectorhash.FieldTypeText,
				TextField: &redisvectorhash.TextFieldConfig{
					Name: vectordb.IDColumnName,
				},
			},
			{
				Type: redisvectorhash.FieldTypeText,
				TextField: &redisvectorhash.TextFieldConfig{
					Name: vectordb.DocumentIDColumnName,
				},
			},
			{
				Type: redisvectorhash.FieldTypeText,
				TextField: &redisvectorhash.TextFieldConfig{
					Name: vectordb.ChunkIDColumnName,
				},
			},
			{
				Type: redisvectorhash.FieldTypeText,
				TextField: &redisvectorhash.TextFieldConfig{
					Name: vectordb.ContentColumnName,
				},
			},
			{
				Type: redisvectorhash.FieldTypeVector,
				VectorField: &redisvectorhash.VectorFieldConfig{
					Name:       vectordb.EmbeddingColumnName,
					Dimension:  TEST_VECTOR_DIMENSION,
					Algorithm:  redisvectorhash.HNSW,
					Metric:     redisvectorhash.COSINE,
					Type:       redisvectorhash.VECTOR_FLOAT32,
					InitialCap: 1000,
					HNSW: &redisvectorhash.HNSWConfig{
						M:              redisvectorhash.M_SMALL,
						EFConstruction: redisvectorhash.EF_CONSTRUCTION_SMALL,
						EFRuntime:      redisvectorhash.EF_RUNTIME_SMALL,
					},
				},
			},
			{
				Type: redisvectorhash.FieldTypeTag,
				TagField: &redisvectorhash.TagFieldConfig{
					Name:      "person",
					Separator: ",",
				},
			},
			{
				Type: redisvectorhash.FieldTypeTag,
				TagField: &redisvectorhash.TagFieldConfig{
					Name:      "place",
					Separator: ",",
				},
			},
			{
				Type: redisvectorhash.FieldTypeTag,
				TagField: &redisvectorhash.TagFieldConfig{
					Name:      "organization",
					Separator: ",",
				},
			},
			{
				Type: redisvectorhash.FieldTypeTag,
				TagField: &redisvectorhash.TagFieldConfig{
					Name:      "product",
					Separator: ",",
				},
			},
			{
				Type: redisvectorhash.FieldTypeTag,
				TagField: &redisvectorhash.TagFieldConfig{
					Name:      "event",
					Separator: ",",
				},
			},
		},
	}
	return collectionConfig
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
