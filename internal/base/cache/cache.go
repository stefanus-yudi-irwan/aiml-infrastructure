package cache

import "aiml-infrastructure/internal/base/cache/redis"

type CacheConfig struct {
	Backend           Backend
	Address           string
	Username          string
	Passsword         string
	DB                int64
	DefaultExpiration int64
}

func NewCacheConnector(config CacheConfig) (ICacheConnector, error) {
	switch config.Backend {
	case BackendRedis, BackendValkey, BackendDragonflyDB:
		return redis.NewRedisClient()
	}
}
