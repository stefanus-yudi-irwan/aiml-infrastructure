package redisvectorhash

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

func (v *VectorDBConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {

	pipe := v.client.Pipeline()

	if err := queueUpsert(ctx, pipe, collection, data); err != nil {
		return v.error(err,
			"Upsert-001",
			fmt.Sprintf("fail to insert hash data into pipeline: %v", data.ID))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return v.error(err,
			"Upsert-002",
			fmt.Sprintf("fail to execute pipeline to insert data: %v", data.ID),
		)
	}

	return nil
}

func (v *VectorDBConnector) Delete(ctx context.Context, collection string, id string) error {

	pipe := v.client.Pipeline()

	if err := queueDelete(ctx, pipe, collection, id); err != nil {
		return v.error(err,
			"Delete-001",
			fmt.Sprintf("fail to insert hash data into pipeline: %v", id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return v.error(err,
			"Delete-002",
			fmt.Sprintf("fail to execute pipeline to delete data: %v", id),
		)
	}

	return nil
}

func (v *VectorDBConnector) UpsertBatch(ctx context.Context, collection string, arrayData []vectordb.Data, batchNumber int) error {

	if batchNumber <= 0 {
		return v.error(
			errors.New("batch number must be greater than zero"),
			"UpsertBatch-001",
			"invalid batch number",
		)
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

		for _, data := range arrayData[start:end] {
			if err := queueUpsert(ctx, pipe, collection, data); err != nil {
				return v.error(
					err,
					"UpsertBatch-002",
					fmt.Sprintf("fail to insert data into pipeline: %v", data.ID),
				)
			}
		}

		_, err := pipe.Exec(ctx)
		if err != nil {
			return v.error(err,
				"UpsertBatch-003",
				fmt.Sprintf("fail to insert batch data from index %d to %d", start, end),
			)
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

func (v *VectorDBConnector) GetByID(ctx context.Context, collection string, id string) (vectordb.Data, error) {
	dataKey, err := createDataKey(collection, id)
	if err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-001",
			fmt.Sprintf("fail to create vectorKey: %v", id))
	}

	fields, err := v.client.HGetAll(ctx, dataKey).Result()
	if err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-002",
			fmt.Sprintf("fail to get vector: %v", id))
	}

	if len(fields) == 0 {
		return vectordb.Data{}, v.error(fmt.Errorf("data not found"),
			"GetByID-003",
			fmt.Sprintf("data not found: %v", id))
	}

	var vectorFloat32 []float32
	if vectorBytes, ok := fields[vectordb.EmbeddingColumnName]; ok {
		vectorFloat32, err = convertEmbeddingBytetoEmbeddingFloat32([]byte(vectorBytes))
		if err != nil {
			return vectordb.Data{}, v.error(err,
				"GetByID-004",
				fmt.Sprintf("fail to decode vector: %v", id))
		}
	}

	var metadataMap map[string]interface{}
	if metadata, ok := fields[vectordb.MetadataColumnName]; ok {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			return vectordb.Data{}, v.error(err,
				"GetByID-005",
				fmt.Sprintf("fail to unmarshall hash metadata: %v", id))
		}
	}

	return vectordb.Data{
		ID:         fields[vectordb.IDColumnName],
		DocumentID: fields[vectordb.DocumentIDColumnName],
		ChunkID:    fields[vectordb.ChunkIDColumnName],
		Content:    fields[vectordb.ContentColumnName],
		Embedding:  vectorFloat32,
		Metadata:   metadataMap,
	}, nil
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
		return v.error(err,
			"Close-001",
			"failed to close vector db client")
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
