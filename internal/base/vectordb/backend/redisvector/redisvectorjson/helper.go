package redisvectorjson

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/go-redis/redis/v8"
)

func convertVectorFloat32toVectorByte(vectorFloat32 []float32) ([]byte, error) {
	if len(vectorFloat32) == 0 {
		return nil, errors.New("empty vector value")
	}
	vectorByte := new(bytes.Buffer)
	if err := binary.Write(vectorByte, binary.LittleEndian, vectorFloat32); err != nil {
		return nil, fmt.Errorf("failed to encode vector: %w", err)
	}
	return vectorByte.Bytes(), nil
}

func convertVectorBytetoVectorFloat32(vectorByte []byte) ([]float32, error) {
	if len(vectorByte) == 0 {
		return nil, errors.New("empty vector value")
	}
	vectorFloat32 := make([]float32, len(vectorByte)/4)
	if err := binary.Read(bytes.NewReader(vectorByte), binary.LittleEndian, &vectorFloat32); err != nil {
		return nil, fmt.Errorf("failed to decode vector: %w", err)
	}
	return vectorFloat32, nil
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
