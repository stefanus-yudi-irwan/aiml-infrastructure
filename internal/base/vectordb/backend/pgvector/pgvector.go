package pgvector

import (
	"aiml-infrastructure/internal/base/vectordb"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type VectorDBConnector struct {
	client *pgxpool.Pool
}

func NewVectorDBConnector(config VectorDBConnectorConfig) (*VectorDBConnector, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgvector config: %w", err)
	}

	if config.MaxConns > 0 {
		poolConfig.MaxConns = config.MaxConns
	}
	if config.MinConns > 0 {
		poolConfig.MinConns = config.MinConns
	}
	if config.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = config.MaxConnLifetime
	}
	if config.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	}

	client, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with pgvector: %w", err)
	}

	return &VectorDBConnector{
		client: client,
	}, nil
}

func (v *VectorDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("VectorDBConnector.(%v)(%v) %w", method, params, err)
}

func (v *VectorDBConnector) CreateVectorExtension(ctx context.Context) error {
	_, err := v.client.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS vector`)
	if err != nil {
		return v.error(err,
			"CreateVectorExtension-001",
			fmt.Errorf("failed to create vector extension: %w", err))
	}
	return nil
}

func (v *VectorDBConnector) Ping(ctx context.Context) error {
	if err := v.client.Ping(ctx); err != nil {
		return v.error(err,
			"Ping-001")
	}
	return nil
}

func (v *VectorDBConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {
	collectionQuery, err := createCollectionQuery(collectionConfig)

	if err != nil {
		return v.error(err,
			"CreateCollection-001",
			"fail to create collection query")
	}

	_, err = v.client.Exec(ctx, collectionQuery)
	if err != nil {
		return v.error(err,
			"CreateCollection-001",
			fmt.Sprintf("fail to create collection: %s", collectionConfig.CollectionName))
	}

	return nil
}

func (v *VectorDBConnector) IsCollectionExists(ctx context.Context, schema string, collection string) (bool, error) {
	var exists bool

	query := `SELECT to_regclass($1) IS NOT NULL`

	tableName := pgx.Identifier{schema, collection}.Sanitize()

	err := v.client.QueryRow(ctx, query, tableName).Scan(&exists)
	if err != nil {
		return false, v.error(err,
			"IsCollectionExists-001",
			"fail to check if collection exists",
		)
	}

	return exists, nil
}

func (v *VectorDBConnector) FlushCollection(ctx context.Context, schema string, collection string) error {
	tableName := pgx.Identifier{schema, collection}.Sanitize()

	query := fmt.Sprintf(`TRUNCATE TABLE %s`, tableName)

	_, err := v.client.Exec(ctx, query)
	if err != nil {
		return v.error(err,
			"FlushCollection-001",
			fmt.Sprintf("failed to flush collection: %q.%q", schema, collection),
		)
	}

	return nil
}

func (v *VectorDBConnector) DeleteCollection(ctx context.Context, schema string, collection string) error {
	tableName := pgx.Identifier{schema, collection}.Sanitize()

	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tableName)

	_, err := v.client.Exec(ctx, query)
	if err != nil {
		return v.error(err,
			"DeleteCollection-001",
			fmt.Sprintf("failed to delete collection : %q.%q", schema, collection),
		)
	}

	return nil
}

func (v *VectorDBConnector) FlushAll(ctx context.Context, schema string) error {

	query := fmt.Sprintf(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = '%s'
	`, schema)

	rows, err := v.client.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}
	defer rows.Close()

	var collections []string

	for rows.Next() {
		var collection string
		if err := rows.Scan(&collection); err != nil {
			return fmt.Errorf("failed to scan collection: %w", err)
		}
		collections = append(collections, collection)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate collections: %w", err)
	}

	for _, collection := range collections {
		query := fmt.Sprintf(
			`DROP TABLE IF EXISTS %s`,
			pgx.Identifier{"public", collection}.Sanitize(),
		)

		if _, err := v.client.Exec(ctx, query); err != nil {
			return fmt.Errorf(
				"failed to delete collection %q: %w",
				collection,
				err,
			)
		}
	}

	return nil
}

