package pgvector

import (
	"aiml-infrastructure/internal/base/vectordb"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
)

func createCollectionQuery(collectionConfig CollectionConfig) (string, error) {

	schemaName := pgx.Identifier{collectionConfig.SchemaName}.Sanitize()
	tableName := pgx.Identifier{
		collectionConfig.SchemaName,
		collectionConfig.CollectionName,
	}.Sanitize()

	var query strings.Builder

	fmt.Fprintf(&query, `
		CREATE SCHEMA IF NOT EXISTS %s;
	`, schemaName)

	fmt.Fprintf(&query, `
		CREATE TABLE IF NOT EXISTS %s (
			%s UUID PRIMARY KEY,
			%s UUID NOT NULL,
			%s UUID NOT NULL,
			%s TEXT NOT NULL,
			%s vector(%d),
			%s JSONB
		);
	`,
		tableName,
		vectordb.IDColumnName,
		vectordb.DocumentIDColumnName,
		vectordb.ChunkIDColumnName,
		vectordb.ContentColumnName,
		vectordb.EmbeddingColumnName,
		collectionConfig.EmbeddingDimension,
		vectordb.MetadataColumnName,
	)

	// Select the distance operator class.
	var operatorClass string
	switch collectionConfig.Distance {
	case DistanceCosine:
		operatorClass = "vector_cosine_ops"
	case DistanceEuclidean:
		operatorClass = "vector_l2_ops"
	case DistanceInnerProduct:
		operatorClass = "vector_ip_ops"
	case DistanceManhattan:
		operatorClass = "vector_l1_ops"
	default:
		return "", fmt.Errorf("unsupported distance type: %s", collectionConfig.Distance)
	}

	// Select the index type and its parameters.
	var indexQuery string

	switch collectionConfig.Index {
	case HNSWIndex:
		indexName := pgx.Identifier{
			collectionConfig.CollectionName + "_embedding_hnsw_idx",
		}.Sanitize()
		indexQuery = fmt.Sprintf(`
			CREATE INDEX IF NOT EXISTS %s
			ON %s
			USING hnsw (%s %s)
			WITH (m = %d, ef_construction = %d);
		`,
			indexName,
			tableName,
			vectordb.EmbeddingColumnName,
			operatorClass,
			collectionConfig.HNSW.M,
			collectionConfig.HNSW.EFConstruction,
		)

	case IVFFlatIndex:
		if collectionConfig.Distance == DistanceManhattan {
			return "", fmt.Errorf(
				"ivfflat does not support distance type: %s",
				collectionConfig.Distance,
			)
		}

		indexName := pgx.Identifier{
			collectionConfig.CollectionName + "_embedding_ivfflat_idx",
		}.Sanitize()

		indexQuery = fmt.Sprintf(`
			CREATE INDEX IF NOT EXISTS %s
			ON %s
			USING ivfflat (%s %s)
			WITH (lists = %d);
		`,
			indexName,
			tableName,
			vectordb.EmbeddingColumnName,
			operatorClass,
			collectionConfig.IVF.IVFFlatLists,
		)

	default:
		return "", fmt.Errorf("unsupported index type: %s", collectionConfig.Index)
	}

	query.WriteString(indexQuery)

	return query.String(), nil
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
