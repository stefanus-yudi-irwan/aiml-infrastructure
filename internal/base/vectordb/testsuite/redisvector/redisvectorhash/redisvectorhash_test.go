package redisvector

import (
	"aiml-infrastructure/internal/base/vectordb/backend/redisvector/redisvectorhash"
	"os"
	"strconv"
	"testing"

	"github.com/google/uuid"
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
	dataKey := uuid.New().String()
	dataInsert := generateTestData(dataKey)

	// validate mock test data
	err = dataInsert.Validate()
	assert.NoError(v.T(), err)

	// insert test data into vector db
	err = v.VectorDBConnector.Upsert(v.T().Context(), v.TestCollection.Name, dataInsert)
	assert.NoError(v.T(), err)

	// check if test data is inside vector db
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 1, count)

	dataGet, err := v.VectorDBConnector.GetByID(v.T().Context(), v.TestCollection.Name, dataKey)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), dataInsert.ID, dataGet.ID)
	assert.Equal(v.T(), dataInsert.DocumentID, dataGet.DocumentID)
	assert.Equal(v.T(), dataInsert.ChunkID, dataGet.ChunkID)
	assert.Equal(v.T(), dataInsert.Content, dataGet.Content)
	assert.Equal(v.T(), dataInsert.SparseEmbeddings, dataGet.SparseEmbeddings)

	for key, vector := range dataInsert.DenseEmbeddings {
		cosineSimilarityScore, err := redisvectorhash.CosineSimilarity(
			vector,
			dataGet.DenseEmbeddings[key])
		assert.NoError(v.T(), err)
		assert.LessOrEqual(v.T(), float32(0.00001), cosineSimilarityScore)
	}
	assert.Equal(v.T(), dataInsert.Metadata, dataGet.Metadata)

	// test delete stored data
	err = v.VectorDBConnector.Delete(v.T().Context(), v.TestCollection.Name, dataKey)
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
	batchNumber := 10
	dataInsertID := generateBulkTestDataID(100)
	dataInsertID500 := dataInsertID[:50]
	dataInsert := generateBulkTestData(dataInsertID)

	// validate all mock test data
	for _, data := range dataInsert {
		err = data.Validate()
		assert.NoError(v.T(), err)
	}

	// insert test data into vector db
	err = v.VectorDBConnector.UpsertBatch(v.T().Context(), v.TestCollection.Name, dataInsert, batchNumber)
	assert.NoError(v.T(), err)

	// check if test data is inside vector db
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), len(dataInsertID), count)

	// check if all stored test data comply with input data
	for _, data := range dataInsert {
		dataGet, err := v.VectorDBConnector.GetByID(v.T().Context(), v.TestCollection.Name, data.ID)
		assert.NoError(v.T(), err)
		assert.Equal(v.T(), data.ID, dataGet.ID)
		assert.Equal(v.T(), data.DocumentID, dataGet.DocumentID)
		assert.Equal(v.T(), data.ChunkID, dataGet.ChunkID)
		assert.Equal(v.T(), data.Content, dataGet.Content)
		assert.Equal(v.T(), data.SparseEmbeddings, dataGet.SparseEmbeddings)

		for key, vector := range data.DenseEmbeddings {
			cosineSimilarityScore, err := redisvectorhash.CosineSimilarity(
				vector,
				dataGet.DenseEmbeddings[key])
			assert.NoError(v.T(), err)
			assert.LessOrEqual(v.T(), float32(0.00001), cosineSimilarityScore)
		}

		assert.Equal(v.T(), data.Metadata, dataGet.Metadata)
	}

	// delete all key in batch
	err = v.VectorDBConnector.DeleteBatch(v.T().Context(), v.TestCollection.Name, dataInsertID500, batchNumber)
	assert.NoError(v.T(), err)

	// check collection after delete half number of keys
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), len(dataInsertID)-len(dataInsertID500), count)

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

