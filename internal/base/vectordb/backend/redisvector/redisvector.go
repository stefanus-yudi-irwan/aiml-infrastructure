package redisvector

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

type VectorDBConnector struct {
	client      *redis.Client
	collections map[string]CollectionConfig
}

func NewVectorDBConnector(config VectorDBConnectorConfig) (*VectorDBConnector, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Password,
		DB:       int(config.Db),
	})

	collections := make(map[string]CollectionConfig, len(config.Collections))

	for _, collection := range config.Collections {
		if err := collection.Validate(); err != nil {
			return nil, err
		}
		collections[collection.Name] = collection
	}

	return &VectorDBConnector{
		client:      client,
		collections: collections,
	}, nil
}

func (v *VectorDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("VectorDBConnector.(%v)(%v) %w", method, params, err)
}

func (v *VectorDBConnector) CreateCollection(ctx context.Context, name string) error {
	config, isExists := v.collections[name]
	if !isExists {
		return v.error(fmt.Errorf("collection %s not found", name), "CreateCollection-001")
	}

	creationQuery := config.CreateCollectionQuery()

	_, err := v.client.Do(ctx, creationQuery...).Result()
	if err != nil {
		return v.error(err, "CreateCollection-002", "fail to create collection "+name)
	}

	return nil
}

func (v *VectorDBConnector) DeleteCollection(ctx context.Context, name string) error {
	_, err := v.client.Do(ctx, "FT.DROPINDEX", name, "DD").Result()
	if err != nil {
		return v.error(err, "DeleteCollection", "fail to delete colleciton "+name)
	}
	return nil
}

func (v *VectorDBConnector) AddCollectionConfig(config CollectionConfig) error {
	if err := config.Validate(); err != nil {
		return v.error(err, "AddCollectionConfig-001", "fail to validate new collection config")
	}

	if v.collections == nil {
		v.collections = make(map[string]CollectionConfig)
	}

	if _, exists := v.collections[config.Name]; exists {
		return v.error(fmt.Errorf("collection %q already exists", config.Name), "AddCollectionConfig-002")
	}

	v.collections[config.Name] = config

	return nil
}

func (v *VectorDBConnector) IsCollectionExists(ctx context.Context, name string) (bool, error) {
	_, err := v.client.Do(ctx, "FT.INFO", name).Result()
	switch {
	case err == nil:
		return true, nil
	case strings.Contains(err.Error(), "unknown Index name"):
		return false, nil
	default:
		return false, v.error(err, "IsCollectionExists", name)
	}
}

func (f *VectorDBConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {
	indexType := f.collections[collection].On

	switch indexType {
	case HASH_TYPE:
		return f.upsertHash(ctx, collection, data)
	case JSON_TYPE:
		return f.upsertJSON(ctx, collection, data)
	default:
		return fmt.Errorf("unsupported index type: %s", indexType)
	}
}

func (v *VectorDBConnector) upsertJSON(ctx context.Context, collection string, data vectordb.Data) error {
	dataKey, err := createDataKey(collection, data.ID)
	if err != nil {
		return v.error(err, "upsertJSON-001")
	}

	if data.Embedding == nil && data.Metadata == nil {
		return v.error(errors.New("empty vector and fields"), "upsertJSON-002", data.ID)
	}

	if data.Embedding != nil && data.Metadata != nil {
		json, err := json.Marshal(data)
		if err != nil {
			return v.error(err, "upsertJSON-003", "fail to marshal JSON document", data.ID)
		}

		if err := v.client.Do(ctx, "JSON.SET", dataKey, "$", string(json)).Err(); err != nil {
			return v.error(err, "upsertJSON-004", "fail to insert JSON document", data.ID)
		}

		return nil
	}

	if data.Embedding != nil {
		json, err := json.Marshal(data)
		if err != nil {
			return v.error(err, "upsertJSON-005", "fail to marshal JSON vector", data.ID)
		}

		result, err := v.client.Do(ctx, "JSON.SET", dataKey, "$", string(json), "NX").Result()
		if err != nil {
			return v.error(err, "upsertJSON-006", "fail to insert JSON vector", data.ID)
		}

		if result == nil {
			if err := v.client.Do(ctx, "JSON.SET", dataKey, "$.vector", string(json)).Err(); err != nil {
				return v.error(err, "upsertJSON-007", "fail to update vector", data.ID)
			}
		}
		return nil
	}

	if data.Metadata != nil {
		json, err := json.Marshal(data)
		if err != nil {
			return v.error(err, "upsertJSON-008", "fail to marshal JSON fields", data.ID)
		}

		result, err := v.client.Do(ctx, "JSON.SET", dataKey, "$", string(json), "NX").Result()
		if err != nil {
			return v.error(err, "upsertJSON-009", "fail to insert JSON fields", data.ID)
		}

		if result == nil {
			if err := v.client.Do(ctx, "JSON.SET", dataKey, "$.metadata", string(json)).Err(); err != nil {
				return v.error(err, "upsertJSON-010", "fail to update fields", data.ID)
			}
		}
		return nil
	}

	return nil
}

func (v *VectorDBConnector) upsertHash(ctx context.Context, collection string, data vectordb.Data) error {
	hashKey, err := createDataKey(collection, data.ID)
	if err != nil {
		return v.error(err, "upsertHash-001")
	}

	hashValues := make(map[string]interface{})

	if data.Embedding != nil {
		vectorHash, err := createVectorHash(data)
		if err != nil {
			return v.error(err, "upsertHash-002", "fail to create vectorHash", data.ID)
		}

		for key, value := range vectorHash {
			hashValues[key] = value
		}
	}

	if data.Metadata != nil {
		fieldHashes, err := createMetadataHash(data)
		if err != nil {
			return v.error(err, "upsertHash-003", "fail to create fieldHashes", data.ID)
		}

		for key, value := range fieldHashes {
			hashValues[key] = value
		}
	}

	if len(hashValues) == 0 {
		return v.error(errors.New("empty vector and fields"), "upsertHash-004", data.ID)
	}

	if err := v.client.HSet(ctx, hashKey, hashValues).Err(); err != nil {
		return v.error(err, "upsertHash-005", "fail to insert hash", data.ID)
	}
	return nil
}

func (v *VectorDBConnector) Delete(ctx context.Context, collection string, dataID string) error {
	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return v.error(err, "Delete-001", "fail to create dataKey", dataID)
	}

	if err := v.client.Del(ctx, dataKey).Err(); err != nil {
		return v.error(err, "Delete-002", "fail to delete data", dataID)
	}
	return nil
}

