package redisvectorhash

import (
	"aiml-infrastructure/internal/base/vectordb"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

func createEmbeddingHash(data vectordb.Data) (map[string]interface{}, error) {
	embeddingByte, err := convertEmbeddingFloat32toEmbeddingByte(data.Embedding)
	if err != nil {
		return nil, err
	}

	embeddingHash := make(map[string]interface{})
	embeddingHash[vectordb.EmbeddingColumnName] = embeddingByte

	return embeddingHash, nil
}

func createMetadataHash(data vectordb.Data) (map[string]interface{}, error) {
	metadataHash := make(map[string]interface{}, len(data.Metadata))

	metadataJson, err := json.Marshal(data.Metadata)
	if err != nil {
		return nil, err
	}
	metadataHash[vectordb.MetadataColumnName] = metadataJson

	for key, value := range data.Metadata {
		switch value := value.(type) {
		case []string:
			metadataHash[key] = strings.Join(value, ",")
		default:
			metadataHash[key] = value
		}

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

	if data.Embedding != nil {
		vectorHash, err := createEmbeddingHash(data)
		if err != nil {
			return fmt.Errorf("fail to create vectorHash: %v", data.ID)
		}
		for key, value := range vectorHash {
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
