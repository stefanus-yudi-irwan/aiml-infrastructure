package cache

type Backend string

const (
	BackendRedis       Backend = "redis"
	BackendValkey      Backend = "valkey"
	BackendDragonflyDB Backend = "dragonflydb"
	BackendMemcached   Backend = "memcached"
)
