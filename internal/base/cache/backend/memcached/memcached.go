package memcached

import (
	"aiml-infrastructure/internal/base/cache"
	"context"
	"errors"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
)

type memcachedClient struct {
	client *memcache.Client
	config cache.Config
}

func NewMemcachedClient(config cache.Config) (*memcachedClient, error) {
	client := memcache.New(config.Address)

	return &memcachedClient{
		client: client,
		config: config,
	}, nil
}

func (r *memcachedClient) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("memcachedCache.(%v)(%v) %w", method, params, err)
}

func (m *memcachedClient) Set(ctx context.Context, keyValue cache.KeyValue) error {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	item := &memcache.Item{
		Key:        keyValue.Key,
		Value:      []byte(keyValue.Value.Data),
		Expiration: int32(m.config.DefaultSecondExpiration),
	}

	if err := m.client.Set(item); err != nil {
		return m.error(err, "Set-001", keyValue.Key)
	}

	return nil
}

func (m *memcachedClient) SetWithTTL(ctx context.Context, ttlSecond int64, keyValue cache.KeyValue) error {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	item := &memcache.Item{
		Key:        keyValue.Key,
		Value:      []byte(keyValue.Value.Data),
		Expiration: int32(ttlSecond),
	}

	if err := m.client.Set(item); err != nil {
		return m.error(err, "SetWithTTL-001", keyValue.Key)
	}

	return nil
}

func (m *memcachedClient) SetBatch(ctx context.Context, keyValues ...cache.KeyValue) error {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	if len(keyValues) == 0 {
		return m.error(errors.New("keyValues is empty"), "SetBatch-001")
	}

	for _, keyValue := range keyValues {
		item := &memcache.Item{
			Key:        keyValue.Key,
			Value:      []byte(keyValue.Value.Data),
			Expiration: int32(m.config.DefaultSecondExpiration),
		}

		if err := m.client.Set(item); err != nil {
			return m.error(err, "SetBatch-002", keyValue.Key)
		}
	}

	return nil
}

func (m *memcachedClient) SetBatchWithTTL(ctx context.Context, ttlSecond int64, KeyValues ...cache.KeyValue) error {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	if len(KeyValues) == 0 {
		return m.error(errors.New("keyValues is empty"), "SetBatchWithTTL-001")
	}

	for _, keyValue := range KeyValues {
		item := &memcache.Item{
			Key:        keyValue.Key,
			Value:      []byte(keyValue.Value.Data),
			Expiration: int32(ttlSecond),
		}

		if err := m.client.Set(item); err != nil {
			return m.error(err, "SetBatchWithTTL-002", keyValue.Key)
		}
	}

	return nil
}

func (m *memcachedClient) Get(ctx context.Context, key string) (cache.KeyValue, error) {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	item, err := m.client.Get(key)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return cache.KeyValue{}, m.error(err, "Get-001", key)
		}
		return cache.KeyValue{}, m.error(err, "Get-002", key)
	}

	return cache.KeyValue{
		Key: item.Key,
		Value: cache.Value{
			Data: string(item.Value),
		},
	}, nil
}

func (m *memcachedClient) GetBatch(ctx context.Context, keys ...string) ([]cache.KeyValue, error) {

	if len(keys) == 0 {
		return nil, m.error(errors.New("keys is empty"), "GetBatch-001")
	}

	items, err := m.client.GetMulti(keys)
	if err != nil {
		return nil, m.error(err, "GetBatch-002")
	}

	result := make([]cache.KeyValue, 0, len(items))
	for _, key := range keys {

		kv := cache.KeyValue{
			Key: key,
		}

		if item, ok := items[key]; ok {
			kv.Value = cache.Value{
				Data: string(item.Value),
			}
		}

		result = append(result, kv)
	}

	return result, nil
}

func (m *memcachedClient) Delete(ctx context.Context, key string) (bool, error) {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	err := m.client.Delete(key)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return false, m.error(err, "Delete-001", key)
		}
		return false, m.error(err, "Delete-002", key)
	}

	return true, nil
}

func (m *memcachedClient) DeleteBatch(ctx context.Context, keys ...string) (int64, error) {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	if len(keys) == 0 {
		return 0, m.error(errors.New("keys is empty"), "DeleteBatch-001")
	}

	var deletedCount int64
	for _, key := range keys {
		err := m.client.Delete(key)
		if err != nil {
			if errors.Is(err, memcache.ErrCacheMiss) {
				continue // Key not found, consider it as not deleted
			}
			return deletedCount, m.error(err, "DeleteBatch-002", key)
		}
		deletedCount++
	}

	return deletedCount, nil
}

func (m *memcachedClient) Exists(ctx context.Context, key string) (bool, error) {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	_, err := m.client.Get(key)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return false, nil // Key does not exist
		}
		return false, m.error(err, "Exists-001", key)
	}

	return true, nil
}

func (m *memcachedClient) FlushAll(ctx context.Context) error {

	_ = ctx // context is not used in memcached client, but we keep it for interface compatibility

	err := m.client.FlushAll()
	if err != nil {
		return m.error(err, "FlushAll-001")
	}

	return nil
}

func (m *memcachedClient) Close() error {
	// gomemcache does not expose a Close method.
	return nil
}
