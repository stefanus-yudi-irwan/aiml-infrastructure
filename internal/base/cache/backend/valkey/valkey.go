package valkey

import (
	"aiml-infrastructure/internal/base/cache"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type valkeyClient struct {
	client *redis.Client
	config cache.Config
}

func NewValkeyClient(config cache.Config) (*valkeyClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Password,
		DB:       int(config.Db),
	})

	return &valkeyClient{
		client: client,
		config: config,
	}, nil
}

func (v *valkeyClient) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("valkeyClient.(%v)(%v) %w", method, params, err)
}

func (v *valkeyClient) Set(ctx context.Context, keyValue cache.KeyValue) error {

	if err := v.client.Set(
		ctx,
		keyValue.Key,
		keyValue.Value.Data,
		time.Duration(v.config.DefaultSecondExpiration)*time.Second,
	).Err(); err != nil {
		return v.error(err, "Set-001", keyValue.Key)
	}
	return nil

}

func (v *valkeyClient) SetWithTTL(ctx context.Context, ttlSecond int64, keyValue cache.KeyValue) error {

	if err := v.client.Set(
		ctx,
		keyValue.Key,
		keyValue.Value.Data,
		time.Duration(ttlSecond)*time.Second,
	).Err(); err != nil {
		return v.error(err, "SetWithTTL-001", keyValue.Key)
	}
	return nil

}

func (v *valkeyClient) SetBatch(ctx context.Context, keyValues ...cache.KeyValue) error {

	batchPipe := v.client.Pipeline()

	for _, keyValue := range keyValues {
		if err := batchPipe.Set(
			ctx,
			keyValue.Key,
			keyValue.Value.Data,
			time.Duration(v.config.DefaultSecondExpiration)*time.Second,
		).Err(); err != nil {
			return v.error(err, "SetBatch-001", keyValue.Key)
		}
	}

	_, err := batchPipe.Exec(ctx)
	if err != nil {
		return v.error(err, "SetBatch-002", "all batch fail to execute")
	}

	return nil

}

func (v *valkeyClient) SetBatchWithTTL(ctx context.Context, ttlSecond int64, keyValues ...cache.KeyValue) error {

	batchPipe := v.client.Pipeline()

	for _, keyValue := range keyValues {
		if err := batchPipe.Set(
			ctx,
			keyValue.Key,
			keyValue.Value.Data,
			time.Duration(ttlSecond)*time.Second,
		).Err(); err != nil {
			return v.error(err, "SetBatchWithTTL-001", keyValue.Key)
		}
	}

	_, err := batchPipe.Exec(ctx)
	if err != nil {
		return v.error(err, "SetBatchWithTTL-002", "all batch fail to execute")
	}

	return nil
}

func (v *valkeyClient) Get(ctx context.Context, key string) (cache.KeyValue, error) {

	value, err := v.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return cache.KeyValue{}, v.error(err, "Get-001", key)
		}
		return cache.KeyValue{}, v.error(err, "Get-002", key)
	}
	return cache.KeyValue{
		Key:   key,
		Value: cache.Value{Data: value},
	}, nil

}

func (v *valkeyClient) GetBatch(ctx context.Context, keys ...string) ([]cache.KeyValue, error) {

	if len(keys) == 0 {
		return nil, v.error(errors.New("keys is empty"), "GetBatch-001", "keys is empty")
	}

	values, err := v.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, v.error(err, "GetBatch-002", keys)
	}

	result := make([]cache.KeyValue, 0, len(keys))
	for i, value := range values {
		kv := cache.KeyValue{
			Key: keys[i],
		}

		valueStr, ok := value.(string)
		if value != nil && !ok {
			return nil, v.error(fmt.Errorf("value is not string: %v", value), "GetBatch-003", keys[i])
		}

		if value != nil {
			kv.Value = cache.Value{Data: valueStr}
		}

		result = append(result, kv)
	}

	return result, nil
}

func (v *valkeyClient) Delete(ctx context.Context, key string) (bool, error) {
	nDeleted, err := v.client.Del(ctx, key).Result()
	if err != nil {
		return false, v.error(err, "Delete-001", key)
	}

	return nDeleted > 0, nil
}

func (v *valkeyClient) DeleteBatch(ctx context.Context, keys ...string) (int64, error) {

	if len(keys) == 0 {
		return 0, v.error(errors.New("keys is empty"), "DeleteBatch-001", "keys is empty")
	}

	nDeleted, err := v.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, v.error(err, "DeleteBatch-002", keys)
	}

	return nDeleted, nil
}

func (v *valkeyClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := v.client.Exists(ctx, key).Result()
	if err != nil {
		return false, v.error(err, "Exists", key)
	}

	return count > 0, nil
}

func (v *valkeyClient) FlushAll(ctx context.Context) error {
	if err := v.client.FlushAll(ctx).Err(); err != nil {
		return v.error(err, "FlushAll-001", "failed to flush all keys")
	}
	return nil
}

func (v *valkeyClient) Close() error {
	err := v.client.Close()
	if err != nil {
		return v.error(err, "Close-001", "failed to close valkey client")
	}
	return nil
}
