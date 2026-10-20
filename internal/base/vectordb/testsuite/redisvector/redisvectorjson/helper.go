package redisvectorjson

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/redisvector/redisvectorjson"
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
)

var (
	TEST_VECTOR_DIMENSION int = 512
	TEST_CONTENT_LENGTH   int = 3000
)

func createTestCollectionConfig(collection string) redisvectorjson.CollectionConfig {
	collectionConfig := redisvectorjson.CollectionConfig{
		Name: collection,
		EmbeddingFields: []*redisvectorjson.EmbeddingFieldConfig{
			{
				Name: fmt.Sprintf("%s.%s",
					vectordb.DenseEmbeddingColumnName,
					vectordb.DenseEmbedding1Name),
				DataType:   redisvectorjson.VECTOR_FLOAT32,
				Dimension:  TEST_VECTOR_DIMENSION,
				Distance:   redisvectorjson.COSINE,
				Index:      redisvectorjson.HNSW,
				InitialCap: 1000,
				HNSW: &redisvectorjson.HNSWConfig{
					M:              redisvectorjson.HNSW_M_SMALL,
					EFConstruction: redisvectorjson.HNSW_EF_CONSTRUCTION_SMALL,
					EFRuntime:      redisvectorjson.HNSW_EF_RUNTIME_SMALL,
				},
			},
			{
				Name: fmt.Sprintf("%s.%s",
					vectordb.DenseEmbeddingColumnName,
					vectordb.DenseEmbedding2Name),
				DataType:   redisvectorjson.VECTOR_FLOAT32,
				Dimension:  TEST_VECTOR_DIMENSION,
				Distance:   redisvectorjson.COSINE,
				Index:      redisvectorjson.HNSW,
				InitialCap: 1000,
				HNSW: &redisvectorjson.HNSWConfig{
					M:              redisvectorjson.HNSW_M_SMALL,
					EFConstruction: redisvectorjson.HNSW_EF_CONSTRUCTION_SMALL,
					EFRuntime:      redisvectorjson.HNSW_EF_RUNTIME_SMALL,
				},
			},
		},
		DataFields: []*redisvectorjson.DataFieldConfig{
			{
				Type: redisvectorjson.FieldTypeText,
				TextField: &redisvectorjson.TextFieldConfig{
					Name: vectordb.IDColumnName,
				},
			},
			{
				Type: redisvectorjson.FieldTypeText,
				TextField: &redisvectorjson.TextFieldConfig{
					Name: vectordb.DocumentIDColumnName,
				},
			},
			{
				Type: redisvectorjson.FieldTypeText,
				TextField: &redisvectorjson.TextFieldConfig{
					Name: vectordb.ChunkIDColumnName,
				},
			},
			{
				Type: redisvectorjson.FieldTypeText,
				TextField: &redisvectorjson.TextFieldConfig{
					Name: vectordb.ContentColumnName,
				},
			},
			{
				Type: redisvectorjson.FieldTypeText,
				TextField: &redisvectorjson.TextFieldConfig{
					Name: fmt.Sprintf("%s_%s",
						vectordb.SparseEmbeddingColumnName,
						vectordb.SparseEmbedding1Name),
				},
			},
			{
				Type: redisvectorjson.FieldTypeTag,
				TagField: &redisvectorjson.TagFieldConfig{
					Name:      "person",
					Separator: ",",
				},
			},
			{
				Type: redisvectorjson.FieldTypeTag,
				TagField: &redisvectorjson.TagFieldConfig{
					Name:      "place",
					Separator: ",",
				},
			},
			{
				Type: redisvectorjson.FieldTypeTag,
				TagField: &redisvectorjson.TagFieldConfig{
					Name:      "organization",
					Separator: ",",
				},
			},
			{
				Type: redisvectorjson.FieldTypeTag,
				TagField: &redisvectorjson.TagFieldConfig{
					Name:      "product",
					Separator: ",",
				},
			},
			{
				Type: redisvectorjson.FieldTypeTag,
				TagField: &redisvectorjson.TagFieldConfig{
					Name:      "event",
					Separator: ",",
				},
			},
		},
	}
	return collectionConfig
}

func generateTestData(id string) vectordb.Data {

	denseEmbeddings := make(map[vectordb.DenseEmbeddingName][]float32)
	sparseEmbeddings := make(map[vectordb.SparseEmbeddingName]vectordb.SparseEmbedding)

	denseEmbeddings[vectordb.DenseEmbedding1Name] = generateRandomVector(TEST_VECTOR_DIMENSION)
	denseEmbeddings[vectordb.DenseEmbedding2Name] = generateRandomVector(TEST_VECTOR_DIMENSION)
	sparseEmbeddings[vectordb.SparseEmbedding1Name] = vectordb.SparseEmbedding{
		Indices: generateRandomIndices(TEST_VECTOR_DIMENSION),
		Values:  generateRandomVector(TEST_VECTOR_DIMENSION),
	}
	sparseEmbeddings[vectordb.SparseEmbedding2Name] = vectordb.SparseEmbedding{
		Indices: generateRandomIndices(TEST_VECTOR_DIMENSION),
		Values:  generateRandomVector(TEST_VECTOR_DIMENSION),
	}

	return vectordb.Data{
		ID:               id,
		DocumentID:       uuid.New().String(),
		ChunkID:          uuid.New().String(),
		Content:          generateRandomString(TEST_CONTENT_LENGTH),
		DenseEmbeddings:  denseEmbeddings,
		SparseEmbeddings: sparseEmbeddings,
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

func generateRandomIndices(dim int) []uint32 {
	indices := make([]uint32, dim)

	for i := range indices {
		indices[i] = rand.Uint32()
	}

	return indices
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}

	return string(result)
}
