package redisvector

import (
	"aiml-infrastructure/internal/base/vectordb/backend/redisvector"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VectorDBTestSuite struct {
	suite.Suite
	VectorDBConnector *redisvector.RedisVectorConnector
}

func (v *VectorDBTestSuite) SetupSuite() {
	err := godotenv.Load(".env")
	assert.NoError(v.T(), err)

	clientAddress := os.Getenv("CLIENT_ADDRESS")
	clientPassword := os.Getenv("CLIENT_PASSWORD")
	clientUsername := os.Getenv("CLIENT_USERNAME")
	clientDB, err := strconv.Atoi(os.Getenv("CLIENT_DB"))
	assert.NoError(v.T(), err)

	redisVectorConfig := redisvector.VectorDBClientConfig{
		Address:  clientAddress,
		Username: clientUsername,
		Password: clientPassword,
		Db:       clientDB,
		Collections: []redisvector.CollectionConfig{
			createTestCollectionConfig(TEST_HASH_COLLECTION, redisvector.HASH_TYPE),
			createTestCollectionConfig(TEST_JSON_COLLECTION, redisvector.JSON_TYPE),
		},
	}

	err = redisVectorConfig.Validate()
	assert.NoError(v.T(), err)

	v.VectorDBConnector, err = redisvector.NewRedisVectorConnector(redisVectorConfig)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) TearDownSuite() {
	err := v.VectorDBConnector.FlushAll(v.T().Context())
	assert.NoError(v.T(), err)

	err = v.VectorDBConnector.Close()
	assert.NoError(v.T(), err)
}

func TestRedisVectorTestSuite(t *testing.T) {
	suite.Run(t, new(VectorDBTestSuite))
}

func (v *VectorDBTestSuite) Test001CreateCollection() {
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), "hash-collection")
	assert.NoError(v.T(), err)

	err = v.VectorDBConnector.CreateCollection(v.T().Context(), "json-collection")
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) Test002IsCollectionExists() {
	isExists, err := v.VectorDBConnector.IsCollectionExists(v.T().Context(), "hash-collection")
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), isExists, true)

	isExists, err = v.VectorDBConnector.IsCollectionExists(v.T().Context(), "json-collection")
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), isExists, true)
}

func (v *VectorDBTestSuite) Test003Upsert() {
	data := generateTestData("test-data-001")
	err := v.VectorDBConnector.Upsert(v.T().Context(), "hash-collection", data)
	assert.NoError(v.T(), err)

	err = v.VectorDBConnector.Upsert(v.T().Context(), "json-collection", data)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) Test004GetByID() {
	data, err := v.VectorDBConnector.GetByID(v.T().Context(), "hash-collection", "test-data-001")
	assert.NoError(v.T(), err)
	fmt.Println(data)

	data, err = v.VectorDBConnector.GetByID(v.T().Context(), "json-collection", "test-data-001")
	assert.NoError(v.T(), err)
	fmt.Println(data)
}

func (v *VectorDBTestSuite) Test005CountVector() {
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), "hash-collection")
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), count, 1)

	count, err = v.VectorDBConnector.CountVector(v.T().Context(), "json-collection")
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), count, 1)
}

func (v *VectorDBTestSuite) Test006Delete() {
	err := v.VectorDBConnector.Delete(v.T().Context(), "hash-collection", "test-data-001")
	assert.NoError(v.T(), err)

	err = v.VectorDBConnector.Delete(v.T().Context(), "json-collection", "test-data-001")
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) Test007DeleteCollection() {
	err := v.VectorDBConnector.DeleteCollection(v.T().Context(), "hash-collection")
	assert.NoError(v.T(), err)

	err = v.VectorDBConnector.DeleteCollection(v.T().Context(), "json-collection")
	assert.NoError(v.T(), err)
}
