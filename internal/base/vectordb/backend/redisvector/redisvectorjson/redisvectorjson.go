package redisvectorjson

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
	client *redis.Client
}

func NewVectorDBConnector(config VectorDBConnectorConfig) (*VectorDBConnector, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Password,
		DB:       int(config.Db),
	})

	return &VectorDBConnector{
		client: client,
	}, nil
}

func (v *VectorDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("VectorDBConnector.(%v)(%v) %w", method, params, err)
}

func (v *VectorDBConnector) Ping(ctx context.Context) error {
	if err := v.client.Ping(ctx).Err(); err != nil {
		return v.error(err, "Ping-001")
	}
	return nil
}

func (v *VectorDBConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {
	creationQuery := collectionConfig.CreateCollectionQuery()

	_, err := v.client.Do(ctx, creationQuery...).Result()
	if err != nil {
		return v.error(err,
			"CreateCollection-001",
			fmt.Sprintf("fail to create collection: %v", collectionConfig.Name))
	}

	return nil
}

func (v *VectorDBConnector) IsCollectionExists(ctx context.Context, collection string) (bool, error) {
	_, err := v.client.Do(ctx, "FT.INFO", collection).Result()
	switch {
	case err == nil:
		return true, nil
	case strings.Contains(err.Error(), "Unknown index name"):
		return false, nil
	default:
		return false, v.error(err,
			"IsCollectionExists-001",
			fmt.Sprintf("error when executing FT.INFO for collection:%v", collection))
	}
}

func (v *VectorDBConnector) FlushCollection(ctx context.Context, collection string) error {
	if collection == "" {
		return v.error(errors.New("collection cannot be empty"), "FlushCollection-001", collection)
	}

	const batchSize = 1000
	var cursor uint64
	pattern := collection + ":*"

	for {
		keys, nextCursor, err := v.client.Scan(ctx, cursor, pattern, batchSize).Result()
		if err != nil {
			return v.error(err,
				"FlushCollection-002",
				fmt.Sprintf("failed to scan collection: %v", collection))
		}

		if len(keys) > 0 {
			if err := v.client.Unlink(ctx, keys...).Err(); err != nil {
				return v.error(err,
					"FlushCollection-003",
					fmt.Sprintf("failed to delete collection data: %v", collection))
			}
		}

		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return nil
}

func (v *VectorDBConnector) DeleteCollection(ctx context.Context, collection string) error {
	_, err := v.client.Do(ctx, "FT.DROPINDEX", collection, "DD").Result()
	if err != nil {
		return v.error(err,
			"DeleteCollection-001",
			fmt.Sprintf("fail to delete colleciton: %v", collection))
	}
	return nil
}

func (v *VectorDBConnector) FlushAll(ctx context.Context) error {
	if err := v.client.FlushAll(ctx).Err(); err != nil {
		return v.error(err, "FlushAll-001", "failed to flush all keys")
	}
	return nil
}

func (f *VectorDBConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {

	dataKey, err := createDataKey(collection, data.ID)
	if err != nil {
		return f.error(err, "Upsert-001")
	}

	// update both embedding and metadata
	if data.Embedding != nil && data.Metadata != nil {
		json, err := json.Marshal(data)
		if err != nil {
			return f.error(err, "Upsert-002", "fail to marshal JSON document", data.ID)
		}

		if err := f.client.Do(ctx, "JSON.SET", dataKey, "$", string(json)).Err(); err != nil {
			return f.error(err, "Upsert-003", "fail to insert JSON document", data.ID)
		}

		return nil
	}

	// if data is incomplete check for existence of the key in the database
	isExists, err := f.IsKeyExists(ctx, collection, data.ID)
	if err != nil {
		return f.error(err, "Upsert-002", "fail to check key existence", data.ID)
	}

	switch {
	case isExists && data.Embedding != nil:
		// update embedding only
		embeddingJSON, err := json.Marshal(data.Embedding)
		if err != nil {
			return f.error(err,
				"Upsert-003",
				fmt.Sprintf("fail to marshal JSON embedding: %v", data.ID))
		}

		if err := f.client.Do(ctx,
			"JSON.SET",
			dataKey,
			fmt.Sprintf("$.%s", vectordb.EmbeddingColumnName),
			string(embeddingJSON)).Err(); err != nil {
			return f.error(err,
				"Upsert-004",
				fmt.Sprintf("fail to update JSON embedding: %v", data.ID))
		}

	case isExists && data.Metadata != nil:
		// update metadata only
		metadataJSON, err := json.Marshal(data.Metadata)
		if err != nil {
			return f.error(err,
				"Upsert-005",
				fmt.Sprintf("fail to marshal JSON metadata: %v", data.ID))
		}

		if err := f.client.Do(ctx,
			"JSON.SET",
			dataKey,
			fmt.Sprintf("$.%s", vectordb.MetadataColumnName),
			string(metadataJSON)).Err(); err != nil {
			return f.error(err,
				"Upsert-006",
				fmt.Sprintf("fail to update JSON metadata: %v", data.ID))
		}

	case !isExists && data.Embedding != nil:
		// insert embedding only
		jsonData, err := json.Marshal(data)
		if err != nil {
			return f.error(err,
				"Upsert-007",
				fmt.Sprintf("fail to marshal JSON document: %v", data.ID))
		}

		if err := f.client.Do(ctx,
			"JSON.SET",
			dataKey,
			"$",
			string(jsonData),
			"NX").Err(); err != nil {
			return f.error(err,
				"Upsert-008",
				fmt.Sprintf("fail to insert JSON document: %v", data.ID))
		}

	case !isExists && data.Metadata != nil:
		// insert metadata only
		jsonData, err := json.Marshal(data)
		if err != nil {
			return f.error(err,
				"Upsert-009",
				fmt.Sprintf("fail to marshal JSON document: %v", data.ID))
		}

		if err := f.client.Do(ctx,
			"JSON.SET",
			dataKey,
			"$",
			string(jsonData),
			"NX").Err(); err != nil {
			return f.error(err,
				"Upsert-010",
				fmt.Sprintf("fail to insert JSON document: %v", data.ID))
		}
	}

	return nil
}

func (v *VectorDBConnector) Delete(ctx context.Context, collection string, dataID string) error {
	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return v.error(err,
			"Delete-001",
			fmt.Sprintf("fail to create dataKey: %v", dataID))
	}

	if err := v.client.Del(ctx, dataKey).Err(); err != nil {
		return v.error(err,
			"Delete-002",
			fmt.Sprintf("fail to delete data: %v", dataID))
	}
	return nil
}

func (v *VectorDBConnector) UpsertBatch(ctx context.Context, collection string, arrayData []vectordb.Data, batchNumber int) error {

	if batchNumber <= 0 {
		return v.error(
			errors.New("batchNumber must be greater than 0"),
			"UpsertBatch-001",
			"invalid batch number")
	}

	for start := 0; start < len(arrayData); start += batchNumber {

		if err := ctx.Err(); err != nil {
			return err
		}

		end := start + batchNumber
		if end > len(arrayData) {
			end = len(arrayData)
		}

		pipe := v.client.Pipeline()

		dataMap := make(map[string]vectordb.Data, len(arrayData[start:end]))
		dataKeys := make([]string, len(arrayData[start:end]))
		for index, data := range arrayData[start:end] {
			dataKey, err := createDataKey(collection, data.ID)
			if err != nil {
				return v.error(err,
					"UpsertBatch-001",
					fmt.Sprintf("fail to create dataKey: %v", data.ID))
			}
			dataMap[dataKey] = data
			dataKeys[index] = dataKey
		}

		existenceMap, err := v.IsKeyExistsBatch(ctx, collection, dataKeys)
		if err != nil {
			return v.error(err,
				"UpsertBatch-002",
				"fail to check key existence")
		}

		for dataKey, isExists := range existenceMap {
			switch {
			case dataMap[dataKey].Embedding != nil && dataMap[dataKey].Metadata != nil:
				jsonData, err := json.Marshal(dataMap[dataKey])
				if err != nil {
					return v.error(err,
						"UpsertBatch-003",
						fmt.Sprintf("fail to marshal JSON document: %v", dataMap[dataKey].ID))
				}

				if err := pipe.Do(ctx,
					"JSON.SET",
					dataKey,
					"$",
					string(jsonData)).Err(); err != nil {
					return v.error(err,
						"UpsertBatch-004",
						fmt.Sprintf("fail to insert JSON document: %v", dataMap[dataKey].ID))
				}

			case isExists && dataMap[dataKey].Embedding != nil:
				embeddingJSON, err := json.Marshal(dataMap[dataKey].Embedding)
				if err != nil {
					return v.error(err,
						"UpsertBatch-005",
						fmt.Sprintf("fail to marshal JSON embedding: %v", dataMap[dataKey].ID))
				}

				if err := pipe.Do(ctx,
					"JSON.SET",
					dataKey,
					fmt.Sprintf("$.%s", vectordb.EmbeddingColumnName),
					string(embeddingJSON)).Err(); err != nil {
					return v.error(err,
						"UpsertBatch-006",
						fmt.Sprintf("fail to update JSON embedding: %v", dataMap[dataKey].ID))
				}

			case isExists && dataMap[dataKey].Metadata != nil:
				metadataJSON, err := json.Marshal(dataMap[dataKey].Metadata)
				if err != nil {
					return v.error(err,
						"UpsertBatch-007",
						fmt.Sprintf("fail to marshal JSON metadata: %v", dataMap[dataKey].ID))
				}

				if err := pipe.Do(ctx,
					"JSON.SET",
					dataKey,
					fmt.Sprintf("$.%s", vectordb.MetadataColumnName),
					string(metadataJSON)).Err(); err != nil {
					return v.error(err,
						"UpsertBatch-008",
						fmt.Sprintf("fail to update JSON metadata: %v", dataMap[dataKey].ID))
				}

			case !isExists && dataMap[dataKey].Embedding != nil:
				jsonData, err := json.Marshal(dataMap[dataKey])
				if err != nil {
					return v.error(err,
						"UpsertBatch-009",
						fmt.Sprintf("fail to marshal JSON document: %v", dataMap[dataKey].ID))
				}

				if err := pipe.Do(ctx,
					"JSON.SET",
					dataKey,
					"$",
					string(jsonData),
					"NX").Err(); err != nil {
					return v.error(err,
						"UpsertBatch-010",
						fmt.Sprintf("fail to insert JSON document: %v", dataMap[dataKey].ID))
				}

			case !isExists && dataMap[dataKey].Metadata != nil:
				jsonData, err := json.Marshal(dataMap[dataKey])
				if err != nil {
					return v.error(err,
						"UpsertBatch-011",
						fmt.Sprintf("fail to marshal JSON document: %v", dataMap[dataKey].ID))
				}

				if err := pipe.Do(ctx,
					"JSON.SET",
					dataKey,
					"$",
					string(jsonData),
					"NX").Err(); err != nil {
					return v.error(err,
						"UpsertBatch-012",
						fmt.Sprintf("fail to insert JSON document: %v", dataMap[dataKey].ID))
				}
			}
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			return v.error(err,
				"UpsertBatch-013",
				"fail to execute pipeline to insert data")
		}
	}

	return nil
}

func (v *VectorDBConnector) DeleteBatch(ctx context.Context, collection string, arrayID []string, batchNumber int) error {

	if batchNumber <= 0 {
		return v.error(
			errors.New("batch number must be greater than zero"),
			"DeleteBatch-001",
			"invalid batch number",
		)
	}

	for start := 0; start < len(arrayID); start += batchNumber {

		if err := ctx.Err(); err != nil {
			return err
		}

		end := start + batchNumber
		if end > len(arrayID) {
			end = len(arrayID)
		}

		pipe := v.client.Pipeline()

		if err := queueDelete(ctx, pipe, collection, arrayID[start:end]...); err != nil {
			return v.error(
				err,
				"DeleteBatch-002",
				fmt.Sprintf("fail to insert data into pipeline from index %d to %d", start, end),
			)
		}

		_, err := pipe.Exec(ctx)
		if err != nil {
			return v.error(err,
				"DeleteBatch-003",
				fmt.Sprintf("fail to delete batch data from index %d to %d", start, end),
			)
		}
	}

	return nil
}
func (v *VectorDBConnector) GetByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {

	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-001",
			fmt.Sprintf("fail to create dataKey: %v", dataID))
	}

	result, err := v.client.Do(ctx, "JSON.GET", dataKey).Text()
	if err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-002",
			fmt.Sprintf("fail to get data: %v", dataID))
	}

	if result == "" {
		return vectordb.Data{}, v.error(
			errors.New("data not found"),
			"GetByID-003",
			dataID)
	}

	var data vectordb.Data
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-004",
			fmt.Sprintf("fail to decode JSON data: %v", dataID))
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