func (v *VectorDBTestSuite) Test004ManageSinglePartialData() {

	// create test collection inside vector db
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), v.TestCollection)
	assert.NoError(v.T(), err)

	dataKey := uuid.New().String()
	dataIdentity := generateTestDataIdentity(dataKey)

	// insert test data with dense embedding only into vector db
	dataDense := insertDenseEmbeddings(dataIdentity)
	err = v.VectorDBConnector.Upsert(v.T().Context(), v.TestCollection.Name, dataDense)
	assert.NoError(v.T(), err)

	// insert test data with sparse embedding only into vector db
	dataSparse := insertSparseEmbeddings(dataIdentity)
	err = v.VectorDBConnector.Upsert(v.T().Context(), v.TestCollection.Name, dataSparse)
	assert.NoError(v.T(), err)

	// insert test data with metadata only into vector db
	dataMetadata := insertMetadata(dataIdentity)
	err = v.VectorDBConnector.Upsert(v.T().Context(), v.TestCollection.Name, dataMetadata)
	assert.NoError(v.T(), err)

	// check if test data is inside vector db
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 1, count)

	dataGet, err := v.VectorDBConnector.GetByID(v.T().Context(), v.TestCollection.Name, dataKey)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), dataIdentity.ID, dataGet.ID)
	assert.Equal(v.T(), dataIdentity.DocumentID, dataGet.DocumentID)
	assert.Equal(v.T(), dataIdentity.ChunkID, dataGet.ChunkID)
	assert.Equal(v.T(), dataIdentity.Content, dataGet.Content)
	assert.Equal(v.T(), dataSparse.SparseEmbeddings, dataGet.SparseEmbeddings)

	for key, vector := range dataDense.DenseEmbeddings {
		cosineSimilarityScore, err := redisvectorhash.CosineSimilarity(
			vector,
			dataGet.DenseEmbeddings[key])
		assert.NoError(v.T(), err)
		assert.LessOrEqual(v.T(), float32(0.00001), cosineSimilarityScore)
	}
	assert.Equal(v.T(), dataMetadata.Metadata, dataGet.Metadata)

	// test delete stored data
	err = v.VectorDBConnector.Delete(v.T().Context(), v.TestCollection.Name, dataKey)
	assert.NoError(v.T(), err)

	// check if collection is already empty
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 0, count)

	// delete collection inside vector db
	err = v.VectorDBConnector.DeleteCollection(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) Test005ManageBatchPartialData() {

	// create test collection inside vector db
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), v.TestCollection)
	assert.NoError(v.T(), err)

	// create bulk mock test data
	batchNumber := 10
	dataInsertID := generateBulkTestDataID(100)
	dataInsertID500 := dataInsertID[:50]

	dataIdentity := generateBulkTestDataIdentity(dataInsertID)

	// insert test data with dense embedding only into vector db
	dataDense := insertBulkDenseEmbeddings(dataIdentity)
	err = v.VectorDBConnector.UpsertBatch(v.T().Context(), v.TestCollection.Name, dataDense, batchNumber)
	assert.NoError(v.T(), err)

	// insert test data with sparse embedding only into vector db
	dataSparse := insertBulkSparseEmbeddings(dataIdentity)
	err = v.VectorDBConnector.UpsertBatch(v.T().Context(), v.TestCollection.Name, dataSparse, batchNumber)
	assert.NoError(v.T(), err)

	// insert test data with metadata only into vector db
	dataMetadata := insertBulkMetadata(dataIdentity)
	err = v.VectorDBConnector.UpsertBatch(v.T().Context(), v.TestCollection.Name, dataMetadata, batchNumber)
	assert.NoError(v.T(), err)

	// check if test data is inside vector db
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), len(dataInsertID), count)

	// check if all stored test data comply with input data
	for index, data := range dataIdentity {
		dataGet, err := v.VectorDBConnector.GetByID(v.T().Context(), v.TestCollection.Name, data.ID)
		assert.NoError(v.T(), err)
		assert.Equal(v.T(), data.ID, dataGet.ID)
		assert.Equal(v.T(), data.DocumentID, dataGet.DocumentID)
		assert.Equal(v.T(), data.ChunkID, dataGet.ChunkID)
		assert.Equal(v.T(), data.Content, dataGet.Content)
		assert.Equal(v.T(), dataSparse[index].SparseEmbeddings, dataGet.SparseEmbeddings)

		for key, vector := range dataDense[index].DenseEmbeddings {
			cosineSimilarityScore, err := redisvectorhash.CosineSimilarity(
				vector,
				dataGet.DenseEmbeddings[key])
			assert.NoError(v.T(), err)
			assert.LessOrEqual(v.T(), float32(0.00001), cosineSimilarityScore)
		}

		assert.Equal(v.T(), dataMetadata[index].Metadata, dataGet.Metadata)
	}

	// delete all key in batch
	err = v.VectorDBConnector.DeleteBatch(v.T().Context(), v.TestCollection.Name, dataInsertID500, batchNumber)
	assert.NoError(v.T(), err)

	// check collection after delete half number of keys
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.Name)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), len(dataInsertID)-len(dataInsertID500), count)

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
