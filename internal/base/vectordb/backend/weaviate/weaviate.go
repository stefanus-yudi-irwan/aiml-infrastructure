package weaviate

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"errors"
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
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
		return nil, fmt.Errorf("failed to create vectordb client: %v", err)
	}

	return &VectorDBConnector{
		client: client,
	}, nil
}

func (v *VectorDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("VectorDBConnector.(%v)(%v) %v", method, params, err)
}

func (v *VectorDBConnector) Ping(ctx context.Context) error {
	ready, err := v.client.Misc().ReadyChecker().Do(ctx)
	if err != nil {
		return v.error(err, "Ping-001", "failed to ping Weaviate")
	}

	if !ready {
		return v.error(errors.New("weaviate is not ready"),
			"Ping-002",
			"weaviate is not ready",
		)
	}

	return nil
}

func (v *VectorDBConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {
	// create weaviate model and properties
	class := &models.Class{
		Class:           collectionConfig.Name,
		Description:     collectionConfig.Description,
		Vectorizer:      collectionConfig.Vectorizer,
		VectorIndexType: string(collectionConfig.VectorIndex),
		VectorIndexConfig: map[string]interface{}{
			"ef":             collectionConfig.HNSW.EF,
			"efConstruction": collectionConfig.HNSW.EFConstruction,
			"maxConnections": collectionConfig.HNSW.MaxConnections,
			"distance":       collectionConfig.Distance,
		},
	}

	for _, propertyConfig := range collectionConfig.Fields {
		property := &models.Property{
			Name:         propertyConfig.Name,
			Description:  propertyConfig.Description,
			DataType:     propertyConfig.DataType,
			Tokenization: propertyConfig.Tokenization,
		}

		if len(propertyConfig.Fields) > 0 {
			for _, nestedPropertyConfig := range propertyConfig.Fields {
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

func (v *VectorDBConnector) IsCollectionExists(ctx context.Context, collection string) (bool, error) {
	exists, err := v.client.Schema().ClassExistenceChecker().WithClassName(collection).Do(ctx)
	if err != nil {
		return false, v.error(err,
			"CreateCollection-001",
			fmt.Sprintf("failed to create collection %s", collection))
	}

	if exists {
		return true, nil
	}

	return false, nil
}

func (v *VectorDBConnector) FlushCollection(ctx context.Context, collection string) error {
	whereFilter := filters.Where().
		WithPath([]string{vectordb.IDColumnName}).
		WithOperator(filters.NotEqual).
		WithValueString("00000000-0000-0000-0000-000000000000")

	totalDeleted := 0
	fmt.Println("starting data purge...")

	//Loop to handle collection larger than 10,000 data
	for {
		response, err := v.client.Batch().ObjectsBatchDeleter().
			WithClassName(collection).
			WithOutput("minimal").
			WithWhere(whereFilter).
			Do(ctx)
		if err != nil {
			v.error(err,
				"FlushCollection-001",
				"fail to execute batch deletion step")
		}

		currentDeleted := response.Results.Successful
		totalDeleted += int(currentDeleted)

		if currentDeleted > 0 {
			fmt.Printf("deleted %d objects in this batch... (total: %d)\n", currentDeleted, totalDeleted)
		}

		if response.Results.Failed > 0 {
			fmt.Printf("warning: %d objects failed to delete in this iteration.\n", response.Results.Failed)
		}

		if currentDeleted == 0 {
			break
		}
	}

	fmt.Printf("finished. safely deleted a total of %d objects while keeping the collection schema.\n", totalDeleted)
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

func (v *VectorDBConnector) FlushAll(ctx context.Context) error {
	err := v.client.Schema().AllDeleter().Do(ctx)

	if err != nil {
		return v.error(err,
			"FlushAll-001",
			"failed to delete all collections",
		)
	}

	return nil
}

func (v *VectorDBConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {

	// Check whether the object already exists.
	isExists, err := v.IsKeyExists(ctx, collection, data.ID)
	if err != nil {
		return v.error(err,
			"Upsert-001",
			fmt.Sprintf("fail to check key existence: %v", data.ID),
		)
	}

	// Both embedding and metadata are present.
	if data.Embedding != nil && data.Metadata != nil && len(data.Metadata) > 0 {
		properties := map[string]interface{}{
			vectordb.DocumentIDColumnName: data.DocumentID,
			vectordb.ChunkIDColumnName:    data.ChunkID,
			vectordb.ContentColumnName:    data.Content,
			vectordb.MetadataColumnName:   data.Metadata,
		}

		if isExists {
			err = v.client.Data().Updater().
				WithClassName(collection).
				WithID(data.ID).
				WithProperties(properties).
				WithVector(data.Embedding).
				Do(ctx)

			if err != nil {
				return v.error(err,
					"Upsert-002",
					fmt.Sprintf("fail to update embedding and metadata: %v", data.ID),
				)
			}
			return nil
		}

		_, err = v.client.Data().Creator().
			WithClassName(collection).
			WithID(data.ID).
			WithProperties(properties).
			WithVector(data.Embedding).
			Do(ctx)

		if err != nil {
			return v.error(err,
				"Upsert-003",
				fmt.Sprintf("fail to insert embedding and metadata: %v", data.ID),
			)
		}

		return nil
	}

	switch {
	case isExists && data.Embedding != nil:
		// Update embedding only.
		err = v.client.Data().Updater().
			WithClassName(collection).
			WithID(data.ID).
			WithVector(data.Embedding).
			Do(ctx)

		if err != nil {
			return v.error(err,
				"Upsert-004",
				fmt.Sprintf("fail to update embedding: %v", data.ID),
			)
		}

	case isExists && data.Metadata != nil:
		// Update metadata only.
		err = v.client.Data().Updater().
			WithClassName(collection).
			WithID(data.ID).
			WithProperties(map[string]interface{}{vectordb.MetadataColumnName: data.Metadata}).
			Do(ctx)

		if err != nil {
			return v.error(err,
				"Upsert-005",
				fmt.Sprintf("fail to update metadata: %v", data.ID),
			)
		}

	case !isExists && data.Embedding != nil:
		// Insert embedding only.
		_, err = v.client.Data().Creator().
			WithClassName(collection).
			WithID(data.ID).
			WithVector(data.Embedding).
			Do(ctx)

		if err != nil {
			return v.error(err,
				"Upsert-006",
				fmt.Sprintf("fail to insert embedding: %v", data.ID),
			)
		}

	case !isExists && data.Metadata != nil:
		// Insert metadata only without embedding.
		properties := map[string]interface{}{
			vectordb.DocumentIDColumnName: data.DocumentID,
			vectordb.ChunkIDColumnName:    data.ChunkID,
			vectordb.ContentColumnName:    data.Content,
			vectordb.MetadataColumnName:   data.Metadata,
		}

		_, err = v.client.Data().Creator().
			WithClassName(collection).
			WithID(data.ID).
			WithProperties(properties).
			Do(ctx)

		if err != nil {
			return v.error(err,
				"Upsert-008",
				fmt.Sprintf("fail to insert metadata-only object: %v", data.ID),
			)
		}

	default:
		return v.error(
			errors.New("either embedding or metadata must be provided"),
			"Upsert-009",
			fmt.Sprintf("cannot upsert data object: %v", data.ID),
		)
	}

	return nil
}

func (v *VectorDBConnector) Delete(ctx context.Context, collection string, dataID string) error {
	err := v.client.Data().Deleter().
		WithClassName(collection).
		WithID(dataID).
		Do(ctx)

	if err != nil {
		return v.error(err,
			"Delete-001",
			fmt.Sprintf("fail to delete object: %v", dataID),
		)
	}

	return nil
}

func (v *VectorDBConnector) UpsertBatch(ctx context.Context, collection string, arrayData []vectordb.Data, batchNumber int) error {

	if batchNumber <= 0 {
		return v.error(
			errors.New("batchNumber must be greater than 0"),
			"UpsertBatch-001",
			"invalid batch number",
		)
	}

	for start := 0; start < len(arrayData); start += batchNumber {
		if err := ctx.Err(); err != nil {
			return v.error(err,
				"UpsertBatch-002",
				"context is cancelled")
		}

		end := start + batchNumber
		if end > len(arrayData) {
			end = len(arrayData)
		}

		batch := arrayData[start:end]
		existenceMap := make(map[string]bool, len(batch))

		for _, data := range batch {
			if err := ctx.Err(); err != nil {
				return v.error(err,
					"UpsertBatch-003",
					"context is cancelled")
			}

			isExists, err := v.IsKeyExists(ctx, collection, data.ID)
			if err != nil {
				return v.error(err,
					"UpsertBatch-004",
					fmt.Sprintf("fail to check object existence: collection=%s id=%s", collection, data.ID))
			}

			existenceMap[data.ID] = isExists
		}

		var createObjects []*models.Object

		for _, data := range batch {
			if err := ctx.Err(); err != nil {
				return v.error(err,
					"UpsertBatch-005",
					"context is cancelled")
			}

			hasEmbedding := data.Embedding != nil
			hasMetadata := len(data.Metadata) > 0
			isExists := existenceMap[data.ID]

			switch {
			case !isExists && hasEmbedding && hasMetadata:
				// create object with embedding + properties
				properties := map[string]interface{}{
					vectordb.DocumentIDColumnName: data.DocumentID,
					vectordb.ChunkIDColumnName:    data.ChunkID,
					vectordb.ContentColumnName:    data.Content,
					vectordb.MetadataColumnName:   data.Metadata,
				}

				createObjects = append(createObjects, &models.Object{
					Class:      collection,
					ID:         strfmt.UUID(data.ID),
					Properties: properties,
					Vector:     data.Embedding,
				})

			case !isExists && hasEmbedding:
				// create object with embedding only
				createObjects = append(createObjects, &models.Object{
					Class:  collection,
					ID:     strfmt.UUID(data.ID),
					Vector: data.Embedding,
				})

			case !isExists && hasMetadata:
				// create object with metadata only
				properties := map[string]interface{}{
					vectordb.DocumentIDColumnName: data.DocumentID,
					vectordb.ChunkIDColumnName:    data.ChunkID,
					vectordb.ContentColumnName:    data.Content,
					vectordb.MetadataColumnName:   data.Metadata,
				}

				createObjects = append(createObjects, &models.Object{
					Class:      collection,
					ID:         strfmt.UUID(data.ID),
					Properties: properties,
				})

			case !isExists:
				return v.error(
					errors.New("either embedding or metadata must be provided"),
					"UpsertBatch-006",
					fmt.Sprintf("cannot upsert data object: %v", data.ID),
				)
			}
		}

		if len(createObjects) > 0 {
			_, err := v.client.Batch().ObjectsBatcher().
				WithObjects(createObjects...).
				Do(ctx)

			if err != nil {
				return v.error(err,
					"UpsertBatch-007",
					fmt.Sprintf("fail to create objects: collection=%s", collection),
				)
			}
		}

		for _, data := range batch {
			if err := ctx.Err(); err != nil {
				return v.error(err,
					"UpsertBatch-008",
					"context is cancelled",
				)
			}

			if !existenceMap[data.ID] {
				continue
			}

			hasEmbedding := data.Embedding != nil
			hasMetadata := len(data.Metadata) > 0

			switch {
			case hasEmbedding && hasMetadata:
				// Update embedding + metadata.
				properties := map[string]interface{}{
					vectordb.DocumentIDColumnName: data.DocumentID,
					vectordb.ChunkIDColumnName:    data.ChunkID,
					vectordb.ContentColumnName:    data.Content,
					vectordb.MetadataColumnName:   data.Metadata,
				}

				err := v.client.Data().Updater().
					WithClassName(collection).
					WithID(data.ID).
					WithProperties(properties).
					WithVector(data.Embedding).
					Do(ctx)

				if err != nil {
					return v.error(err,
						"UpsertBatch-009",
						fmt.Sprintf("fail to update embedding and metadata: %v", data.ID),
					)
				}

			case hasEmbedding:
				// Update embedding only.
				err := v.client.Data().Updater().
					WithClassName(collection).
					WithID(data.ID).
					WithVector(data.Embedding).
					Do(ctx)

				if err != nil {
					return v.error(err,
						"UpsertBatch-010",
						fmt.Sprintf("fail to update embedding: %v", data.ID),
					)
				}

			case hasMetadata:
				// Update metadata only.
				err := v.client.Data().Updater().
					WithClassName(collection).
					WithID(data.ID).
					WithProperties(map[string]interface{}{
						"metadata": data.Metadata,
					}).
					Do(ctx)

				if err != nil {
					return v.error(err,
						"UpsertBatch-011",
						fmt.Sprintf("fail to update metadata: %v", data.ID),
					)
				}

			default:
				return v.error(
					errors.New("either embedding or metadata must be provided"),
					"UpsertBatch-012",
					fmt.Sprintf("cannot upsert data object: %v", data.ID),
				)
			}
		}

	}

	return nil
}

func (v *VectorDBConnector) DeleteBatch(ctx context.Context, collection string, arrayID []string, batchNumber int) error {
	if batchNumber <= 0 {
		return v.error(errors.New("batchNumber must be greater than 0"),
			"DeleteBatch-001",
			"invalid batch number",
		)
	}

	for start := 0; start < len(arrayID); start += batchNumber {
		if err := ctx.Err(); err != nil {
			return v.error(err,
				"DeleteBatch-002",
				"context is cancelled",
			)
		}

		end := start + batchNumber
		if end > len(arrayID) {
			end = len(arrayID)
		}

		batch := arrayID[start:end]

		for _, dataID := range batch {
			if err := ctx.Err(); err != nil {
				return v.error(err,
					"DeleteBatch-003",
					"context is cancelled",
				)
			}

			err := v.client.Data().Deleter().
				WithClassName(collection).
				WithID(dataID).
				Do(ctx)

			if err != nil {
				return v.error(err,
					"DeleteBatch-004",
					fmt.Sprintf("fail to delete object: collection=%s id=%s", collection, dataID),
				)
			}
		}
	}

	return nil
}

func (v *VectorDBConnector) GetByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {
	if err := ctx.Err(); err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-001",
			"context is cancelled",
		)
	}

	objects, err := v.client.Data().ObjectsGetter().
		WithClassName(collection).
		WithID(dataID).
		WithVector().
		Do(ctx)

	if err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-002",
			fmt.Sprintf("fail to get object: collection=%s id=%s", collection, dataID),
		)
	}

	if len(objects) == 0 || objects[0] == nil {
		return vectordb.Data{}, v.error(errors.New("object not found"),
			"GetByID-003",
			fmt.Sprintf("object not found: collection=%s id=%s", collection, dataID),
		)
	}

	object := objects[0]

	data := vectordb.Data{
		ID:        object.ID.String(),
		Embedding: object.Vector,
	}

	if object.Properties == nil {
		return data, nil
	}

	properties, ok := object.Properties.(map[string]interface{})
	if !ok {
		return vectordb.Data{}, v.error(fmt.Errorf("unexpected properties type: %T", object.Properties),
			"GetByID-004",
			fmt.Sprintf("fail to parse object properties: collection=%s id=%s", collection, dataID),
		)
	}

	if value, ok := properties["document_id"].(string); ok {
		data.DocumentID = value
	}

	if value, ok := properties["chunk_id"].(string); ok {
		data.ChunkID = value
	}

	if value, ok := properties["content"].(string); ok {
		data.Content = value
	}

	if value, ok := properties["metadata"].(map[string]interface{}); ok {
		data.Metadata = value
	}

	return data, nil
}

func (v *VectorDBConnector) CountVector(ctx context.Context, collection string) (int, error) {
	metaFields := graphql.Field{
		Name: "meta",
		Fields: []graphql.Field{
			{Name: "count"},
		},
	}

	result, err := v.client.GraphQL().Aggregate().
		WithClassName(collection).
		WithFields(metaFields).
		Do(ctx)

	if err != nil {
		return 0, v.error(err,
			"CountVector-001",
			fmt.Errorf("fail to count vector from collection: %v", collection),
		)
	}

	for i, gqlErr := range result.Errors {
		fmt.Printf("GraphQL error %d: %+v\n", i, gqlErr)
	}

	aggregate, ok := result.Data["Aggregate"].(map[string]interface{})
	if !ok {
		return 0, v.error(errors.New("invalid aggregate response"),
			"CountVector-002",
		)
	}

	classResult, ok := aggregate[collection].([]interface{})
	if !ok || len(classResult) == 0 {
		return 0, nil
	}

	item, ok := classResult[0].(map[string]interface{})
	if !ok {
		return 0, v.error(errors.New("invalid aggregate item"),
			"CountVector-003",
		)
	}

	meta, ok := item["meta"].(map[string]interface{})
	if !ok {
		return 0, v.error(errors.New("missing meta in aggregate response"),
			"CountVector-004",
		)
	}

	count, ok := meta["count"].(float64)
	if !ok {
		return 0, v.error(errors.New("invalid count in aggregate response"),
			"CountVector-005",
		)
	}

	return int(count), nil
}

func (v *VectorDBConnector) Close() error {
	// weaviate client does not expose a close method
	return nil
}

func (v *VectorDBConnector) IsKeyExists(ctx context.Context, collection string, dataID string) (bool, error) {
	isExists, err := v.client.Data().Checker().
		WithClassName(collection).
		WithID(dataID).
		Do(ctx)

	if err != nil {
		return false, v.error(err,
			"IsKeyExists-001",
			fmt.Sprintf("fail to check if dataID exists: %v", dataID),
		)
	}

	return isExists, nil
}
