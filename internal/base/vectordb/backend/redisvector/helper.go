package redisvector

import (
	"aiml-infrastructure/internal/base/vectordb"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
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

func createVectorHash(data vectordb.Data) (map[string]interface{}, error) {
	vectorByte, err := convertVectorFloat32toVectorByte(data.Vector)
	if err != nil {
		return nil, err
	}

	vectorHash := make(map[string]interface{})
	vectorHash["vector"] = vectorByte

	return vectorHash, nil
}

func createMetadataHash(data vectordb.Data) (map[string]interface{}, error) {
	metadataHash := make(map[string]interface{}, len(data.Fields))
	for key, value := range data.Fields {
		metadataHash[key] = value
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
