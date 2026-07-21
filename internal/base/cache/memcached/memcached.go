package memcached

import (
	"aiml-infrastructure/internal/base/cache"
	"context"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
)

type memcachedClient struct {
	client *memcache.Client
	cfg    cache.CacheConfig
}

func NewMemcachedClient(cfg cache.CacheConfig) (cache.ICacheConnector, error) {

	client := memcache.New(cfg.Address)

	return &memcachedClient{
		client: client,
		cfg:    cfg,
	}, nil
}

func (r *memcachedClient) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("memcachedCache.(%v)(%v) %w", method, params, err)
}

func (m *memcachedClient) Set(ctx context.Context, pair cache.Pair) error {
	return nil
}
func (m *memcachedClient) Get(ctx context.Context, key string) (cache.Pair, error) {
	return cache.Pair{}, nil
}

func (m *memcachedClient) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *memcachedClient) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (r *memcachedClient) Close() error {
	return nil
}
