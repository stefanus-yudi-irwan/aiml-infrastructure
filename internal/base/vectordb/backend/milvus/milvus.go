package milvus

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"encoding/json"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

type milvusData struct {
	ID       string    `milvus:"name:id;PRIMARY_KEY"`
	Vector   []float32 `milvus:"name:vector;DIM:128"`
	Metadata []byte    `milvus:"name:metadata"`
}
type VectorDBConnector struct {
	client *milvusclient.Client
}

func NewVectorDBConnector(ctx context.Context, config VectorDBConnectorConfig) (*VectorDBConnector, error) {
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: config.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus client: %w", err)
	}

	return &VectorDBConnector{
		client: client,
	}, nil
}

func (v *VectorDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("VectorDBConnector.(%v)(%v) %w", method, params, err)
}

func (v *VectorDBConnector) CreateCollection(ctx context.Context, config CollectionConfig) error {
	collectionOption := createCollectionOption(config)
	if err := v.client.CreateCollection(ctx, collectionOption); err != nil {
		return v.error(err, "CreateCollection-001", fmt.Sprintf("failed to create collection %q", config.Name))
	}

	return nil
}

func (v *VectorDBConnector) CreateVectorIndex(ctx context.Context, collectionName string, vectorConfig FieldConfig) error {
	switch vectorConfig.VectorConfig.Algorithm {
	case HNSW:
		return v.CreateHNSWIndex(ctx, collectionName, vectorConfig)
	default:
		return fmt.Errorf("unsupported index algorithm: %s", vectorConfig.VectorConfig.Algorithm)
	}
}

func (v *VectorDBConnector) CreateHNSWIndex(ctx context.Context, collectionName string, vectorConfig FieldConfig) error {
	hnswIndex := index.NewHNSWIndex(
		vectorConfig.VectorConfig.Metric,
		vectorConfig.VectorConfig.HNSW.M,
		vectorConfig.VectorConfig.HNSW.EFConstruction,
	)

	task, err := v.client.CreateIndex(ctx,
		milvusclient.NewCreateIndexOption(collectionName, vectorConfig.Name, hnswIndex).
			WithIndexName(fmt.Sprintf("%s_%s_hnsw_index", collectionName, vectorConfig.Name)),
	)
	if err != nil {
		return v.error(err, "CreateHNSWIndex-001", fmt.Sprintf("failed to create HNSW index for collection %q and field %q", collectionName, vectorConfig.Name))
	}

	if err := task.Await(ctx); err != nil {
		return v.error(err, "CreateHNSWIndex-002", fmt.Sprintf("failed to build index for collection %q and field %q", collectionName, vectorConfig.Name))
	}

	return nil
}

func (v *VectorDBConnector) LoadCollection(ctx context.Context, collectionName string) error {
	task, err := v.client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collectionName))
	if err != nil {
		return v.error(err, "LoadCollection-001", fmt.Sprintf("failed to load collection %q", collectionName))
	}

	if err := task.Await(ctx); err != nil {
		return v.error(err, "LoadCollection-002", fmt.Sprintf("failed to load collection %q", collectionName))
	}

	return nil
}

func (v *VectorDBConnector) DeleteCollection(ctx context.Context, collection string) error {
	err := v.client.DropCollection(ctx, milvusclient.NewDropCollectionOption(collection))
	if err != nil {
		return v.error(err, "DeleteCollection-001", fmt.Sprintf("failed to delete collection %q", collection))
	}
	return nil
}

func (v *VectorDBConnector) IsCollectionExists(ctx context.Context, collection string) (bool, error) {
	result, err := v.client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collection))
	if err != nil {
		return false, v.error(err, "IsCollectionExists-001", fmt.Sprintf("failed to check if collection %q exists", collection))
	}
	return result, nil
}

