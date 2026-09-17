package milvus

import (
	"aiml-infrastructure/internal/base/vectordb/backend/milvus"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VectorDBTestSuite struct {
	suite.Suite
	MilvusConnector *milvus.MilvusConnector
}

func (v *VectorDBTestSuite) SetupSuite() {
	err := godotenv.Load(".env")
	assert.NoError(v.T(), err)

	clientAddress := os.Getenv("CLIENT_ADDRESS")

	milvusConfig := milvus.MilvusConnectorConfig{
		Address: clientAddress,
	}

	v.MilvusConnector, err = milvus.NewMilvusConnector(
		v.T().Context(),
		milvusConfig,
	)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) TearDownSuite() {
	err := v.MilvusConnector.FlushAll(v.T().Context())
	assert.NoError(v.T(), err)

	err = v.MilvusConnector.Close(v.T().Context())
	assert.NoError(v.T(), err)
}

func TestMilvusTestSuite(t *testing.T) {
	suite.Run(t, new(VectorDBTestSuite))
}

func (v *VectorDBTestSuite) Test001CreateCollection() {
	configMilvus := createTestCollectionConfig()
	err := v.MilvusConnector.CreateCollection(v.T().Context(), configMilvus)
	assert.NoError(v.T(), err)

	for _, fieldConfig := range configMilvus.Fields {
		if fieldConfig.VectorConfig != nil {
			err = v.MilvusConnector.CreateVectorIndex(v.T().Context(), configMilvus.Name, fieldConfig)
			assert.NoError(v.T(), err)
		}
	}

	err = v.MilvusConnector.LoadCollection(v.T().Context(), configMilvus.Name)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) Test002IsCollectionExists() {
	collectionName := "test_collection"
	exists, err := v.MilvusConnector.IsCollectionExists(v.T().Context(), collectionName)
	assert.NoError(v.T(), err)
	assert.True(v.T(), exists)
}

func (v *VectorDBTestSuite) Test003Upsert() {
	vectorData := createTestVector("test-id-001")
	err := v.MilvusConnector.Upsert(v.T().Context(), "test_collection", vectorData)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) Test004GetByID() {
	vectorData := createTestVector("test-id-002")
	err := v.MilvusConnector.Upsert(v.T().Context(), "test_collection", vectorData)
	assert.NoError(v.T(), err)

	time.Sleep(1 * time.Second)

	retrievedData, err := v.MilvusConnector.GetByID(v.T().Context(), "test_collection", "test-id-002")
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), vectorData.ID, retrievedData.ID)
	assert.Equal(v.T(), vectorData.Vector, retrievedData.Vector)
	assert.Equal(v.T(), vectorData.Fields, retrievedData.Fields)
}

func (v *VectorDBTestSuite) Test006Delete() {
	collection := "test_collection"
	err := v.MilvusConnector.Delete(v.T().Context(), collection, "test-id-002")
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) Test005CountVector() {
	collection := "test_collection"
	numRecord, err := v.MilvusConnector.CountVector(v.T().Context(), collection)
	assert.NoError(v.T(), err)
	assert.Equal(v.T(), int64(2), numRecord)
}

func (v *VectorDBTestSuite) Test007DeleteCollection() {
	collectionName := "test_collection"
	err := v.MilvusConnector.DeleteCollection(v.T().Context(), collectionName)
	assert.NoError(v.T(), err)
}
