package memcached

import (
	"aiml-infrastructure/internal/base/cache"
	"aiml-infrastructure/internal/base/cache/factory"
	"aiml-infrastructure/internal/base/cache/testsuite"
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MemcachedTestSuite struct {
	testsuite.CacheTestSuite
}

func (d *MemcachedTestSuite) SetupSuite() {
	err := godotenv.Load(".env")
	assert.NoError(d.T(), err)

	clientAddress := os.Getenv("CLIENT_ADDRESS")
	clientPassword := os.Getenv("CLIENT_PASSWORD")
	clientDB, err := strconv.Atoi(os.Getenv("CLIENT_DB"))
	assert.NoError(d.T(), err)

	cacheConfig := cache.Config{
		Backend:  factory.BackendMemcached,
		Address:  clientAddress,
		Password: clientPassword,
		Db:       int64(clientDB),
	}

	d.CacheConnector, err = factory.NewCacheConnector(cacheConfig)
	assert.NoError(d.T(), err)
}

func (r *MemcachedTestSuite) TearDownSuite() {
	err := r.CacheConnector.FlushAll(r.T().Context())
	assert.NoError(r.T(), err)

	err = r.CacheConnector.Close()
	assert.NoError(r.T(), err)
}

func TestMemcachedTestSuite(t *testing.T) {
	suite.Run(t, new(MemcachedTestSuite))
}
