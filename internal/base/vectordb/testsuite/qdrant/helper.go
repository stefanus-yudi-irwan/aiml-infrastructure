package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb"
	"aiml-infrastructure/internal/base/vectordb/backend/qdrant"
	"math/rand/v2"
	"sort"

	"github.com/google/uuid"
)

var (
	TEST_VECTOR_DIMENSION int = 3
	TEST_CONTENT_LENGTH   int = 1000
)

func createTestCollectionConfig(collection string) qdrant.CollectionConfig {
	return qdrant.CollectionConfig{
		Name: collection,
		DenseEmbeddingFields: []*qdrant.DenseEmbeddingFieldConfig{
			{
				Name:      vectordb.DenseEmbedding1Name,
				Dimension: TEST_VECTOR_DIMENSION,
				Distance:  qdrant.COSINE,
				Index:     qdrant.HNSW,
				HNSW: &qdrant.HNSWConfig{
					M:                 qdrant.HNSW_M_SMALL,
					EFConstruction:    qdrant.HNSW_EF_CONSTRUCTION_SMALL,
					FullScanThreshold: qdrant.HNSW_FULL_SCAN_THRESHOLD_SMALL,
				},
			},
			{
				Name:      vectordb.DenseEmbedding2Name,
				Dimension: TEST_VECTOR_DIMENSION,
				Distance:  qdrant.COSINE,
				Index:     qdrant.HNSW,
				HNSW: &qdrant.HNSWConfig{
					M:                 qdrant.HNSW_M_SMALL,
					EFConstruction:    qdrant.HNSW_EF_CONSTRUCTION_SMALL,
					FullScanThreshold: qdrant.HNSW_FULL_SCAN_THRESHOLD_SMALL,
				},
			},
		},
		SparseEmbeddingFields: []*qdrant.SparseEmbeddingFieldConfig{
			{
				Name: vectordb.SparseEmbedding1Name,
			},
			{
				Name: vectordb.SparseEmbedding2Name,
			},
		},
		DataFields: []*qdrant.DataFieldConfig{
			{Name: "person", Type: qdrant.DataFieldKeyword},
			{Name: "place", Type: qdrant.DataFieldKeyword},
			{Name: "organization", Type: qdrant.DataFieldKeyword},
			{Name: "product", Type: qdrant.DataFieldKeyword},
			{Name: "event", Type: qdrant.DataFieldKeyword},
			{Name: "title", Type: qdrant.DataFieldText},
			{Name: "tags", Type: qdrant.DataFieldKeyword},
			{Name: "description", Type: qdrant.DataFieldText},
			{Name: "float_data", Type: qdrant.DataFieldFloat},
			{Name: "boolean_data", Type: qdrant.DataFieldBool},
			{Name: "float_array", Type: qdrant.DataFieldFloat},
			{Name: "boolean_array", Type: qdrant.DataFieldBool},
		},
		ShardNumber:            2,
		ReplicationFactor:      2,
		WriteConsistencyFactor: 2,
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

func normalizeSparseEmbedding(embedding vectordb.SparseEmbedding) vectordb.SparseEmbedding {
	indices := append([]uint32(nil), embedding.Indices...)
	values := append([]float32(nil), embedding.Values...)

	if len(indices) != len(values) {
		return embedding // Alternatively, return an error.
	}

	order := make([]int, len(indices))
	for i := range order {
		order[i] = i
	}

	sort.Slice(order, func(i, j int) bool {
		return indices[order[i]] < indices[order[j]]
	})

	normalized := vectordb.SparseEmbedding{
		Indices: make([]uint32, len(indices)),
		Values:  make([]float32, len(values)),
	}

	for i, originalIndex := range order {
		normalized.Indices[i] = indices[originalIndex]
		normalized.Values[i] = values[originalIndex]
	}

	return normalized
}

func normalizeSparseEmbeddingsMap(
	embeddings map[vectordb.SparseEmbeddingName]vectordb.SparseEmbedding,
) map[vectordb.SparseEmbeddingName]vectordb.SparseEmbedding {
	normalized := make(
		map[vectordb.SparseEmbeddingName]vectordb.SparseEmbedding,
		len(embeddings),
	)

	for name, embedding := range embeddings {
		normalized[name] = normalizeSparseEmbedding(embedding)
	}

	return normalized
}
