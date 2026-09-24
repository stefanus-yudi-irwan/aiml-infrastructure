package vectordb

import "context"

type ICollectionConfig interface {
	Validate() error
}

type IVectorDB interface {
	CreateCollection(ctx context.Context, config ICollectionConfig) error
	DeleteCollection(ctx context.Context, collection string) error
	IsCollectionExists(ctx context.Context, collection string) (bool, error)
	Upsert(ctx context.Context, collection string, data Data) error
	Delete(ctx context.Context, collection string, id string) error
	GetByID(ctx context.Context, collection string, id string) (Data, error)
	CountVector(ctx context.Context, collection string) (int, error)
	FlushAll(ctx context.Context, collection string) error
	Ping(ctx context.Context) error
	Close() error
	UpsertBatch(ctx context.Context, collection string, data []Data) error
	DeleteBatch(ctx context.Context, collection string, ids []string) error
}

type IVectorSearch interface {
	Search(ctx context.Context, collection string, request SearchRequest) ([]SearchResult, error)
}