func (v *VectorDBConnector) GetByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	indexType := v.collections[collection].On

	switch indexType {
	case HASH_TYPE:
		return v.getHashByID(ctx, collection, dataID)
	case JSON_TYPE:
		return v.getJSONByID(ctx, collection, dataID)
	default:
		return vectordb.Data{}, fmt.Errorf("unsupported index type: %s", indexType)
	}
}

func (v *VectorDBConnector) getHashByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return vectordb.Data{}, v.error(err, "GetByID-001", "fail to create vectorKey", dataID)
	}

	fields, err := v.client.HGetAll(ctx, dataKey).Result()
	if err != nil {
		return vectordb.Data{}, v.error(err, "GetByID-002", "fail to get vector", dataID)
	}

	if len(fields) == 0 {
		return vectordb.Data{}, v.error(fmt.Errorf("vector not found"), "GetByID-003", dataID)
	}

	var vectorFloat32 []float32
	if vectorBytes, ok := fields["vector"]; ok {
		vectorFloat32, err = convertVectorBytetoVectorFloat32([]byte(vectorBytes))
		if err != nil {
			return vectordb.Data{}, v.error(err, "GetByID-005", "fail to decode vector", dataID)
		}
	}

	metadata := make(vectordb.MetadataMap, len(fields)-1)
	for field, value := range fields {
		if field == "vector" {
			continue
		}
		metadata[field] = value
	}

	return vectordb.Data{
		ID:        dataID,
		Embedding: vectorFloat32,
		Metadata:  metadata,
	}, nil
}

func (v *VectorDBConnector) getJSONByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return vectordb.Data{}, v.error(err, "GetByID-001", "fail to create vectorKey", dataID)
	}

	result, err := v.client.Do(ctx, "JSON.GET", dataKey).Text()
	if err != nil {
		return vectordb.Data{}, v.error(err, "GetByID-002", "fail to get vector", dataID)
	}

	if result == "" {
		return vectordb.Data{}, v.error(fmt.Errorf("vector not found"), "GetByID-003", dataID)
	}

	var data vectordb.Data
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return vectordb.Data{}, v.error(err, "GetByID-004", "fail to decode JSON data", dataID)
	}

	return data, nil
}

func (v *VectorDBConnector) CountVector(ctx context.Context, collection string) (int, error) {
	if collection == "" {
		return 0, v.error(errors.New("collection cannot be empty"), "CountVector-001", collection)
	}

	result, err := v.client.Do(ctx, "FT.SEARCH", collection, "*", "LIMIT", 0, 0).Result()

	if err != nil {
		return 0, v.error(err, "CountVector-002", collection)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) == 0 {
		return 0, v.error(errors.New("invalid FT.SEARCH response"), "CountVector-003", collection)
	}

	count, ok := resultSlice[0].(int64)
	if !ok {
		return 0, v.error(errors.New("invalid vector count"), "CountVector-004", collection)
	}

	return int(count), nil
}

func (v *VectorDBConnector) FlushAll(ctx context.Context) error {
	if err := v.client.FlushAll(ctx).Err(); err != nil {
		return v.error(err, "FlushAll-001", "failed to flush all keys")
	}
	return nil
}

func (v *VectorDBConnector) Ping(ctx context.Context) error {
	if err := v.client.Ping(ctx).Err(); err != nil {
		return v.error(err, "Ping-001")
	}
	return nil
}

func (v *VectorDBConnector) Close() error {
	err := v.client.Close()
	if err != nil {
		return v.error(err, "Close-001", "failed to close redis vector client")
	}
	return nil
}
