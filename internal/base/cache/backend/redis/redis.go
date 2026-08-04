package redis

import (
	"aiml-infrastructure/internal/base/cache"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisClient struct {
	client *redis.Client
	config cache.Config
}

func NewRedisClient(config cache.Config) (cache.ICacheConnector, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Password,
		DB:       int(config.Db),
	})

	return &redisClient{
		client: client,
		config: config,
	}, nil
}

func (r *redisClient) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("redisClient.(%v)(%v) %w", method, params, err)
}

func (r *redisClient) Set(ctx context.Context, keyValue cache.KeyValue) error {

	if err := r.client.Set(
		ctx,
		keyValue.Key,
		keyValue.Value.Data,
		r.config.DefaultSecondExpiration,
	).Err(); err != nil {
		return r.error(err, "Set-001", keyValue.Key)
	}
	return nil

}

func (r *redisClient) SetWithTTL(ctx context.Context, ttlSecond time.Duration, keyValue cache.KeyValue) error {

	if err := r.client.Set(
		ctx,
		keyValue.Key,
		keyValue.Value.Data,
		ttlSecond,
	).Err(); err != nil {
		return r.error(err, "SetWithTTL-001", keyValue.Key)
	}
	return nil

}

func (r *redisClient) SetBatch(ctx context.Context, keyValues ...cache.KeyValue) error {

	if len(keyValues) == 0 {
		return r.error(errors.New("keyValues is empty"), "SetBatch-001")
	}

	batchPipe := r.client.Pipeline()

	for _, keyValue := range keyValues {
		if err := batchPipe.Set(
			ctx,
			keyValue.Key,
			keyValue.Value.Data,
			r.config.DefaultSecondExpiration,
		).Err(); err != nil {
			return r.error(err, "SetBatch-002", keyValue.Key)
		}
	}

	_, err := batchPipe.Exec(ctx)
	if err != nil {
		return r.error(err, "SetBatch-003", "all batch fail to execute")
	}

	return nil

}

func (r *redisClient) SetBatchWithTTL(ctx context.Context, ttlSecond time.Duration, keyValues ...cache.KeyValue) error {

	if len(keyValues) == 0 {
		return r.error(errors.New("keyValues is empty"), "SetBatchWithTTL-001")
	}

	batchPipe := r.client.Pipeline()

	for _, keyValue := range keyValues {
		if err := batchPipe.Set(
			ctx,
			keyValue.Key,
			keyValue.Value.Data,
			ttlSecond,
		).Err(); err != nil {
			return r.error(err, "SetBatchWithTTL-002", keyValue.Key)
		}
	}

	_, err := batchPipe.Exec(ctx)
	if err != nil {
		return r.error(err, "SetBatchWithTTL-003", "all batch fail to execute")
	}

	return nil
}

func (r *redisClient) Get(ctx context.Context, key string) (cache.KeyValue, error) {

	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return cache.KeyValue{}, r.error(err, "Get-001", key)
		}
		return cache.KeyValue{}, r.error(err, "Get-002", key)
	}
	return cache.KeyValue{
		Key:   key,
		Value: cache.Value{Data: value},
	}, nil

}

func (r *redisClient) GetBatch(ctx context.Context, keys ...string) ([]cache.KeyValue, error) {

	if len(keys) == 0 {
		return nil, r.error(errors.New("keys is empty"), "GetBatch-001", "keys is empty")
	}

	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, r.error(err, "GetBatch-002", keys)
	}

	result := make([]cache.KeyValue, 0, len(keys))
	for i, value := range values {
		kv := cache.KeyValue{
			Key: keys[i],
		}

		valueStr, ok := value.(string)
		if value != nil && !ok {
			return nil, r.error(fmt.Errorf("value is not string: %v", value), "GetBatch-003", keys[i])
		}

		if value != nil {
			kv.Value = cache.Value{Data: valueStr}
		}

		result = append(result, kv)
	}

	return result, nil
}

func (r *redisClient) Delete(ctx context.Context, key string) (bool, error) {
	nDeleted, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return false, r.error(err, "Delete-001", key)
	}

	return nDeleted > 0, nil
}

func (r *redisClient) DeleteBatch(ctx context.Context, keys ...string) (int64, error) {

	if len(keys) == 0 {
		return 0, r.error(errors.New("keys is empty"), "DeleteBatch-001", "keys is empty")
	}

	nDeleted, err := r.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, r.error(err, "DeleteBatch-002", keys)
	}

	return nDeleted, nil
}

func (r *redisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, r.error(err, "Exists", key)
	}

	return count > 0, nil
}

func (r *redisClient) Close() error {
	err := r.client.Close()
	if err != nil {
		return r.error(err, "Close-001", "failed to close redis client")
	}
	return nil
}
