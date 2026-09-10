package dragonfly

import (
	"aiml-infrastructure/internal/base/cache"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type dragonflyClient struct {
	client *redis.Client
	config cache.Config
}

func NewDragonFlyClient(config cache.Config) (*dragonflyClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Username: config.Username,
		Password: config.Password,
		DB:       int(config.Db),
	})

	return &dragonflyClient{
		client: client,
		config: config,
	}, nil
}

func (d *dragonflyClient) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("dragonflyClient.(%v)(%v) %w", method, params, err)
}

func (d *dragonflyClient) Set(ctx context.Context, keyValue cache.KeyValue) error {

	if err := d.client.Set(
		ctx,
		keyValue.Key,
		keyValue.Value.Data,
		time.Duration(d.config.DefaultSecondExpiration)*time.Second,
	).Err(); err != nil {
		return d.error(err, "Set-001", keyValue.Key)
	}
	return nil

}

func (d *dragonflyClient) SetWithTTL(ctx context.Context, ttlSecond int64, keyValue cache.KeyValue) error {

	if err := d.client.Set(
		ctx,
		keyValue.Key,
		keyValue.Value.Data,
		time.Duration(ttlSecond)*time.Second,
	).Err(); err != nil {
		return d.error(err, "SetWithTTL-001", keyValue.Key)
	}
	return nil

}

func (d *dragonflyClient) SetBatch(ctx context.Context, keyValues ...cache.KeyValue) error {

	batchPipe := d.client.Pipeline()

	for _, keyValue := range keyValues {
		if err := batchPipe.Set(
			ctx,
			keyValue.Key,
			keyValue.Value.Data,
			time.Duration(d.config.DefaultSecondExpiration)*time.Second,
		).Err(); err != nil {
			return d.error(err, "SetBatch-001", keyValue.Key)
		}
	}

	_, err := batchPipe.Exec(ctx)
	if err != nil {
		return d.error(err, "SetBatch-002", "all batch fail to execute")
	}

	return nil

}

func (d *dragonflyClient) SetBatchWithTTL(ctx context.Context, ttlSecond int64, keyValues ...cache.KeyValue) error {

	batchPipe := d.client.Pipeline()

	for _, keyValue := range keyValues {
		if err := batchPipe.Set(
			ctx,
			keyValue.Key,
			keyValue.Value.Data,
			time.Duration(ttlSecond)*time.Second,
		).Err(); err != nil {
			return d.error(err, "SetBatchWithTTL-001", keyValue.Key)
		}
	}

	_, err := batchPipe.Exec(ctx)
	if err != nil {
		return d.error(err, "SetBatchWithTTL-002", "all batch fail to execute")
	}

	return nil
}

func (d *dragonflyClient) Get(ctx context.Context, key string) (cache.KeyValue, error) {

	value, err := d.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return cache.KeyValue{}, d.error(err, "Get-001", key)
		}
		return cache.KeyValue{}, d.error(err, "Get-002", key)
	}
	return cache.KeyValue{
		Key:   key,
		Value: cache.Value{Data: value},
	}, nil

}

func (d *dragonflyClient) GetBatch(ctx context.Context, keys ...string) ([]cache.KeyValue, error) {

	if len(keys) == 0 {
		return nil, d.error(errors.New("keys is empty"), "GetBatch-001", "keys is empty")
	}

	values, err := d.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, d.error(err, "GetBatch-002", keys)
	}

	result := make([]cache.KeyValue, 0, len(keys))
	for i, value := range values {
		kv := cache.KeyValue{
			Key: keys[i],
		}

		valueStr, ok := value.(string)
		if value != nil && !ok {
			return nil, d.error(fmt.Errorf("value is not string: %v", value), "GetBatch-003", keys[i])
		}

		if value != nil {
			kv.Value = cache.Value{Data: valueStr}
		}

		result = append(result, kv)
	}

	return result, nil
}

func (d *dragonflyClient) Delete(ctx context.Context, key string) (bool, error) {
	nDeleted, err := d.client.Del(ctx, key).Result()
	if err != nil {
		return false, d.error(err, "Delete-001", key)
	}

	return nDeleted > 0, nil
}

func (d *dragonflyClient) DeleteBatch(ctx context.Context, keys ...string) (int64, error) {

	if len(keys) == 0 {
		return 0, d.error(errors.New("keys is empty"), "DeleteBatch-001", "keys is empty")
	}

	nDeleted, err := d.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, d.error(err, "DeleteBatch-002", keys)
	}

	return nDeleted, nil
}

func (d *dragonflyClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := d.client.Exists(ctx, key).Result()
	if err != nil {
		return false, d.error(err, "Exists", key)
	}

	return count > 0, nil
}

func (d *dragonflyClient) FlushAll(ctx context.Context) error {
	if err := d.client.FlushAll(ctx).Err(); err != nil {
		return d.error(err, "FlushAll-001", "failed to flush all keys")
	}
	return nil
}

func (d *dragonflyClient) Close() error {
	err := d.client.Close()
	if err != nil {
		return d.error(err, "Close-001", "failed to close redis client")
	}
	return nil
}
