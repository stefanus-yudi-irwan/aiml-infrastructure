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

func (c *CacheTestSuite) SetupCache(envFile string, backend cache.BackendCache) {
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

	err = cacheConfig.Validate()
	assert.NoError(c.T(), err)

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
	keyValue := createKeyValueTest("test-set", 1)

	err := keyValue.ValidateKeyValue()
	assert.NoError(c.T(), err)

	err = c.CacheConnector.Set(c.T().Context(), keyValue)
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) Test002SetWithTTL() {
	keyValue := createKeyValueTest("test-set-with-ttl", 1)

	err := keyValue.ValidateKeyValue()
	assert.NoError(c.T(), err)

	err = c.CacheConnector.SetWithTTL(c.T().Context(), 100, keyValue)
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) Test003SetBatch() {
	keyValues := createBatchKeyValueTest("test-set-batch", 10)

	for _, keyValue := range keyValues {
		err := keyValue.ValidateKeyValue()
		assert.NoError(c.T(), err)
	}

	err := c.CacheConnector.SetBatch(c.T().Context(), keyValues...)
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) Test004SetBatchWithTTL() {
	keyValues := createBatchKeyValueTest("test-set-batch-with-ttl", 10)

	for _, keyValue := range keyValues {
		err := keyValue.ValidateKeyValue()
		assert.NoError(c.T(), err)
	}

	err := c.CacheConnector.SetBatchWithTTL(c.T().Context(), 100, keyValues...)
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) Test005Get() {
	keyValue := createKeyValueTest("test-get", 1)

	err := keyValue.ValidateKeyValue()
	assert.NoError(c.T(), err)

	err = c.CacheConnector.Set(c.T().Context(), keyValue)
	assert.NoError(c.T(), err)

	keyValueGet, err := c.CacheConnector.Get(c.T().Context(), keyValue.Key)
	assert.NoError(c.T(), err)

	err = keyValueGet.ValidateKeyValue()
	assert.NoError(c.T(), err)

	assert.Equal(c.T(), keyValue, keyValueGet)
}

func (c *CacheTestSuite) Test006GetBatch() {
	keyValues := createBatchKeyValueTest("test-get-batch", 10)

	for _, keyValue := range keyValues {
		err := keyValue.ValidateKeyValue()
		assert.NoError(c.T(), err)
	}

	err := c.CacheConnector.SetBatch(c.T().Context(), keyValues...)
	assert.NoError(c.T(), err)

	keys := extractKeys(keyValues...)
	keyValuesGet, err := c.CacheConnector.GetBatch(c.T().Context(), keys...)
	assert.NoError(c.T(), err)

	for _, keyValueGet := range keyValuesGet {
		err := keyValueGet.ValidateKeyValue()
		assert.NoError(c.T(), err)
	}

	assert.Equal(c.T(), keyValues, keyValuesGet)
}

func (c *CacheTestSuite) Test007Delete() {
	keyValue := createKeyValueTest("test-delete", 1)

	err := keyValue.ValidateKeyValue()
	assert.NoError(c.T(), err)

	err = c.CacheConnector.Set(c.T().Context(), keyValue)
	assert.NoError(c.T(), err)

	deleteFlag, err := c.CacheConnector.Delete(c.T().Context(), keyValue.Key)
	assert.NoError(c.T(), err)
	assert.Equal(c.T(), deleteFlag, true)
}

func (c *CacheTestSuite) Test008DeleteBatch() {
	keyValues := createBatchKeyValueTest("test-delete-batch", 10)

	for _, keyValue := range keyValues {
		err := keyValue.ValidateKeyValue()
		assert.NoError(c.T(), err)
	}

	err := c.CacheConnector.SetBatch(c.T().Context(), keyValues...)
	assert.NoError(c.T(), err)

	keys := extractKeys(keyValues...)
	countDeleted, err := c.CacheConnector.DeleteBatch(c.T().Context(), keys...)
	assert.NoError(c.T(), err)
	assert.Equal(c.T(), countDeleted, int64(10))
}

func (c *CacheTestSuite) Test009Exists() {
	keyValue := createKeyValueTest("test-exists", 1)

	err := keyValue.ValidateKeyValue()
	assert.NoError(c.T(), err)

	err = c.CacheConnector.Set(c.T().Context(), keyValue)
	assert.NoError(c.T(), err)

	existsFlag, err := c.CacheConnector.Exists(c.T().Context(), keyValue.Key)
	assert.NoError(c.T(), err)
	assert.Equal(c.T(), existsFlag, true)
}
