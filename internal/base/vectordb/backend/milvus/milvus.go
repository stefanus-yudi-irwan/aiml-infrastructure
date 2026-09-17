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
type MilvusConnector struct {
	client *milvusclient.Client
}

func NewMilvusConnector(ctx context.Context, config MilvusConnectorConfig) (*MilvusConnector, error) {
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: config.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Milvus client: %w", err)
	}

	return &MilvusConnector{
		client: client,
	}, nil
}

func (m *MilvusConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("MilvusConnector.(%v)(%v) %w", method, params, err)
}

func (m *MilvusConnector) CreateCollection(ctx context.Context, config CollectionConfig) error {
	collectionOption := createCollectionOption(config)
	if err := m.client.CreateCollection(ctx, collectionOption); err != nil {
		return m.error(err, "CreateCollection-001", fmt.Sprintf("failed to create collection %q", config.Name))
	}

	return nil
}

func (m *MilvusConnector) CreateVectorIndex(ctx context.Context, collectionName string, vectorConfig FieldConfig) error {
	switch vectorConfig.VectorConfig.Algorithm {
	case HNSW:
		return m.CreateHNSWIndex(ctx, collectionName, vectorConfig)
	default:
		return fmt.Errorf("unsupported index algorithm: %s", vectorConfig.VectorConfig.Algorithm)
	}
}

func (m *MilvusConnector) CreateHNSWIndex(ctx context.Context, collectionName string, vectorConfig FieldConfig) error {
	hnswIndex := index.NewHNSWIndex(
		vectorConfig.VectorConfig.Metric,
		vectorConfig.VectorConfig.HNSW.M,
		vectorConfig.VectorConfig.HNSW.EFConstruction,
	)

	task, err := m.client.CreateIndex(ctx,
		milvusclient.NewCreateIndexOption(collectionName, vectorConfig.Name, hnswIndex).
			WithIndexName(fmt.Sprintf("%s_%s_hnsw_index", collectionName, vectorConfig.Name)),
	)
	if err != nil {
		return m.error(err, "CreateHNSWIndex-001", fmt.Sprintf("failed to create HNSW index for collection %q and field %q", collectionName, vectorConfig.Name))
	}

	if err := task.Await(ctx); err != nil {
		return m.error(err, "CreateHNSWIndex-002", fmt.Sprintf("failed to build index for collection %q and field %q", collectionName, vectorConfig.Name))
	}

	return nil
}

func (m *MilvusConnector) LoadCollection(ctx context.Context, collectionName string) error {
	task, err := m.client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collectionName))
	if err != nil {
		return m.error(err, "LoadCollection-001", fmt.Sprintf("failed to load collection %q", collectionName))
	}

	if err := task.Await(ctx); err != nil {
		return m.error(err, "LoadCollection-002", fmt.Sprintf("failed to load collection %q", collectionName))
	}

	return nil
}

func (m *MilvusConnector) DeleteCollection(ctx context.Context, collection string) error {
	err := m.client.DropCollection(ctx, milvusclient.NewDropCollectionOption(collection))
	if err != nil {
		return m.error(err, "DeleteCollection-001", fmt.Sprintf("failed to delete collection %q", collection))
	}
	return nil
}

func (m *MilvusConnector) IsCollectionExists(ctx context.Context, collection string) (bool, error) {
	result, err := m.client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collection))
	if err != nil {
		return false, m.error(err, "IsCollectionExists-001", fmt.Sprintf("failed to check if collection %q exists", collection))
	}
	return result, nil
}

func (m *MilvusConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {
	row := map[string]interface{}{
		"id":       data.ID,
		"vector":   data.Vector,
		"metadata": data.Fields,
	}

	_, err := m.client.Insert(ctx, milvusclient.NewRowBasedInsertOption(collection, row))
	if err != nil {
		return m.error(err, "Upsert-001", fmt.Sprintf("failed to upsert data into collection %q", collection))
	}
	return nil
}

func (m *MilvusConnector) FlushAll(ctx context.Context) error {
	collections, err := m.client.ListCollections(ctx, milvusclient.NewListCollectionOption())
	if err != nil {
		return m.error(err, "FlushAll-001")
	}

	for _, collection := range collections {
		err := m.client.DropCollection(ctx, milvusclient.NewDropCollectionOption(collection))
		if err != nil {
			return m.error(err, "FlushAll-002", fmt.Sprintf("failed to drop collection %q", collection))
		}
	}
	return nil
}

func (m *MilvusConnector) GetByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	queryResult, err := m.client.Query(ctx, milvusclient.NewQueryOption(collection).
		WithFilter(fmt.Sprintf(`id == "%s"`, dataID)).
		WithOutputFields("*"),
	)
	if err != nil {
		return vectordb.Data{}, m.error(err, "GetByID-001", fmt.Sprintf("failed to query entity %q from collection %q", dataID, collection))
	}

	if queryResult.Len() == 0 {
		return vectordb.Data{}, m.error(nil, "GetByID-002", fmt.Sprintf("entity %q not found in collection %q", dataID, collection))
	}

	var rows []*milvusData
	if err := queryResult.Unmarshal(&rows); err != nil {
		return vectordb.Data{}, m.error(err, "GetByID-003", fmt.Sprintf("failed to unmarshal query result for entity %q from collection %q", dataID, collection))
	}

	if len(rows) == 0 {
		return vectordb.Data{}, m.error(nil, "GetByID-004", fmt.Sprintf("entity %q not found in collection %q", dataID, collection))
	}

	var metadata map[string]interface{}

	if err := json.Unmarshal(rows[0].Metadata, &metadata); err != nil {
		return vectordb.Data{}, m.error(err, "GetByID-005", fmt.Sprintf("failed to unmarshal metadata for entity %q", dataID))
	}

	return vectordb.Data{
		ID:     rows[0].ID,
		Vector: rows[0].Vector,
		Fields: metadata,
	}, nil
}

func (m *MilvusConnector) Delete(ctx context.Context, collection string, dataID string) error {
	_, err := m.client.Delete(
		ctx,
		milvusclient.NewDeleteOption(collection).
			WithExpr(fmt.Sprintf(`id == "%s"`, dataID)),
	)

	if err != nil {
		return m.error(err, "Delete-001",
			fmt.Sprintf("failed to delete entity %s from collection %s", dataID, collection))
	}
	return nil
}

func (m *MilvusConnector) CountVector(ctx context.Context, collection string) (int64, error) {

	result, err := m.client.Query(ctx, milvusclient.NewQueryOption(collection).
		WithFilter("").
		WithOutputFields("count(*)"),
	)
	if err != nil {
		return 0, m.error(err, "failed to count vectors in collection: %s", collection)
	}

	countColumn := result.GetColumn("count(*)")
	if countColumn == nil {
		return 0, m.error(err, "count(*) column not found in query result for collection: %q", collection)
	}

	if countColumn.Len() == 0 {
		return 0, m.error(err, "count(*) returned no result for collection: %q", collection)
	}

	count, err := countColumn.GetAsInt64(0)
	if err != nil {
		return 0, m.error(err, "failed to extract count(*) for collection %q: %w", collection)
	}

	return count, nil
}

func (m *MilvusConnector) Close(ctx context.Context) error {
	return m.client.Close(ctx)
}

func (m *MilvusConnector) ReleaseCollection(ctx context.Context, collection string) error {
	err := m.client.ReleaseCollection(ctx, milvusclient.NewReleaseCollectionOption(collection))
	if err != nil {
		return m.error(err, "ReleaseCollection-001", fmt.Sprintf("failed to release collection %q", collection))
	}
	return nil
}