func (v *VectorDBConnector) Close() error {
	err := v.client.Close()
	if err != nil {
		return v.error(err, "Close-001", "failed to close redis vector client")
	}
	return nil
}

func (v *VectorDBConnector) IsKeyExists(ctx context.Context, collection string, dataID string) (bool, error) {
	dataKey, err := createDataKey(collection, dataID)
	if err != nil {
		return false, v.error(err,
			"IsKeyExits-001",
			fmt.Sprintf("fail to create key: %v", dataID),
		)
	}

	exists, err := v.client.Exists(ctx, dataKey).Result()
	if err != nil {
		return false, v.error(err,
			"IsKeyExists-002",
			fmt.Sprintf("fail to check key existence: %v", dataID),
		)
	}

	return exists > 0, nil
}

func (v *VectorDBConnector) IsKeyExistsBatch(ctx context.Context, collection string, dataKeys []string) (map[string]bool, error) {

	pipe := v.client.Pipeline()

	commands := make([]*redis.IntCmd, len(dataKeys))
	for index, dataKey := range dataKeys {
		commands[index] = pipe.Exists(ctx, dataKey)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, v.error(err,
			"IsKeyExistsBatch-002",
			"fail to execute pipeline for key existence check",
		)
	}

	existenceMap := make(map[string]bool, len(dataKeys))
	for index, dataKey := range dataKeys {
		existenceMap[dataKey] = commands[index].Val() > 0
	}

	return existenceMap, nil
}
