package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"errors"
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
		return nil, fmt.Errorf("failed to create vectordb client: %w", err)
	}

	return &VectorDBConnector{
		client: client,
	}, nil
}

func (v *VectorDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("VectorDBConnector.(%v)(%v) %w", method, params, err)
}

func (v *VectorDBConnector) Ping(ctx context.Context) error {
	return nil
}

func (v *VectorDBConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {

	denseVectorConfig := make(map[string]*qdrant.VectorParams)
	sparseVectorConfig := make(map[string]*qdrant.SparseVectorParams)

	for _, denseEmbeddingField := range collectionConfig.DenseEmbeddingFields {
		denseVectorConfig[string(denseEmbeddingField.Name)] = &qdrant.VectorParams{
			Size:     uint64(denseEmbeddingField.Dimension),
			Distance: toQdrantDistance(denseEmbeddingField.Distance),
		}

		if denseEmbeddingField.HNSW != nil {
			denseVectorConfig[string(denseEmbeddingField.Name)].HnswConfig = &qdrant.HnswConfigDiff{
				M:           qdrant.PtrOf(uint64(denseEmbeddingField.HNSW.M)),
				EfConstruct: qdrant.PtrOf(uint64(denseEmbeddingField.HNSW.EFConstruction)),
			}
		}
	}

	for _, sparseEmbeddingField := range collectionConfig.SparseEmbeddingFields {
		sparseVectorConfig[string(sparseEmbeddingField.Name)] = &qdrant.SparseVectorParams{}
	}

	collection := &qdrant.CreateCollection{
		CollectionName:      collectionConfig.Name,
		VectorsConfig:       qdrant.NewVectorsConfigMap(denseVectorConfig),
		SparseVectorsConfig: qdrant.NewSparseVectorsConfig(sparseVectorConfig),
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

func (v *VectorDBConnector) IsCollectionExists(ctx context.Context, collection string) (bool, error) {
	exists, err := v.client.CollectionExists(ctx, collection)
	if err != nil {
		return false, v.error(err,
			"IsCollectionExists-001",
			fmt.Sprintf("fail to check collection exists: %s", collection))
	}

	return exists, nil
}

func (v *VectorDBConnector) FlushCollection(ctx context.Context, collection string) error {
	_, err := v.client.Delete(ctx,
		&qdrant.DeletePoints{
			CollectionName: collection,
			Points: qdrant.NewPointsSelectorFilter(
				&qdrant.Filter{},
			),
		})

	if err != nil {
		return v.error(err,
			"FlushCollection-001",
			fmt.Sprintf("fail to flush collection: %s", collection))
	}

	return nil
}

func (v *VectorDBConnector) DeleteCollection(ctx context.Context, collection string) error {
	err := v.client.DeleteCollection(ctx, collection)
	if err != nil {
		return v.error(err,
			"DeleteCollection-001",
			fmt.Sprintf("fail to delete collection: %s", collection))
	}

	return nil
}

func (v *VectorDBConnector) FlushAll(ctx context.Context) error {
	response, err := v.client.ListCollections(ctx)
	if err != nil {
		return v.error(err,
			"FlushAll-001",
			"fail to list collections")
	}

	for _, collection := range response {
		if err := v.client.DeleteCollection(ctx, collection); err != nil {
			return v.error(err,
				"FlushAll-002",
				fmt.Sprintf("fail to delete collection: %s", collection))
		}
	}

	return nil
}

func (v *VectorDBConnector) Upsert(ctx context.Context, collection string, data vectordb.Data) error {

	hasVectors := len(data.DenseEmbeddings) > 0 || len(data.SparseEmbeddings) > 0
	hasMetadata := len(data.Metadata) > 0

	pointStruct, err := createQdrantPoint(data)
	if err != nil {
		return v.error(err,
			"Upsert-001",
			fmt.Sprintf("fail to create qdrant point: %v", data.ID))
	}

	if hasVectors && hasMetadata {
		_, err = v.client.Upsert(ctx, &qdrant.UpsertPoints{
			CollectionName: collection,
			Points:         []*qdrant.PointStruct{pointStruct},
		})

		if err != nil {
			return v.error(err,
				"Upsert-002",
				fmt.Sprintf("fail to upsert point: %v", data.ID))
		}

		return nil
	}

	// if data is incomplete check for existence of the key in the database
	isExists, err := v.IsKeyExists(ctx, collection, data.ID)
	if err != nil {
		return v.error(err,
			"Upsert-003",
			fmt.Sprintf("fail to check key existence: %v", data.ID))
	}

	switch {
	case isExists && hasVectors:
		// update embedding only
		_, err = v.client.UpdateVectors(ctx,
			&qdrant.UpdatePointVectors{
				CollectionName: collection,
				Points: []*qdrant.PointVectors{
					{
						Id:      pointStruct.Id,
						Vectors: pointStruct.Vectors,
					},
				},
			},
		)

		if err != nil {
			return v.error(err,
				"Upsert-005",
				fmt.Sprintf("fail to update embedding: %v", data.ID))
		}

	case isExists && hasMetadata:
		// update metadata only
		_, err = v.client.SetPayload(ctx,
			&qdrant.SetPayloadPoints{
				CollectionName: collection,
				Payload:        pointStruct.Payload,
				PointsSelector: &qdrant.PointsSelector{
					PointsSelectorOneOf: &qdrant.PointsSelector_Points{
						Points: &qdrant.PointsIdsList{
							Ids: []*qdrant.PointId{
								pointStruct.Id,
							},
						},
					},
				},
			},
		)
		if err != nil {
			return v.error(err,
				"Upsert-006",
				fmt.Sprintf("fail to update metadata: %v", data.ID),
			)
		}

	case !isExists && hasVectors:
		// insert embedding only
		_, err = v.client.Upsert(ctx,
			&qdrant.UpsertPoints{
				CollectionName: collection,
				Points:         []*qdrant.PointStruct{pointStruct},
			},
		)
		if err != nil {
			return v.error(err,
				"Upsert-007",
				fmt.Sprintf("fail to insert embedding: %v", data.ID),
			)
		}

	case !isExists && hasMetadata:
		// insert metadata only
		return v.error(
			errors.New("cannot insert metadata without embedding for a new point"),
			"Upsert-008",
			fmt.Sprintf("cannot insert metadata-only point: %v", data.ID),
		)
	default:
		return v.error(errors.New("either embedding or metadata must be provided"),
			"Upsert-009",
			fmt.Sprintf("cannot upsert data point: %v", data.ID))
	}

	return nil
}

func (v *VectorDBConnector) Delete(ctx context.Context, collection string, dataID string) error {
	_, err := v.client.Delete(ctx,
		&qdrant.DeletePoints{
			CollectionName: collection,
			Points: &qdrant.PointsSelector{
				PointsSelectorOneOf: &qdrant.PointsSelector_Points{
					Points: &qdrant.PointsIdsList{
						Ids: []*qdrant.PointId{qdrant.NewID(dataID)},
					},
				},
			},
		},
	)
	if err != nil {
		return v.error(err,
			"Delete-001",
			fmt.Sprintf("fail to delete point: collection=%s, id=%s", collection, dataID),
		)
	}

	return nil
}

func (v *VectorDBConnector) UpsertBatch(ctx context.Context, collection string, arrayData []vectordb.Data, batchNumber int) error {

	if batchNumber <= 0 {
		return v.error(
			errors.New("batchNumber must be greater than 0"),
			"UpsertBatch-001",
			"invalid batch number")
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
		pointIDs := make([]*qdrant.PointId, 0, len(batch))

		for _, data := range batch {
			pointIDs = append(pointIDs, qdrant.NewID(data.ID))
		}

		existingPoints, err := v.client.Get(ctx,
			&qdrant.GetPoints{
				CollectionName: collection,
				Ids:            pointIDs,
			})
		if err != nil {
			return v.error(err,
				"UpsertBatch-003",
				fmt.Sprintf("fail to get existing points: collection=%s", collection),
			)
		}

		existenceMap := make(map[string]bool, len(existingPoints))
		for _, point := range existingPoints {
			if point == nil || point.Id == nil {
				continue
			}
			if uuid := point.Id.GetUuid(); uuid != "" {
				existenceMap[uuid] = true
			}
		}

		var upsertPoints []*qdrant.PointStruct
		var updateVectors []*qdrant.PointVectors
		var updatePayloads []*qdrant.PointsUpdateOperation

		for _, data := range batch {
			if err := ctx.Err(); err != nil {
				return err
			}

			pointStruct, err := createQdrantPoint(data)
			if err != nil {
				return v.error(err,
					"UpsertBatch-004",
					fmt.Sprintf("fail to create qdrant point: %v", data.ID))
			}

			hasEmbedding := len(data.DenseEmbeddings) > 0 || len(data.SparseEmbeddings) > 0
			hasMetadata := len(data.Metadata) > 0
			isExists := existenceMap[data.ID]

			switch {
			case isExists && hasEmbedding && hasMetadata:
				upsertPoints = append(upsertPoints, pointStruct)

			case isExists && hasEmbedding:
				updateVectors = append(updateVectors,
					&qdrant.PointVectors{
						Id:      pointStruct.Id,
						Vectors: pointStruct.Vectors,
					},
				)

			case isExists && hasMetadata:
				updatePayloads = append(updatePayloads,
					qdrant.NewPointsUpdateSetPayload(
						&qdrant.PointsUpdateOperation_SetPayload{
							Payload: pointStruct.Payload,
							PointsSelector: &qdrant.PointsSelector{
								PointsSelectorOneOf: &qdrant.PointsSelector_Points{
									Points: &qdrant.PointsIdsList{
										Ids: []*qdrant.PointId{pointStruct.Id},
									},
								},
							},
						},
					))

			case !isExists && hasEmbedding:
				upsertPoints = append(upsertPoints, pointStruct)

			case !isExists && hasMetadata:
				return v.error(
					errors.New("cannot insert metadata without embedding for a new point"),
					"UpsertBatch-005",
					fmt.Sprintf("cannot insert metadata-only point: %v", data.ID),
				)

			default:
				return v.error(errors.New("either embedding or metadata must be provided"),
					"UpsertBatch-006",
					fmt.Sprintf("cannot upsert data point: %v", data.ID))
			}

			if len(upsertPoints) > 0 {
				_, err = v.client.Upsert(ctx, &qdrant.UpsertPoints{
					CollectionName: collection,
					Points:         upsertPoints,
				},
				)
				if err != nil {
					return v.error(err,
						"UpsertBatch-007",
						fmt.Sprintf("fail to upsert points: collection=%s", collection),
					)
				}
			}

			if len(updateVectors) > 0 {
				_, err = v.client.UpdateVectors(ctx, &qdrant.UpdatePointVectors{
					CollectionName: collection,
					Points:         updateVectors,
				},
				)
				if err != nil {
					return v.error(err,
						"UpsertBatch-008",
						fmt.Sprintf("fail to update vectors: collection=%s", collection),
					)
				}
			}

			if len(updatePayloads) > 0 {
				_, err := v.client.UpdateBatch(ctx,
					&qdrant.UpdateBatchPoints{
						CollectionName: collection,
						Operations:     updatePayloads,
					})

				if err != nil {
					return v.error(err,
						"UpsertBatch-009",
						fmt.Sprintf("fail to update payloads: collection=%s", collection),
					)
				}
			}
		}
	}

	return nil
}

func (v *VectorDBConnector) DeleteBatch(ctx context.Context, collection string, arrayID []string, batchNumber int) error {

	if batchNumber <= 0 {
		return v.error(
			errors.New("batchNumber must be greater than 0"),
			"DeleteBatch-001",
			"invalid batch number")
	}

	for start := 0; start < len(arrayID); start += batchNumber {
		if err := ctx.Err(); err != nil {
			return err
		}

		end := start + batchNumber
		if end > len(arrayID) {
			end = len(arrayID)
		}

		batch := arrayID[start:end]
		pointIDs := make([]*qdrant.PointId, 0, len(batch))

		for _, dataID := range batch {
			pointIDs = append(pointIDs, qdrant.NewID(dataID))
		}

		_, err := v.client.Delete(ctx,
			&qdrant.DeletePoints{
				CollectionName: collection,
				Points: &qdrant.PointsSelector{
					PointsSelectorOneOf: &qdrant.PointsSelector_Points{
						Points: &qdrant.PointsIdsList{
							Ids: pointIDs,
						},
					},
				},
			},
		)
		if err != nil {
			return v.error(err,
				"DeleteBatch-002",
				fmt.Sprintf("fail to delete points: collection=%s", collection),
			)
		}
	}

	return nil
}

func (v *VectorDBConnector) GetByID(ctx context.Context, collection string, dataID string) (vectordb.Data, error) {

	result, err := v.client.Get(ctx, &qdrant.GetPoints{
		CollectionName: collection,
		Ids:            []*qdrant.PointId{qdrant.NewID(dataID)},
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(true),
	},
	)

	if err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-001",
			fmt.Sprintf("fail to get data: collection=%s, id=%s", collection, dataID),
		)
	}

	if len(result) == 0 {
		return vectordb.Data{}, v.error(errors.New("data not found"),
			"GetByID-002",
			fmt.Sprintf("data does not exist: collection=%s, id=%s", collection, dataID),
		)
	}

	point := result[0]
	data := vectordb.Data{
		ID:               dataID,
		DenseEmbeddings:  make(map[vectordb.DenseEmbeddingName][]float32),
		SparseEmbeddings: make(map[vectordb.SparseEmbeddingName]vectordb.SparseEmbedding),
		Metadata:         make(map[string]interface{}),
	}

	if point.Vectors != nil {
		namedVectors := point.Vectors.GetVectors()
		for name, vector := range namedVectors.GetVectors() {
			if vector == nil {
				continue
			}

			if dense := vector.GetDenseVector(); dense != nil {
				data.DenseEmbeddings[vectordb.DenseEmbeddingName(name)] = dense.GetData()
				continue
			}

			if sparse := vector.GetSparseVector(); sparse != nil {
				data.SparseEmbeddings[vectordb.SparseEmbeddingName(name)] = vectordb.SparseEmbedding{
					Indices: sparse.GetIndices(),
					Values:  sparse.GetValues(),
				}
			}
		}
	}

	for key, value := range point.Payload {
		switch key {
		case vectordb.IDColumnName:
			if value != nil {
				data.ID = value.GetStringValue()
			}

		case vectordb.DocumentIDColumnName:
			if value != nil {
				data.DocumentID = value.GetStringValue()
			}

		case vectordb.ChunkIDColumnName:
			if value != nil {
				data.ChunkID = value.GetStringValue()
			}

		case vectordb.ContentColumnName:
			if value != nil {
				data.Content = value.GetStringValue()
			}

		default:
			data.Metadata[key] = qdrantValueToInterface(value)
		}
	}

	return data, nil
}

func (v *VectorDBConnector) CountVector(ctx context.Context, collection string) (int, error) {
	result, err := v.client.Count(ctx,
		&qdrant.CountPoints{
			CollectionName: collection,
			Exact:          &[]bool{true}[0],
		},
	)
	if err != nil {
		return 0, v.error(err,
			"CountVector-001",
			fmt.Sprintf("fail to count vectors: collection=%s", collection),
		)
	}

	return int(result), nil
}

func (v *VectorDBConnector) Close() error {
	// qdrant client does not expose a close method
	return nil
}

func (v *VectorDBConnector) IsKeyExists(ctx context.Context, collection string, dataID string) (bool, error) {
	result, err := v.client.Get(ctx,
		&qdrant.GetPoints{
			CollectionName: collection,
			Ids: []*qdrant.PointId{
				qdrant.NewID(dataID),
			},
		})

	if err != nil {
		return false, v.error(err,
			"IsKeyExists-001",
			fmt.Sprintf("fail to get point: %v", dataID))
	}

	return len(result) > 0, nil
}
