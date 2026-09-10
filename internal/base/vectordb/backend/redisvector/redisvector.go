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

type RedisVectorConnector struct {
	client      *redis.Client
	collections map[string]CollectionConfig
}

func NewRedisVectorConnector(config VectorDBClientConfig) (*RedisVectorConnector, error) {
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

	return &RedisVectorConnector{
		client:      client,
		collections: collections,
	}, nil
}

func (r *RedisVectorConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("RedisVectorConnector.(%v)(%v) %w", method, params, err)
}

func (r *RedisVectorConnector) AddCollectionConfig(config CollectionConfig) error {
	if err := config.Validate(); err != nil {
		return r.error(err, "AddCollectionConfig-001", "fail to validate new collection config")
	}

	if r.collections == nil {
		r.collections = make(map[string]CollectionConfig)
	}

	if _, exists := r.collections[config.Name]; exists {
		return r.error(fmt.Errorf("collection %q already exists", config.Name), "AddCollectionConfig-002")
	}

	r.collections[config.Name] = config

	return nil
}

func (r *RedisVectorConnector) CreateCollection(ctx context.Context, name string) error {
	config, isExists := r.collections[name]
	if !isExists {
		return r.error(fmt.Errorf("collection %s not found", name), "CreateCollection-001")
	}

	creationQuery := config.CreateCollectionQuery()

	_, err := r.client.Do(ctx, creationQuery...).Result()
	if err != nil {
		return r.error(err, "CreateCollection-002", "fail to create collection "+name)
	}

	return nil
}

func (r *RedisVectorConnector) DeleteCollection(ctx context.Context, name string) error {
	_, err := r.client.Do(ctx, "FT.DROPINDEX", name, "DD").Result()
	if err != nil {
		return r.error(err, "DeleteCollection", "fail to delete colleciton "+name)
	}
	return nil
}

func (r *RedisVectorConnector) IsCollectionExists(ctx context.Context, name string) (bool, error) {
	_, err := r.client.Do(ctx, "FT.INFO", name).Result()
	switch {
	case err == nil:
		return true, nil
	case strings.Contains(err.Error(), "unknown Index name"):
		return false, nil
	default:
		return false, r.error(err, "IsCollectionExists", name)
	}
}

func (f *RedisVectorConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {
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

func (r *RedisVectorConnector) upsertJSON(ctx context.Context, collection string, data vectordb.Data) error {
	dataKey, err := createDataKey(collection, data.ID)
	if err != nil {
		return r.error(err, "upsertJSON-001")
	}

	if data.Vector == nil && data.Fields == nil {
		return r.error(errors.New("empty vector and fields"), "upsertJSON-002", data.ID)
	}

	if data.Vector != nil && data.Fields != nil {
		json, err := json.Marshal(data)
		if err != nil {
			return r.error(err, "upsertJSON-003", "fail to marshal JSON document", data.ID)
		}

		if err := r.client.Do(ctx, "JSON.SET", dataKey, "$", string(json)).Err(); err != nil {
			return r.error(err, "upsertJSON-004", "fail to insert JSON document", data.ID)
		}

		return nil
	}

	if data.Vector != nil {
		json, err := json.Marshal(data)
		if err != nil {
			return r.error(err, "upsertJSON-005", "fail to marshal JSON vector", data.ID)
		}

		result, err := r.client.Do(ctx, "JSON.SET", dataKey, "$", string(json), "NX").Result()
		if err != nil {
			return r.error(err, "upsertJSON-006", "fail to insert JSON vector", data.ID)
		}

		if result == nil {
			if err := r.client.Do(ctx, "JSON.SET", dataKey, "$.vector", string(json)).Err(); err != nil {
				return r.error(err, "upsertJSON-007", "fail to update vector", data.ID)
			}
		}
		return nil
	}

	if data.Fields != nil {
		json, err := json.Marshal(data)
		if err != nil {
			return r.error(err, "upsertJSON-008", "fail to marshal JSON fields", data.ID)
		}

		result, err := r.client.Do(ctx, "JSON.SET", dataKey, "$", string(json), "NX").Result()
		if err != nil {
			return r.error(err, "upsertJSON-009", "fail to insert JSON fields", data.ID)
		}

		if result == nil {
			if err := r.client.Do(ctx, "JSON.SET", dataKey, "$.metadata", string(json)).Err(); err != nil {
				return r.error(err, "upsertJSON-010", "fail to update fields", data.ID)
			}
		}
		return nil
	}

	return nil
}

func (r *RedisVectorConnector) upsertHash(ctx context.Context, collection string, data vectordb.Data) error {
	hashKey, err := createDataKey(collection, data.ID)
	if err != nil {
		return r.error(err, "upsertHash-001")
	}

	hashValues := make(map[string]interface{})

	if data.Vector != nil {
		vectorHash, err := createVectorHash(data)
		if err != nil {
			return r.error(err, "upsertHash-002", "fail to create vectorHash", data.ID)
		}

		for key, value := range vectorHash {
			hashValues[key] = value
		}
	}

	if data.Fields != nil {
		fieldHashes, err := createMetadataHash(data)
		if err != nil {
			return r.error(err, "upsertHash-003", "fail to create fieldHashes", data.ID)
		}

		for key, value := range fieldHashes {
			hashValues[key] = value
		}
	}

	if len(hashValues) == 0 {
		return r.error(errors.New("empty vector and fields"), "upsertHash-004", data.ID)
	}

	if err := r.client.HSet(ctx, hashKey, hashValues).Err(); err != nil {
		return r.error(err, "upsertHash-005", "fail to insert hash", data.ID)
	}
	return nil
}

func (r *RedisVectorConnector) Delete(ctx context.Context, collection string, dataID string) error {
	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return r.error(err, "Delete-001", "fail to create dataKey", dataID)
	}

	if err := r.client.Del(ctx, dataKey).Err(); err != nil {
		return r.error(err, "Delete-002", "fail to delete data", dataID)
	}
	return nil
}

func (r *RedisVectorConnector) GetByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	indexType := r.collections[collection].On

	switch indexType {
	case HASH_TYPE:
		return r.getHashByID(ctx, collection, dataID)
	case JSON_TYPE:
		return r.getJSONByID(ctx, collection, dataID)
	default:
		return vectordb.Data{}, fmt.Errorf("unsupported index type: %s", indexType)
	}
}

