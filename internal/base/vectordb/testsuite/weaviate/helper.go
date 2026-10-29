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
	TEST_CONTENT_LENGTH   int = 65535
)

func createTestCollectionConfig(collection string) weaviate.CollectionConfig {
	return weaviate.CollectionConfig{
		Name:        collection,
		Description: fmt.Sprintf("test collection for collection %v unit testing", collection),
		DenseEmbeddingFields: []*weaviate.DenseEmbeddingFieldConfig{
			{
				Name:      vectordb.DenseEmbedding1Name,
				Dimension: TEST_VECTOR_DIMENSION,
				Distance:  weaviate.COSINE,
				Index:     weaviate.HNSW,
				HNSW: &weaviate.HNSWConfig{
					M:              weaviate.HNSW_M_SMALL,
					EFRuntime:      weaviate.HNSW_EF_RUNTIME_SMALL,
					EFConstruction: weaviate.HNSW_EF_CONSTRUCTION_SMALL,
				},
			},
			{
				Name:      vectordb.DenseEmbedding2Name,
				Dimension: TEST_VECTOR_DIMENSION,
				Distance:  weaviate.COSINE,
				Index:     weaviate.HNSW,
				HNSW: &weaviate.HNSWConfig{
					M:              weaviate.HNSW_M_SMALL,
					EFRuntime:      weaviate.HNSW_EF_RUNTIME_SMALL,
					EFConstruction: weaviate.HNSW_EF_CONSTRUCTION_SMALL,
				},
			},
		},
		SparseEmbeddingFields: []*weaviate.SparseEmbeddingFieldConfig{
			{
				Name:        vectordb.SparseEmbedding1Name,
				Description: fmt.Sprintf("description for %s", vectordb.SparseEmbedding1Name),
				DataType:    weaviate.DataFieldText,
			},
			{
				Name:        vectordb.SparseEmbedding2Name,
				Description: fmt.Sprintf("description for %s", vectordb.SparseEmbedding2Name),
				DataType:    weaviate.DataFieldText,
			},
		},
		DataFields: []*weaviate.DataFieldConfig{
			{
				Name: vectordb.DocumentIDColumnName,
				Description: fmt.Sprintf("description for %s",
					vectordb.DocumentIDColumnName),
				DataType: weaviate.DataFieldText,
			},
			{
				Name: vectordb.ChunkIDColumnName,
				Description: fmt.Sprintf("description for %s",
					vectordb.ChunkIDColumnName),
				DataType: weaviate.DataFieldText,
			},
			{
				Name: vectordb.ContentColumnName,
				Description: fmt.Sprintf("description for %s",
					vectordb.ContentColumnName),
				DataType: weaviate.DataFieldText,
			},
			{
				Name: vectordb.MetadataColumnName,
				Description: fmt.Sprintf("description for %v",
					vectordb.MetadataColumnName),
				DataType: weaviate.DataFieldObject,
				Fields: []weaviate.DataFieldConfig{
					{Name: "person", DataType: weaviate.DataFieldTextArray},
					{Name: "place", DataType: weaviate.DataFieldTextArray},
					{Name: "organization", DataType: weaviate.DataFieldTextArray},
					{Name: "product", DataType: weaviate.DataFieldTextArray},
					{Name: "event", DataType: weaviate.DataFieldTextArray},
					{Name: "title", DataType: weaviate.DataFieldText},
					{Name: "tags", DataType: weaviate.DataFieldTextArray},
					{Name: "description", DataType: weaviate.DataFieldText},
					{Name: "float_data", DataType: weaviate.DataFieldFloat},
					{Name: "float_array", DataType: weaviate.DataFieldFloatArray},
					{Name: "boolean_data", DataType: weaviate.DataFieldBool},
					{Name: "boolean_array", DataType: weaviate.DataFieldBoolArray},
				},
			},
		},
	}
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
			"person":        []string{"personA", "personB", "personC"},
			"place":         []string{"placeA", "placeB", "placeC"},
			"organization":  []string{"orgA", "orgB", "orgC"},
			"product":       []string{"productA", "productB", "productC"},
			"event":         []string{"eventA", "eventB", "eventC"},
			"title":         "titleA",
			"tags":          []string{"tagsA", "tagsB", "tagsC"},
			"description":   "descriptionA",
			"float_data":    1234.567,
			"boolean_data":  true,
			"float_array":   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			"boolean_array": []bool{true, true, true, false, false},
		},
	}
}

