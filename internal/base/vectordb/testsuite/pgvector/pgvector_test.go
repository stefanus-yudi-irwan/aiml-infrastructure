package pgvector

import (
	"aiml-infrastructure/internal/base/vectordb/backend/pgvector"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VectorDBTestSuite struct {
	suite.Suite
	VectorDBConnector *pgvector.VectorDBConnector
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

	p.VectorDBConnector, err = pgvector.NewVectorDBConnector(p.T().Context(), connectorConfig)
	assert.NoError(p.T(), err)
}

func (p *VectorDBTestSuite) TearDownSuite() {
	p.VectorDBConnector.Close()
}

func TestSuiteVectorDB(t *testing.T) {
	suite.Run(t, new(VectorDBTestSuite))
}

func (p *VectorDBTestSuite) Test001CreateCollection() {

	err := p.VectorDBConnector.CreateVectorExtension(p.T().Context())
	assert.NoError(p.T(), err)

	collectionConfig := createTestCollectionConfig()
	err = collectionConfig.Validate()
	assert.NoError(p.T(), err)
	err = p.VectorDBConnector.CreateCollection(p.T().Context(), collectionConfig)
	assert.NoError(p.T(), err)
}
