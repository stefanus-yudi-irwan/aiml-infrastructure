package valkey

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

type ValkeyTestSuite struct {
	testsuite.CacheTestSuite
}

func (r *ValkeyTestSuite) SetupSuite() {
	err := godotenv.Load(".env")
	assert.NoError(r.T(), err)

	clientAddress := os.Getenv("CLIENT_ADDRESS")
	clientPassword := os.Getenv("CLIENT_PASSWORD")
	clientDB, err := strconv.Atoi(os.Getenv("CLIENT_DB"))
	assert.NoError(r.T(), err)

	cacheConfig := cache.Config{
		Backend:  factory.BackendValkey,
		Address:  clientAddress,
		Password: clientPassword,
		Db:       int64(clientDB),
	}

	r.CacheConnector, err = factory.NewCacheConnector(cacheConfig)
	assert.NoError(r.T(), err)
}

func (r *ValkeyTestSuite) TearDownSuite() {
	// err := r.CacheConnector.FlushAll(r.T().Context())
	// assert.NoError(r.T(), err)

	// err = r.CacheConnector.Close()
	// assert.NoError(r.T(), err)
}

func TestValkeyTestSuite(t *testing.T) {
	suite.Run(t, new(ValkeyTestSuite))
}