func (v *VectorDBConnector) Upsert(ctx context.Context, schema string, collection string, data vectordb.Data) error {
	tableName := pgx.Identifier{schema, collection}.Sanitize()

	idColumn := pgx.Identifier{vectordb.IDColumnName}.Sanitize()
	documentIDColumn := pgx.Identifier{vectordb.DocumentIDColumnName}.Sanitize()
	chunkIDColumn := pgx.Identifier{vectordb.ChunkIDColumnName}.Sanitize()
	contentColumn := pgx.Identifier{vectordb.ContentColumnName}.Sanitize()
	embeddingColumn := pgx.Identifier{vectordb.EmbeddingColumnName}.Sanitize()
	metadataColumn := pgx.Identifier{vectordb.MetadataColumnName}.Sanitize()

	// A nil embedding is stored as SQL NULL.
	var embedding any
	if data.Embedding != nil {
		embedding = pgvector.NewVector(data.Embedding)
	}

	// A nil metadata map is stored as SQL NULL.
	var metadata any
	if data.Metadata != nil {
		metadataJSON, err := json.Marshal(data.Metadata)
		if err != nil {
			return v.error(err,
				"Upsert-001",
				"failed to marshal metadata",
			)
		}
		metadata = string(metadataJSON)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			%s, %s, %s, %s, %s, %s
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::vector, $6::jsonb)
		ON CONFLICT (%s)
		DO UPDATE SET
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s
	`,
		tableName,
		idColumn,
		documentIDColumn,
		chunkIDColumn,
		contentColumn,
		embeddingColumn,
		metadataColumn,
		idColumn,
		documentIDColumn, documentIDColumn,
		chunkIDColumn, chunkIDColumn,
		contentColumn, contentColumn,
		embeddingColumn, embeddingColumn,
		metadataColumn, metadataColumn,
	)

	_, err := v.client.Exec(ctx,
		query,
		data.ID,
		data.DocumentID,
		data.ChunkID,
		data.Content,
		embedding,
		metadata,
	)
	if err != nil {
		return v.error(err,
			"Upsert-002",
			"failed to upsert vector",
		)
	}

	return nil
}

func (v *VectorDBConnector) Delete(ctx context.Context, schema string, collection string, id string) error {
	tableName := pgx.Identifier{schema, collection}.Sanitize()
	idColumn := pgx.Identifier{vectordb.IDColumnName}.Sanitize()

	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE %s = $1::uuid
	`, tableName, idColumn)

	_, err := v.client.Exec(ctx, query, id)
	if err != nil {
		return v.error(err,
			"Delete-001",
			"failed to delete vector",
		)
	}

	return nil
}

func (v *VectorDBConnector) UpsertBatch(ctx context.Context, schema string, collection string, arrayData []vectordb.Data, batchNumber int) error {
	if batchNumber <= 0 {
		return v.error(fmt.Errorf("batch number must be greater than zero"),
			"UpsertBatch-001",
			"invalid batch number",
		)
	}

	if len(arrayData) == 0 {
		return nil
	}

	tableName := pgx.Identifier{schema, collection}.Sanitize()

	idColumn := pgx.Identifier{vectordb.IDColumnName}.Sanitize()
	documentIDColumn := pgx.Identifier{vectordb.DocumentIDColumnName}.Sanitize()
	chunkIDColumn := pgx.Identifier{vectordb.ChunkIDColumnName}.Sanitize()
	contentColumn := pgx.Identifier{vectordb.ContentColumnName}.Sanitize()
	embeddingColumn := pgx.Identifier{vectordb.EmbeddingColumnName}.Sanitize()
	metadataColumn := pgx.Identifier{vectordb.MetadataColumnName}.Sanitize()

	query := fmt.Sprintf(`
		INSERT INTO %s (
			%s, %s, %s, %s, %s, %s
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::vector, $6::jsonb)
		ON CONFLICT (%s)
		DO UPDATE SET
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s
	`,
		tableName,
		idColumn,
		documentIDColumn,
		chunkIDColumn,
		contentColumn,
		embeddingColumn,
		metadataColumn,
		idColumn,
		documentIDColumn, documentIDColumn,
		chunkIDColumn, chunkIDColumn,
		contentColumn, contentColumn,
		embeddingColumn, embeddingColumn,
		metadataColumn, metadataColumn,
	)

	for start := 0; start < len(arrayData); start += batchNumber {
		if err := ctx.Err(); err != nil {
			return v.error(err,
				"UpsertBatch-002",
				"context cancelled",
			)
		}

		end := min(start+batchNumber, len(arrayData))

		var batch pgx.Batch

		for i := start; i < end; i++ {
			if err := ctx.Err(); err != nil {
				return v.error(err,
					"UpsertBatch-003",
					"context cancelled while preparing batch",
				)
			}

			data := arrayData[i]

			var embedding any
			if data.Embedding != nil {
				embedding = pgvector.NewVector(data.Embedding)
			}

			var metadata any
			if data.Metadata != nil {
				metadataJSON, err := json.Marshal(data.Metadata)
				if err != nil {
					return v.error(err,
						"UpsertBatch-004",
						"failed to marshal metadata",
					)
				}
				metadata = string(metadataJSON)
			}

			batch.Queue(
				query,
				data.ID,
				data.DocumentID,
				data.ChunkID,
				data.Content,
				embedding,
				metadata,
			)
		}

		results := v.client.SendBatch(ctx, &batch)

		var batchErr error

		for i := start; i < end; i++ {
			_, err := results.Exec()
			if err != nil && batchErr == nil {
				batchErr = err
			}
		}

		closeErr := results.Close()

		if batchErr != nil {
			return v.error(batchErr,
				"UpsertBatch-005",
				"failed to execute batch upsert",
			)
		}

		if closeErr != nil {
			return v.error(closeErr,
				"UpsertBatch-006",
				"failed to close batch results",
			)
		}
	}

	return nil
}

