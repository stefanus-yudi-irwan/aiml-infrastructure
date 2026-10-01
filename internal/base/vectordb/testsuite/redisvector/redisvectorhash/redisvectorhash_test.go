package redisvector

import (
	"aiml-infrastructure/internal/base/vectordb/backend/redisvector/redisvectorhash"
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
	VectorDBConnector *redisvectorhash.VectorDBConnector
	TestCollection    redisvectorhash.CollectionConfig
}

func (v *VectorDBTestSuite) SetupSuite() {
	err := godotenv.Load(".env")
	assert.NoError(v.T(), err)

	clientAddress := os.Getenv("CLIENT_ADDRESS")
	clientPassword := os.Getenv("CLIENT_PASSWORD")
	clientUsername := os.Getenv("CLIENT_USERNAME")
	clientDB, err := strconv.Atoi(os.Getenv("CLIENT_DB"))
	assert.NoError(v.T(), err)

	vectorDBConnectorConfig := redisvectorhash.VectorDBConnectorConfig{
		Address:  clientAddress,
		Username: clientUsername,
		Password: clientPassword,
		Db:       clientDB,
	}

	err = vectorDBConnectorConfig.Validate()
	assert.NoError(v.T(), err)

	v.VectorDBConnector, err = redisvectorhash.NewVectorDBConnector(vectorDBConnectorConfig)
	assert.NoError(v.T(), err)

	v.TestCollection = createTestCollectionConfig("test_collection")
}

func (v *VectorDBTestSuite) TearDownSuite() {
	err := v.VectorDBConnector.FlushAll(v.T().Context())
	assert.NoError(v.T(), err)

	err = v.VectorDBConnector.Close()
	assert.NoError(v.T(), err)
}

func TestSuiteVectorDBRedisVectorHash(t *testing.T) {
	suite.Run(t, new(VectorDBTestSuite))
}

func (v *VectorDBTestSuite) Test001ManageCollection() {

	// create test collection inside vector db
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), v.TestCollection)
	assert.NoError(v.T(), err)

	// check if collection exists in vector db
	isExists, err := v.VectorDBConnector.IsCollectionExists(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), true, isExists)

	// delete collection inside vector db
	err = v.VectorDBConnector.DeleteCollection(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)

	// check if collection still exists in vector db
	isExists, err = v.VectorDBConnector.IsCollectionExists(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), false, isExists)
}

func (v *VectorDBTestSuite) Test002ManageSingleData() {

	// create test collection inside vector db
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), v.TestCollection)
	assert.NoError(v.T(), err)

	// create mock test data
	dataInsert := generateTestData("test-data-001")

	// validate mock test data
	err = dataInsert.ValidateIdentifier()
	assert.NoError(v.T(), err)
	err = dataInsert.ValidateContent()
	assert.NoError(v.T(), err)
	err = dataInsert.ValidateEmbedding()
	assert.NoError(v.T(), err)
	err = dataInsert.ValidateMetadata()
	assert.NoError(v.T(), err)

	// insert test data into vector db
	err = v.VectorDBConnector.Upsert(v.T().Context(), v.TestCollection.Name, dataInsert)
	assert.NoError(v.T(), err)

	// check if test data is inside vector db
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 1, count)

	dataGet, err := v.VectorDBConnector.GetByID(v.T().Context(), v.TestCollection.Name, "test-data-001")
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), dataInsert.ID, dataGet.ID)
	assert.Equal(v.T(), dataInsert.DocumentID, dataGet.DocumentID)
	assert.Equal(v.T(), dataInsert.ChunkID, dataGet.ChunkID)
	assert.Equal(v.T(), dataInsert.Content, dataGet.Content)
	assert.Equal(v.T(), dataInsert.Embedding, dataGet.Embedding)
	//assert.Equal(v.T(), dataInsert.Metadata, dataGet.Metadata)

	// test delete stored data
	err = v.VectorDBConnector.Delete(v.T().Context(), v.TestCollection.Name, "test-data-001")
	assert.NoError(v.T(), err)

	// check if collection is already empty
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 0, count)

	// delete collection inside vector db
	err = v.VectorDBConnector.DeleteCollection(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)

}

func (v *VectorDBTestSuite) Test003ManageBatchdata() {

	// create test collection inside vector db
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), v.TestCollection)
	assert.NoError(v.T(), err)

	// create bulk mock test data
	dataInsert := generateBulkTestData(1000, "test-data")
	dataInsertID := generateBulkTestDataID(500, "test-data")

	// validate all mock test data
	for _, data := range dataInsert {
		err = data.ValidateIdentifier()
		assert.NoError(v.T(), err)
		err = data.ValidateContent()
		assert.NoError(v.T(), err)
		err = data.ValidateEmbedding()
		assert.NoError(v.T(), err)
		err = data.ValidateMetadata()
		assert.NoError(v.T(), err)
	}

	// insert test data into vector db
	err = v.VectorDBConnector.UpsertBatch(v.T().Context(), v.TestCollection.Name, dataInsert, 100)
	assert.NoError(v.T(), err)

	// check if test data is inside vector db
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 1000, count)

	// check if all stored test data comply with input data
	for index, data := range dataInsert {
		dataGet, err := v.VectorDBConnector.GetByID(v.T().Context(), v.TestCollection.Name, fmt.Sprintf("%s-%d", "test-data", index))
		assert.NoError(v.T(), err)
		assert.Equal(v.T(), data.ID, dataGet.ID)
		assert.Equal(v.T(), data.DocumentID, dataGet.DocumentID)
		assert.Equal(v.T(), data.ChunkID, dataGet.ChunkID)
		assert.Equal(v.T(), data.Content, dataGet.Content)
		assert.Equal(v.T(), data.Embedding, dataGet.Embedding)
		//assert.Equal(v.T(), data.Metadata, dataGet.Metadata)
	}

	// delete all key in batch
	err = v.VectorDBConnector.DeleteBatch(v.T().Context(), v.TestCollection.Name, dataInsertID, 100)
	assert.NoError(v.T(), err)

	// check collection after delete half number of keys
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 500, count)

	// test delete all stored data
	err = v.VectorDBConnector.FlushCollection(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)

	// check if collection is already empty
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 0, count)

	// delete collection inside vector db
	err = v.VectorDBConnector.DeleteCollection(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)

}
