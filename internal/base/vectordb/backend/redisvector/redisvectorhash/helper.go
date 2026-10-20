package redisvectorhash

import (
	"aiml-infrastructure/internal/base/vectordb"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/go-redis/redis/v8"
)

func convertEmbeddingFloat32toEmbeddingByte(embeddingFloat32 []float32) ([]byte, error) {
	if len(embeddingFloat32) == 0 {
		return nil, errors.New("empty vector value")
	}
	embeddingByte := new(bytes.Buffer)
	if err := binary.Write(embeddingByte, binary.LittleEndian, embeddingFloat32); err != nil {
		return nil, fmt.Errorf("failed to encode vector: %w", err)
	}
	return embeddingByte.Bytes(), nil
}

func convertEmbeddingBytetoEmbeddingFloat32(embeddingByte []byte) ([]float32, error) {
	if len(embeddingByte) == 0 {
		return nil, errors.New("empty vector value")
	}
	embeddingFloat32 := make([]float32, len(embeddingByte)/4)
	if err := binary.Read(bytes.NewReader(embeddingByte), binary.LittleEndian, &embeddingFloat32); err != nil {
		return nil, fmt.Errorf("failed to decode vector: %w", err)
	}
	return embeddingFloat32, nil
}

func createIndexPrefix(indexName string) string {
	return indexName + ":"
}

func createDataKey(collection string, id string) (string, error) {
	if collection == "" {
		return "", errors.New("empty collection name")
	}
	if id == "" {
		return "", errors.New("empty vector ID")
	}
	return collection + ":" + id, nil
}

func createIdentityHash(data vectordb.Data) map[string]interface{} {
	identityHash := make(map[string]interface{})
	identityHash[vectordb.IDColumnName] = data.ID
	identityHash[vectordb.DocumentIDColumnName] = data.DocumentID
	identityHash[vectordb.ChunkIDColumnName] = data.ChunkID
	identityHash[vectordb.ContentColumnName] = data.Content

	return identityHash
}

func createDenseEmbeddingHash(data vectordb.Data) (map[string]interface{}, error) {

	denseEmbeddingHash := make(map[string]interface{})

	for embeddingName, embedding := range data.DenseEmbeddings {
		embeddingByte, err := convertEmbeddingFloat32toEmbeddingByte(
			embedding,
		)
		if err != nil {
			return nil, err
		}

		denseEmbeddingHash[fmt.Sprintf("%s:%s",
			vectordb.DenseEmbeddingsColumnName,
			embeddingName)] = embeddingByte
	}

	return denseEmbeddingHash, nil
}

func createSparseEmbeddingHash(data vectordb.Data) (map[string]interface{}, error) {
	sparseEmbeddingHash := make(map[string]interface{})

	for embeddingName, embedding := range data.SparseEmbeddings {
		sparseEmbeddingJson, err := json.Marshal(embedding)
		if err != nil {
			return nil, err
		}
		sparseEmbeddingHash[fmt.Sprintf("%s:%s",
			vectordb.SparseEmbeddingsColumnName,
			embeddingName)] = sparseEmbeddingJson
	}

	return sparseEmbeddingHash, nil
}

func createMetadataHash(data vectordb.Data) (map[string]interface{}, error) {
	metadataHash := make(map[string]interface{}, len(data.Metadata))

	metadataJson, err := json.Marshal(data.Metadata)
	if err != nil {
		return nil, err
	}
	metadataHash[vectordb.MetadataColumnName] = metadataJson

	for key, value := range data.Metadata {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		metadataHash[fmt.Sprintf("%s:%s",
			vectordb.MetadataColumnName,
			key)] = string(encoded)
	}
	return metadataHash, nil
}

func createVectorMetadataHash(vectorHash, metadataHash map[string]interface{}) map[string]interface{} {
	vectorMetadataHash := make(map[string]interface{}, len(vectorHash)+len(metadataHash))
	for vectorKey, vectorValue := range vectorHash {
		vectorMetadataHash[vectorKey] = vectorValue
	}
	for metadataKey, metadataValue := range vectorHash {
		vectorMetadataHash[metadataKey] = metadataValue
	}
	return vectorMetadataHash
}

func queueUpsert(ctx context.Context, pipe redis.Pipeliner, collection string, data vectordb.Data) error {
	hashKey, err := createDataKey(collection, data.ID)
	if err != nil {
		return fmt.Errorf("fail to create data key: %v", data.ID)
	}

	hashValues := createIdentityHash(data)

	if data.DenseEmbeddings != nil {
		vectorHash, err := createDenseEmbeddingHash(data)
		if err != nil {
			return fmt.Errorf("fail to create denseEmbeddingHash: %v", data.ID)
		}
		for key, value := range vectorHash {
			hashValues[key] = value
		}
	}

	if data.SparseEmbeddings != nil {
		sparseEmbeddingHash, err := createSparseEmbeddingHash(data)
		if err != nil {
			return fmt.Errorf("fail to create sparseEmbeddingHash: %v", data.ID)
		}
		for key, value := range sparseEmbeddingHash {
			hashValues[key] = value
		}
	}

	if data.Metadata != nil {
		metadataHash, err := createMetadataHash(data)
		if err != nil {
			return fmt.Errorf("fail to create metadataHash: %v", data.ID)
		}
		for key, value := range metadataHash {
			hashValues[key] = value
		}
	}

	pipe.HSet(ctx, hashKey, hashValues)

	return nil
}

func queueDelete(ctx context.Context, pipe redis.Pipeliner, collection string, ids ...string) error {

	keys := make([]string, 0, len(ids))

	for _, id := range ids {
		hashKey, err := createDataKey(collection, id)
		if err != nil {
			return fmt.Errorf("fail to create dataKey: %v", id)
		}
		keys = append(keys, hashKey)

	}

	if len(keys) > 0 {
		pipe.Del(ctx, keys...)
	}

	return nil
}

func normalizeMetadata(metadata map[string]interface{}) {
	for key, value := range metadata {
		items, ok := value.([]interface{})
		if !ok {
			continue
		}

		if values, ok := toTypedSlice[string](items); ok {
			metadata[key] = values
		} else if values, ok := toTypedSlice[bool](items); ok {
			metadata[key] = values
		} else if values, ok := toTypedSlice[float64](items); ok {
			metadata[key] = values
		}
	}
}

func toTypedSlice[T any](items []interface{}) ([]T, bool) {
	result := make([]T, len(items))

	for index, item := range items {
		value, ok := item.(T)
		if !ok {
			return nil, false
		}
		result[index] = value
	}

	return result, true
}

func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("vectors must have the same dimension")
	}

	if len(a) == 0 {
		return 0, errors.New("vectors must not be empty")
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, errors.New("cosine similarity is undefined for zero vectors")
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB)))), nil
}
