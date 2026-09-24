package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

type VectorDBConnector struct {
	client *qdrant.Client
}

func NewVectorDBConnector(config VectorDBConnectorConfig) (*VectorDBConnector, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:                   config.Host,
		Port:                   config.Port,
		APIKey:                 config.APIKey,
		SkipCompatibilityCheck: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	return &VectorDBConnector{
		client: client,
	}, nil
}

func (v *VectorDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("VectorDBConnector.(%v)(%v) %w", method, params, err)
}

func (v *VectorDBConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {
	collection := &qdrant.CreateCollection{
		CollectionName: collectionConfig.Name,
		VectorsConfig: qdrant.NewVectorsConfig(
			&qdrant.VectorParams{
				Size:     collectionConfig.Dimension,
				Distance: qdrant.Distance_Cosine,
			},
		),
	}

	// Optional configuration
	if collectionConfig.ShardNumber > 0 {
		collection.ShardNumber = qdrant.PtrOf(collectionConfig.ShardNumber)
	}
	if collectionConfig.ReplicationFactor > 0 {
		collection.ReplicationFactor = qdrant.PtrOf(collectionConfig.ReplicationFactor)
	}
	if collectionConfig.WriteConsistencyFactor > 0 {
		collection.WriteConsistencyFactor = qdrant.PtrOf(collectionConfig.WriteConsistencyFactor)
	}

	if err := v.client.CreateCollection(ctx, collection); err != nil {
		return v.error(err,
			"CreateCollection-001",
			fmt.Sprintf("fail to craete collection name %s", collectionConfig.Name))
	}

	return nil
}

func (v *VectorDBConnector) DeleteCollection(ctx context.Context, collection string) error {

	return nil
}

func (v *VectorDBConnector) IsCollectionExists(ctx context.Context, collectionName string) (bool, error) {

	exists, err := v.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return false, v.error(err,
			"checkCollectionExists-001",
			fmt.Sprintf("fail to check collection exists: %s", collectionName))
	}

	if exists {
		return true, v.error(err,
			"checkCollectionExists-002",
			fmt.Sprintf("collectin %s already exists", collectionName))
	}

	return false, nil
}

func (v *VectorDBConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {
	return nil
}

func (v *VectorDBConnector) Delete(ctx context.Context, collection string, dataID string) error {
	return nil
}

func (v *VectorDBConnector) GetByID(ctx context.Context, collection string, dataID string) error {
	return nil
}

func (v *VectorDBConnector) CountVector(ctx context.Context, collection string) (int, error) {
	return 0, nil
}

func (v *VectorDBConnector) Close() error {
	// qdrant client does not expose a close method
	return nil
}
