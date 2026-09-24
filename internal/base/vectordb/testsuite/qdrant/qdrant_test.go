package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb/backend/qdrant"
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VectorDBTestSuite struct {
	suite.Suite
	VectorDBConnector *qdrant.VectorDBConnector
}

func (v *VectorDBTestSuite) SetupSuite() {
	err := godotenv.Load(".env")
	assert.NoError(v.T(), err)

	port, err := strconv.Atoi(os.Getenv("PORT"))
	assert.NoError(v.T(), err)
	connectorConfig := qdrant.VectorDBConnectorConfig{
		Host:   os.Getenv("HOST"),
		Port:   port,
		APIKey: os.Getenv("API_KEY"),
	}

	v.VectorDBConnector, err = qdrant.NewVectorDBConnector(connectorConfig)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) TearDownSuite() {
}

func TestSuiteVectorDB(t *testing.T) {
	suite.Run(t, new(VectorDBTestSuite))
}

func (v *VectorDBTestSuite) Test001CreateCollection() {
	collectionConfig := createTestCollectionConfig()
	err := collectionConfig.Validate()
	assert.NoError(v.T(), err)
	err = v.VectorDBConnector.CreateCollection(v.T().Context(), collectionConfig)
	assert.NoError(v.T(), err)
}
