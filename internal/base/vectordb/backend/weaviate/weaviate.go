package weaviate

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

type VectorDBConnector struct {
	client *weaviate.Client
}

func NewVectorDBConnector(config VectorDBConnectorConfig) (*VectorDBConnector, error) {
	client, err := weaviate.NewClient(weaviate.Config{
		Host:   config.Host,
		Scheme: config.Scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create weaviate client: %v", err)
	}

	return &VectorDBConnector{
		client: client,
	}, nil
}

func (v *VectorDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("VectorDBConnector.(%v)(%v) %v", method, params, err)
}

func (v *VectorDBConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {
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

		if len(propertyConfig.Properties) > 0 {
			for _, nestedPropertyConfig := range propertyConfig.Properties {
				nestedProperty := &models.NestedProperty{
					Name:         nestedPropertyConfig.Name,
					Description:  nestedPropertyConfig.Description,
					DataType:     nestedPropertyConfig.DataType,
					Tokenization: nestedPropertyConfig.Tokenization,
				}

				property.NestedProperties = append(property.NestedProperties, nestedProperty)
			}
		}

		class.Properties = append(class.Properties, property)
	}

	// create collection
	err := v.client.Schema().ClassCreator().WithClass(class).Do(ctx)
	if err != nil {
		return v.error(err, "CreateCollection-001", collectionConfig.Name)
	}

	return nil
}

func (v *VectorDBConnector) DeleteCollection(ctx context.Context, collection string) error {
	if err := v.client.Schema().ClassDeleter().
		WithClassName(collection).
		Do(ctx); err != nil {
		return v.error(err,
			"DeleteCollection-003",
			fmt.Sprintf("failed to delete collection: %s", collection),
		)
	}

	return nil
}

func (v *VectorDBConnector) IsCollectionExists(ctx context.Context, collection string) (bool, error) {
	exists, err := v.client.Schema().ClassExistenceChecker().WithClassName(collection).Do(ctx)
	if err != nil {
		return false, v.error(err,
			"CreateCollection-001",
			fmt.Sprintf("failed to create collection %s", collection))
	}
	if exists {
		return true, v.error(fmt.Errorf("collection %s already exists", collection),
			"CreateCollection-002")
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
	// weaviate client does not expose a close method
	return nil
}
