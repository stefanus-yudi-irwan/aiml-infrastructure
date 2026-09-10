package vectordb

import "context"

type IVectorDB interface {
	CreateCollection(ctx context.Context, config interface{}) error
	DeleteCollection(ctx context.Context, collection string) error
	IsCollectionExists(ctx context.Context, collection string) (bool, error)
	Upsert(ctx context.Context, collection string, data Data) error
	Delete(ctx context.Context, collection string, dataID string) error
	GetByID(ctx context.Context, collection string, dataID string) error
	CountVector(ctx context.Context, collection string) (int, error)
	Close() error
	//FlushAll(ctx context.Context, colleciton string) error
	//FlushAll(ctx context.Context, collection string) error
	//Ping(ctx context.Context) error
	// UpsertBatch(ctx context.Context, collection string, vector []Vector) error
	// DeleteBatch(ctx context.Context, collection string, vector []Vector) error
	// GetByIDs(ctx context.Context, collection string, vectorIds ...string) error
}

// type SearchVectorDB interface {
// 	Search(ctx context.Context, collection string, vector []float32, query string, topK int, filter Filter) ([]SearchResult, error)
// }
