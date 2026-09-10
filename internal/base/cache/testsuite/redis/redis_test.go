package redis_test

import (
	"aiml-infrastructure/internal/base/cache"
	"aiml-infrastructure/internal/base/cache/testsuite"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RedisTestSuite struct {
	testsuite.CacheTestSuite
}

func (r *RedisTestSuite) SetupSuite() {
	r.CacheTestSuite.SetupCache(".env", cache.BackendRedis)
}

func (r *RedisTestSuite) TearDownSuite() {
	r.CacheTestSuite.TearDownCache()
}

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}
