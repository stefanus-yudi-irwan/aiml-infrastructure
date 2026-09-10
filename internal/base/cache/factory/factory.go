package factory

import (
	"aiml-infrastructure/internal/base/cache"
	"aiml-infrastructure/internal/base/cache/backend/dragonfly"
	"aiml-infrastructure/internal/base/cache/backend/memcached"
	"aiml-infrastructure/internal/base/cache/backend/redis"
	"aiml-infrastructure/internal/base/cache/backend/valkey"
	"fmt"
)

func NewCacheConnector(config cache.Config) (cache.ICacheConnector, error) {
	switch config.Backend {
	case cache.BackendRedis:
		return redis.NewRedisClient(config)
	case cache.BackendMemcached:
		return memcached.NewMemcachedClient(config)
	case cache.BackendValkey:
		return valkey.NewValkeyClient(config)
	case cache.BackendDragonfly:
		return dragonfly.NewDragonFlyClient(config)
	default:
		return nil, fmt.Errorf("unsupported cache backend: %s", config.Backend)
	}
}
