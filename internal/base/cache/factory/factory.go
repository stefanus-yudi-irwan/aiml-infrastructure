package factory

import (
	"aiml-infrastructure/internal/base/cache"
	"aiml-infrastructure/internal/base/cache/backend/dragonfly"
	"aiml-infrastructure/internal/base/cache/backend/memcached"
	"aiml-infrastructure/internal/base/cache/backend/redis"
	"aiml-infrastructure/internal/base/cache/backend/valkey"
	"fmt"
)

const (
	BackendRedis     string = "redis"
	BackendValkey    string = "valkey"
	BackendDragonfly string = "dragonfly"
	BackendMemcached string = "memcached"
)

func NewCacheConnector(config cache.Config) (cache.ICacheConnector, error) {

	switch config.Backend {
	case BackendRedis:
		return redis.NewRedisClient(config)
	case BackendMemcached:
		return memcached.NewMemcachedClient(config)
	case BackendValkey:
		return valkey.NewValkeyClient(config)
	case BackendDragonfly:
		return dragonfly.NewDragonFlyClient(config)
	default:
		return nil, fmt.Errorf("unsupported cache backend: %s", config.Backend)
	}

}
