package cache

import "context"

type Value struct {
	Data  string
	Score float64
}

type Pair struct {
	Key        string
	Value      Value
	Expiration int64
}

type IPair interface {
	ConstructKey(keys ...string) string
	CreatePair(key string, value Value) Pair
}

type ICacheConnector interface {
	Set(ctx context.Context, pair Pair) error
	Get(ctx context.Context, key string) (Pair, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
}