func (v *VectorDBConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {
	row := map[string]interface{}{
		"id":       data.ID,
		"vector":   data.Embedding,
		"metadata": data.Metadata,
	}

	_, err := v.client.Insert(ctx, milvusclient.NewRowBasedInsertOption(collection, row))
	if err != nil {
		return v.error(err, "Upsert-001", fmt.Sprintf("failed to upsert data into collection %q", collection))
	}
	return nil
}

func (v *VectorDBConnector) FlushAll(ctx context.Context) error {
	collections, err := v.client.ListCollections(ctx, milvusclient.NewListCollectionOption())
	if err != nil {
		return v.error(err, "FlushAll-001")
	}

	for _, collection := range collections {
		err := v.client.DropCollection(ctx, milvusclient.NewDropCollectionOption(collection))
		if err != nil {
			return v.error(err, "FlushAll-002", fmt.Sprintf("failed to drop collection %q", collection))
		}
	}
	return nil
}

func (v *VectorDBConnector) GetByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	queryResult, err := v.client.Query(ctx, milvusclient.NewQueryOption(collection).
		WithFilter(fmt.Sprintf(`id == "%s"`, dataID)).
		WithOutputFields("*"),
	)
	if err != nil {
		return vectordb.Data{}, v.error(err, "GetByID-001", fmt.Sprintf("failed to query entity %q from collection %q", dataID, collection))
	}

	if queryResult.Len() == 0 {
		return vectordb.Data{}, v.error(nil, "GetByID-002", fmt.Sprintf("entity %q not found in collection %q", dataID, collection))
	}

	var rows []*milvusData
	if err := queryResult.Unmarshal(&rows); err != nil {
		return vectordb.Data{}, v.error(err, "GetByID-003", fmt.Sprintf("failed to unmarshal query result for entity %q from collection %q", dataID, collection))
	}

	if len(rows) == 0 {
		return vectordb.Data{}, v.error(nil, "GetByID-004", fmt.Sprintf("entity %q not found in collection %q", dataID, collection))
	}

	var metadata map[string]interface{}

	if err := json.Unmarshal(rows[0].Metadata, &metadata); err != nil {
		return vectordb.Data{}, v.error(err, "GetByID-005", fmt.Sprintf("failed to unmarshal metadata for entity %q", dataID))
	}

	return vectordb.Data{
		ID:        rows[0].ID,
		Embedding: rows[0].Vector,
		Metadata:  metadata,
	}, nil
}

func (v *VectorDBConnector) Delete(ctx context.Context, collection string, dataID string) error {
	_, err := v.client.Delete(
		ctx,
		milvusclient.NewDeleteOption(collection).
			WithExpr(fmt.Sprintf(`id == "%s"`, dataID)),
	)

	if err != nil {
		return v.error(err, "Delete-001",
			fmt.Sprintf("failed to delete entity %s from collection %s", dataID, collection))
	}
	return nil
}

func (v *VectorDBConnector) CountVector(ctx context.Context, collection string) (int64, error) {

	result, err := v.client.Query(ctx, milvusclient.NewQueryOption(collection).
		WithFilter("").
		WithOutputFields("count(*)"),
	)
	if err != nil {
		return 0, v.error(err, "failed to count vectors in collection: %s", collection)
	}

	countColumn := result.GetColumn("count(*)")
	if countColumn == nil {
		return 0, v.error(err, "count(*) column not found in query result for collection: %q", collection)
	}

	if countColumn.Len() == 0 {
		return 0, v.error(err, "count(*) returned no result for collection: %q", collection)
	}

	count, err := countColumn.GetAsInt64(0)
	if err != nil {
		return 0, v.error(err, "failed to extract count(*) for collection %q: %w", collection)
	}

	return count, nil
}

func (v *VectorDBConnector) Close(ctx context.Context) error {
	return v.client.Close(ctx)
}

func (v *VectorDBConnector) ReleaseCollection(ctx context.Context, collection string) error {
	err := v.client.ReleaseCollection(ctx, milvusclient.NewReleaseCollectionOption(collection))
	if err != nil {
		return v.error(err, "ReleaseCollection-001", fmt.Sprintf("failed to release collection %q", collection))
	}
	return nil
}
