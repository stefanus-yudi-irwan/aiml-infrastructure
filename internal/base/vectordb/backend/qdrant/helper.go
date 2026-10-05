package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb"
	"encoding/json"
	"errors"
	"math"

	"github.com/qdrant/go-client/qdrant"
)

func createQdrantPoint(data vectordb.Data) (*qdrant.PointStruct, error) {

	var err error
	var id *qdrant.PointId
	var embeddings *qdrant.Vectors
	var metadata map[string]*qdrant.Value

	switch {
	case data.Embedding != nil && data.Metadata != nil && len(data.Metadata) > 0:
		id = qdrant.NewID(data.ID)
		embeddings = createQdrantMapEmbedding(data)
		metadata, err = createQdrantMapMetadata(data)
		if err != nil {
			return nil, err
		}
	case data.Embedding != nil && data.Metadata == nil:
		id = qdrant.NewID(data.ID)
		embeddings = createQdrantMapEmbedding(data)
		metadata = nil
	case data.Embedding == nil && data.Metadata != nil && len(data.Metadata) > 0:
		id = qdrant.NewID(data.ID)
		embeddings = nil
		metadata, err = createQdrantMapMetadata(data)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("either embedding or metadata must be provided")
	}

	return &qdrant.PointStruct{
		Id:      id,
		Vectors: embeddings,
		Payload: metadata,
	}, nil
}

func createQdrantMapIdentity(data vectordb.Data) map[string]*qdrant.Value {
	identityMap := map[string]*qdrant.Value{
		vectordb.DocumentIDColumnName: qdrant.NewValueString(data.DocumentID),
		vectordb.ChunkIDColumnName:    qdrant.NewValueString(data.ChunkID),
		vectordb.ContentColumnName:    qdrant.NewValueString(data.Content),
	}

	return identityMap
}

func createQdrantMapEmbedding(data vectordb.Data) *qdrant.Vectors {
	return qdrant.NewVectors(data.Embedding...)
}

func createQdrantMapMetadata(data vectordb.Data) (map[string]*qdrant.Value, error) {

	metadataJson, err := json.Marshal(data.Metadata)
	if err != nil {
		return nil, err
	}

	metadataMap := map[string]*qdrant.Value{
		vectordb.DocumentIDColumnName: qdrant.NewValueString(data.DocumentID),
		vectordb.ChunkIDColumnName:    qdrant.NewValueString(data.ChunkID),
		vectordb.ContentColumnName:    qdrant.NewValueString(data.Content),
		vectordb.MetadataColumnName:   qdrant.NewValueString(string(metadataJson)),
	}

	for key, value := range data.Metadata {
		switch v := value.(type) {
		case []string:
			metadataMap[key] = createStringListValue(v)
		case string:
			metadataMap[key] = &qdrant.Value{
				Kind: &qdrant.Value_StringValue{
					StringValue: v,
				},
			}
		case int:
			metadataMap[key] = &qdrant.Value{
				Kind: &qdrant.Value_IntegerValue{
					IntegerValue: int64(v),
				},
			}
		case float64:
			metadataMap[key] = &qdrant.Value{
				Kind: &qdrant.Value_DoubleValue{
					DoubleValue: v,
				},
			}
		case bool:
			metadataMap[key] = &qdrant.Value{
				Kind: &qdrant.Value_BoolValue{
					BoolValue: v,
				},
			}
		default:
			return nil, errors.New("unsupported metadata type")
		}
	}

	return metadataMap, nil
}

func createStringListValue(values []string) *qdrant.Value {
	list := make([]*qdrant.Value, len(values))

	for i, value := range values {
		list[i] = qdrant.NewValueString(value)
	}

	return &qdrant.Value{
		Kind: &qdrant.Value_ListValue{
			ListValue: &qdrant.ListValue{
				Values: list,
			},
		},
	}
}

func qdrantValueToInterface(value *qdrant.Value) interface{} {

	if value == nil {
		return nil
	}

	switch kind := value.Kind.(type) {

	case *qdrant.Value_StringValue:
		return kind.StringValue

	case *qdrant.Value_IntegerValue:
		return kind.IntegerValue

	case *qdrant.Value_DoubleValue:
		return kind.DoubleValue

	case *qdrant.Value_BoolValue:
		return kind.BoolValue

	case *qdrant.Value_ListValue:
		values := make([]interface{}, len(kind.ListValue.Values))

		for i, item := range kind.ListValue.Values {
			values[i] = qdrantValueToInterface(item)
		}

		return values

	case *qdrant.Value_StructValue:
		result := make(map[string]interface{})

		for key, item := range kind.StructValue.Fields {
			result[key] = qdrantValueToInterface(item)
		}

		return result

	default:
		return nil
	}
}

func NormalizeVector(vector []float32) []float32 {
	var normValue float32
	for _, value := range vector {
		normValue += value * value
	}
	normValue = float32(math.Sqrt(float64(normValue)))

	normVector := make([]float32, len(vector))
	for index, value := range vector {
		normVector[index] = value / normValue
	}

	return normVector
}

func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("vectors must have the same dimension")
	}

	if len(a) == 0 {
		return 0, errors.New("vectors must not be empty")
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, errors.New("cosine similarity is undefined for zero vectors")
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB)))), nil
}
