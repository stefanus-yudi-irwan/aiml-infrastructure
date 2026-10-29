package weaviate

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	// create vector configuration
	vectorConfig := make(map[string]models.VectorConfig, len(collectionConfig.DenseEmbeddingFields))
	for _, embeddingField := range collectionConfig.DenseEmbeddingFields {

		indexType, err := toWeaviateIndexType(embeddingField.Index)
		if err != nil {
			v.error(err,
				"CreateCollection-001",
				"unsupported index type")
		}

		distanceType, err := toWeaviateDistanceType(embeddingField.Distance)
		if err != nil {
			v.error(err,
				"CreateCollection-002",
				"unsupported distance type")
		}

		vectorConfig[fmt.Sprintf("%s_%s",
			vectordb.DenseEmbeddingsColumnName,
			embeddingField.Name)] = models.VectorConfig{
			Vectorizer: map[string]interface{}{
				"none": map[string]interface{}{},
			},
			VectorIndexType: indexType,
			VectorIndexConfig: map[string]interface{}{
				"ef":             embeddingField.HNSW.EFRuntime,
				"efConstruction": embeddingField.HNSW.EFConstruction,
				"maxConnections": embeddingField.HNSW.M,
				"distance":       distanceType,
			},
		}
	}

	// create metadata configuration
	metadataConfig := make([]*models.Property, 0,
		len(collectionConfig.DataFields)+len(collectionConfig.SparseEmbeddingFields))

	for _, sparseField := range collectionConfig.SparseEmbeddingFields {
		weaviateDataType, err := toWeaviateDataType(sparseField.DataType)
		if err != nil {
			v.error(err,
				"CreateCollection-001",
				"fail to create weaviateDataType")
		}

		property := &models.Property{
			Name: fmt.Sprintf("%s_%s",
				vectordb.SparseEmbeddingsColumnName,
				sparseField.Name),
			Description: sparseField.Description,
			DataType:    weaviateDataType,
		}

		metadataConfig = append(metadataConfig, property)
	}

	for _, dataField := range collectionConfig.DataFields {
		weaviateDataType, err := toWeaviateDataType(dataField.DataType)
		if err != nil {
			v.error(err,
				"CreateCollection-001",
				"fail to create weaviateDataType")
		}

		property := &models.Property{
			Name:        dataField.Name,
			Description: dataField.Description,
			DataType:    weaviateDataType,
		}

		if len(dataField.Fields) > 0 {
			for _, field := range dataField.Fields {
				weaviateDataType, err := toWeaviateDataType(field.DataType)
				if err != nil {
					v.error(err,
						"CreateCollection-002",
						"fail to create nested weaviateDataType")
				}

				nestedProperty := &models.NestedProperty{
					Name:        field.Name,
					Description: field.Description,
					DataType:    weaviateDataType,
				}

				property.NestedProperties = append(
					property.NestedProperties,
					nestedProperty)
			}
		}

		metadataConfig = append(metadataConfig, property)
	}

	// create weaviate model and properties
	class := &models.Class{
		Class:        collectionConfig.Name,
		Description:  collectionConfig.Description,
		VectorConfig: vectorConfig,
		Properties:   metadataConfig,
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

	var err error
	hasVectors := len(data.DenseEmbeddings) > 0
	hasMetadata := len(data.Metadata) > 0 || len(data.SparseEmbeddings) > 0
	vectors := make(models.Vectors)
	properties := make(map[string]interface{})

	if hasVectors {
		vectors = createWeaviateVectors(data)
	}

	if hasMetadata {
		properties, err = createWeaviateProperties(data)
		if err != nil {
			return v.error(err,
				"Upsert-002",
				fmt.Sprintf("fail to create properties: %v", data.ID))
		}
	}

	// Check whether the object already exists.
	isExists, err := v.IsKeyExists(ctx, collection, data.ID)
	if err != nil {
		return v.error(err,
			"Upsert-001",
			fmt.Sprintf("fail to check key existence: %v", data.ID),
		)
	}

	// Both embedding and metadata are present.
	if hasVectors && hasMetadata {
		if isExists {
			err = v.client.Data().Updater().
				WithMerge().
				WithClassName(collection).
				WithID(data.ID).
				WithProperties(properties).
				WithVectors(vectors).
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
			WithVectors(vectors).
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
	case isExists && hasVectors:
		// Update embedding only.
		err = v.client.Data().Updater().
			WithMerge().
			WithClassName(collection).
			WithID(data.ID).
			WithVectors(vectors).
			Do(ctx)

		if err != nil {
			return v.error(err,
				"Upsert-004",
				fmt.Sprintf("fail to update embedding: %v", data.ID),
			)
		}

	case isExists && hasMetadata:
		// Update metadata only.
		err = v.client.Data().Updater().
			WithMerge().
			WithClassName(collection).
			WithID(data.ID).
			WithProperties(properties).
			Do(ctx)

		if err != nil {
			return v.error(err,
				"Upsert-005",
				fmt.Sprintf("fail to update metadata: %v", data.ID),
			)
		}

	case !isExists && hasVectors:
		// Insert embedding only.
		_, err = v.client.Data().Creator().
			WithClassName(collection).
			WithID(data.ID).
			WithVectors(vectors).
			Do(ctx)

		if err != nil {
			return v.error(err,
				"Upsert-006",
				fmt.Sprintf("fail to insert embedding: %v", data.ID),
			)
		}

	case !isExists && hasMetadata:
		// Insert metadata only without embedding.
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

			var err error
			hasVectors := len(data.DenseEmbeddings) > 0
			hasMetadata := len(data.Metadata) > 0 || len(data.SparseEmbeddings) > 0
			vectors := make(models.Vectors)
			properties := make(map[string]interface{})
			isExists := existenceMap[data.ID]

			if hasVectors {
				vectors = createWeaviateVectors(data)
			}

			if hasMetadata {
				properties, err = createWeaviateProperties(data)
				if err != nil {
					return v.error(err,
						"Upsert-002",
						fmt.Sprintf("fail to create properties: %v", data.ID))
				}
			}

			switch {
			case isExists && hasVectors && hasMetadata:
				// update embedding + metadata
				err := v.client.Data().Updater().
					WithMerge().
					WithClassName(collection).
					WithID(data.ID).
					WithProperties(properties).
					WithVectors(vectors).
					Do(ctx)

				if err != nil {
					return v.error(err,
						"UpsertBatch-009",
						fmt.Sprintf("fail to update embedding and metadata: %v", data.ID),
					)
				}

			case isExists && hasVectors:
				// Update embedding only.
				err := v.client.Data().Updater().
					WithMerge().
					WithClassName(collection).
					WithID(data.ID).
					WithVectors(vectors).
					Do(ctx)

				if err != nil {
					return v.error(err,
						"UpsertBatch-010",
						fmt.Sprintf("fail to update embedding: %v", data.ID),
					)
				}

			case isExists && hasMetadata:
				// Update metadata only.
				err := v.client.Data().Updater().
					WithMerge().
					WithClassName(collection).
					WithID(data.ID).
					WithProperties(properties).
					Do(ctx)

				if err != nil {
					return v.error(err,
						"UpsertBatch-011",
						fmt.Sprintf("fail to update metadata: %v", data.ID),
					)
				}

			case !isExists && hasVectors && hasMetadata:
				// create object with embedding + properties
				createObjects = append(createObjects, &models.Object{
					Class:      collection,
					ID:         strfmt.UUID(data.ID),
					Properties: properties,
					Vectors:    vectors,
				})

			case !isExists && hasVectors:
				// create object with embedding only
				createObjects = append(createObjects, &models.Object{
					Class:   collection,
					ID:      strfmt.UUID(data.ID),
					Vectors: vectors,
				})

			case !isExists && hasMetadata:
				// create object with metadata only
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
	data := vectordb.Data{ID: object.ID.String()}

	if len(object.Vectors) > 0 {
		denseEmbedding := make(map[vectordb.DenseEmbeddingName][]float32, len(object.Vectors))
		for name, embedding := range object.Vectors {
			vector, ok := embedding.([]float32)
			if !ok {
				return vectordb.Data{}, v.error(fmt.Errorf("unexpected vector type of %q: %T", name, embedding),
					"GetByID-004",
					fmt.Sprintf("failed to parse named vector: collection=%s id=%s", collection, dataID))
			}
			denseEmbedding[vectordb.DenseEmbeddingName(
				strings.TrimPrefix(name,
					fmt.Sprintf("%s_", vectordb.DenseEmbeddingsColumnName)))] = vector
		}
		data.DenseEmbeddings = denseEmbedding
	}

	if object.Properties != nil {
		properties, ok := object.Properties.(map[string]interface{})
		sparseEmbeddings := make(map[vectordb.SparseEmbeddingName]vectordb.SparseEmbedding)
		if !ok {
			return vectordb.Data{}, v.error(fmt.Errorf("unexpected properties type: %T", object.Properties),
				"GetByID-005",
				fmt.Sprintf("fail to parse object properties: collection=%s id=%s", collection, dataID),
			)
		}

		for name, property := range properties {
			if name == vectordb.DocumentIDColumnName {
				payload, ok := property.(string)
				if !ok {
					return vectordb.Data{}, v.error(
						fmt.Errorf("fail to parse document ID"),
						"GetByID-006",
					)
				}
				data.DocumentID = payload
			}

			if name == vectordb.ChunkIDColumnName {
				payload, ok := property.(string)
				if !ok {
					return vectordb.Data{}, v.error(
						fmt.Errorf("fail to parse document ID"),
						"GetByID-007",
					)
				}
				data.ChunkID = payload
			}

			if name == vectordb.ContentColumnName {
				payload, ok := property.(string)
				if !ok {
					return vectordb.Data{}, v.error(
						fmt.Errorf("fail to parse document ID"),
						"GetByID-008",
					)
				}
				data.Content = payload
			}

			if name == vectordb.MetadataColumnName {
				payload, ok := property.(map[string]interface{})
				if !ok {
					return vectordb.Data{}, v.error(
						fmt.Errorf("fail to parse document ID"),
						"GetByID-009",
					)
				}
				normalizeMetadata(payload)
				data.Metadata = payload
			}

			if strings.Contains(name, vectordb.SparseEmbeddingsColumnName) {
				var jsonData []byte
				switch payload := property.(type) {
				case string:
					jsonData = []byte(payload)
				case []byte:
					jsonData = payload
				default:
					return vectordb.Data{}, v.error(
						fmt.Errorf("unexpected sparse embedding property type %q: %T", name, property),
						"GetByID-010",
						fmt.Sprintf("failed to parse sparse embedding: collection=%s id=%s", collection, dataID),
					)
				}

				var sparseEmbedding vectordb.SparseEmbedding
				if err := json.Unmarshal(jsonData, &sparseEmbedding); err != nil {
					return vectordb.Data{}, v.error(err,
						"GetByID-010",
						fmt.Sprintf("fail to unmarshall sparseEmbedding: %v", data.ID),
					)
				}

				sparseEmbeddings[vectordb.SparseEmbeddingName(
					strings.TrimPrefix(name,
						fmt.Sprintf("%s_", vectordb.SparseEmbeddingsColumnName)))] = sparseEmbedding
			}
		}

		data.SparseEmbeddings = sparseEmbeddings
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
