package memcached

import (
	"aiml-infrastructure/internal/base/cache/factory"
	"aiml-infrastructure/internal/base/cache/testsuite"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MemcachedTestSuite struct {
	testsuite.CacheTestSuite
}

func (d *MemcachedTestSuite) SetupSuite() {
	d.CacheTestSuite.SetupCache(".env", factory.BackendMemcached)
}

func (d *MemcachedTestSuite) TearDownSuite() {
	d.CacheTestSuite.TearDownCache()
}

func TestMemcachedTestSuite(t *testing.T) {
	suite.Run(t, new(MemcachedTestSuite))
}
