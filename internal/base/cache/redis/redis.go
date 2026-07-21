package redis

import (
	"aiml-infrastructure/internal/base/cache"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type redisClient struct {
	client *redis.Client
	config cache.CacheConfig
}

func NewRedisClient(config cache.CacheConfig) (cache.ICacheConnector, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Passsword,
		DB:       int(config.DB),
	})

	return &redisClient{
		client: client,
		config: config,
	}, nil
}

func (r *redisClient) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("redisClient.(%v)(%v) %w", method, params, err)
}

func (r *redisClient) Set(ctx context.Context, pair cache.Pair) error {
	return nil
}

func (r *redisClient) Get(ctx context.Context, key string) (cache.Pair, error) {
	return cache.Pair{}, nil
}

func (r *redisClient) Delete(ctx context.Context, key string) error {
	return nil
}

func (r *redisClient) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (r *redisClient) Close() error {
	return nil
}