func (v *VectorDBConnector) DeleteBatch(ctx context.Context, schema string, collection string, arrayID []string, batchNumber int) error {
	if batchNumber <= 0 {
		return v.error(fmt.Errorf("batch number must be greater than zero"),
			"DeleteBatch-001",
			"invalid batch number",
		)
	}

	if len(arrayID) == 0 {
		return nil
	}

	tableName := pgx.Identifier{schema, collection}.Sanitize()
	idColumn := pgx.Identifier{vectordb.IDColumnName}.Sanitize()

	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE %s = $1::uuid
	`, tableName, idColumn)

	for start := 0; start < len(arrayID); start += batchNumber {
		if err := ctx.Err(); err != nil {
			return v.error(err,
				"DeleteBatch-002",
				"context cancelled",
			)
		}

		end := min(start+batchNumber, len(arrayID))

		var batch pgx.Batch

		for i := start; i < end; i++ {
			if err := ctx.Err(); err != nil {
				return v.error(err,
					"DeleteBatch-003",
					"context cancelled while preparing batch",
				)
			}

			batch.Queue(query, arrayID[i])
		}

		results := v.client.SendBatch(ctx, &batch)

		var batchErr error

		for i := start; i < end; i++ {
			_, err := results.Exec()
			if err != nil && batchErr == nil {
				batchErr = err
			}
		}

		closeErr := results.Close()

		if batchErr != nil {
			return v.error(batchErr,
				"DeleteBatch-004",
				"failed to execute batch delete",
			)
		}

		if closeErr != nil {
			return v.error(closeErr,
				"DeleteBatch-005",
				"failed to close batch results",
			)
		}
	}

	return nil
}

func (v *VectorDBConnector) GetByID(ctx context.Context, schema string, collection string, id string) (vectordb.Data, error) {
	tableName := pgx.Identifier{schema, collection}.Sanitize()

	query := fmt.Sprintf(`
		SELECT
			%s, %s, %s, %s, %s::text, %s::text
		FROM %s
		WHERE %s = $1::uuid
	`,
		pgx.Identifier{vectordb.IDColumnName}.Sanitize(),
		pgx.Identifier{vectordb.DocumentIDColumnName}.Sanitize(),
		pgx.Identifier{vectordb.ChunkIDColumnName}.Sanitize(),
		pgx.Identifier{vectordb.ContentColumnName}.Sanitize(),
		pgx.Identifier{vectordb.EmbeddingColumnName}.Sanitize(),
		pgx.Identifier{vectordb.MetadataColumnName}.Sanitize(),
		tableName,
		pgx.Identifier{vectordb.IDColumnName}.Sanitize(),
	)

	var (
		data         vectordb.Data
		embedding    *string
		metadataJSON *string
	)

	err := v.client.QueryRow(ctx, query, id).Scan(
		&data.ID,
		&data.DocumentID,
		&data.ChunkID,
		&data.Content,
		&embedding,
		&metadataJSON,
	)
	if err != nil {
		return vectordb.Data{}, v.error(err,
			"GetByID-001",
			"failed to retrieve vector",
		)
	}

	if embedding != nil {
		var values []float32
		if err := json.Unmarshal([]byte(*embedding), &values); err != nil {
			return vectordb.Data{}, v.error(err,
				"GetByID-002",
				"failed to decode embedding",
			)
		}
		data.Embedding = values
	}

	if metadataJSON != nil {
		if err := json.Unmarshal([]byte(*metadataJSON), &data.Metadata); err != nil {
			return vectordb.Data{}, v.error(
				err,
				"GetByID-003",
				"failed to decode metadata",
			)
		}
	}

	return data, nil
}

func (v *VectorDBConnector) CountVector(ctx context.Context, schema string, collection string) (int, error) {
	var count int

	tableName := pgx.Identifier{schema, collection}.Sanitize()

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
	`, tableName)

	err := v.client.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, v.error(err,
			"CountVector-001",
			"failed to count vectors",
		)
	}

	return count, nil
}

func (v *VectorDBConnector) Close() error {
	if v.client != nil {
		v.client.Close()
	}

	return nil
}

func (v *VectorDBConnector) IsKeyExists(ctx context.Context, schema string, collection string, dataID string) (bool, error) {
	var exists bool

	tableName := pgx.Identifier{schema, collection}.Sanitize()

	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1
			FROM %s
			WHERE %s = $1::UUID
		)
	`, tableName, pgx.Identifier{vectordb.IDColumnName}.Sanitize())

	err := v.client.QueryRow(ctx, query, dataID).Scan(&exists)
	if err != nil {
		return false, v.error(
			err,
			"IsKeyExists-001",
			"failed to check if key exists",
		)
	}

	return exists, nil
}
