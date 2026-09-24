package pgvector

import "fmt"

func createCollectionQuery(collectionConfig CollectionConfig) string {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s" (
			id UUID PRIMARY KEY,
			document_id UUID NOT NULL,
			chunk_id UUID NOT NULL,
			content TEXT NOT NULL,
			embedding vector(%d),
			metadata JSONB
		)
	`, collectionConfig.Name, collectionConfig.EmbeddingDimension)

	return query
}