func (r *RedisVectorConnector) getHashByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return vectordb.Data{}, r.error(err, "GetByID-001", "fail to create vectorKey", dataID)
	}

	fields, err := r.client.HGetAll(ctx, dataKey).Result()
	if err != nil {
		return vectordb.Data{}, r.error(err, "GetByID-002", "fail to get vector", dataID)
	}

	if len(fields) == 0 {
		return vectordb.Data{}, r.error(fmt.Errorf("vector not found"), "GetByID-003", dataID)
	}

	var vectorFloat32 []float32
	if vectorBytes, ok := fields["vector"]; ok {
		vectorFloat32, err = convertVectorBytetoVectorFloat32([]byte(vectorBytes))
		if err != nil {
			return vectordb.Data{}, r.error(err, "GetByID-005", "fail to decode vector", dataID)
		}
	}

	metadata := make(vectordb.Metadata, len(fields)-1)
	for field, value := range fields {
		if field == "vector" {
			continue
		}
		metadata[field] = value
	}

	return vectordb.Data{
		ID:     dataID,
		Vector: vectorFloat32,
		Fields: metadata,
	}, nil
}

func (r *RedisVectorConnector) getJSONByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return vectordb.Data{}, r.error(err, "GetByID-001", "fail to create vectorKey", dataID)
	}

	result, err := r.client.Do(ctx, "JSON.GET", dataKey).Text()
	if err != nil {
		return vectordb.Data{}, r.error(err, "GetByID-002", "fail to get vector", dataID)
	}

	if result == "" {
		return vectordb.Data{}, r.error(fmt.Errorf("vector not found"), "GetByID-003", dataID)
	}

	var data vectordb.Data
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return vectordb.Data{}, r.error(err, "GetByID-004", "fail to decode JSON data", dataID)
	}

	return data, nil
}

func (r *RedisVectorConnector) CountVector(ctx context.Context, collection string) (int, error) {
	if collection == "" {
		return 0, r.error(errors.New("collection cannot be empty"), "CountVector-001", collection)
	}

	result, err := r.client.Do(ctx, "FT.SEARCH", collection, "*", "LIMIT", 0, 0).Result()

	if err != nil {
		return 0, r.error(err, "CountVector-002", collection)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) == 0 {
		return 0, r.error(errors.New("invalid FT.SEARCH response"), "CountVector-003", collection)
	}

	count, ok := resultSlice[0].(int64)
	if !ok {
		return 0, r.error(errors.New("invalid vector count"), "CountVector-004", collection)
	}

	return int(count), nil
}

func (r *RedisVectorConnector) FlushAll(ctx context.Context) error {
	if err := r.client.FlushAll(ctx).Err(); err != nil {
		return r.error(err, "FlushAll-001", "failed to flush all keys")
	}
	return nil
}

func (r *RedisVectorConnector) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return r.error(err, "Ping-001")
	}
	return nil
}

func (r *RedisVectorConnector) Close() error {
	err := r.client.Close()
	if err != nil {
		return r.error(err, "Close-001", "failed to close redis vector client")
	}
	return nil
}
