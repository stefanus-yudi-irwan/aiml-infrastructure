package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb"
	"errors"
	"fmt"
	"math"

	"github.com/qdrant/go-client/qdrant"
)

func createQdrantPoint(data vectordb.Data) (*qdrant.PointStruct, error) {
	metadata, err := createQdrantMapMetadata(data)
	if err != nil {
		return nil, err
	}

	embeddings := createQdrantMapEmbeddings(data)

	return &qdrant.PointStruct{
		Id:      qdrant.NewID(data.ID),
		Vectors: embeddings,
		Payload: metadata,
	}, nil
}

func createQdrantMapEmbeddings(data vectordb.Data) *qdrant.Vectors {
	embeddings := make(map[string]*qdrant.Vector, len(data.DenseEmbeddings))

	for name, embedding := range data.DenseEmbeddings {
		embeddings[fmt.Sprintf("%s.%s",
			vectordb.DenseEmbeddingsColumnName,
			name)] = qdrant.NewVector(embedding...)
	}

	for name, embedding := range data.SparseEmbeddings {
		embeddings[fmt.Sprintf("%s.%s",
			vectordb.SparseEmbeddingsColumnName,
			name)] = qdrant.NewVectorSparse(
			embedding.Indices,
			embedding.Values,
		)
	}

	return qdrant.NewVectorsMap(embeddings)
}

func createQdrantMapMetadata(data vectordb.Data) (map[string]*qdrant.Value, error) {
	metadataMap := map[string]*qdrant.Value{
		vectordb.DocumentIDColumnName: qdrant.NewValueString(data.DocumentID),
		vectordb.ChunkIDColumnName:    qdrant.NewValueString(data.ChunkID),
		vectordb.ContentColumnName:    qdrant.NewValueString(data.Content),
	}

	for key, value := range data.Metadata {
		metadataKey := fmt.Sprintf("%s.%s",
			vectordb.MetadataColumnName,
			key,
		)
		switch v := value.(type) {
		case []string:
			metadataMap[metadataKey] = createStringListValue(v)
		case []float64:
			metadataMap[metadataKey] = createFloatListValue(v)
		case []bool:
			metadataMap[metadataKey] = createBoolListValue(v)
		case string:
			metadataMap[metadataKey] = createStringValue(v)
		case float64:
			metadataMap[metadataKey] = createFloat64Value(v)
		case bool:
			metadataMap[metadataKey] = createBoolValue(v)
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

func createFloatListValue(values []float64) *qdrant.Value {
	list := make([]*qdrant.Value, len(values))
	for i, item := range values {
		list[i] = &qdrant.Value{
			Kind: &qdrant.Value_DoubleValue{
				DoubleValue: item,
			},
		}
	}

	return &qdrant.Value{
		Kind: &qdrant.Value_ListValue{
			ListValue: &qdrant.ListValue{
				Values: list,
			},
		},
	}
}

func createBoolListValue(values []bool) *qdrant.Value {
	list := make([]*qdrant.Value, len(values))
	for i, item := range values {
		list[i] = &qdrant.Value{
			Kind: &qdrant.Value_BoolValue{
				BoolValue: item,
			},
		}
	}

	return &qdrant.Value{
		Kind: &qdrant.Value_ListValue{
			ListValue: &qdrant.ListValue{
				Values: list,
			},
		},
	}
}

func createStringValue(value string) *qdrant.Value {
	return &qdrant.Value{
		Kind: &qdrant.Value_StringValue{
			StringValue: value,
		},
	}
}

func createFloat64Value(value float64) *qdrant.Value {
	return &qdrant.Value{
		Kind: &qdrant.Value_DoubleValue{
			DoubleValue: value,
		},
	}
}

func createBoolValue(value bool) *qdrant.Value {
	return &qdrant.Value{
		Kind: &qdrant.Value_BoolValue{
			BoolValue: value,
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

	case *qdrant.Value_DoubleValue:
		return kind.DoubleValue

	case *qdrant.Value_BoolValue:
		return kind.BoolValue

	case *qdrant.Value_ListValue:
		values := kind.ListValue.Values

		if len(values) == 0 {
			return []interface{}{}
		}

		switch values[0].Kind.(type) {
		case *qdrant.Value_StringValue:
			result := make([]string, len(values))
			for i, item := range values {
				if v, ok := item.Kind.(*qdrant.Value_StringValue); ok {
					result[i] = v.StringValue
				} else {
					return nil
				}
			}
			return result
		case *qdrant.Value_DoubleValue:
			result := make([]float64, len(values))
			for i, item := range values {
				if v, ok := item.Kind.(*qdrant.Value_DoubleValue); ok {
					result[i] = v.DoubleValue
				} else {
					return nil
				}
			}
			return result
		case *qdrant.Value_BoolValue:
			result := make([]bool, len(values))
			for i, item := range values {
				if v, ok := item.Kind.(*qdrant.Value_BoolValue); ok {
					result[i] = v.BoolValue
				} else {
					return nil
				}
			}
			return result
		default:
			return nil
		}

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
