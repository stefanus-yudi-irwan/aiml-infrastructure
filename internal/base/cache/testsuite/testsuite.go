package testsuite

import (
	"aiml-infrastructure/internal/base/cache"
	"aiml-infrastructure/internal/base/cache/factory"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
	CacheConnector cache.ICacheConnector
}

func (c *CacheTestSuite) SetupCache(envFile string, backend string) {
	err := godotenv.Load(envFile)
	assert.NoError(c.T(), err)

	clientAddress := os.Getenv("CLIENT_ADDRESS")
	clientPassword := os.Getenv("CLIENT_PASSWORD")
	clientDB, err := strconv.Atoi(os.Getenv("CLIENT_DB"))
	assert.NoError(c.T(), err)

	cacheConfig := cache.Config{
		Backend:  backend,
		Address:  clientAddress,
		Password: clientPassword,
		Db:       int64(clientDB),
	}

	c.CacheConnector, err = factory.NewCacheConnector(cacheConfig)
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) TearDownCache() {
	err := c.CacheConnector.FlushAll(c.T().Context())
	assert.NoError(c.T(), err)

	err = c.CacheConnector.Close()
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) Test001Set() {

	key, err := cache.CreateCacheKey("test", "key", "001")
	assert.NoError(c.T(), err)
	value, err := cache.CreateCacheValue("test-value-001")
	assert.NoError(c.T(), err)
	keyValue, err := cache.CreateCacheKeyValue(key, value)
	assert.NoError(c.T(), err)

	err = c.CacheConnector.Set(c.T().Context(), keyValue)
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) Test002SetWithTTL() {

	key, err := cache.CreateCacheKey("test", "key", "002")
	assert.NoError(c.T(), err)
	value, err := cache.CreateCacheValue("test-value-002")
	assert.NoError(c.T(), err)
	keyValue, err := cache.CreateCacheKeyValue(key, value)
	assert.NoError(c.T(), err)

	err = c.CacheConnector.SetWithTTL(c.T().Context(), 100, keyValue)
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) Test003SetBatch()        {}
func (c *CacheTestSuite) Test004SetBatchWithTTL() {}
func (c *CacheTestSuite) Test005Get()             {}
func (c *CacheTestSuite) Test006GetBatch()        {}
func (c *CacheTestSuite) Test007Delete()          {}
func (c *CacheTestSuite) Test008DeleteBatch()     {}
func (c *CacheTestSuite) Test009Exists()          {}
