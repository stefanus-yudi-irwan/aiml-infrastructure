package weaviate

import (
	"aiml-infrastructure/internal/base/vectordb/backend/weaviate"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VectorDBTestSuite struct {
	suite.Suite
	WeaviateConnector *weaviate.WeaviateConnector
}

func (v *VectorDBTestSuite) SetupSuite() {
	err := godotenv.Load(".env")
	assert.NoError(v.T(), err)

	weaviateConfig := weaviate.WeaviateConnectorConfig{
		Host:   os.Getenv("HOST"),
		Scheme: os.Getenv("SCHEME"),
	}

	v.WeaviateConnector, err = weaviate.NewWeaviateConnector(weaviateConfig)
	assert.NoError(v.T(), err)
}

func (v *VectorDBTestSuite) TearDownSuite() {

}

func TestWeaviateTestSuite(t *testing.T) {
	suite.Run(t, new(VectorDBTestSuite))
}

func (v *VectorDBTestSuite) Test001CreateCollection() {
	collectionConfig := createTestCollectionConfig()
	err := collectionConfig.Validate()
	assert.NoError(v.T(), err)
	err = v.WeaviateConnector.CreateCollection(v.T().Context(), collectionConfig)
	assert.NoError(v.T(), err)
}