func generateTestDataIdentity(id string) vectordb.Data {
	return vectordb.Data{
		ID:         id,
		DocumentID: uuid.New().String(),
		ChunkID:    uuid.New().String(),
		Content:    generateRandomString(TEST_CONTENT_LENGTH),
	}
}

func insertDenseEmbeddings(data vectordb.Data) vectordb.Data {
	denseEmbeddings := make(map[vectordb.DenseEmbeddingName][]float32)
	denseEmbeddings[vectordb.DenseEmbedding1Name] = generateRandomVector(TEST_VECTOR_DIMENSION)
	denseEmbeddings[vectordb.DenseEmbedding2Name] = generateRandomVector(TEST_VECTOR_DIMENSION)

	return vectordb.Data{
		ID:              data.ID,
		DocumentID:      data.DocumentID,
		ChunkID:         data.ChunkID,
		Content:         data.Content,
		DenseEmbeddings: denseEmbeddings,
	}
}

func insertSparseEmbeddings(data vectordb.Data) vectordb.Data {
	sparseEmbeddings := make(map[vectordb.SparseEmbeddingName]vectordb.SparseEmbedding)

	sparseEmbeddings[vectordb.SparseEmbedding1Name] = vectordb.SparseEmbedding{
		Indices: generateRandomIndices(TEST_VECTOR_DIMENSION),
		Values:  generateRandomVector(TEST_VECTOR_DIMENSION),
	}
	sparseEmbeddings[vectordb.SparseEmbedding2Name] = vectordb.SparseEmbedding{
		Indices: generateRandomIndices(TEST_VECTOR_DIMENSION),
		Values:  generateRandomVector(TEST_VECTOR_DIMENSION),
	}

	return vectordb.Data{
		ID:               data.ID,
		DocumentID:       data.DocumentID,
		ChunkID:          data.ChunkID,
		Content:          data.Content,
		SparseEmbeddings: sparseEmbeddings,
	}
}

func insertMetadata(data vectordb.Data) vectordb.Data {
	return vectordb.Data{
		ID:         data.ID,
		DocumentID: data.DocumentID,
		ChunkID:    data.ChunkID,
		Content:    data.Content,
		Metadata: map[string]interface{}{
			"person":        []string{"personA", "personB", "personC"},
			"place":         []string{"placeA", "placeB", "placeC"},
			"organization":  []string{"orgA", "orgB", "orgC"},
			"product":       []string{"productA", "productB", "productC"},
			"event":         []string{"eventA", "eventB", "eventC"},
			"title":         "titleA",
			"tags":          []string{"tagsA", "tagsB", "tagsC"},
			"description":   "descriptionA",
			"float_data":    1234.567,
			"boolean_data":  true,
			"float_array":   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			"boolean_array": []bool{true, true, true, false, false},
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

func generateBulkTestDataIdentity(dataID []string) []vectordb.Data {
	bulkData := make([]vectordb.Data, len(dataID))

	for index, id := range dataID {
		bulkData[index] = generateTestDataIdentity(id)
	}

	return bulkData
}

func insertBulkDenseEmbeddings(arrayData []vectordb.Data) []vectordb.Data {
	bulkData := make([]vectordb.Data, len(arrayData))

	for index, data := range arrayData {
		bulkData[index] = insertDenseEmbeddings(data)
	}

	return bulkData
}

func insertBulkSparseEmbeddings(arrayData []vectordb.Data) []vectordb.Data {
	bulkData := make([]vectordb.Data, len(arrayData))

	for index, data := range arrayData {
		bulkData[index] = insertSparseEmbeddings(data)
	}

	return bulkData
}

func insertBulkMetadata(arrayData []vectordb.Data) []vectordb.Data {
	bulkData := make([]vectordb.Data, len(arrayData))

	for index, data := range arrayData {
		bulkData[index] = insertMetadata(data)
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
