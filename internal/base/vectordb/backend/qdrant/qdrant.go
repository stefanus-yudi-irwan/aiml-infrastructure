package qdrant

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

type QdrantConnector struct {
	client *qdrant.Client
}

func NewQdrantConnector(config QdrantConnectorConfig) (*QdrantConnector, error) {

	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   config.Host,
		Port:   config.Port,
		APIKey: config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	return &QdrantConnector{
		client: client,
	}, nil
}

func (q *QdrantConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("QdrantConnector.(%v)(%v) %w", method, params, err)
}

func (q *QdrantConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {

	_, err := q.checkCollectionExists(ctx, collectionConfig.Name)
	if err != nil {
		return q.error(err,
			"CreateCollection-001",
			"fail checking existing collection")
	}

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

	if err = q.client.CreateCollection(ctx, collection); err != nil {
		return q.error(err,
			"CreateCollection-002",
			fmt.Sprintf("fail to craete collection name %s", collectionConfig.Name))
	}

	return nil
}

func (q *QdrantConnector) checkCollectionExists(ctx context.Context, collectionName string) (bool, error) {

	exists, err := q.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return false, q.error(err,
			"checkCollectionExists-001",
			fmt.Sprintf("fail to check collection exists: %s", collectionName))
	}

	if exists {
		return true, q.error(err,
			"checkCollectionExists-002",
			fmt.Sprintf("collectin %s already exists", collectionName))
	}

	return false, nil
}
