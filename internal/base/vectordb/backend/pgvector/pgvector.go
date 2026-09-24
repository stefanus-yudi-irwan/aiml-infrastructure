package pgvector

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VectorDBConnector struct {
	client *pgxpool.Pool
}

func NewVectorDBConnector(ctx context.Context, config VectorDBConnectorConfig) (*VectorDBConnector, error) {

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
		return nil, fmt.Errorf("failed to parse PostgreSQL config: %w", err)
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

	client, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL client: %w", err)
	}

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
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

func (v *VectorDBConnector) CreateCollection(ctx context.Context, collectionConfig CollectionConfig) error {
	collectionQuery := createCollectionQuery(collectionConfig)
	_, err := v.client.Exec(ctx, collectionQuery)
	if err != nil {
		return v.error(err,
			"CreateCollection-001",
			fmt.Sprintf("fail to create collection: %s", collectionConfig.Name))
	}
	return nil
}

func (v *VectorDBConnector) Close() {
	if v.client != nil {
		v.client.Close()
	}
}
