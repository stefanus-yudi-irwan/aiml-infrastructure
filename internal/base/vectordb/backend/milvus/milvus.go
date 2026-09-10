package milvus

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)


type MilvusConnector struct {
	client *milvusclient.Client
}

func NewMilvusConnector(ctx context.Context, config MilvusConfig) (*MilvusConnector, error) {
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

func (m *MilvusConnector) CreateCollection(ctx context.Context, config interface{}) error {
	// TODO: convert config into Milvus collection schema
	return nil
}

func (m *MilvusConnector) DeleteCollection(ctx context.Context, collection string) error {
	// TODO: implement collection deletion
	return nil
}

func (m *MilvusConnector) IsCollectionExists(ctx context.Context, collection string) (bool, error) {
	// TODO: implement collection existence check
	return false, nil
}

func (m *MilvusConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {
	// TODO: convert Data into Milvus fields and upsert
	return nil
}

func (m *MilvusConnector) Delete(ctx context.Context, collection string, dataID string) error {
	// TODO: delete entity by primary key
	return nil
}

func (m *MilvusConnector) GetByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	// TODO: query entity by primary key
	return vectordb.Data{}, nil
}

func (m *MilvusConnector) CountVector(ctx context.Context, collection string) (int, error) {
	// TODO: query collection row count
	return 0, nil
}

func (m *MilvusConnector) Close() error {
	if m.client == nil {
		return nil
	}

	return m.client.Close(context.Background())
}
