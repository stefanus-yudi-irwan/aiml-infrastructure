package cache

import (
	"context"
)

type ICacheConnector interface {
	Set(ctx context.Context, keyValue KeyValue) error
	SetWithTTL(ctx context.Context, ttlSecond int64, KeyValue KeyValue) error
	SetBatch(ctx context.Context, keyValues ...KeyValue) error
	SetBatchWithTTL(ctx context.Context, ttlSecond int64, KeyValues ...KeyValue) error
	Get(ctx context.Context, key string) (KeyValue, error)
	GetBatch(ctx context.Context, keys ...string) ([]KeyValue, error)
	Delete(ctx context.Context, key string) (bool, error)
	DeleteBatch(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, key string) (bool, error)
	FlushAll(ctx context.Context) error
	Close() error
}

// chunk size for batch get and delete
