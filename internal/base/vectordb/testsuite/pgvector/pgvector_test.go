package pgvector

import (
	"aiml-infrastructure/internal/base/vectordb/backend/pgvector"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VectorDBTestSuite struct {
	suite.Suite
	VectorDBConnector *pgvector.VectorDBConnector
	TestCollection    pgvector.CollectionConfig
}

func (p *VectorDBTestSuite) SetupSuite() {
	err := godotenv.Load(".env")
	assert.NoError(p.T(), err)

	port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	assert.NoError(p.T(), err)

	maxConns, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_CONNS"))
	assert.NoError(p.T(), err)

	minConns, err := strconv.Atoi(os.Getenv("POSTGRES_MIN_CONNS"))
	assert.NoError(p.T(), err)

	maxConnLifetime, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_CONN_LIFETIME_MINUTE"))
	assert.NoError(p.T(), err)

	maxConnIdletime, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_CONN_IDLETIME_MINUTE"))
	assert.NoError(p.T(), err)

	connectorConfig := pgvector.VectorDBConnectorConfig{
		Host:            os.Getenv("POSTGRES_HOST"),
		Port:            port,
		User:            os.Getenv("POSTGRES_USER"),
		Password:        os.Getenv("POSTGRES_PASSWORD"),
		Database:        os.Getenv("POSTGRES_DATABASE"),
		SSLMode:         os.Getenv("POSTGRES_SSLMODE"),
		MaxConns:        int32(maxConns),
		MinConns:        int32(minConns),
		MaxConnLifetime: time.Duration(maxConnLifetime) * time.Minute,
		MaxConnIdleTime: time.Duration(maxConnIdletime) * time.Minute,
	}

	p.VectorDBConnector, err = pgvector.NewVectorDBConnector(connectorConfig)
	assert.NoError(p.T(), err)

	p.TestCollection = createTestCollectionConfig()
}

func (v *VectorDBTestSuite) TearDownSuite() {
	err := v.VectorDBConnector.FlushAll(v.T().Context(), v.TestCollection.SchemaName)
	assert.NoError(v.T(), err)

	err = v.VectorDBConnector.Close()
	assert.NoError(v.T(), err)
}

func TestSuiteVectorDBPGVector(t *testing.T) {
	suite.Run(t, new(VectorDBTestSuite))
}

func (v *VectorDBTestSuite) Test001ManageCollection() {

	// create test collection inside vector db
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), v.TestCollection)
	assert.NoError(v.T(), err)

	// check if collection exists in vector db
	isExists, err := v.VectorDBConnector.IsCollectionExists(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), true, isExists)

	// delete collection inside vector db
	err = v.VectorDBConnector.DeleteCollection(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)

	// check if collection still exists in vector db
	isExists, err = v.VectorDBConnector.IsCollectionExists(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), false, isExists)
}

func (v *VectorDBTestSuite) Test002ManageSingleData() {

	// create test collection inside vector db
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), v.TestCollection)
	assert.NoError(v.T(), err)

	// create mock test data
	testDataID := uuid.New().String()
	dataInsert := generateTestData(testDataID)

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
	err = v.VectorDBConnector.Upsert(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName, dataInsert)
	assert.NoError(v.T(), err)

	// check if test data is inside vector db
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 1, count)

	dataGet, err := v.VectorDBConnector.GetByID(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName, testDataID)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), dataInsert.ID, dataGet.ID)
	assert.Equal(v.T(), dataInsert.DocumentID, dataGet.DocumentID)
	assert.Equal(v.T(), dataInsert.ChunkID, dataGet.ChunkID)
	assert.Equal(v.T(), dataInsert.Content, dataGet.Content)

	// qdrant vector being autonormalized if using COSINE
	cosineSimilarityScore, err := pgvector.CosineSimilarity(pgvector.NormalizeVector(dataInsert.Embedding), dataGet.Embedding)
	assert.NoError(v.T(), err)
	assert.LessOrEqual(v.T(), float32(0.00001), cosineSimilarityScore)
	//assert.Equal(v.T(), dataInsert.Metadata, dataGet.Metadata)

	// test delete stored data
	err = v.VectorDBConnector.Delete(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName, testDataID)
	assert.NoError(v.T(), err)

	// check if collection is already empty
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 0, count)

	// delete collection inside vector db
	err = v.VectorDBConnector.DeleteCollection(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) Test003ManageBatchData() {

	// create test collection inside vector db
	err := v.VectorDBConnector.CreateCollection(v.T().Context(), v.TestCollection)
	assert.NoError(v.T(), err)

	// create bulk mock test data
	batchNumber := 10
	dataInsertID := generateBulkTestDataID(1000)
	dataInsertID500 := dataInsertID[:500]
	dataInsert := generateBulkTestData(dataInsertID)

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
	err = v.VectorDBConnector.UpsertBatch(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName, dataInsert, batchNumber)
	assert.NoError(v.T(), err)

	// check if test data is inside vector db
	count, err := v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), len(dataInsertID), count)

	// check if all stored test data comply with input data
	for _, data := range dataInsert {
		dataGet, err := v.VectorDBConnector.GetByID(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName, data.ID)
		assert.NoError(v.T(), err)
		assert.Equal(v.T(), data.ID, dataGet.ID)
		assert.Equal(v.T(), data.DocumentID, dataGet.DocumentID)
		assert.Equal(v.T(), data.ChunkID, dataGet.ChunkID)
		assert.Equal(v.T(), data.Content, dataGet.Content)

		// qdrant vector being autonormalized if using COSINE
		cosineSimilarityScore, err := pgvector.CosineSimilarity(pgvector.NormalizeVector(data.Embedding), dataGet.Embedding)
		assert.NoError(v.T(), err)
		assert.LessOrEqual(v.T(), float32(0.00001), cosineSimilarityScore)
		//assert.Equal(v.T(), data.Metadata, dataGet.Metadata)
	}

	// delete all key in batch
	err = v.VectorDBConnector.DeleteBatch(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName, dataInsertID500, batchNumber)
	assert.NoError(v.T(), err)

	// check collection after delete half number of keys
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), len(dataInsertID)-len(dataInsertID500), count)

	// test delete all stored data
	err = v.VectorDBConnector.FlushCollection(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)

	// check if collection is already empty
	count, err = v.VectorDBConnector.CountVector(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), 0, count)

	// delete collection inside vector db
	err = v.VectorDBConnector.DeleteCollection(v.T().Context(), v.TestCollection.SchemaName, v.TestCollection.CollectionName)
	assert.NoError(v.T(), err)

}
