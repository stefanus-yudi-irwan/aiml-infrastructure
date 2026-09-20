package weaviate

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

type WeaviateConnector struct {
	client *weaviate.Client
}

func NewWeaviateConnector(config WeaviateConnectorConfig) (*WeaviateConnector, error) {
	weaviateClient, err := weaviate.NewClient(weaviate.Config{
		Host:   config.Host,
		Scheme: config.Scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create weaviate client: %w", err)
	}

	return &WeaviateConnector{
		client: weaviateClient,
	}, nil
}

func (w *WeaviateConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("WeaviateConnector.(%v)(%v) %w", method, params, err)
}

func (w *WeaviateConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {

	_, err := w.checkCollectionExists(ctx, collectionConfig.Name)
	if err != nil {
		return w.error(err, "CreateCollection-001")
	}

	// create weaviate model and properties
	class := &models.Class{
		Class:       collectionConfig.Name,
		Description: collectionConfig.Description,
		Vectorizer:  collectionConfig.Vectorizer,
	}

	for _, propertyConfig := range collectionConfig.Properties {
		property := &models.Property{
			Name:         propertyConfig.Name,
			Description:  propertyConfig.Description,
			DataType:     propertyConfig.DataType,
			Tokenization: propertyConfig.Tokenization,
		}

		class.Properties = append(class.Properties, property)
	}

	// create collection
	err = w.client.Schema().ClassCreator().WithClass(class).Do(ctx)
	if err != nil {
		return w.error(err, "CreateCollection-003", collectionConfig.Name)
	}

	return nil
}

func (w *WeaviateConnector) checkCollectionExists(ctx context.Context, collectionName string) (bool, error) {

	exists, err := w.client.Schema().ClassExistenceChecker().WithClassName(collectionName).Do(ctx)
	if err != nil {
		return false, w.error(err,
			"CreateCollection-001",
			fmt.Sprintf("failed to create collection %s", collectionName))
	}
	if exists {
		return true, w.error(fmt.Errorf("collection %s already exists", collectionName),
			"CreateCollection-002")
	}

	return false, nil
}

func (w *WeaviateConnector) DeleteCollection(ctx context.Context, collection string) error {
	return nil
}

func (w *WeaviateConnector) IsCollectionExists(ctx context.Context, collection string) (bool, error) {
	return false, nil
}

func (w *WeaviateConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {
	return nil
}

func (w *WeaviateConnector) Delete(ctx context.Context, collection string, dataID string) error {
	return nil
}

func (w *WeaviateConnector) GetByID(ctx context.Context, collection string, dataID string) error {
	return nil
}

func (w *WeaviateConnector) CountVector(ctx context.Context, collection string) (int, error) {
	return 0, nil
}

func (w *WeaviateConnector) Close() error {
	return nil
}
