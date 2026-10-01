package vectordb

import "context"

type IVectorDB interface {
	Ping(ctx context.Context) error
	CreateCollection(ctx context.Context, collectionConfig interface{}) error
	IsCollectionExists(ctx context.Context, collection string) (bool, error)
	FlushCollection(ctx context.Context, collection string) error
	DeleteCollection(ctx context.Context, collection string) error
	FlushAll(ctx context.Context) error
	Upsert(ctx context.Context, collection string, data Data) error
	Delete(ctx context.Context, collection string, id string) error
	UpsertBatch(ctx context.Context, collection string, arrayData []Data, batchNumber int) error
	DeleteBatch(ctx context.Context, collection string, arrayID []string, batchNumber int) error
	IsKeyExists(ctx context.Context, collection string, id string) (bool, error)
	GetByID(ctx context.Context, collection string, id string) (Data, error)
	CountVector(ctx context.Context, collection string) (int, error)
	Close() error
}

type IVectorSearch interface {
	Search(ctx context.Context, collection string, request SearchRequest) ([]SearchResult, error)
}
