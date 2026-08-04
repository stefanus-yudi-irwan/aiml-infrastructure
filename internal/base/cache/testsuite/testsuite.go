package testsuite

import (
	"aiml-infrastructure/internal/base/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
	CacheConnector cache.ICacheConnector
}

func (c *CacheTestSuite) Test001Set() {

	keyValue := cache.KeyValue{
		Key: "test-key-001",
		Value: cache.Value{
			Data: "test-value-001",
		},
	}

	err := c.CacheConnector.Set(c.T().Context(), keyValue)
	assert.NoError(c.T(), err)
}

func (c *CacheTestSuite) Test002SetWithTTL() {

	keyValue := cache.KeyValue{
		Key: "test-key-002",
		Value: cache.Value{
			Data: "test-value-002",
		},
	}

	err := c.CacheConnector.SetWithTTL(c.T().Context(), 100, keyValue)
	assert.NoError(c.T(), err)
}
